package lint_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
)

// clean is a zone holding a and b, one numbered connection from a to b on port TCP 80,
// and a card below. It passes every rule, so each case below breaks one thing.
func clean() *diagram.Spec {
	return &diagram.Spec{
		ID: "t", Width: 800, Height: 480,
		Zones: []diagram.Zone{{ID: "z", Kind: "region", Label: "Z", Color: "slate", X: 40, Y: 40, W: 700, H: 260}},
		Nodes: []diagram.Node{
			{ID: "a", Kind: "backend", Title: "A", X: 80, Y: 120, W: 200, H: 80},
			{ID: "b", Kind: "backend", Title: "B", X: 460, Y: 120, W: 200, H: 80},
		},
		Edges: []diagram.Edge{
			{ID: "e", From: "a", To: "b", Kind: "flow", Label: "[1] GET · HTTP", Port: "TCP 80", Path: []diagram.Point{{280, 160}, {460, 160}}},
		},
		Cards: []diagram.Card{{ID: "c", Title: "Legend", Legend: true, Color: "slate", X: 40, Y: 340, W: 700, H: 100}},
	}
}

func TestLint_EveryRuleFlagsItsCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rule, name string
		breakIt    func(s *diagram.Spec)
		where, msg string
	}{
		{"ids", "unknown zone colour", func(s *diagram.Spec) { s.Zones[0].Color = "purple" },
			"z", `unknown zone color "purple"`},
		{"ids", "duplicate id", func(s *diagram.Spec) { s.Nodes[1].ID = "a"; s.Edges[0].To = "a" },
			"a", "duplicate id"},
		{"node text", "title wider than the box", func(s *diagram.Spec) { s.Nodes[0].Title = "A title far too long for this box" },
			"a", "title overflows"},
		{"node text", "box too short for its lines", func(s *diagram.Spec) { s.Nodes[0].Tech, s.Nodes[0].Desc, s.Nodes[0].H = "[t]", "d", 40 },
			"a", "text needs 60px height, box is 40px"},
		{"node spacing", "boxes closer than 24 px", func(s *diagram.Spec) { s.Nodes[1].X = 290 },
			"a", "closer than 24px to b"},
		{"node in zone", "box on a zone title", func(s *diagram.Spec) { s.Nodes[0].X, s.Nodes[0].Y = 50, 50 },
			"a", "covers title of zone z"},
		{"node in zone", "box across a zone border", func(s *diagram.Spec) { s.Nodes[0].X = 20 },
			"a", "straddles border of zone z"},
		{"zone spacing", "sibling zones closer than 60 px", func(s *diagram.Spec) {
			s.Zones = append(s.Zones, diagram.Zone{ID: "z2", Kind: "region", Label: "Z2", Color: "cyan", X: 770, Y: 40, W: 20, H: 100})
		}, "z", "closer than 60px to zone z2"},
		{"ports", "badge over the server's icon", func(s *diagram.Spec) {
			s.Nodes[1].Icon = "k8s-svc"
			s.Edges[0].Path = []diagram.Point{{280, 145}, {460, 145}}
		}, "e", `port badge "TCP 80" covers the icon of b`},
		{"edge sides", "ingress through the bottom", func(s *diagram.Spec) {
			s.Edges[0].Port = ""
			s.Edges[0].Path = []diagram.Point{{180, 200}, {180, 240}, {560, 240}, {560, 200}}
		}, "e", "ingress enters b from the bottom side, want left or top"},
		{"label_at", "pinned label off the path", func(s *diagram.Spec) { s.Edges[0].LabelAt = &diagram.Point{300, 100} },
			"e", "label_at [300 100] is not on the drawn path"},
		{"edge crossings", "diagonal segment", func(s *diagram.Spec) { s.Edges[0].Path = []diagram.Point{{280, 160}, {460, 170}} },
			"e", "segment 0 is diagonal"},
		{"edge crossings", "segment through a box", func(s *diagram.Spec) {
			s.Nodes = append(s.Nodes, diagram.Node{ID: "m", Kind: "external", Title: "M", X: 330, Y: 210, W: 60, H: 60})
			s.Edges[0].Path = []diagram.Point{{180, 200}, {180, 240}, {560, 240}, {560, 200}}
			s.Edges[0].Port, s.Edges[0].Kind = "", "link"
		}, "e", "segment 1 crosses node m"},
		{"label overlaps", "label over a note", func(s *diagram.Spec) {
			s.Notes = append(s.Notes, diagram.Note{ID: "n", Text: "x", X: 340, Y: 100, W: 40, H: 56})
		}, "e", "label overlaps note n"},
		{"collinear", "two edges on the same line", func(s *diagram.Spec) {
			s.Edges = append(s.Edges, diagram.Edge{ID: "e2", From: "a", To: "b", Kind: "link", Path: []diagram.Point{{280, 162}, {460, 162}}})
		}, "e", "runs on top of e2 (horizontal, y≈160)"},
		{"steps", "a step used twice", func(s *diagram.Spec) {
			s.Edges = append(s.Edges, diagram.Edge{ID: "e2", From: "a", To: "b", Kind: "link", Label: "[1] ref", Path: []diagram.Point{{180, 200}, {180, 240}, {560, 240}, {560, 200}}})
		}, "e2", "step [1] already used on e"},
		{"steps", "two steps on one edge", func(s *diagram.Spec) { s.Edges[0].Label = "[1] [2] GET" },
			"e", "carries 2 steps, want at most one"},
	}
	for _, tt := range tests {
		t.Run(tt.rule+": "+tt.name, func(t *testing.T) {
			t.Parallel()
			s := clean()
			require.Empty(t, lint.Lint(s), "the clean spec must pass every rule")
			tt.breakIt(s)

			findings := lint.Lint(s)

			var hit bool
			for _, f := range findings {
				hit = hit || (f.Where == tt.where && strings.Contains(f.Msg, tt.msg))
			}
			assert.True(t, hit, "want %s: %s…, got %v", tt.where, tt.msg, findings)
		})
	}
}
