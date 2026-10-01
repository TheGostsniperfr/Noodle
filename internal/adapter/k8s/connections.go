package k8s

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// graph collects connections and references, keeping the first of each id so a pair
// found by several routes is one line pointing at the first route read.
type graph struct {
	conns []model.DiscoveredConnection
	refs  []model.DiscoveredReference
	seen  map[string]bool
}

// edgeID is a short hash of what the edge joins: from and to already say it, and a
// readable id repeating both doubled the size of every line (ADR-0017 token guards).
// The prefix keeps connection and reference ids apart.
func edgeID(prefix, from, to, port string) string {
	sum := sha256.Sum256([]byte(from + "\x00" + to + "\x00" + port))
	return "k8s:" + prefix + "/" + hex.EncodeToString(sum[:5])
}

func (g *graph) connect(from, to, port, protocol string, src model.Src) {
	cid := edgeID("c", from, to, port)
	if g.seen == nil {
		g.seen = map[string]bool{}
	}
	if g.seen[cid] {
		return
	}
	g.seen[cid] = true
	g.conns = append(g.conns, model.DiscoveredConnection{Connection: model.Connection{
		ID: cid, From: from, To: to, Port: port, Protocol: protocol, Kind: "flow"}, Src: src})
}

func (g *graph) refer(from, to, kind string, src model.Src) {
	rid := edgeID("r", from, to, kind)
	if g.seen == nil {
		g.seen = map[string]bool{}
	}
	if g.seen[rid] {
		return
	}
	g.seen[rid] = true
	g.refs = append(g.refs, model.DiscoveredReference{Reference: model.Reference{ID: rid, From: from, To: to, Kind: kind}, Src: src})
}

// selectsPods connects each Service to the workloads its selector matches, once per
// target port. A Service whose selector matches nothing in the input usually fronts
// pods an operator creates, such as a database cluster, and is reported.
func selectsPods(idx *index, g *graph, u *unresolvedSet) {
	for _, svc := range idx.byKind["Service"] {
		selector := strMap(svc.Body, "spec", "selector")
		if len(selector) == 0 {
			continue
		}
		matched := false
		for _, w := range idx.workloadsIn(svc.Namespace) {
			if !matchLabels(selector, podLabels(w)) {
				continue
			}
			matched = true
			for _, p := range slice(svc.Body, "spec", "ports") {
				pm, _ := p.(map[string]any)
				g.connect(svc.id(), w.id(), targetPortName(pm, w), "", svc.Src)
			}
			if len(slice(svc.Body, "spec", "ports")) == 0 {
				g.connect(svc.id(), w.id(), "", "", svc.Src)
			}
		}
		if !matched {
			u.add("unmatched-selector", labelString(selector), "", svc.id(), svc.Src)
		}
	}
}

// targetPortName names the container port a Service port forwards to. targetPort is a
// container port name, a number, or absent, meaning the Service port itself. A number
// the workload does not declare is still named by its number.
func targetPortName(servicePort map[string]any, w *object) string {
	target := or(str(servicePort, "targetPort"), str(servicePort, "port"))
	n, err := strconv.Atoi(target)
	if err != nil {
		return target
	}
	for _, p := range containerPorts(obj(w.Body, podSpecPath[w.Kind]...)) {
		if p.Number == n {
			return p.Name
		}
	}
	return target
}

func matchLabels(selector, labels map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}

func labelString(labels map[string]string) string {
	parts := make([]string, 0, len(labels))
	for _, k := range sortedKeys(labels) {
		parts = append(parts, k+"="+labels[k])
	}
	return strings.Join(parts, ",")
}

// servicePortName names the Service port a route or an Ingress sends to, by number
// or by name. With no port given, a single-port Service is unambiguous.
func servicePortName(svc *object, port string) string {
	ports := slice(svc.Body, "spec", "ports")
	if port == "" {
		if len(ports) == 1 {
			pm, _ := ports[0].(map[string]any)
			return port1Name(pm)
		}
		return ""
	}
	for _, p := range ports {
		pm, _ := p.(map[string]any)
		if str(pm, "port") == port || str(pm, "name") == port {
			return port1Name(pm)
		}
	}
	return port
}

func port1Name(pm map[string]any) string {
	if name := str(pm, "name"); name != "" {
		return name
	}
	return str(pm, "port")
}
