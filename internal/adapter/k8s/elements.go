package k8s

import (
	"sort"
	"strconv"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

func workloads(idx *index) []model.DiscoveredElement {
	var out []model.DiscoveredElement
	for _, o := range idx.objects {
		path, ok := podSpecPath[o.Kind]
		if !ok {
			continue
		}
		spec := obj(o.Body, path...)
		e := model.Element{ID: o.id(), Kind: "backend", Parent: id(o.Namespace, "Namespace", o.Namespace),
			Tech: o.Kind, Ports: containerPorts(spec)}
		if image := firstImage(spec); image != "" {
			e.Tech += " · " + image
		}
		if o.Kind == "DaemonSet" {
			e.Multiplicity = "one per node"
		}
		out = append(out, model.DiscoveredElement{Element: e, Src: o.Src})
	}
	return out
}

// containerPorts merges the ports of every container, as the pod listens on all of them.
func containerPorts(podSpec map[string]any) []model.Port {
	var ports []model.Port
	for _, c := range slice(podSpec, "containers") {
		cm, _ := c.(map[string]any)
		for _, p := range slice(cm, "ports") {
			pm, _ := p.(map[string]any)
			if n, ok := num(pm, "containerPort"); ok {
				ports = append(ports, port(str(pm, "name"), str(pm, "protocol"), n))
			}
		}
	}
	return uniquePorts(ports)
}

func services(idx *index) []model.DiscoveredElement {
	var out []model.DiscoveredElement
	for _, o := range idx.byKind["Service"] {
		tech := "Service"
		switch t := str(o.Body, "spec", "type"); {
		case str(o.Body, "spec", "clusterIP") == "None":
			tech += " · headless"
		case t != "" && t != "ClusterIP":
			tech += " · " + t
		}
		var ports []model.Port
		for _, p := range slice(o.Body, "spec", "ports") {
			pm, _ := p.(map[string]any)
			if n, ok := num(pm, "port"); ok {
				ports = append(ports, port(str(pm, "name"), str(pm, "protocol"), n))
			}
		}
		out = append(out, model.DiscoveredElement{Element: model.Element{ID: o.id(), Kind: "backend",
			Parent: id(o.Namespace, "Namespace", o.Namespace), Tech: tech, Ports: uniquePorts(ports)}, Src: o.Src})
	}
	return out
}

// namespaces are zones: one per namespace that holds an element or is declared. A zone
// without its Namespace object points at the first object read in it.
func namespaces(idx *index, elements []model.DiscoveredElement) []model.DiscoveredElement {
	srcs := map[string]model.Src{}
	for _, o := range idx.byKind["Namespace"] {
		srcs[o.Name] = o.Src
	}
	used := map[string]bool{}
	for _, e := range elements {
		used[e.Parent] = true
	}
	for _, o := range idx.objects {
		if _, ok := srcs[o.Namespace]; !ok && o.Namespace != "" && used[id(o.Namespace, "Namespace", o.Namespace)] {
			srcs[o.Namespace] = o.Src
		}
	}
	var out []model.DiscoveredElement
	for _, ns := range sortedKeys(srcs) {
		out = append(out, model.DiscoveredElement{Element: model.Element{ID: id(ns, "Namespace", ns), Kind: "group"}, Src: srcs[ns]})
	}
	return out
}

// firstImage is the main container's image without registry and path, which is what
// tells a reader what a generically named workload runs: keycloak:26.1, not the chart.
func firstImage(podSpec map[string]any) string {
	containers := slice(podSpec, "containers")
	if len(containers) == 0 {
		return ""
	}
	c, _ := containers[0].(map[string]any)
	image, _, _ := strings.Cut(str(c, "image"), "@")
	return image[strings.LastIndex(image, "/")+1:]
}

func port(name, protocol string, number int) model.Port {
	if name == "" {
		name = strconv.Itoa(number)
	}
	if protocol == "" {
		protocol = "TCP"
	}
	return model.Port{Name: name, Protocol: protocol, Number: number}
}

func uniquePorts(ports []model.Port) []model.Port {
	sort.SliceStable(ports, func(i, j int) bool {
		if ports[i].Number != ports[j].Number {
			return ports[i].Number < ports[j].Number
		}
		return ports[i].Name < ports[j].Name
	})
	out := ports[:0]
	for i, p := range ports {
		if i == 0 || p != ports[i-1] {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
