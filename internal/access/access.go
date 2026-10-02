// Package access answers who may act on what (ADR-0019): the groups a subject belongs
// to, the grants it holds through them, and its strongest level on each resource in a
// state. It keeps its own adjacency lists until internal/graph exists (ADR-0020).
package access

import (
	"sort"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// States filter memberships and grants by status: current drops planned items, target
// drops deprecated ones, diff keeps both.
const (
	Current = "current"
	Target  = "target"
	Diff    = "diff"
)

var rank = map[string]int{"read": 1, "write": 2, "admin": 3}

// Stronger reports whether level a gives more than level b. Break-glass sits below any
// daily level and above none: it only counts where nothing else applies.
func Stronger(a, b string) bool {
	switch {
	case a == b || a == "":
		return false
	case b == "":
		return true
	case a == "breakglass":
		return false
	case b == "breakglass":
		return true
	}
	return rank[a] > rank[b]
}

// CanChange is true for levels that let a subject change a resource.
func CanChange(level string) bool { return level == "write" || level == "admin" }

type Graph struct {
	m      *model.Model
	groups map[string][]model.Membership // by subject
	grants []model.Grant
}

// New keeps the memberships and grants that count in state. An empty state is current.
func New(m *model.Model, state string) *Graph {
	g := &Graph{m: m, groups: map[string][]model.Membership{}}
	for _, ms := range m.Memberships {
		if counts(ms.Status, state) {
			g.groups[ms.Subject] = append(g.groups[ms.Subject], ms)
		}
	}
	for _, gr := range m.Grants {
		if counts(gr.Status, state) {
			g.grants = append(g.grants, gr)
		}
	}
	return g
}

func counts(status, state string) bool {
	switch state {
	case Target:
		return status != "deprecated"
	case Diff:
		return true
	default:
		return status != "planned"
	}
}

// Groups returns every group subject belongs to, directly or through other groups,
// sorted. Check rejects membership cycles; a visited set guards anyway.
func (g *Graph) Groups(subject string) []string {
	seen := map[string]bool{}
	var walk func(string)
	walk = func(id string) {
		for _, ms := range g.groups[id] {
			if !seen[ms.Group] {
				seen[ms.Group] = true
				walk(ms.Group)
			}
		}
	}
	walk(subject)
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// Grants returns the grants subject holds itself or through its groups, in model order.
func (g *Graph) Grants(subject string) []model.Grant {
	holders := map[string]bool{subject: true}
	for _, id := range g.Groups(subject) {
		holders[id] = true
	}
	var out []model.Grant
	for _, gr := range g.grants {
		if holders[gr.Subject] {
			out = append(out, gr)
		}
	}
	return out
}

// Levels maps each resource subject reaches to its strongest level.
func (g *Graph) Levels(subject string) map[string]string {
	out := map[string]string{}
	for _, gr := range g.Grants(subject) {
		if Stronger(gr.Level, out[gr.Resource]) {
			out[gr.Resource] = gr.Level
		}
	}
	return out
}

// GrantIDs maps each resource subject reaches to the sorted ids of the grants that give
// it, so two states with the same level but different grants can be told apart.
func (g *Graph) GrantIDs(subject string) map[string]string {
	ids := map[string][]string{}
	for _, gr := range g.Grants(subject) {
		ids[gr.Resource] = append(ids[gr.Resource], gr.ID)
	}
	out := map[string]string{}
	for res, list := range ids {
		sort.Strings(list)
		out[res] = strings.Join(list, ",")
	}
	return out
}

// Phase returns the targets of the planned or deprecated items that change subject's
// access to resource: the grants on it and the memberships leading to their holders.
// Unique, sorted, joined with "+"; empty when nothing names a phase. The graph should
// be built in the diff state.
func (g *Graph) Phase(subject, resource string) string {
	seen := map[string]bool{}
	holders := map[string]bool{subject: true}
	for _, id := range g.Groups(subject) {
		holders[id] = true
	}
	for _, gr := range g.grants {
		if gr.Resource == resource && holders[gr.Subject] && gr.Status != "" && gr.Target != "" {
			seen[gr.Target] = true
		}
	}
	for id := range holders {
		for _, ms := range g.groups[id] {
			if ms.Status != "" && ms.Target != "" && g.reaches(ms.Group, resource) {
				seen[ms.Target] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Strings(out)
	return strings.Join(out, "+")
}

// reaches is true when group, or a group it belongs to, holds a grant on resource.
func (g *Graph) reaches(group, resource string) bool {
	for _, gr := range g.Grants(group) {
		if gr.Resource == resource {
			return true
		}
	}
	return false
}
