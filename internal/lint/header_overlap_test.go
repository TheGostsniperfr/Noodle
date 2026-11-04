package lint_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
)

func headed(meta ...string) diagram.Spec {
	return diagram.Spec{ID: "t", Width: 1600, Height: 900, Title: "Title", Subtitle: "Subtitle", Meta: meta}
}

func TestLint_HeaderOverlap(t *testing.T) {
	t.Parallel()
	threeMeta := []string{"one", "two", "three"}
	withZone := func(s diagram.Spec, x, y float64) diagram.Spec {
		s.Zones = []diagram.Zone{{ID: "z", Kind: "region", Label: "Z", Color: "slate", X: x, Y: y, W: 300, H: 200}}
		return s
	}
	finding := func(where, bottom string) []lint.Finding {
		return []lint.Finding{{Where: where, Msg: "overlaps the header, which ends at y " + bottom}}
	}
	tests := []struct {
		name     string
		spec     diagram.Spec
		expected []lint.Finding
	}{
		{"zone_at_40", withZone(headed(threeMeta...), 40, 40), finding("z", "131.5")},
		{"zone_at_bottom", withZone(headed(threeMeta...), 40, 131.5), nil},
		{"zone_1px_above", withZone(headed(threeMeta...), 40, 130.5), finding("z", "131.5")},
		{"meta_wraps", withZone(headed(strings.Repeat("m", 190)), 40, 104), finding("z", "117.75")},
		{"zone_right_of_header", withZone(headed(threeMeta...), 1265, 40), nil},
		{"node", func() diagram.Spec {
			s := headed(threeMeta...)
			s.Nodes = []diagram.Node{{ID: "n", Kind: "backend", Title: "N", X: 100, Y: 40, W: 100, H: 60}}
			return s
		}(), finding("n", "131.5")},
		{"card", func() diagram.Spec {
			s := headed(threeMeta...)
			s.Cards = []diagram.Card{{ID: "c", Title: "C", X: 100, Y: 40, W: 300, H: 200}}
			return s
		}(), finding("c", "131.5")},
		{"note", func() diagram.Spec {
			s := headed(threeMeta...)
			s.Notes = []diagram.Note{{ID: "n", Text: "x", X: 100, Y: 40, W: 100, H: 40}}
			return s
		}(), finding("n", "131.5")},
		{"no_header", diagram.Spec{ID: "t", Width: 1600, Height: 900,
			Nodes: []diagram.Node{{ID: "n", Kind: "backend", Title: "N", X: 100, Y: 0, W: 100, H: 60}}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings := lint.Lint(&tt.spec)

			assert.Equal(t, tt.expected, findings)
		})
	}
}
