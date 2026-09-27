package model

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Finding names the file and the id at fault, as FR-005 requires.
type Finding struct {
	File, ID, Msg string
}

func (f Finding) String() string { return fmt.Sprintf("%s: %s: %s", f.File, f.ID, f.Msg) }

var (
	elementKinds    = set("frontend", "backend", "database", "cloud", "security", "bus", "external", "region", "group")
	connectionKinds = set("flow", "auth", "tunnel", "async")
	statuses        = set("", "planned", "deprecated")
	shapes          = set("", "box", "cylinder", "actor")
	viewTypes       = set("topology", "sequence")
	sides           = set("left", "right", "top", "bottom")
)

func set(vs ...string) map[string]bool {
	m := map[string]bool{}
	for _, v := range vs {
		m[v] = true
	}
	return m
}

type checker struct {
	s        *System
	findings []Finding
	elements map[string]Element
	conns    map[string]Connection
	edgeIDs  map[string]bool
}

func (c *checker) errf(file, id, format string, args ...any) {
	c.findings = append(c.findings, Finding{file, id, fmt.Sprintf(format, args...)})
}

// Check reports every broken reference and invalid value across the model, its views
// and their layouts. An empty result means the system can be resolved.
func Check(s *System) []Finding {
	c := &checker{s: s, elements: map[string]Element{}, conns: map[string]Connection{}, edgeIDs: map[string]bool{}}
	c.checkModel()
	for _, id := range s.ViewIDs() {
		c.checkView(s.Views[id])
	}
	for _, id := range sortedKeys(s.Layouts) {
		c.checkLayout(s.Layouts[id])
	}
	sort.SliceStable(c.findings, func(i, j int) bool {
		if c.findings[i].File != c.findings[j].File {
			return c.findings[i].File < c.findings[j].File
		}
		return c.findings[i].ID < c.findings[j].ID
	})
	return c.findings
}

func (c *checker) checkModel() {
	const file = "model.yaml"
	seen := map[string]bool{}
	unique := func(id string) {
		if id == "" {
			c.errf(file, "(empty)", "missing id")
		} else if seen[id] {
			c.errf(file, id, "duplicate id")
		}
		seen[id] = true
	}
	for _, e := range c.s.Model.Elements {
		unique(e.ID)
		c.elements[e.ID] = e
	}
	for _, e := range c.s.Model.Elements {
		if !elementKinds[e.Kind] {
			c.errf(file, e.ID, "unknown element kind %q", e.Kind)
		}
		if !statuses[e.Status] {
			c.errf(file, e.ID, "unknown status %q, want planned or deprecated", e.Status)
		}
		if !shapes[e.Shape] {
			c.errf(file, e.ID, "unknown shape %q", e.Shape)
		}
		if e.Parent != "" {
			if p, ok := c.elements[e.Parent]; !ok {
				c.errf(file, e.ID, "parent %q does not exist", e.Parent)
			} else if !p.IsZone() {
				c.errf(file, e.ID, "parent %q is not a zone", e.Parent)
			}
		}
	}
	for _, cn := range c.s.Model.Connections {
		unique(cn.ID)
		c.conns[cn.ID] = cn
		c.edgeIDs[cn.ID] = true
		if !connectionKinds[cn.Kind] {
			c.errf(file, cn.ID, "unknown connection kind %q", cn.Kind)
		}
		c.mustElement(file, cn.ID, "from", cn.From)
		if to, ok := c.mustElement(file, cn.ID, "to", cn.To); ok && cn.Port != "" && !hasPort(to, cn.Port) {
			c.errf(file, cn.ID, "port %q is not a port of %s", cn.Port, cn.To)
		}
	}
	for _, r := range c.s.Model.References {
		unique(r.ID)
		c.edgeIDs[r.ID] = true
		c.mustElement(file, r.ID, "from", r.From)
		c.mustElement(file, r.ID, "to", r.To)
	}
	for _, a := range c.s.Model.Annotations {
		unique(a.ID)
		for _, t := range a.Targets {
			if _, ok := c.elements[t]; !ok && !c.edgeIDs[t] {
				c.errf(file, a.ID, "target %q does not exist in the model", t)
			}
		}
	}
}

func (c *checker) mustElement(file, id, role, ref string) (Element, bool) {
	e, ok := c.elements[ref]
	if !ok {
		c.errf(file, id, "%s %q does not exist in the model", role, ref)
	}
	return e, ok
}

func hasPort(e Element, name string) bool {
	for _, p := range e.Ports {
		if p.Name == name {
			return true
		}
	}
	return false
}

