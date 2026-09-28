package resolve_test

import (
	"os"
	"path/filepath"
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

func TestTopology_CarriesInheritedStatusAndTarget(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	files := map[string]string{
		"model.yaml": `apiVersion: noodle/v1alpha1
kind: Model
elements:
  - {id: obs, kind: region, title: Observability, status: planned, target: SP2}
  - {id: loki, kind: backend, parent: obs, title: Loki}
`,
		"views/v.yaml": "apiVersion: noodle/v1alpha1\nkind: View\ntype: topology\n",
		"layouts/v.yaml": `apiVersion: noodle/v1alpha1
kind: Layout
canvas: {width: 400, height: 300}
elements:
  obs: {x: 0, y: 0, w: 300, h: 200}
  loki: {x: 40, y: 60, w: 200, h: 80}
`,
	}
	for rel, content := range files {
		p := filepath.Join(dir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	}
	s, err := model.LoadSystem(dir)
	require.NoError(t, err)
	require.Empty(t, model.Check(s))

	got, err := resolve.Topology(s, "v")

	require.NoError(t, err)
	assert.Equal(t, [4]string{"planned", "SP2", "planned", "SP2"},
		[4]string{got.Zones[0].Status, got.Zones[0].Target, got.Nodes[0].Status, got.Nodes[0].Target})
}

func TestLandscape_PlacesItemsInsideTheirSectionAndBand(t *testing.T) {
	t.Parallel()
	s, err := model.LoadSystem("../../examples/landscape")
	require.NoError(t, err)
	require.Empty(t, model.Check(s))

	got, err := resolve.View(s, "stack")

	require.NoError(t, err)
	zones := map[string]diagram.Rect{}
	for _, z := range got.Zones {
		zones[z.ID] = z.Rect()
	}
	nodes := map[string]diagram.Rect{}
	for _, n := range got.Nodes {
		nodes[n.ID] = n.Rect()
	}
	contains := func(zone, node string) bool { return zones[zone].Contains(nodes[node]) }
	assert.Equal(t, []bool{true, true, true, true, true}, []bool{
		contains("delivery", "argocd"), contains("delivery-1", "argocd"),
		contains("runtime-0", "cilium"), contains("side-0", "keycloak"),
		zones["delivery"].Y+zones["delivery"].H < zones["services"].Y,
	})
}

func TestLandscape_DrawsFlowArrowsBetweenSectionsOnly(t *testing.T) {
	t.Parallel()
	s, err := model.LoadSystem("../../examples/landscape")
	require.NoError(t, err)

	got, err := resolve.View(s, "stack")

	require.NoError(t, err)
	assert.Len(t, got.Arrows, 1, "delivery has flow: true and two sections; other bands none")
}
