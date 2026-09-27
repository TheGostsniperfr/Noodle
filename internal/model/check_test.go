package model_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

func TestCheck_ReportsNothing_OnEveryExample(t *testing.T) {
	t.Parallel()
	models, err := filepath.Glob("../../examples/*/model.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, models)

	for _, m := range models {
		t.Run(filepath.Dir(m), func(t *testing.T) {
			s, err := model.LoadSystem(filepath.Dir(m))
			require.NoError(t, err)

			findings := model.Check(s)

			assert.Empty(t, findings)
		})
	}
}

func TestCheck_ReportsNothing_OnAValidSystem(t *testing.T) {
	t.Parallel()
	s, err := model.LoadSystem("testdata/minimal")
	require.NoError(t, err)

	findings := model.Check(s)

	assert.Empty(t, findings)
}

const (
	modelHead  = "apiVersion: noodle/v1alpha1\nkind: Model\n"
	viewHead   = "apiVersion: noodle/v1alpha1\nkind: View\n"
	layoutHead = "apiVersion: noodle/v1alpha1\nkind: Layout\n"
	baseModel  = modelHead + `elements:
  - {id: z, kind: region, title: Z}
  - {id: a, kind: backend, parent: z, title: A}
  - {id: b, kind: database, parent: z, title: B, ports: [{name: pg, protocol: TCP, port: 5432}]}
connections:
  - {id: c-ab, from: a, to: b, port: pg, protocol: SQL, verb: query, kind: flow}
`
	baseView   = viewHead + "type: topology\nsteps: [c-ab]\ncards: [{id: card, title: C}]\n"
	baseLayout = layoutHead + `canvas: {width: 800, height: 400}
elements:
  z: {x: 0, y: 0, w: 600, h: 300}
  a: {x: 20, y: 40, w: 200, h: 80}
  b: {x: 300, y: 40, w: 200, h: 80}
cards:
  card: {x: 0, y: 320, w: 600, h: 60}
edges:
  c-ab: {from: a.right, to: b.left}
`
)

// system writes a model, one view "v" and, when layout is non-empty, its layout.
func system(t *testing.T, modelYAML, viewYAML, layoutYAML string) *model.System {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(dir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	}
	write("model.yaml", modelYAML)
	write("views/v.yaml", viewYAML)
	if layoutYAML != "" {
		write("layouts/v.yaml", layoutYAML)
	}
	s, err := model.LoadSystem(dir)
	require.NoError(t, err)
	return s
}

func TestCheck_ReportsFileIDAndReason(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		model, view, layout string
		wantFile, wantID    string
		wantMsg             string
	}{
		{"the base system is clean, so each case below fails for one reason only",
			baseModel, baseView, baseLayout, "", "", ""},
		{"duplicate id", baseModel + "  - {id: a, from: a, to: b, kind: flow}\n", baseView, baseLayout,
			"model.yaml", "a", "duplicate id"},
		{"unknown element kind", modelHead + "elements: [{id: x, kind: server}]\n", viewHead + "type: sequence\n", "",
			"model.yaml", "x", `unknown element kind "server"`},
		{"unknown status", modelHead + "elements: [{id: x, kind: backend, status: gone}]\n", viewHead + "type: sequence\n", "",
			"model.yaml", "x", `unknown status "gone"`},
		{"parent that is not a zone", modelHead + "elements: [{id: x, kind: backend}, {id: y, kind: backend, parent: x}]\n", viewHead + "type: sequence\n", "",
			"model.yaml", "y", `parent "x" is not a zone`},
		{"connection to a missing element", baseModel + "  - {id: c-ax, from: a, to: x, kind: flow}\n", baseView, baseLayout,
			"model.yaml", "c-ax", `to "x" does not exist in the model`},
		{"port the target does not listen on", baseModel + "  - {id: c-ba, from: b, to: a, port: pg, kind: flow}\n", baseView, baseLayout,
			"model.yaml", "c-ba", `port "pg" is not a port of a`},
		{"blocked is no longer a connection kind", baseModel + "  - {id: c-ba, from: b, to: a, kind: blocked}\n", baseView, baseLayout,
			"model.yaml", "c-ba", `unknown connection kind "blocked"`},
		{"enforcement by a missing element", baseModel + "  - {id: c-ba, from: b, to: a, kind: flow, denied: true, enforced_by: [np]}\n", baseView, baseLayout,
			"model.yaml", "c-ba", `enforced_by "np" does not exist in the model`},
		{"enforcement on an allowed connection", baseModel + "  - {id: c-ba, from: b, to: a, kind: flow, enforced_by: [z]}\n", baseView, baseLayout,
			"model.yaml", "c-ba", "enforced_by is for denied connections only"},
		{"annotation target missing", baseModel + "annotations: [{id: G1, targets: [x]}]\n", baseView, baseLayout,
			"model.yaml", "G1", `target "x" does not exist in the model`},
		{"unknown view type", baseModel, viewHead + "type: context\n", baseLayout,
			filepath.Join("views", "v.yaml"), "v", `type "context", want topology or sequence`},
		{"step that is not a connection", baseModel, baseView + "background: [a]\n", baseLayout,
			filepath.Join("views", "v.yaml"), "a", "step is not a connection of the model"},
		{"include of a missing element", baseModel, baseView + "include: [x/**]\n", baseLayout,
			filepath.Join("views", "v.yaml"), "v", `include "x" does not exist in the model`},
		{"participants on a topology view", baseModel, baseView + "participants: [a]\n", baseLayout,
			filepath.Join("views", "v.yaml"), "v", "participants are for sequence views only"},
		{"topology view without layout", baseModel, baseView, "",
			filepath.Join("views", "v.yaml"), "v", "topology view has no layouts/v.yaml"},
		{"sequence view with a layout", baseModel, viewHead + "type: sequence\n", baseLayout,
			filepath.Join("views", "v.yaml"), "v", "sequence views are computed and take no layout"},
		{"included element without position", baseModel, baseView, layoutHead + "elements:\n  z: {x: 0, y: 0, w: 600, h: 300}\n  a: {x: 20, y: 40, w: 200, h: 80}\n",
			filepath.Join("layouts", "v.yaml"), "b", "included element has no position"},
		{"position relative to an unplaced zone", baseModel, baseView + "include: [a, b]\n", layoutHead + "elements:\n  a: {x: 20, y: 40, w: 200, h: 80}\n  b: {x: 300, y: 40, w: 200, h: 80}\n",
			filepath.Join("layouts", "v.yaml"), "a", `position is relative to zone "z", which has no position`},
		{"edge that is not in the model", baseModel, baseView, baseLayout + "  c-zz: {from: a.right, to: b.left}\n",
			filepath.Join("layouts", "v.yaml"), "c-zz", "edge is not a connection or reference of the model"},
		{"endpoint with an unknown side", baseModel, baseView, baseLayout + "  c-ab2: {from: a.east, to: b.left}\n",
			filepath.Join("layouts", "v.yaml"), "c-ab2", `endpoint "a.east": side "east"`},
		{"undefined lane", baseModel, baseView, baseLayout + "  c-ab3: {from: a.right, to: b.left, waypoints: [lane:bus]}\n",
			filepath.Join("layouts", "v.yaml"), "c-ab3", `lane "bus" is not defined`},
		{"lane with both x and y", baseModel, baseView, baseLayout + "lanes: {bus: {x: 1, y: 2}}\n",
			filepath.Join("layouts", "v.yaml"), "bus", "lane needs exactly one of x or y"},
		{"route drawn backwards", baseModel, baseView, layoutHead + "canvas: {width: 800, height: 400}\nelements:\n  z: {x: 0, y: 0, w: 600, h: 300}\n  a: {x: 20, y: 40, w: 200, h: 80}\n  b: {x: 300, y: 40, w: 200, h: 80}\nedges:\n  c-ab: {from: b.left, to: a.right}\n",
			filepath.Join("layouts", "v.yaml"), "c-ab", `endpoint "b.left" must be on a, the from of the edge`},
		{"edge between shown elements without route", baseModel, baseView, layoutHead + "elements:\n  z: {x: 0, y: 0, w: 600, h: 300}\n  a: {x: 20, y: 40, w: 200, h: 80}\n  b: {x: 300, y: 40, w: 200, h: 80}\n",
			filepath.Join("layouts", "v.yaml"), "c-ab", "edge between shown elements has no route"},
		{"card without position", baseModel, strings.Replace(baseView, "{id: card, title: C}", "{id: card, title: C}, {id: card2, title: D}", 1), baseLayout,
			filepath.Join("layouts", "v.yaml"), "card2", "card has no position"},
		{"card the view does not define", baseModel, baseView, strings.Replace(baseLayout, "cards:\n", "cards:\n  other: {x: 0, y: 0, w: 1, h: 1}\n", 1),
			filepath.Join("layouts", "v.yaml"), "other", "card is not defined in the view"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := system(t, tt.model, tt.view, tt.layout)

			findings := model.Check(s)

			if tt.wantMsg == "" {
				assert.Empty(t, findings)
				return
			}
			var match []model.Finding
			for _, f := range findings {
				if f.File == tt.wantFile && f.ID == tt.wantID && strings.Contains(f.Msg, tt.wantMsg) {
					match = append(match, f)
				}
			}
			assert.NotEmpty(t, match, "want %s: %s: %s…, got %v", tt.wantFile, tt.wantID, tt.wantMsg, findings)
		})
	}
}
