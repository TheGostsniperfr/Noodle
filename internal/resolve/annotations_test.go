package resolve_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/TheGostsniperfr/Noodle/internal/model"
	"github.com/TheGostsniperfr/Noodle/internal/resolve"
)

const (
	annotationHead = "apiVersion: noodle/v1alpha1\nkind: "
	annotatedModel = annotationHead + `Model
elements:
  - {id: a, kind: frontend, title: A}
  - {id: b, kind: backend, title: B}
connections:
  - {id: c1, from: a, to: b, kind: flow, protocol: HTTPS}
  - {id: c2, from: a, to: b, kind: flow}
annotations:
  - {id: G1, targets: [c1]}
  - {id: G2, targets: [a]}
  - {id: G3, targets: [c2]}
`
	annotatedLayout = annotationHead + `Layout
canvas: {width: 800, height: 400}
elements:
  a: {x: 40, y: 160, w: 200, h: 80}
  b: {x: 500, y: 160, w: 200, h: 80}
edges:
  c1: {from: a.right, to: b.left}
  c2: {from: a.bottom, to: b.bottom, waypoints: [[140, 320], [600, 320]]}
`
)

// bareModel has a connection with neither verb nor protocol, annotated G1.
const bareModel = annotationHead + `Model
elements:
  - {id: a, kind: frontend, title: A}
  - {id: b, kind: backend, title: B}
connections:
  - {id: c1, from: a, to: b, kind: flow}
annotations:
  - {id: G1, targets: [c1]}
`

func annotatedSystem(t *testing.T, view string, sequence bool) *model.System {
	t.Helper()
	return systemOf(t, annotatedModel, view, sequence)
}

func systemOf(t *testing.T, modelYAML, view string, sequence bool) *model.System {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{"model.yaml": modelYAML, "views/v.yaml": annotationHead + view}
	if !sequence {
		files["layouts/v.yaml"] = annotatedLayout
	}
	for rel, content := range files {
		p := filepath.Join(dir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	}
	s, err := model.LoadSystem(dir)
	require.NoError(t, err)
	require.Empty(t, model.Check(s))
	return s
}

func edgeLabel(t *testing.T, view string) (diagram.Edge, *diagram.Spec) {
	t.Helper()
	spec, err := resolve.Topology(annotatedSystem(t, view, false), "v")
	require.NoError(t, err)
	for _, e := range spec.Edges {
		if e.ID == "c1" {
			return e, spec
		}
	}
	require.Fail(t, "c1 is not drawn")
	return diagram.Edge{}, nil
}

// messageLabel is a sequence message's label without its step number: a sequence
// numbers every message (ADR-0011), so the label is "[1] " and then what is asserted.
func messageLabel(t *testing.T, view string) string {
	t.Helper()
	return unnumbered(t, annotatedSystem(t, view, true))
}

func unnumbered(t *testing.T, s *model.System) string {
	t.Helper()
	seq, err := resolve.Sequence(s, "v")
	require.NoError(t, err)
	require.Len(t, seq.Messages, 1)
	rest, found := strings.CutPrefix(seq.Messages[0].Label, "[1] ")
	require.True(t, found, "label %q starts with its step number", seq.Messages[0].Label)
	return rest
}

func TestResolve_AnnotationLabel(t *testing.T) {
	t.Parallel()
	topology := "View\ntype: topology\n"
	tests := []struct {
		name, view string
		isSequence bool
		expected   string
	}{
		{"empty_label", topology + "labels: {c1: \"\"}\n", false, "!!⚠ G1!!"},
		{"label", topology, false, "HTTPS · !!⚠ G1!!"},
		{"sequence_label", "View\ntype: sequence\nparticipants: [a, b]\nsteps:\n  - {id: m, from: a, to: b, over: c1}\nhidden_annotations: {G3: other message}\n", true, "HTTPS · !!⚠ G1!!"},
		{"sequence_text", "View\ntype: sequence\nparticipants: [a, b]\nsteps:\n  - {id: m, from: a, to: b, over: c1, text: GET}\nhidden_annotations: {G3: other message}\n", true, "GET · !!⚠ G1!!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var label string
			if tt.isSequence {
				label = messageLabel(t, tt.view)
			} else {
				e, _ := edgeLabel(t, tt.view)
				label = e.Label
			}

			assert.Equal(t, tt.expected, label)
		})
	}
}

func TestResolve_AnnotationLabel_SequenceMessageWithAnEmptyText(t *testing.T) {
	t.Parallel()
	view := "View\ntype: sequence\nparticipants: [a, b]\nsteps:\n  - {id: m, from: a, to: b, over: c1, text: \"\"}\n"

	label := unnumbered(t, systemOf(t, bareModel, view, true))

	assert.Equal(t, "!!⚠ G1!!", label)
}

func TestResolve_HideAnnotations(t *testing.T) {
	t.Parallel()

	t.Run("edge_empty", func(t *testing.T) {
		t.Parallel()

		e, _ := edgeLabel(t, "View\ntype: topology\nhide_annotations: true\nlabels: {c1: \"\"}\n")

		_, hasBox := house.LabelBox(e, e.Path)
		assert.Equal(t, "", e.Label)
		assert.False(t, hasBox)
	})
	t.Run("edge_label", func(t *testing.T) {
		t.Parallel()

		e, _ := edgeLabel(t, "View\ntype: topology\nhide_annotations: true\n")

		assert.Equal(t, "HTTPS", e.Label)
	})
	t.Run("node_badge", func(t *testing.T) {
		t.Parallel()

		_, hidden := edgeLabel(t, "View\ntype: topology\nhide_annotations: true\n")
		_, shown := edgeLabel(t, "View\ntype: topology\n")

		assert.Equal(t, "", hidden.Nodes[0].Badge)
		assert.Equal(t, "G2", shown.Nodes[0].Badge)
	})
	t.Run("sequence", func(t *testing.T) {
		t.Parallel()

		label := messageLabel(t, "View\ntype: sequence\nhide_annotations: true\nparticipants: [a, b]\nsteps:\n  - {id: m, from: a, to: b, over: c1, text: HTTPS}\nhidden_annotations: {G3: other message}\n")

		assert.Equal(t, "HTTPS", label)
	})
	t.Run("sequence_participant_badge", func(t *testing.T) {
		t.Parallel()
		seq, err := resolve.Sequence(annotatedSystem(t, "View\ntype: sequence\nhide_annotations: true\nparticipants: [a, b]\nsteps:\n  - {id: m, from: a, to: b, over: c1, text: HTTPS}\nhidden_annotations: {G3: other message}\n", true), "v")
		require.NoError(t, err)

		assert.Equal(t, "", seq.Frame.Nodes[0].Badge)
	})
}
