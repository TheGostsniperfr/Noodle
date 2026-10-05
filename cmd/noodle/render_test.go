package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
)

func TestNoLigatures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, input, expected string
	}{
		{"empty value stays empty so blank cells keep no markup", "", ""},
		{"text is wrapped so pairs like >- stay two glyphs", "&lt;app&gt;-frontend", `<span style="font-variant-ligatures:none">&lt;app&gt;-frontend</span>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, noLigatures(tt.input))
		})
	}
}

func TestLegendNotes_ExplainOnlyWhatTheDiagramDraws(t *testing.T) {
	t.Parallel()
	zone := []diagram.Zone{{ID: "z"}}
	gapNode := []diagram.Node{{ID: "a", Badge: "G1"}}
	gapEdge := []diagram.Edge{{ID: "e", Label: "GET · HTTP · !!⚠ G2!!"}}

	tests := []struct {
		name               string
		spec               diagram.Spec
		wantFrame, wantGap bool
	}{
		{"no zone and no gap leaves the notes empty", diagram.Spec{}, false, false},
		{"zones without gaps explain the frames only", diagram.Spec{Zones: zone}, true, false},
		{"a gap badge on a node adds the gap line", diagram.Spec{Zones: zone, Nodes: gapNode}, true, true},
		{"a gap on an edge label adds the gap line", diagram.Spec{Edges: gapEdge}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &renderer{spec: &tt.spec, th: house.Themes["light"]}

			notes := r.legendNotes()

			assert.Equal(t, [2]bool{tt.wantFrame, tt.wantGap},
				[2]bool{strings.Contains(notes, "dashed frame"), strings.Contains(notes, "Gx")})
		})
	}
}
