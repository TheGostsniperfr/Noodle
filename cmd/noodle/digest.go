package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

const digestUsage = "usage: noodle digest DIR"

// digestCommand prints what an agent needs from the fragments of DIR, one line per
// fact (ADR-0031). It is a view, not a contract: the fragments stay the source. Stderr
// gets both token estimates, so the saving is visible on every run.
func digestCommand(args []string, stdout, stderr io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf(digestUsage)
	}
	frags, err := model.LoadFragments(args[0])
	if err != nil {
		return fmt.Errorf("digest: %w", err)
	}
	if len(frags) == 0 {
		return fmt.Errorf("digest: %s has no fragment; run noodle discover first", args[0])
	}
	out := digest(frags)
	fragmentBytes := 0
	for _, f := range frags {
		info, err := os.Stat(filepath.Join(args[0], f.ID+".yaml"))
		if err != nil {
			return fmt.Errorf("digest: %w", err)
		}
		fragmentBytes += int(info.Size())
	}
	fmt.Fprintf(stderr, "digest: %d fragments ~%s tokens, digest ~%s tokens\n", len(frags), kTokens(fragmentBytes), kTokens(len(out)))
	_, err = io.WriteString(stdout, out)
	return err
}

var podKinds = map[string]bool{"deployment": true, "statefulset": true, "daemonset": true,
	"replicaset": true, "job": true, "cronjob": true, "pod": true}

// hiddenKinds are elements the digest folds into others or leaves out: a namespace is in
// every short id, and a route is the label of the connection it creates.
var hiddenKinds = map[string]bool{"namespace": true, "httproute": true, "grpcroute": true,
	"tlsroute": true, "tcproute": true, "ingress": true}

// kindOf is the kind segment of k8s:<ns>/<kind>/<name>.
func kindOf(id string) string {
	parts := strings.Split(strings.TrimPrefix(id, "k8s:"), "/")
	if len(parts) == 3 {
		return parts[1]
	}
	return ""
}

// short drops the adapter prefix, and the kind of a workload: the digest names a
// component the way a diagram would.
func short(id string) string {
	parts := strings.Split(strings.TrimPrefix(id, "k8s:"), "/")
	if parts[0] == "_external" {
		return "external/" + parts[len(parts)-1]
	}
	if len(parts) != 3 {
		return strings.TrimPrefix(id, "k8s:")
	}
	ns, kind, name := parts[0], parts[1], parts[2]
	switch {
	case podKinds[kind]:
		return ns + "/" + name
	case ns == "_cluster":
		return kind + "/" + name
	}
	return ns + "/" + kind + "/" + name
}

type digester struct {
	elements map[string]model.DiscoveredElement
	conns    []model.DiscoveredConnection
	refs     []model.DiscoveredReference
	open     []model.Unresolved
	folded   map[string]string
	lines    map[string][]string
}

func digest(frags []*model.Fragment) string {
	d := &digester{elements: map[string]model.DiscoveredElement{}, folded: map[string]string{}, lines: map[string][]string{}}
	for _, f := range frags {
		for _, e := range f.Elements {
			d.elements[e.ID] = e
		}
		d.conns = append(d.conns, f.Connections...)
		d.refs = append(d.refs, f.References...)
		d.open = append(d.open, f.Unresolved...)
	}
	d.foldServices()
	d.connections()
	d.secrets()
	d.unresolved()
	d.components()

	var b strings.Builder
	fmt.Fprintf(&b, "# noodle digest of %d fragments: %d components, %d connections, %d secrets, %d open\n",
		len(frags), len(d.lines["components"]), len(d.lines["connections"]), len(d.lines["secrets"]), len(d.lines["open"]))
	for _, section := range []string{"components", "connections", "secrets", "open"} {
		lines := d.lines[section]
		if len(lines) == 0 {
			continue
		}
		sort.Strings(lines)
		fmt.Fprintf(&b, "# %s\n", section)
		for i, l := range lines {
			if i == 0 || l != lines[i-1] {
				b.WriteString(l + "\n")
			}
		}
	}
	return b.String()
}

// foldServices maps a Service that selects exactly one workload to that workload, so a
// request path reads gateway -> api, not gateway -> service -> api.
func (d *digester) foldServices() {
	targets := map[string][]string{}
	for _, c := range d.conns {
		if kindOf(c.From) == "service" && podKinds[kindOf(c.To)] {
			targets[c.From] = append(targets[c.From], c.To)
		}
	}
	for svc, ws := range targets {
		if len(ws) == 1 {
			d.folded[svc] = ws[0]
		}
	}
}

func (d *digester) fold(id string) string {
	if w, ok := d.folded[id]; ok {
		return w
	}
	return id
}

func (d *digester) add(section, text string, src model.Src) {
	d.lines[section] = append(d.lines[section], text+"  src "+srcText(src))
}