func (c *checker) checkView(v *View) {
	file := filepath.Join("views", v.ID+".yaml")
	if !viewTypes[v.Type] {
		c.errf(file, v.ID, "type %q, want topology or sequence", v.Type)
	}
	for _, inc := range v.Include {
		c.mustElement(file, v.ID, "include", strings.TrimSuffix(inc, "/**"))
	}
	for _, list := range [][]string{v.Steps, v.Background} {
		for _, id := range list {
			if _, ok := c.conns[id]; !ok {
				c.errf(file, id, "step is not a connection of the model")
			}
		}
	}
	for _, id := range sortedKeys(v.Labels) {
		if !c.edgeIDs[id] {
			c.errf(file, id, "label for an unknown connection or reference")
		}
	}
	if len(v.Participants) > 0 && v.Type != "sequence" {
		c.errf(file, v.ID, "participants are for sequence views only")
	}
	for _, p := range v.Participants {
		c.mustElement(file, v.ID, "participant", p)
	}
	_, hasLayout := c.s.Layouts[v.ID]
	if v.Type == "topology" && !hasLayout {
		c.errf(file, v.ID, "topology view has no layouts/%s.yaml", v.ID)
	}
	if v.Type == "sequence" && hasLayout {
		c.errf(file, v.ID, "sequence views are computed and take no layout (ADR-0007)")
	}
}

func (c *checker) checkLayout(l *Layout) {
	file := filepath.Join("layouts", l.ID+".yaml")
	v, ok := c.s.Views[l.ID]
	if !ok {
		c.errf(file, l.ID, "no view views/%s.yaml for this layout", l.ID)
		return
	}
	for _, id := range sortedKeys(l.Elements) {
		e, ok := c.mustElement(file, id, "element", id)
		if _, placed := l.Elements[e.Parent]; ok && e.Parent != "" && !placed {
			c.errf(file, id, "position is relative to zone %q, which has no position", e.Parent)
		}
	}
	for _, id := range c.included(v) {
		if _, ok := l.Elements[id]; !ok {
			c.errf(file, id, "included element has no position")
		}
	}
	for _, id := range sortedKeys(l.Edges) {
		r := l.Edges[id]
		if !c.edgeIDs[id] {
			c.errf(file, id, "edge is not a connection or reference of the model")
		}
		for _, end := range []string{r.From, r.To} {
			c.checkEndpoint(file, id, end)
		}
		for _, w := range r.Waypoints {
			if w.Lane != "" {
				if _, ok := l.Lanes[w.Lane]; !ok {
					c.errf(file, id, "lane %q is not defined", w.Lane)
				}
			}
		}
	}
	for _, name := range sortedKeys(l.Lanes) {
		if lane := l.Lanes[name]; (lane.X == nil) == (lane.Y == nil) {
			c.errf(file, name, "lane needs exactly one of x or y")
		}
	}
	cards, notes := map[string]bool{}, map[string]bool{}
	for _, cd := range v.Cards {
		cards[cd.ID] = true
	}
	for _, n := range v.Notes {
		notes[n.ID] = true
	}
	for _, id := range sortedKeys(l.Cards) {
		if !cards[id] {
			c.errf(file, id, "card is not defined in the view")
		}
	}
	for _, id := range sortedKeys(l.Notes) {
		if !notes[id] {
			c.errf(file, id, "note is not defined in the view")
		}
	}
}

// checkEndpoint validates the element and side of "id.side[@NN%]"; the ratio's range
// is the resolver's concern.
func (c *checker) checkEndpoint(file, edgeID, end string) {
	id, anchor := Endpoint(end)
	c.mustElement(file, edgeID, "endpoint", id)
	side, _, _ := strings.Cut(anchor, "@")
	if !sides[side] {
		c.errf(file, edgeID, "endpoint %q: side %q, want left, right, top or bottom", end, side)
	}
}

// included returns the element ids a view shows, sorted. No include means all.
func (c *checker) included(v *View) []string {
	if len(v.Include) == 0 {
		return sortedKeys(c.elements)
	}
	children := map[string][]string{}
	for _, e := range c.s.Model.Elements {
		if e.Parent != "" {
			children[e.Parent] = append(children[e.Parent], e.ID)
		}
	}
	out := map[string]bool{}
	var walk func(string)
	walk = func(id string) {
		if out[id] {
			return
		}
		out[id] = true
		for _, ch := range children[id] {
			walk(ch)
		}
	}
	for _, inc := range v.Include {
		if root, ok := strings.CutSuffix(inc, "/**"); ok {
			walk(root)
		} else if _, ok := c.elements[inc]; ok {
			out[inc] = true
		}
	}
	return sortedKeys(out)
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
