package access_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/TheGostsniperfr/Noodle/internal/access"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// platform: a member of tenant A through two nested groups, one membership the plan
// removes, a provisioning service losing its static admin token for scoped grants, and
// a root account kept for break-glass.
const platform = `
memberships:
  - {id: m-alice, subject: alice, group: team-a, auth: {method: oidc, mfa: true}}
  - {id: m-team-a, subject: team-a, group: devs}
  - {id: m-alice-all, subject: alice, group: everyone, status: deprecated, target: P3}
  - {id: m-cmp, subject: cmp, group: cmp-role, status: planned, target: P2}
grants:
  - {id: g-a, subject: team-a, resource: vault-a, level: write, via: [policy-a]}
  - {id: g-a-read, subject: devs, resource: vault-a, level: read}
  - {id: g-all, subject: everyone, resource: apps, level: read}
  - {id: g-cmp-root, subject: cmp, resource: vault, level: admin, status: deprecated, target: P2}
  - {id: g-cmp-new, subject: cmp-role, resource: vault, level: write, status: planned, target: P2}
  - {id: g-root, subject: root, resource: vault, level: breakglass}
  - {id: g-root-daily, subject: root, resource: vault, level: admin, status: deprecated, target: P1}
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
	g := access.New(load(t), access.Target)

	groups := g.Groups("alice")

	assert.Equal(t, []string{"devs", "team-a"}, groups)
}

func TestLevels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, state, subject string
		want                 map[string]string
	}{
		{"strongest level wins across nested groups", access.Target, "alice", map[string]string{"vault-a": "write"}},
		{"a membership the plan removes still counts today", access.Current, "alice", map[string]string{"vault-a": "write", "apps": "read"}},
		{"current keeps deprecated grants and drops planned ones", access.Current, "cmp", map[string]string{"vault": "admin"}},
		{"target drops deprecated grants and keeps planned ones", access.Target, "cmp", map[string]string{"vault": "write"}},
		{"break-glass loses to a daily level", access.Current, "root", map[string]string{"vault": "admin"}},
		{"break-glass alone is reported as such", access.Target, "root", map[string]string{"vault": "breakglass"}},
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

func TestTopGrants_KeepsOnlyTheGrantsAtTheStrongestLevel(t *testing.T) {
	t.Parallel()
	m := load(t)
	m.Grants = append(m.Grants, model.Grant{ID: "g-alice-admin", Subject: "alice", Resource: "apps", Level: "admin"})
	g := access.New(m, access.Current)

	top := g.TopGrants("alice")["apps"]

	assert.Equal(t, map[string]bool{"g-alice-admin": true}, top)
}

func TestPhase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, subject, resource, want string
	}{
		{"from the grants that change", "cmp", "vault", "P2"},
		{"from a membership that changes", "alice", "apps", "P3"},
		{"several phases on one cell", "root", "vault", "P1"},
		{"a grant's own phase wins over its membership's", "cmp", "vault", "P2"},
		{"nothing changes", "alice", "vault-a", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := access.New(load(t), access.Diff)

			phase := g.Phase(tt.subject, tt.resource)

			assert.Equal(t, tt.want, phase)
		})
	}
}

func TestStronger(t *testing.T) {
	t.Parallel()
	assert.True(t, access.Stronger("read", ""))
	assert.True(t, access.Stronger("breakglass", ""))
	assert.True(t, access.Stronger("read", "breakglass"))
	assert.False(t, access.Stronger("breakglass", "read"))
	assert.True(t, access.Stronger("admin", "write"))
}