func srcText(s model.Src) string {
	if s.File == "" {
		return s.Object
	}
	return fmt.Sprintf("%s:%d", s.File, s.Line)
}

func (d *digester) connections() {
	for _, c := range d.conns {
		if d.folded[c.From] == c.To {
			continue
		}
		var labels []string
		for _, l := range []string{c.Port, c.Protocol} {
			if l != "" {
				labels = append(labels, l)
			}
		}
		if c.Object != nil && hiddenKinds[strings.ToLower(c.Object.Kind)] {
			if host := d.elements[routeID(c.Object)].Desc; host != "" {
				labels = append(labels, "host "+host)
			}
		}
		if c.Inferred {
			labels = append(labels, "inferred")
		}
		d.connection(c.From, c.To, labels, c.Src)
	}
}

func name(id string) string { return id[strings.LastIndex(id, "/")+1:] }

func routeID(o *model.Object) string {
	return "k8s:" + o.Namespace + "/" + strings.ToLower(o.Kind) + "/" + o.Name
}

func (d *digester) connection(from, to string, labels []string, src model.Src) {
	from, to = d.fold(from), d.fold(to)
	if from == to {
		return
	}
	text := short(from) + " -> " + short(to)
	if len(labels) > 0 {
		text += "  " + strings.Join(labels, " ")
	}
	d.add("connections", text, src)
}

// refsFrom are the targets of from's references of one kind.
func (d *digester) refsFrom(from, kind string) []string {
	var out []string
	for _, r := range d.refs {
		if r.From == from && r.Kind == kind {
			out = append(out, r.To)
		}
	}
	return out
}

// secrets prints the path from a Secret written by an ExternalSecret to the provider that
// holds its values: where an agent would look for an address the manifests lack.
func (d *digester) secrets() {
	for _, r := range d.refs {
		if r.Kind != "target" || kindOf(r.From) != "externalsecret" {
			continue
		}
		text := short(r.To) + " <- externalsecret"
		if name(r.From) != name(r.To) {
			text += " " + name(r.From)
		}
		for _, store := range d.refsFrom(r.From, "store") {
			text += " <- " + short(store)
			for _, p := range d.refsFrom(store, "provider") {
				text += " <- " + short(p)
			}
		}
		d.add("secrets", text, r.Src)
	}
}

// unresolved settles a missing backend that another fragment defines into a connection,
// as drift does (FR-012), and keeps the rest open.
func (d *digester) unresolved() {
	for _, u := range d.open {
		if _, defined := d.elements[u.Value]; u.Kind == "missing-backend" && defined {
			for _, about := range u.About {
				d.settle(u, about)
			}
			continue
		}
		text := u.Kind
		switch u.Kind {
		case "secret-endpoint":
			ns := strings.Split(strings.TrimPrefix(u.About[0], "k8s:"), "/")[0]
			text += " " + ns + "/" + strings.Replace(u.Value, "/", ": ", 1)
		default:
			if u.Value != "" {
				text += " " + short(u.Value)
			}
		}
		if u.Count > 1 {
			text += fmt.Sprintf(" ×%d", u.Count)
		}
		if len(u.About) > 0 {
			about := make([]string, len(u.About))
			for i, a := range u.About {
				about[i] = short(d.fold(a))
			}
			text += "  about " + strings.Join(about, ", ")
		}
		d.add("open", text, u.Src)
	}
}

// settle turns "route R names gateway G, defined elsewhere" into G -> R's backends, and
// "R names service S, defined elsewhere" into R's gateways -> S.
func (d *digester) settle(u model.Unresolved, route string) {
	host := d.elements[route].Desc
	var labels []string
	if host != "" {
		labels = []string{"host " + host}
	}
	switch u.Hint {
	case "gateway":
		for _, svc := range d.refsFrom(route, "backendRef") {
			d.connection(u.Value, svc, labels, u.Src)
		}
	default:
		gateways := d.refsFrom(route, "parentRef")
		if len(gateways) == 0 {
			gateways = []string{route}
		}
		for _, g := range gateways {
			d.connection(g, u.Value, labels, u.Src)
		}
	}
}

func (d *digester) components() {
	for id, e := range d.elements {
		kind := kindOf(id)
		if e.Kind == "group" || hiddenKinds[kind] {
			continue
		}
		if _, folded := d.folded[id]; folded {
			continue
		}
		text := short(id)
		if e.Tech != "" {
			text += "  " + e.Tech
		}
		if len(e.Ports) > 0 {
			ports := make([]string, len(e.Ports))
			for i, p := range e.Ports {
				ports[i] = fmt.Sprintf("%s:%d", p.Name, p.Number)
			}
			text += "  ports " + strings.Join(ports, ",")
		}
		d.add("components", text, e.Src)
	}
}
