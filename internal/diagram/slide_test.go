package diagram_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
)

func TestForSlide_DropsTheFrameAndCropsToTheDrawing(t *testing.T) {
	t.Parallel()
	s := &diagram.Spec{
		Title: "T", Subtitle: "S", Meta: []string{"m"}, Width: 3000, Height: 2000,
		Zones: []diagram.Zone{{ID: "z", X: 200, Y: 300, W: 1000, H: 500}},
		Nodes: []diagram.Node{{ID: "n", X: 250, Y: 350, W: 200, H: 80}},
		Edges: []diagram.Edge{{ID: "e", Path: []diagram.Point{{450, 390}, {1400, 390}}}},
		Cards: []diagram.Card{{ID: "legend", X: 0, Y: 1800, W: 800, H: 150}},
	}

	s.ForSlide(20)

	assert.Equal(t, []any{"", 0, 1240.0, 540.0, 20.0, 20.0, 1220.0},
		[]any{s.Title, len(s.Cards), s.Width, s.Height, s.Zones[0].X, s.Zones[0].Y, s.Edges[0].Path[1].X()})
}
