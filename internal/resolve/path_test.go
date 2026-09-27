package resolve_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/model"
	"github.com/TheGostsniperfr/Noodle/internal/resolve"
)

// Zone z at (100, 100); a at (20, 20) and b at (320, 40) inside it, both 200 x 80.
// So a is [120, 120, 200, 80] and b is [420, 140, 200, 80] on the canvas.
const twoBoxes = `apiVersion: noodle/v1alpha1
kind: Model
elements:
  - {id: z, kind: region, title: Z}
  - {id: a, kind: backend, parent: z, title: A}
  - {id: b, kind: backend, parent: z, title: B, ports: [{name: http, protocol: TCP, port: 80}]}
connections:
  - {id: c, from: a, to: b, port: http, protocol: HTTP, verb: GET, kind: flow}
`

func resolveOne(t *testing.T, route string, lanes string) (*diagram.Spec, error) {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(dir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	}
	write("model.yaml", twoBoxes)
	write("views/v.yaml", "apiVersion: noodle/v1alpha1\nkind: View\ntype: topology\nsteps: [c]\n")
	write("layouts/v.yaml", `apiVersion: noodle/v1alpha1
kind: Layout
canvas: {width: 800, height: 500}
`+lanes+`elements:
  z: {x: 100, y: 100, w: 600, h: 300}
  a: {x: 20, y: 20, w: 200, h: 80}
  b: {x: 320, y: 40, w: 200, h: 80}
edges:
  c: `+route+"\n")
	s, err := model.LoadSystem(dir)
	require.NoError(t, err)
	require.Empty(t, model.Check(s))
	return resolve.Topology(s, "v")
}

func TestTopology_PlacesEndpointsAndWaypoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, route, lanes string
		want               []diagram.Point
	}{
		{"without waypoints, from sits at 50 % and to lines up with it",
			"{from: a.right, to: b.left}", "",
			[]diagram.Point{{320, 160}, {420, 160}}},
		{"endpoints line up with the neighbouring waypoints",
			"{from: a.right, to: b.top, waypoints: [[370, 130], [470, 130]]}", "",
			[]diagram.Point{{320, 130}, {370, 130}, {470, 130}, {470, 140}}},
		{"a percentage is measured from the top of a vertical side",
			"{from: a.right@25%, to: b.left, waypoints: [[370, 140], [370, 200]]}", "",
			[]diagram.Point{{320, 140}, {370, 140}, {370, 200}, {420, 200}}},
		{"pixel offsets are exact on both ends",
			"{from: a.right@55px, to: b.left@5px}", "",
			[]diagram.Point{{320, 175}, {420, 145}}},
		{"an x lane adds the bends in and out of it",
			"{from: a.right, to: b.left, waypoints: [lane:spine]}", "lanes: {spine: {x: 370}}\n",
			[]diagram.Point{{320, 160}, {370, 160}, {370, 180}, {420, 180}}},
		{"a y lane followed by a point turns at the lane",
			"{from: a.bottom, to: b.left, waypoints: [lane:bus, [380, 200]]}", "lanes: {bus: {y: 230}}\n",
			[]diagram.Point{{220, 200}, {220, 230}, {380, 230}, {380, 200}, {420, 200}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveOne(t, tt.route, tt.lanes)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Edges[0].Path)
		})
	}
}

func TestTopology_RejectsEndpointsOffTheBox(t *testing.T) {
	t.Parallel()

	tests := []struct{ name, route, wantErr string }{
		{"offset past the side", "{from: a.right@90px, to: b.left}", "offset 90px is past the right side (80px long)"},
		{"offset without unit", "{from: a.right@40, to: b.left}", `want @0-100% or @NNpx`},
		{"percentage above 100", "{from: a.right@120%, to: b.left}", `want @0-100% or @NNpx`},
		{"an alignment that lands beside the box", "{from: a.right@5px, to: b.left}", `endpoint "b.left" lines up at y=125, outside its side (140 to 220)`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := resolveOne(t, tt.route, "")

			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestTopology_KeepsEdgesOnALaneParallelWhenTheLaneMoves(t *testing.T) {
	t.Parallel()
	const route = "{from: a.right, to: b.left, waypoints: [lane:spine]}"

	before, err := resolveOne(t, route, "lanes: {spine: {x: 360}}\n")
	require.NoError(t, err)
	after, err := resolveOne(t, route, "lanes: {spine: {x: 390}}\n")
	require.NoError(t, err)

	vertical := func(p []diagram.Point) float64 { require.Equal(t, p[1].X(), p[2].X()); return p[1].X() }
	assert.Equal(t, 30.0, vertical(after.Edges[0].Path)-vertical(before.Edges[0].Path))
}

func TestTopology_DerivesLabelsBadgesAndBlockedKind(t *testing.T) {
	t.Parallel()
	s, err := model.LoadSystem("../../examples/cnp-runtime")
	require.NoError(t, err)

	got, err := resolve.Topology(s, "runtime")

	require.NoError(t, err)
	edges := map[string]diagram.Edge{}
	for _, e := range got.Edges {
		edges[e.ID] = e
	}
	assert.Equal(t, "[3] forward · HTTP", edges["e-3-connector-envoy"].Label, "step number and verb · protocol")
	assert.Equal(t, "TCP 80", edges["e-3-connector-envoy"].Port, "badge from the target's named port")
	assert.Equal(t, "[B] write Secret · K8s API", edges["e-b-vso-secret"].Label, "background steps are lettered")
	assert.Equal(t, "OIDC token call · HTTPS · !!⚠ G6!!", edges["e-oidc-token-hairpin"].Label, "annotation on an edge")
	assert.Equal(t, "targetRef", edges["r-ctp-gateway"].Label, "a reference is labelled with its kind")
	assert.Equal(t, "link", edges["r-ctp-gateway"].Kind)
	assert.Equal(t, "blocked", edges["x-cross-tenant"].Kind, "denied connections draw as blocked")
}
