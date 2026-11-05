package k8s

import (
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// builtinGroups are the API groups that ship with Kubernetes. An object outside them is
// a custom resource, whose spec may name the Services it talks to (FR-003).
var builtinGroups = set("core", "apps", "batch", "networking.k8s.io", "policy", "rbac.authorization.k8s.io",
	"autoscaling", "storage.k8s.io", "apiextensions.k8s.io", "admissionregistration.k8s.io", "scheduling.k8s.io",
	"node.k8s.io", "coordination.k8s.io", "discovery.k8s.io", "certificates.k8s.io", "events.k8s.io",
	"apiregistration.k8s.io", "snapshot.storage.k8s.io", "gateway.networking.k8s.io")

// dnsAddress is <svc>.<ns>[.svc[.cluster.local]][:port], matched at a position the
// caller has checked starts a word.
var dnsAddress = regexp.MustCompile(`^([a-z0-9](?:[-a-z0-9]*[a-z0-9])?)\.([a-z0-9](?:[-a-z0-9]*[a-z0-9])?)(?:\.svc(?:\.cluster\.local)?)?(?::([0-9]{1,5}))?`)

func isNameChar(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || b == '-' || b == '_' || b == '.'
}

type address struct{ service, namespace, port string }

// addressesIn finds the Service addresses a value holds. Values are scanned, never
// evaluated; a longer dotted name such as shop.example.com is not an address.
func addressesIn(value string) []address {
	var out []address
	for i := 0; i < len(value); i++ {
		if (i > 0 && isNameChar(value[i-1])) || !(value[i] >= 'a' && value[i] <= 'z' || value[i] >= '0' && value[i] <= '9') {
			continue
		}
		m := dnsAddress.FindStringSubmatch(value[i:])
		if m == nil {
			continue
		}
		end := i + len(m[0])
		if end < len(value) && (value[end] == '-' || value[end] == '_' || value[end] >= 'a' && value[end] <= 'z' || value[end] >= 'A' && value[end] <= 'Z' || value[end] >= '0' && value[end] <= '9' ||
			value[end] == '.' && end+1 < len(value) && isNameChar(value[end+1])) {
			continue
		}
		out = append(out, address{service: m[1], namespace: m[2], port: m[3]})
		i = end - 1
	}
	return out
}

type urlHost struct{ host, port, scheme string }

// urlHostsIn finds the scheme://host[:port] URLs of a value, without any credentials.
func urlHostsIn(value string) []urlHost {
	var out []urlHost
	rest := value
	for {
		i := strings.Index(rest, "://")
		if i < 0 {
			return out
		}
		start := i
		for start > 0 && (rest[start-1] >= 'a' && rest[start-1] <= 'z' || rest[start-1] >= 'A' && rest[start-1] <= 'Z' || rest[start-1] >= '0' && rest[start-1] <= '9' || strings.IndexByte("+.-", rest[start-1]) >= 0) {
			start--
		}
		scheme := strings.ToLower(rest[start:i])
		after := rest[i+3:]
		end := strings.IndexAny(after, "/?# \t\r\n\"',;")
		if end < 0 {
			end = len(after)
		}
		authority := after[:end]
		if at := strings.LastIndexByte(authority, '@'); at >= 0 {
			authority = authority[at+1:]
		}
		host, port, _ := strings.Cut(authority, ":")
		if host != "" && scheme != "" {
			out = append(out, urlHost{host: strings.ToLower(host), port: port, scheme: scheme})
		}
		rest = after[end:]
	}
}

type entryPoint struct {
	id    string
	ports []model.Port
}

// entryPoints maps each public hostname of the input to what serves it: the Gateways of
// an HTTPRoute, the controller behind an Ingress (FR-004).
func entryPoints(idx *index) map[string][]entryPoint {
	out := map[string][]entryPoint{}
	for _, kind := range sortedKeys(routeProtocols) {
		for _, r := range idx.byKind[kind] {
			var gateways []entryPoint
			for _, p := range slice(r.Body, "spec", "parentRefs") {
				pm, _ := p.(map[string]any)
				if k := str(pm, "kind"); k != "" && k != "Gateway" {
					continue
				}
				if gw := idx.get(or(str(pm, "namespace"), r.Namespace), "Gateway", str(pm, "name")); gw != nil {
					gateways = append(gateways, entryPoint{id: gw.id(), ports: gatewayPorts(gw)})
				}
			}
			for _, host := range strSlice(r.Body, "spec", "hostnames") {
				out[strings.ToLower(host)] = append(out[strings.ToLower(host)], gateways...)
			}
		}
	}
	for _, ing := range idx.byKind["Ingress"] {
		controller := ing.id()
		if class := or(str(ing.Body, "spec", "ingressClassName"), ing.Annotations["kubernetes.io/ingress.class"]); class != "" {
			controller = classID(class)
		}
		for _, rule := range slice(ing.Body, "spec", "rules") {
			rm, _ := rule.(map[string]any)
			if h := str(rm, "host"); h != "" {
				out[strings.ToLower(h)] = append(out[strings.ToLower(h)], entryPoint{id: controller})
			}
		}
	}
	return out
}

func (m entryPointMap) serving(host string) []entryPoint {
	if eps, ok := m[host]; ok {
		return eps
	}
	if _, rest, ok := strings.Cut(host, "."); ok {
		return m["*."+rest]
	}
	return nil
}

type entryPointMap map[string][]entryPoint

// target is where a value points: a Service or entry point, and the port name on it.
type target struct{ to, port string }

type scanner struct {
	idx    *index
	public entryPointMap
}

// scan resolves one string value to the known Services and public entry points it names.
func (s *scanner) scan(value string) []target {
	var out []target
	for _, a := range addressesIn(value) {
		if svc := s.idx.get(a.namespace, "Service", a.service); svc != nil {
			if port, ok := addressPort(svc, a.port); ok {
				out = append(out, target{svc.id(), port})
			}
		}
	}
	for _, u := range urlHostsIn(value) {
		for _, ep := range s.public.serving(u.host) {
			out = append(out, target{ep.id, entryPortName(ep, u)})
		}
	}
	return out
}

// addressPort picks the Service port a value means: the one numbered like its port, or,
// with none written, the only port, else the one named http, else the first (plan 009).
func addressPort(svc *object, port string) (string, bool) {
	ports := slice(svc.Body, "spec", "ports")
	if port != "" {
		for _, p := range ports {
			if pm, _ := p.(map[string]any); str(pm, "port") == port {
				return port1Name(pm), true
			}
		}
		return "", false
	}
	for _, p := range ports {
		if pm, _ := p.(map[string]any); len(ports) == 1 || str(pm, "name") == "http" {
			return port1Name(pm), true
		}
	}
	if len(ports) > 0 {
		pm, _ := ports[0].(map[string]any)
		return port1Name(pm), true
	}
	return "", true
}

// entryPortName names the listener a URL reaches: its explicit port, else its scheme's.
func entryPortName(ep entryPoint, u urlHost) string {
	number := u.port
	if number == "" {
		number = map[string]string{"http": "80", "https": "443"}[u.scheme]
	}
	n, err := strconv.Atoi(number)
	if err != nil {
		return ""
	}
	for _, p := range ep.ports {
		if p.Number == n {
			return p.Name
		}
	}
	return ""
}

// addresses emits an inferred connection for every Service address or public URL that an
// env value, a ConfigMap read by a workload, or a custom resource spec holds (FR-003,
// FR-004). Secrets are never scanned.
func addresses(idx *index, g *graph, uses []configUse) []model.DiscoveredElement {
	s := &scanner{idx: idx, public: entryPoints(idx)}
	emit := func(from string, o *object, line int, value string) bool {
		found := false
		for _, t := range s.scan(value) {
			g.infer(from, t.to, t.port, o, o.srcAt(line))
			found = true
		}
		return found
	}
	for _, w := range idx.objects {
		if !isWorkload(w.Kind) {
			continue
		}
		for _, c := range w.containers() {
			for _, env := range seq(mapValue(c, "env")) {
				if v := mapValue(env, "value"); v != nil {
					emit(w.id(), w, v.Line, text(v))
				}
			}
		}
	}
	for _, cm := range idx.byKind["ConfigMap"] {
		data := at(cm.Node, "data")
		if data == nil || data.Kind != yaml.MappingNode {
			continue
		}
		for i := 0; i+1 < len(data.Content); i += 2 {
			key, v := data.Content[i].Value, resolve(data.Content[i+1])
			readers := map[*object]bool{}
			for _, u := range uses {
				if u.target == cm.id() && (u.key == "" || u.key == key) && !readers[u.workload] {
					readers[u.workload] = true
					emit(u.workload.id(), cm, v.Line, text(v))
				}
			}
		}
	}
	var elements []model.DiscoveredElement
	for _, o := range idx.objects {
		if builtinGroups[group(o.APIVersion)] || isWorkload(o.Kind) || o.Kind == "CustomResourceDefinition" {
			continue
		}
		found := false
		walkStrings(at(o.Node, "spec"), func(v *yaml.Node) {
			if emit(o.id(), o, v.Line, v.Value) {
				found = true
			}
		})
		if found {
			e := model.Element{ID: o.id(), Kind: "backend", Tech: o.Kind}
			if o.Namespace != "" {
				e.Parent = id(o.Namespace, "Namespace", o.Namespace)
			}
			elements = append(elements, model.DiscoveredElement{Element: e, Src: o.Src, Object: o.identity()})
		}
	}
	return elements
}

// walkStrings visits every string value under n, in document order.
func walkStrings(n *yaml.Node, visit func(*yaml.Node)) {
	switch n = resolve(n); {
	case n == nil:
	case n.Kind == yaml.ScalarNode && n.Tag == "!!str":
		visit(n)
	case n.Kind == yaml.SequenceNode:
		for _, c := range n.Content {
			walkStrings(c, visit)
		}
	case n.Kind == yaml.MappingNode:
		for i := 1; i < len(n.Content); i += 2 {
			walkStrings(n.Content[i], visit)
		}
	}
}
