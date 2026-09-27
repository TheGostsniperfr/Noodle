package resolve_test

import (
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/model"
	"github.com/TheGostsniperfr/Noodle/internal/resolve"
)

func loadV0(t *testing.T, path string) *diagram.Spec {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var s diagram.Spec
	require.NoError(t, yaml.Unmarshal(raw, &s))
	return &s
}

func byID(edges []diagram.Edge) []diagram.Edge {
	out := append([]diagram.Edge(nil), edges...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// The split CNP example must describe exactly the diagram the v0 file draws: same
// boxes, same paths, same labels and badges. Only the id (the view's file stem) and
// the edge order (connections, then references) differ.
func TestTopology_ReproducesTheV0CNPDiagram(t *testing.T) {
	t.Parallel()
	want := loadV0(t, "../../examples/cnp-runtime/runtime-request-path.yaml")
	s, err := model.LoadSystem("../../examples/cnp-runtime")
	require.NoError(t, err)
	require.Empty(t, model.Check(s))

	got, err := resolve.Topology(s, "runtime")

	require.NoError(t, err)
	assert.Equal(t, "runtime", got.ID)
	got.ID = want.ID
	assert.Equal(t, byID(want.Edges), byID(got.Edges))
	got.Edges, want.Edges = nil, nil
	assert.Equal(t, want, got)
}
