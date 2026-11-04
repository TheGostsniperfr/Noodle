package lint_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
)

func TestLint_ZoneSub(t *testing.T) {
	t.Parallel()
	sub := strings.TrimSuffix(strings.Repeat("abcdefghi ", 9), " ")[:89] + "."
	// e9 leaves a node above the zone and enters one inside it, straight down at x.
	spec := func(x, nodeX, nodeW float64) diagram.Spec {
		return diagram.Spec{ID: "t", Width: 1000, Height: 600,
			Zones: []diagram.Zone{{ID: "platform", Kind: "region", Label: "Platform", Sub: sub, Color: "sky", X: 100, Y: 100, W: 600, H: 300}},
			Nodes: []diagram.Node{
				{ID: "s", Kind: "backend", Title: "S", X: nodeX, Y: 5, W: nodeW, H: 35},
				{ID: "t", Kind: "backend", Title: "T", X: nodeX, Y: 200, W: nodeW, H: 60},
			},
			Edges: []diagram.Edge{{ID: "e9", From: "s", To: "t", Kind: "flow", Path: []diagram.Point{{x, 40}, {x, 200}}}},
		}
	}
	tests := []struct {
		name     string
		spec     diagram.Spec
		expected []lint.Finding
	}{
		{"lane_free", spec(366, 306, 120), nil},
		{"through_title", spec(150, 100, 100),
			[]lint.Finding{{Where: "e9", Msg: "segment 0 crosses title of zone platform"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings := lint.Lint(&tt.spec)

			assert.Equal(t, tt.expected, findings)
		})
	}
}
