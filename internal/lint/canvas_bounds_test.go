package lint_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
)

func TestLint_CanvasBounds(t *testing.T) {
	t.Parallel()
	canvas := func(mutate func(*diagram.Spec)) diagram.Spec {
		s := diagram.Spec{ID: "t", Width: 1600, Height: 900}
		mutate(&s)
		return s
	}
	node := func(x, y float64) diagram.Node {
		return diagram.Node{ID: "n", Kind: "backend", Icon: "k8s-svc", Title: "N", X: x, Y: y, W: 100, H: 60}
	}
	tests := []struct {
		name     string
		spec     diagram.Spec
		expected []lint.Finding
	}{
		{"zone", canvas(func(s *diagram.Spec) {
			s.Zones = []diagram.Zone{{ID: "z", Kind: "region", Label: "Z", Color: "slate", X: 1500, Y: 200, W: 200, H: 100}}
		}), []lint.Finding{{Where: "z", Msg: "extends past the canvas: right edge at 1700, canvas is 1600 wide"}}},
		{"node", canvas(func(s *diagram.Spec) { s.Nodes = []diagram.Node{node(1550, 200)} }),
			[]lint.Finding{{Where: "n", Msg: "extends past the canvas: right edge at 1650, canvas is 1600 wide"}}},
		{"card", canvas(func(s *diagram.Spec) {
			s.Cards = []diagram.Card{{ID: "c", Title: "C", X: 1500, Y: 200, W: 200, H: 100}}
		}), []lint.Finding{{Where: "c", Msg: "extends past the canvas: right edge at 1700, canvas is 1600 wide"}}},
		{"note", canvas(func(s *diagram.Spec) { s.Notes = []diagram.Note{{ID: "n", Text: "x", X: 1550, Y: 200, W: 100, H: 40}} }),
			[]lint.Finding{{Where: "n", Msg: "extends past the canvas: right edge at 1650, canvas is 1600 wide"}}},
		{"tile", canvas(func(s *diagram.Spec) { s.Type = "landscape"; s.Nodes = []diagram.Node{node(1550, 200)} }),
			[]lint.Finding{{Where: "n", Msg: "extends past the canvas: right edge at 1650, canvas is 1600 wide"}}},
		{"side_column", canvas(func(s *diagram.Spec) {
			s.Type = "landscape"
			s.Zones = []diagram.Zone{{ID: "side-0", Kind: "region", Label: "Upkeep", Color: "amber", X: 1450, Y: 170, W: 262, H: 124}}
		}), []lint.Finding{{Where: "side-0", Msg: "extends past the canvas: right edge at 1712, canvas is 1600 wide"}}},
		{"past_height", canvas(func(s *diagram.Spec) { s.Nodes = []diagram.Node{node(200, 850)} }),
			[]lint.Finding{{Where: "n", Msg: "extends past the canvas: bottom edge at 910, canvas is 900 high"}}},
		{"past_the_left_edge", canvas(func(s *diagram.Spec) { s.Nodes = []diagram.Node{node(-10, 200)} }),
			[]lint.Finding{{Where: "n", Msg: "extends past the canvas: left edge at -10"}}},
		{"touching", canvas(func(s *diagram.Spec) { s.Nodes = []diagram.Node{node(1500, 840)} }), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings := lint.Lint(&tt.spec)

			assert.Equal(t, tt.expected, findings)
		})
	}
}
