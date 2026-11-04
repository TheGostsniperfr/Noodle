// Package graph merges the curated model with the fragments adapters discovered and
// answers questions about the result: lookup, neighbours, drift (ADR-0015). It reads
// only the model contracts, so another backend could replace it without touching an
// adapter.
package graph

import (
	"sort"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// Source is where one discovered item behind a curated element or connection comes from.
type Source struct {
	ID     string
	Src    model.Src
	Object *model.Object
}

// Edge is a discovered connection or reference, in discovered ids.
type Edge struct {
	ID        string
	From, To  string
	Kind      string
	Port      string
	Reference bool
	Inferred  bool
	Src       model.Src
	Object    *model.Object
}

// Node is what an id stands for: the merged element for a curated id, the discovered
// element for a discovered id, and just the id when only an edge names it.
type Node struct {
	ID      string
	Element model.Element
	Sources []Source
}

// Graph is the merged model and the discovered facts behind it. The curated model it was
// built from is never modified.
type Graph struct {
	// Model is the curated model with fragments merged in.
	Model *model.Model
	// Sources lists, by curated element or connection id, the discovered items merged into it.
	Sources map[string][]Source

	curated    map[string]model.Element
	discovered map[string]model.DiscoveredElement
	edges      []Edge
	matchedBy  map[string]string
}

// Merge applies ADR-0015: a field the curated element leaves empty is filled from the
// first matched discovered element that has it, a list whole and never item by item; a
// field it sets is kept; a discovered connection between matched elements is mapped onto
// the curated ids. Fragments are read in id order and the first definition of an id wins,
// so the result does not depend on the order the caller passes them in.
func Merge(m *model.Model, fragments ...*model.Fragment) *Graph {
	frags := append([]*model.Fragment{}, fragments...)
	sort.SliceStable(frags, func(i, j int) bool { return frags[i].ID < frags[j].ID })
	g := &Graph{Sources: map[string][]Source{}, curated: map[string]model.Element{},
		discovered: map[string]model.DiscoveredElement{}, matchedBy: map[string]string{}}
	merged := *m
	merged.Elements = make([]model.Element, len(m.Elements))
	merged.Connections = append([]model.Connection{}, m.Connections...)
	g.Model = &merged

	for _, f := range frags {
		for _, e := range f.Elements {
			if _, taken := g.discovered[e.ID]; !taken {
				g.discovered[e.ID] = e
			}
		}
		for _, c := range f.Connections {
			g.edges = append(g.edges, Edge{ID: c.ID, From: c.From, To: c.To, Kind: c.Kind, Port: c.Port, Inferred: c.Inferred, Src: c.Src, Object: c.Object})
		}
		for _, r := range f.References {
			g.edges = append(g.edges, Edge{ID: r.ID, From: r.From, To: r.To, Kind: r.Kind, Reference: true, Src: r.Src, Object: r.Object})
		}
	}
	sort.SliceStable(g.edges, func(i, j int) bool { return g.edges[i].ID < g.edges[j].ID })

	for i, e := range m.Elements {
		e.Ports, e.Tags, e.Matches = clone(e.Ports), clone(e.Tags), clone(e.Matches)
		for _, id := range e.Matches {
			d, ok := g.discovered[id]
			if !ok {
				continue
			}
			g.matchedBy[id] = e.ID
			g.Sources[e.ID] = append(g.Sources[e.ID], Source{ID: id, Src: d.Src, Object: d.Object})
			fill(&e, d.Element)
		}
		merged.Elements[i] = e
		g.curated[e.ID] = e
	}
	g.mapConnections(m, frags)
	return g
}

func clone[T any](s []T) []T {
	if s == nil {
		return nil
	}
	return append([]T{}, s...)
}

// fill takes from d every field e leaves empty. Parent is not one: it names a discovered
// zone, which no curated element is.
func fill(e *model.Element, d model.Element) {
	fillString(&e.Kind, d.Kind)
	fillString(&e.Title, d.Title)
	fillString(&e.Tech, d.Tech)
	fillString(&e.Desc, d.Desc)
	fillString(&e.Icon, d.Icon)
	fillString(&e.Multiplicity, d.Multiplicity)
	if len(e.Ports) == 0 {
		e.Ports = clone(d.Ports)
	}
	if len(e.Tags) == 0 {
		e.Tags = clone(d.Tags)
	}
}

func fillString(dst *string, from string) {
	if *dst == "" {
		*dst = from
	}
}

func (g *Graph) mapConnections(m *model.Model, frags []*model.Fragment) {
	type key struct{ from, to, port string }
	existing := map[[2]string]string{}
	for _, c := range m.Connections {
		if _, dup := existing[[2]string{c.From, c.To}]; !dup {
			existing[[2]string{c.From, c.To}] = c.ID
		}
	}
	added := map[key]string{}
	for _, f := range frags {
		for _, c := range f.Connections {
			from, okFrom := g.matchedBy[c.From]
			to, okTo := g.matchedBy[c.To]
			if !okFrom || !okTo || from == to {
				continue
			}
			src := Source{ID: c.ID, Src: c.Src, Object: c.Object}
			if id, ok := existing[[2]string{from, to}]; ok {
				g.Sources[id] = append(g.Sources[id], src)
				continue
			}
			port := c.Port
			if !hasPort(g.curated[to], port) {
				port = ""
			}
			if id, ok := added[key{from, to, port}]; ok {
				g.Sources[id] = append(g.Sources[id], src)
				continue
			}
			var enforcedBy []string
			for _, id := range c.EnforcedBy {
				if mapped, ok := g.matchedBy[id]; ok {
					enforcedBy = append(enforcedBy, mapped)
				}
			}
			g.Model.Connections = append(g.Model.Connections, model.Connection{ID: c.ID, From: from, To: to, Port: port,
				Protocol: c.Protocol, Verb: c.Verb, Kind: c.Kind, Denied: c.Denied, EnforcedBy: enforcedBy})
			added[key{from, to, port}] = c.ID
			g.Sources[c.ID] = append(g.Sources[c.ID], src)
		}
	}
}

func hasPort(e model.Element, name string) bool {
	if name == "" {
		return false
	}
	for _, p := range e.Ports {
		if p.Name == name {
			return true
		}
	}
	return false
}

// Lookup finds an id: a curated element, a discovered element, or an id only an edge
// names, such as a Secret.
func (g *Graph) Lookup(id string) (Node, bool) {
	if e, ok := g.curated[id]; ok {
		return Node{ID: id, Element: e, Sources: g.Sources[id]}, true
	}
	if d, ok := g.discovered[id]; ok {
		return Node{ID: id, Element: d.Element, Sources: []Source{{ID: id, Src: d.Src, Object: d.Object}}}, true
	}
	for _, e := range g.edges {
		if e.From == id || e.To == id {
			return Node{ID: id, Element: model.Element{ID: id}}, true
		}
	}
	return Node{}, false
}

// Neighbours lists the discovered edges that touch id, of one kind or of every kind when
// kind is empty, sorted by kind, ends and source. A curated id answers for every
// discovered id it matches.
func (g *Graph) Neighbours(id, kind string) []Edge {
	ids := map[string]bool{id: true}
	if e, ok := g.curated[id]; ok {
		for _, match := range e.Matches {
			ids[match] = true
		}
	}
	var out []Edge
	for _, e := range g.edges {
		if (ids[e.From] || ids[e.To]) && (kind == "" || e.Kind == kind) {
			out = append(out, e)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		switch {
		case a.Kind != b.Kind:
			return a.Kind < b.Kind
		case a.From != b.From:
			return a.From < b.From
		case a.To != b.To:
			return a.To < b.To
		case a.Src.File != b.Src.File:
			return a.Src.File < b.Src.File
		}
		return a.Src.Line < b.Src.Line
	})
	return out
}
