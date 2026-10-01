package resolve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
	"github.com/TheGostsniperfr/Noodle/internal/model"
	"github.com/TheGostsniperfr/Noodle/internal/resolve"
)

func accessSpec(t *testing.T, view string) *diagram.Spec {
	t.Helper()
	s, err := model.LoadSystem("../../examples/access")
	require.NoError(t, err)
	require.Empty(t, model.Check(s))
	spec, err := resolve.View(s, view)
	require.NoError(t, err)
	return spec
}

func TestAccess_EveryExampleViewPassesTheLint(t *testing.T) {
	t.Parallel()
	for _, view := range []string{"cmp-current", "cmp-target", "cmp-diff", "alice-diff", "vault-a"} {
		t.Run(view, func(t *testing.T) {
			t.Parallel()
			spec := accessSpec(t, view)

			findings := lint.Lint(spec)

			assert.Empty(t, findings)
		})
	}
}

func TestAccess_PutsTheSubjectFirstAndResourcesLast(t *testing.T) {
	t.Parallel()
	spec := accessSpec(t, "cmp-target")
	x := map[string]float64{}
	for _, n := range spec.Nodes {
		x[n.ID] = n.X
	}

	assert.Less(t, x["cmp"], x["role-cmp"])
	assert.Less(t, x["policy-cmp"], x["tfstate"])
	assert.Equal(t, x["tfstate"], x["vault-a"], "sinks share the last column")
}

func TestAccess_ShowsStatusOnlyInADiff(t *testing.T) {
	t.Parallel()
	statuses := func(spec *diagram.Spec) map[string]bool {
		out := map[string]bool{}
		for _, e := range spec.Edges {
			out[e.Status] = true
		}
		return out
	}

	assert.Equal(t, map[string]bool{"": true}, statuses(accessSpec(t, "cmp-target")))
	assert.True(t, statuses(accessSpec(t, "cmp-diff"))["deprecated"])
}

func TestAccess_DrawsEscalationsTheTargetOpens(t *testing.T) {
	t.Parallel()
	spec := accessSpec(t, "cmp-target")

	var escalations []string
	for _, e := range spec.Edges {
		if e.Kind == "escalation" {
			escalations = append(escalations, e.To)
		}
	}

	assert.ElementsMatch(t, []string{"vault-a", "vault-b"}, escalations)
}

func TestAccess_RegressionFixturePassesTheLint(t *testing.T) {
	t.Parallel()
	s, err := model.LoadSystem("testdata/access-regression")
	require.NoError(t, err)
	require.Empty(t, model.Check(s))
	for _, view := range []string{"dev-a-diff", "cmp-current"} {
		t.Run(view, func(t *testing.T) {
			t.Parallel()
			spec, err := resolve.View(s, view)
			require.NoError(t, err)

			findings := lint.Lint(spec)

			assert.Empty(t, findings)
		})
	}
}
