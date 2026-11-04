package graph

import (
	"sort"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// curatedFile is the source of an entry that comes from the curated model, which keeps
// no line numbers.
const curatedFile = "model.yaml"

// Entry is one finding of a drift report. ID names the element or connection, Detail the
// discovered id or ids behind it, Src where to look.
type Entry struct {
	ID     string
	Detail string
	Src    model.Src
}

// Report is the gap between the curated model and what was discovered (ADR-0015). It
// is a report: nothing in it is an error.
type Report struct {
	// Unmatched are discovered components no curated element matches.
	Unmatched []Entry
	// Gone are matched ids no fragment contains any more.
	Gone []Entry
	// Planned are planned elements still absent from the fragments.
	Planned []Entry
	// Unmodelled are discovered connections between matched elements that the model lacks.
	Unmodelled []Entry
}

// Drift compares the curated model with the fragments. Every family is sorted.
func (g *Graph) Drift() Report {
	var r Report
	for id, d := range g.discovered {
		if _, matched := g.matchedBy[id]; !matched && !d.IsZone() {
			r.Unmatched = append(r.Unmatched, Entry{ID: id, Src: d.Src})
		}
	}
	for _, e := range g.Model.Elements {
		planned := e.Status == "planned"
		present := false
		for _, match := range e.Matches {
			if g.contains(match) {
				present = true
			} else if !planned {
				r.Gone = append(r.Gone, Entry{ID: e.ID, Detail: match, Src: model.Src{File: curatedFile}})
			}
		}
		if planned && !present {
			r.Planned = append(r.Planned, Entry{ID: e.ID, Src: model.Src{File: curatedFile}})
		}
	}
	for _, e := range g.unmodelled {
		r.Unmodelled = append(r.Unmodelled, e)
	}
	for _, family := range [][]Entry{r.Unmatched, r.Gone, r.Planned, r.Unmodelled} {
		sort.Slice(family, func(i, j int) bool {
			if family[i].ID != family[j].ID {
				return family[i].ID < family[j].ID
			}
			return family[i].Detail < family[j].Detail
		})
	}
	return r
}

// contains tells whether a fragment holds id, as an element or as the end of an edge.
func (g *Graph) contains(id string) bool {
	if _, ok := g.discovered[id]; ok {
		return true
	}
	for _, e := range g.edges {
		if e.From == id || e.To == id {
			return true
		}
	}
	return false
}

// settleAcrossFragments resolves what one fragment could not see: a missing-backend
// entry whose target another fragment defines becomes an edge, and is dropped from
// Unresolved (FR-012).
func (g *Graph) settleAcrossFragments(frags []*model.Fragment) {
	kinds := map[string]string{"gateway": "parentRef", "service": "backendRef"}
	for _, f := range frags {
		for _, u := range f.Unresolved {
			kind, linkable := kinds[u.Hint]
			if _, defined := g.discovered[u.Value]; u.Kind != "missing-backend" || !linkable || !defined {
				g.Unresolved = append(g.Unresolved, u)
				continue
			}
			for _, from := range u.About {
				g.edges = append(g.edges, Edge{ID: "xref:" + from + "->" + u.Value, From: from, To: u.Value, Kind: kind, Reference: true, Src: u.Src})
			}
		}
	}
}
