package access_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/TheGostsniperfr/Noodle/internal/access"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// platform: a member of tenant A through two nested groups, a tenant B group, a
// provisioning service losing its static admin token for scoped grants, and a root
// account kept for break-glass.
const platform = `
memberships:
  - {id: m-alice, subject: alice, group: team-a, auth: {method: oidc, mfa: true}}
  - {id: m-team-a, subject: team-a, group: devs}
  - {id: m-bob, subject: bob, group: team-b}
  - {id: m-cmp, subject: cmp, group: cmp-role, auth: {method: k8s-sa, lifetime: 1h}, status: planned, target: P2}
grants:
  - {id: g-a, subject: team-a, resource: vault-a, level: write, via: [policy-a]}
  - {id: g-a-read, subject: devs, resource: vault-a, level: read}
  - {id: g-b, subject: team-b, resource: vault-b, level: write, via: [policy-b]}
  - {id: g-cmp-root, subject: cmp, resource: vault, level: admin, via: [root-token], auth: {method: static-token}, status: deprecated}
  - {id: g-cmp-policies, subject: cmp-role, resource: policy-b, level: write, scope: "project-*", via: [policy-cmp], status: planned, target: P2}
  - {id: g-root, subject: root, resource: vault, level: breakglass}
  - {id: g-root-read, subject: root, resource: vault-a, level: breakglass}
`

func load(t *testing.T) *model.Model {
	t.Helper()
	var m model.Model
	if err := yaml.Unmarshal([]byte(platform), &m); err != nil {
		t.Fatal(err)
	}
	return &m
}

func TestGroups_FollowsNestedMemberships(t *testing.T) {
	t.Parallel()
	g := access.New(load(t), access.Current)

	groups := g.Groups("alice")

	assert.Equal(t, []string{"devs", "team-a"}, groups)
}

func TestMembers_FollowsNestedMembershipsBackward(t *testing.T) {
	t.Parallel()
	g := access.New(load(t), access.Current)

	members := g.Members("devs")

	assert.Equal(t, []string{"alice", "team-a"}, members)
}

func TestLevels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, state, subject string
		want                 map[string]string
	}{
		{"strongest level wins across groups", access.Current, "alice", map[string]string{"vault-a": "write"}},
		{"a tenant reaches only its own resources", access.Current, "bob", map[string]string{"vault-b": "write"}},
		{"current keeps deprecated grants and drops planned ones", access.Current, "cmp", map[string]string{"vault": "admin"}},
		{"target drops deprecated grants and keeps planned ones", access.Target, "cmp", map[string]string{"policy-b": "write"}},
		{"break-glass is reported as such", access.Current, "root", map[string]string{"vault": "breakglass", "vault-a": "breakglass"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := access.New(load(t), tt.state)

			levels := g.Levels(tt.subject)

			assert.Equal(t, tt.want, levels)
		})
	}
}

func TestReachers_IncludesMembersOfGrantedGroups(t *testing.T) {
	t.Parallel()
	g := access.New(load(t), access.Current)

	_, subjects := g.Reachers("vault-a")

	assert.Equal(t, []string{"alice", "devs", "root", "team-a"}, subjects)
}

func TestEscalations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, state, subject string
		want                 []string // through → grant
	}{
		{"write on another tenant's policy opens its grant", access.Target, "cmp", []string{"policy-b → g-b"}},
		{"no write grant, no escalation", access.Current, "alice", nil},
		{"current state has no planned write grant", access.Current, "cmp", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := access.New(load(t), tt.state)

			var got []string
			for _, e := range g.Escalations(tt.subject) {
				got = append(got, e.Through+" → "+e.Grant.ID)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEscalations_ThroughAGroupItCanChange(t *testing.T) {
	t.Parallel()
	m := load(t)
	m.Grants = append(m.Grants, model.Grant{ID: "g-admin-b", Subject: "alice", Resource: "team-b", Level: "admin"})
	g := access.New(m, access.Current)

	escalations := g.Escalations("alice")

	assert.Len(t, escalations, 1)
	assert.Equal(t, "g-b", escalations[0].Grant.ID)
}

func TestSubgraph(t *testing.T) {
	t.Parallel()
	hop := func(h access.Hop) string {
		s := h.Kind + " " + h.From + "→" + h.To
		if h.Leg != "" {
			s += " " + h.Leg
		}
		if h.Level != "" {
			s += " " + h.Level
		}
		if h.Status != "" {
			s += " " + h.Status
		}
		return s
	}
	tests := []struct {
		name, state, subject, resource string
		want                           []string
	}{
		{"forward from a member: memberships, then grants split at their mechanism", access.Current, "alice", "", []string{
			"member alice→team-a", "member team-a→devs",
			"grant team-a→policy-a bind write", "grant policy-a→vault-a grant write", "grant devs→vault-a direct read",
		}},
		{"forward in target: the new role, its grant and the escalation it opens", access.Target, "cmp", "", []string{
			"member cmp→cmp-role planned",
			"grant cmp-role→policy-cmp bind write planned", "grant policy-cmp→policy-b grant write planned",
			"escalation cmp→vault-b write",
		}},
		{"diff keeps removed and planned hops side by side", access.Diff, "cmp", "", []string{
			"member cmp→cmp-role planned",
			"grant cmp→root-token bind admin deprecated", "grant root-token→vault grant admin deprecated",
			"grant cmp-role→policy-cmp bind write planned", "grant policy-cmp→policy-b grant write planned",
			"escalation cmp→vault-b write",
		}},
		{"backward from a resource: grants on it and members of their groups", access.Current, "", "vault-a", []string{
			"member alice→team-a", "member team-a→devs",
			"grant team-a→policy-a bind write", "grant policy-a→vault-a grant write",
			"grant devs→vault-a direct read", "grant root→vault-a direct breakglass",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := access.New(load(t), tt.state)

			var got []string
			for _, h := range g.Subgraph(tt.subject, tt.resource) {
				got = append(got, hop(h))
			}

			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestSubgraph_MergesScopesOnTheSameHop(t *testing.T) {
	t.Parallel()
	m := &model.Model{Grants: []model.Grant{
		{ID: "g1", Subject: "cmp", Resource: "vault", Level: "write", Scope: "project-*", Via: []string{"p"}},
		{ID: "g2", Subject: "cmp", Resource: "vault", Level: "admin", Scope: "sys/mounts", Via: []string{"p"}},
	}}
	g := access.New(m, access.Current)

	hops := g.Subgraph("cmp", "")

	assert.Len(t, hops, 2)
	assert.Equal(t, "admin", hops[1].Level)
	assert.Equal(t, "project-*, sys/mounts", hops[1].Scope)
}
