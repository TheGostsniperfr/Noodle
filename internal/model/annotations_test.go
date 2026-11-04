package model_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

const (
	annotationModel = modelHead + `elements:
  - {id: a, kind: backend, title: A}
  - {id: b, kind: backend, title: B}
  - {id: c, kind: backend, title: C}
  - {id: d, kind: backend, title: D}
  - {id: grp, kind: group, title: G}
  - {id: kid, kind: backend, parent: grp, title: K}
connections:
  - {id: c-ab, from: a, to: b, kind: flow}
  - {id: c9, from: c, to: d, kind: flow}
annotations:
`
	annotationLayout = layoutHead + `canvas: {width: 800, height: 400}
elements:
  a: {x: 20, y: 40, w: 200, h: 80}
  b: {x: 300, y: 40, w: 200, h: 80}
edges:
  c-ab: {from: a.right, to: b.left}
`
	globLayout = layoutHead + `canvas: {width: 800, height: 400}
elements:
  grp: {x: 20, y: 20, w: 400, h: 200}
  kid: {x: 20, y: 60, w: 200, h: 80}
`
	topologyView = viewHead + "type: topology\ninclude: [a, b]\n"
	sequenceView = viewHead + "type: sequence\nparticipants: [a, b]\nsteps:\n  - {id: m, from: a, to: b, over: c-ab, text: go}\n"
)

func TestCheck_UndrawnAnnotation(t *testing.T) {
	t.Parallel()
	viewFile := filepath.Join("views", "v.yaml")
	undrawn := model.Finding{File: viewFile, ID: "G1",
		Msg: "annotation targets nothing drawn in this view (c9): draw a target or list it in hidden_annotations with a reason"}
	tests := []struct {
		name         string
		annotations  string
		view, layout string
		expected     []model.Finding
	}{
		{"undrawn", "  - {id: G1, targets: [c9]}\n", topologyView, annotationLayout, []model.Finding{undrawn}},
		{"acknowledged", "  - {id: G1, targets: [c9]}\n", topologyView + "hidden_annotations: {G1: \"background job, drawn in jobs view\"}\n", annotationLayout, nil},
		{"one_drawn_one_not", "  - {id: G1, targets: [c9, a]}\n", topologyView, annotationLayout, nil},
		{"include_glob", "  - {id: G1, targets: [kid]}\n", viewHead + "type: topology\ninclude: [grp/**]\n", globLayout, nil},
		{"sequence_view", "  - {id: G1, targets: [c9]}\n", sequenceView, "", []model.Finding{undrawn}},
		{"sequence_view_drawn_message", "  - {id: G1, targets: [c-ab]}\n", sequenceView, "", nil},
		{"hide_annotations_still_checks", "  - {id: G1, targets: [c9]}\n", topologyView + "hide_annotations: true\n", annotationLayout, []model.Finding{undrawn}},
		{"hide_annotations_drawn", "  - {id: G1, targets: [c9, a]}\n", topologyView + "hide_annotations: true\n", annotationLayout, nil},
		{"empty_reason", "  - {id: G1, targets: [c9]}\n", topologyView + "hidden_annotations: {G1: \"\"}\n", annotationLayout,
			[]model.Finding{{File: viewFile, ID: "G1", Msg: "hidden_annotations needs a reason"}}},
		{"unknown_id", "  - {id: G1, targets: [c9]}\n", topologyView + "hidden_annotations: {G1: why, G7: x}\n", annotationLayout,
			[]model.Finding{{File: viewFile, ID: "G7", Msg: "hidden_annotations names G7, which is not an annotation"}}},
		{"drawn_id", "  - {id: G1, targets: [a]}\n", topologyView + "hidden_annotations: {G1: why}\n", annotationLayout,
			[]model.Finding{{File: viewFile, ID: "G1", Msg: "hidden_annotations hides G1, which this view draws"}}},
		{"annotation_without_targets", "  - {id: G1}\n", topologyView, annotationLayout, nil},
		{"fields_on_a_landscape_view", "  - {id: G1, targets: [c9]}\n", viewHead + "type: landscape\nhide_annotations: true\nbands: [{id: b1, title: B, sections: [{items: [a]}]}]\n", "",
			[]model.Finding{{File: viewFile, ID: "v", Msg: "hidden_annotations and hide_annotations are for topology and sequence views"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := system(t, annotationModel+tt.annotations, tt.view, tt.layout)

			findings := model.Check(s)

			assert.Equal(t, tt.expected, findings)
		})
	}
}
