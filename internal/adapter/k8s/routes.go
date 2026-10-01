package k8s

import (
	"sort"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// routeProtocols are the Gateway API route kinds and the protocol their traffic uses.
var routeProtocols = map[string]string{"HTTPRoute": "HTTP", "GRPCRoute": "gRPC", "TLSRoute": "TLS", "TCPRoute": "TCP"}

func gateways(idx *index) []model.DiscoveredElement {
	var out []model.DiscoveredElement
	for _, gw := range idx.byKind["Gateway"] {
		tech := "Gateway"
		if class := str(gw.Body, "spec", "gatewayClassName"); class != "" {
			tech += " · " + class
		}
		var ports []model.Port
		for _, l := range slice(gw.Body, "spec", "listeners") {
			lm, _ := l.(map[string]any)
			if n, ok := num(lm, "port"); ok {
				protocol := "TCP"
				if str(lm, "protocol") == "UDP" {
					protocol = "UDP"
				}
				ports = append(ports, port(str(lm, "name"), protocol, n))
			}
		}
		out = append(out, model.DiscoveredElement{Element: model.Element{ID: gw.id(), Kind: "backend",
			Parent: id(gw.Namespace, "Namespace", gw.Namespace), Tech: tech, Ports: uniquePorts(ports)}, Src: gw.Src})
	}
	return out
}

// routes are elements, as a reader looks for the host a route serves, with a
// parentRef to their Gateway and a backendRef to each Service. The traffic itself is a
// connection from the Gateway to the Service.
func routes(idx *index, g *graph, u *unresolvedSet) []model.DiscoveredElement {
	var out []model.DiscoveredElement
	for _, kind := range sortedKeys(routeProtocols) {
		for _, r := range idx.byKind[kind] {
			out = append(out, model.DiscoveredElement{Element: model.Element{ID: r.id(), Kind: "backend",
				Parent: id(r.Namespace, "Namespace", r.Namespace), Tech: kind,
				Desc: strings.Join(strSlice(r.Body, "spec", "hostnames"), ", ")}, Src: r.Src})
			var parents []string
			for _, p := range slice(r.Body, "spec", "parentRefs") {
				pm, _ := p.(map[string]any)
				if k := str(pm, "kind"); k != "" && k != "Gateway" {
					continue
				}
				ns := or(str(pm, "namespace"), r.Namespace)
				gw := idx.get(ns, "Gateway", str(pm, "name"))
				if gw == nil {
					u.add("missing-backend", id(ns, "Gateway", str(pm, "name")), "gateway", r.id(), r.Src)
					continue
				}
				g.refer(r.id(), gw.id(), "parentRef", r.Src)
				parents = append(parents, gw.id())
			}
			for _, rule := range slice(r.Body, "spec", "rules") {
				rm, _ := rule.(map[string]any)
				for _, b := range slice(rm, "backendRefs") {
					bm, _ := b.(map[string]any)
					svc := backendService(idx, u, r, or(str(bm, "kind"), "Service"), or(str(bm, "namespace"), r.Namespace), str(bm, "name"))
					if svc == nil {
						continue
					}
					g.refer(r.id(), svc.id(), "backendRef", r.Src)
					for _, gw := range parents {
						g.connect(gw, svc.id(), servicePortName(svc, str(bm, "port")), routeProtocols[kind], r.Src)
					}
				}
			}
		}
	}
	return out
}

// ingresses are elements like routes. Their traffic comes from the controller behind
// their class, an element of its own; an Ingress without a class sends from itself.
func ingresses(idx *index, g *graph, u *unresolvedSet) []model.DiscoveredElement {
	var out []model.DiscoveredElement
	classes := map[string]model.Src{}
	for _, ing := range idx.byKind["Ingress"] {
		var hosts []string
		for _, rule := range slice(ing.Body, "spec", "rules") {
			rm, _ := rule.(map[string]any)
			if h := str(rm, "host"); h != "" {
				hosts = append(hosts, h)
			}
		}
		out = append(out, model.DiscoveredElement{Element: model.Element{ID: ing.id(), Kind: "backend",
			Parent: id(ing.Namespace, "Namespace", ing.Namespace), Tech: "Ingress", Desc: strings.Join(hosts, ", ")}, Src: ing.Src})
		source := ing.id()
		if class := or(str(ing.Body, "spec", "ingressClassName"), ing.Annotations["kubernetes.io/ingress.class"]); class != "" {
			source = classID(class)
			if _, ok := classes[class]; !ok {
				classes[class] = ing.Src
			}
			g.refer(ing.id(), source, "ingressClassName", ing.Src)
		}
		for _, b := range ingressBackends(ing.Body) {
			svc := backendService(idx, u, ing, "Service", ing.Namespace, b.name)
			if svc == nil {
				continue
			}
			g.refer(ing.id(), svc.id(), "backend", ing.Src)
			g.connect(source, svc.id(), servicePortName(svc, b.port), "HTTP", ing.Src)
		}
	}
	for _, name := range sortedKeys(classes) {
		tech, src := "IngressClass", classes[name]
		if c := idx.get("", "IngressClass", name); c != nil {
			tech += " · " + str(c.Body, "spec", "controller")
			src = c.Src
		}
		out = append(out, model.DiscoveredElement{Element: model.Element{ID: classID(name), Kind: "backend", Tech: tech}, Src: src})
	}
	return out
}

// classID puts a cluster-scoped IngressClass under _cluster, which no namespace can be
// named.
func classID(name string) string { return "k8s:_cluster/ingressclass/" + name }

type ingressBackend struct{ name, port string }

// ingressBackends reads networking.k8s.io/v1 backends and the older
// serviceName/servicePort form still found in charts.
func ingressBackends(body map[string]any) []ingressBackend {
	read := func(b map[string]any) (ingressBackend, bool) {
		if s := obj(b, "service"); s != nil {
			return ingressBackend{str(s, "name"), or(str(s, "port", "number"), str(s, "port", "name"))}, true
		}
		if name := str(b, "serviceName"); name != "" {
			return ingressBackend{name, str(b, "servicePort")}, true
		}
		return ingressBackend{}, false
	}
	var out []ingressBackend
	if b, ok := read(obj(body, "spec", "defaultBackend")); ok {
		out = append(out, b)
	}
	if b, ok := read(obj(body, "spec", "backend")); ok {
		out = append(out, b)
	}
	for _, rule := range slice(body, "spec", "rules") {
		rm, _ := rule.(map[string]any)
		for _, p := range slice(rm, "http", "paths") {
			pm, _ := p.(map[string]any)
			if b, ok := read(obj(pm, "backend")); ok {
				out = append(out, b)
			}
		}
	}
	return out
}

// backendService finds the Service a route or an Ingress sends to, or reports it.
func backendService(idx *index, u *unresolvedSet, from *object, kind, ns, name string) *object {
	if kind != "Service" {
		u.add("missing-backend", id(ns, kind, name), "not-a-service", from.id(), from.Src)
		return nil
	}
	svc := idx.get(ns, "Service", name)
	if svc == nil {
		u.add("missing-backend", id(ns, "Service", name), "service", from.id(), from.Src)
	}
	return svc
}

func strSlice(m map[string]any, path ...string) []string {
	var out []string
	for _, v := range slice(m, path...) {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
