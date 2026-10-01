// Package access answers who may act on what (ADR-0019): the groups a subject belongs
// to, what it reaches through grants, who reaches a resource, and the escalation paths
// a write grant opens. It keeps its own adjacency lists until internal/graph exists
// (ADR-0020).
package access

import (
	"slices"
	"sort"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// States filter grants and memberships by status: current drops planned ones, target
// drops deprecated ones, diff keeps both.
const (
	Current = "current"
	Target  = "target"
	Diff    = "diff"
)

var rank = map[string]int{"read": 1, "write": 2, "admin": 3}

// Stronger reports whether level a gives more than level b. Break-glass is outside the
// order: it is never stronger than a daily grant, and any daily grant beats none.
func Stronger(a, b string) bool { return rank[a] > rank[b] || (b == "" && a != "") }

// CanChange is true for levels that let a subject change what it holds.
func CanChange(level string) bool { return level == "write" || level == "admin" }

type Graph struct {
	m       *model.Model
	state   string
	groups  map[string][]model.Membership // by subject
	members map[string][]model.Membership // by group
	grants  []model.Grant
}

// New keeps the memberships and grants that count in state. An empty state is current.
func New(m *model.Model, state string) *Graph {
	g := &Graph{m: m, state: state, groups: map[string][]model.Membership{}, members: map[string][]model.Membership{}}
	for _, ms := range m.Memberships {
		if counts(ms.Status, state) {
			g.groups[ms.Subject] = append(g.groups[ms.Subject], ms)
			g.members[ms.Group] = append(g.members[ms.Group], ms)
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
	return sortedKeys(seen)
}

// Members returns every subject that belongs to group, directly or through nested
// groups, sorted.
func (g *Graph) Members(group string) []string {
	seen := map[string]bool{}
	var walk func(string)
	walk = func(id string) {
		for _, ms := range g.members[id] {
			if !seen[ms.Subject] {
				seen[ms.Subject] = true
				walk(ms.Subject)
			}
		}
	}
	walk(group)
	return sortedKeys(seen)
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

// Levels maps each resource subject reaches to its strongest level. A resource reached
// only by break-glass maps to "breakglass".
func (g *Graph) Levels(subject string) map[string]string {
	out := map[string]string{}
	for _, gr := range g.Grants(subject) {
		cur, ok := out[gr.Resource]
		switch {
		case gr.Level == "breakglass":
			if !ok {
				out[gr.Resource] = "breakglass"
			}
		case !ok || cur == "breakglass" || Stronger(gr.Level, cur):
			out[gr.Resource] = gr.Level
		}
	}
	return out
}

// Reachers returns the grants on resource and, for each grant held by a group, that
// group's members, sorted.
func (g *Graph) Reachers(resource string) (grants []model.Grant, subjects []string) {
	seen := map[string]bool{}
	for _, gr := range g.grants {
		if gr.Resource != resource {
			continue
		}
		grants = append(grants, gr)
		seen[gr.Subject] = true
		for _, id := range g.Members(gr.Subject) {
			seen[id] = true
		}
	}
	return grants, sortedKeys(seen)
}

// Escalation is a path a write grant opens: Subject can change Through, which gives
// Grant to someone, so Subject can obtain Grant.Resource too.
type Escalation struct {
	Subject string
	Through string
	Grant   model.Grant
}

// Escalations derives, for subject, every grant it can obtain by changing an element it
// holds write or admin on: a mechanism named in another grant's via, or a group that
// holds a grant. It over-approximates: changing a policy may not be enough to obtain
// it. Grants that give no more than subject already reaches are left out.
func (g *Graph) Escalations(subject string) []Escalation {
	held := map[string]bool{}
	for _, gr := range g.Grants(subject) {
		held[gr.ID] = true
	}
	levels := g.Levels(subject)
	var out []Escalation
	seen := map[string]bool{}
	for _, own := range g.Grants(subject) {
		if !CanChange(own.Level) {
			continue
		}
		for _, gr := range g.grants {
			if held[gr.ID] || seen[own.Resource+"/"+gr.ID] || !opens(gr.Level, levels[gr.Resource]) {
				continue
			}
			if gr.Subject == own.Resource || slices.Contains(gr.Via, own.Resource) {
				seen[own.Resource+"/"+gr.ID] = true
				out = append(out, Escalation{Subject: subject, Through: own.Resource, Grant: gr})
			}
		}
	}
	return out
}

// opens is true when a grant at level would give more than held.
func opens(level, held string) bool {
	return held == "" || held == "breakglass" || Stronger(level, held)
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Hop is one edge of an access graph. A grant through a mechanism gives two hops, the
// subject to the mechanism (Leg "bind", carrying Auth) and the mechanism to the
// resource (Leg "grant", carrying Level and Scope); a grant with no via gives one hop
// (Leg "direct"). Escalation hops are derived, never stored in the model.
type Hop struct {
	Kind           string // member, grant or escalation
	Leg            string // bind, grant or direct; empty for member and escalation
	From, To       string
	Level, Scope   string
	Auth           *model.Auth
	Status, Target string
	Through        string // escalation only: the element the subject can change
}

// Subgraph returns the hops an access view draws. With a subject, it walks forward:
// the subject's memberships, the grants it holds, the escalations it opens. With a
// resource, it walks backward: the grants on it and the members of the groups holding
// them. With neither, it returns every membership and grant. Hops are merged when they
// join the same two elements in the same role and status.
func (g *Graph) Subgraph(subject, resource string) []Hop {
	var hops []Hop
	addMemberships := func(keep func(model.Membership) bool) {
		for _, ms := range g.m.Memberships {
			if counted(g.groups[ms.Subject], ms.ID) && keep(ms) {
				hops = append(hops, Hop{Kind: "member", From: ms.Subject, To: ms.Group, Auth: ms.Auth, Status: ms.Status, Target: ms.Target})
			}
		}
	}
	addGrant := func(gr model.Grant) {
		if len(gr.Via) == 0 {
			hops = append(hops, Hop{Kind: "grant", Leg: "direct", From: gr.Subject, To: gr.Resource, Level: gr.Level, Scope: gr.Scope, Auth: gr.Auth, Status: gr.Status, Target: gr.Target})
			return
		}
		for _, v := range gr.Via {
			hops = append(hops,
				Hop{Kind: "grant", Leg: "bind", From: gr.Subject, To: v, Level: gr.Level, Auth: gr.Auth, Status: gr.Status, Target: gr.Target},
				Hop{Kind: "grant", Leg: "grant", From: v, To: gr.Resource, Level: gr.Level, Scope: gr.Scope, Status: gr.Status, Target: gr.Target})
		}
	}
	switch {
	case subject != "":
		holders := map[string]bool{subject: true}
		for _, id := range g.Groups(subject) {
			holders[id] = true
		}
		addMemberships(func(ms model.Membership) bool { return holders[ms.Subject] })
		for _, gr := range g.Grants(subject) {
			addGrant(gr)
		}
		// A diff shows where the migration leads, so its escalations are the target's.
		esc := g
		if g.state == Diff {
			esc = New(g.m, Target)
		}
		for _, e := range esc.Escalations(subject) {
			hops = append(hops, Hop{Kind: "escalation", From: subject, To: e.Grant.Resource, Level: e.Grant.Level, Through: e.Through})
		}
	case resource != "":
		grants, _ := g.Reachers(resource)
		groups := map[string]bool{}
		for _, gr := range grants {
			addGrant(gr)
			groups[gr.Subject] = true
			for _, id := range g.Members(gr.Subject) {
				groups[id] = true
			}
		}
		addMemberships(func(ms model.Membership) bool { return groups[ms.Group] && groups[ms.Subject] })
	default:
		addMemberships(func(model.Membership) bool { return true })
		for _, gr := range g.grants {
			addGrant(gr)
		}
	}
	return merge(hops)
}

func counted(list []model.Membership, id string) bool {
	for _, ms := range list {
		if ms.ID == id {
			return true
		}
	}
	return false
}

// merge folds hops that join the same two elements in the same role and status: the
// strongest level wins and scopes are listed once each.
func merge(hops []Hop) []Hop {
	var out []Hop
	at := map[string]int{}
	for _, h := range hops {
		key := h.Kind + "|" + h.Leg + "|" + h.From + "|" + h.To + "|" + h.Status + "|" + h.Through
		i, ok := at[key]
		if !ok {
			at[key] = len(out)
			out = append(out, h)
			continue
		}
		if Stronger(h.Level, out[i].Level) {
			out[i].Level = h.Level
		}
		if h.Scope != "" && !containsScope(out[i].Scope, h.Scope) {
			if out[i].Scope != "" {
				out[i].Scope += ", "
			}
			out[i].Scope += h.Scope
		}
	}
	return out
}

func containsScope(list, scope string) bool {
	return list != "" && slices.Contains(strings.Split(list, ", "), scope)
}
