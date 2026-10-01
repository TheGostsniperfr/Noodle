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

func lensSpec(t *testing.T, lens string) *diagram.Spec {
	t.Helper()
	s, err := model.LoadSystem("../../examples/access")
	require.NoError(t, err)
	spec, err := resolve.Lens(s, "platform", lens)
	require.NoError(t, err)
	return spec
}

func TestLens_EveryExampleLensPassesTheLint(t *testing.T) {
	t.Parallel()
	for _, lens := range []string{"cmp-current", "cmp-target", "alice-current", "alice-target", "root"} {
		t.Run(lens, func(t *testing.T) {
			t.Parallel()
			spec := lensSpec(t, lens)

			findings := lint.Lint(spec)

			assert.Empty(t, findings)
		})
	}
}

func TestLens_KeepsTheTopologyGeometry(t *testing.T) {
	t.Parallel()
	s, err := model.LoadSystem("../../examples/access")
	require.NoError(t, err)
	plain, err := resolve.Topology(s, "platform")
	require.NoError(t, err)

	lit := lensSpec(t, "cmp-target")

	require.Len(t, lit.Nodes, len(plain.Nodes))
	for i := range plain.Nodes {
		assert.Equal(t, plain.Nodes[i].Rect(), lit.Nodes[i].Rect(), plain.Nodes[i].ID)
	}
}

func TestLens_MarksLevelsRemovalsAndTheRest(t *testing.T) {
	t.Parallel()
	spec := lensSpec(t, "cmp-target")
	got := map[string]string{}
	for _, n := range spec.Nodes {
		switch {
		case n.Focus:
			got[n.ID] = "focus"
		case n.Removed:
			got[n.ID] = "removed"
		case n.Level != "":
			got[n.ID] = n.Level
		}
	}

	assert.Equal(t, map[string]string{
		"cmp": "focus", "realm-platform": "write", "tfstate": "write",
		"vault-sys": "removed", "vault-a": "removed", "vault-b": "removed",
	}, got)
}

func TestLens_RejectsAnUnknownLens(t *testing.T) {
	t.Parallel()
	s, err := model.LoadSystem("../../examples/access")
	require.NoError(t, err)

	_, err = resolve.Lens(s, "platform", "nobody")

	assert.ErrorContains(t, err, `no lens "nobody"`)
}
