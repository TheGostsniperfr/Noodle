package lint_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
)

func TestLint_AnnotationOnlyLabel(t *testing.T) {
	t.Parallel()
	badgeOnly := func() (*diagram.Spec, diagram.Rect) {
		s := clean()
		s.Edges[0].Label = "!!⚠ G1!!"
		s.Edges[0].LabelOffset = diagram.Point{0, -30}
		path := house.DrawnPath(s.Edges[0], house.PortBadges(s))
		box, ok := house.LabelBox(s.Edges[0], path)
		require.True(t, ok)
		return s, box
	}

	t.Run("the box is 32px wide, not the 50px of a leading separator", func(t *testing.T) {
		t.Parallel()
		s, box := badgeOnly()
		old := *s
		old.Edges = []diagram.Edge{s.Edges[0]}
		old.Edges[0].Label = " · !!⚠ G1!!"
		oldBox, _ := house.LabelBox(old.Edges[0], house.DrawnPath(old.Edges[0], house.PortBadges(&old)))

		assert.Equal(t, 32.0, box.W)
		assert.Equal(t, 50.0, oldBox.W)
		assert.InDelta(t, box.X+box.W/2, oldBox.X+oldBox.W/2, 1e-9)
	})
	tests := []struct {
		name     string
		nodeEdge func(box diagram.Rect) float64
		expected []lint.Finding
	}{
		{"one_px_in", func(box diagram.Rect) float64 { return box.X + 1 }, []lint.Finding{{Where: "e", Msg: "label overlaps node m"}}},
		{"one_px_outside", func(box diagram.Rect) float64 { return box.X - 1 }, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, box := badgeOnly()
			right := tt.nodeEdge(box)
			s.Nodes = append(s.Nodes, diagram.Node{ID: "m", Kind: "backend", Title: "M", X: right - 30, Y: 98, W: 30, H: 40})

			findings := lint.Lint(s)

			assert.Equal(t, tt.expected, findings)
		})
	}
}
