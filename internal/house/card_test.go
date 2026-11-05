package house

import (
	"testing"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/stretchr/testify/assert"
)

func TestCardLayout_ListedCard_StacksTitleAndLines(t *testing.T) {
	t.Parallel()
	card := diagram.Card{ID: "gaps", Title: "Gaps", X: 100, Y: 200, W: 400, H: 200,
		Lines: []string{"!!⚠ G1!! first", "[1] a step", "plain"}}

	l := CardLayout(&diagram.Spec{}, card)

	kinds := []string{}
	bullets := []bool{}
	for _, r := range l.Rows {
		kinds = append(kinds, r.Kind)
		bullets = append(bullets, r.Bullet)
	}
	assert.Equal(t, []string{RowTitle, RowLine, RowLine, RowLine}, kinds)
	assert.Equal(t, []bool{false, false, false, true}, bullets)
	assert.Equal(t, diagram.Rect{X: 120, Y: 214, W: 364, H: 16.25}, l.Rows[0].Rect)
	assert.InDelta(t, 14+16.25+13.75+3*17.6+14, l.Needed, 1e-9)
}

func TestCardLayout_LegendCard_HoldsOnlyTheKindsTheSpecUses(t *testing.T) {
	t.Parallel()
	spec := &diagram.Spec{
		Nodes: []diagram.Node{{ID: "a", Kind: "backend"}, {ID: "b", Kind: "bus", Shape: "pipe", Multiplicity: "per tenant"}},
		Edges: []diagram.Edge{{ID: "e", Kind: "link"}},
	}
	card := diagram.Card{ID: "legend", Title: "Legend", Legend: true, X: 0, Y: 0, W: 600, H: 300}

	l := CardLayout(spec, card)

	ids := []string{}
	for _, r := range l.Rows {
		ids = append(ids, r.ID)
	}
	assert.Equal(t, []string{"", "legend-node-backend", "legend-node-bus", "legend-pipe", "legend-steps", "legend-port",
		"legend-edge-link", "legend-stack-1", "legend-notes"}, ids)
}

func TestCardRow_Widest_IsTheLineWhenTheCellDoesNotWrap_AndTheWordWhenItDoes(t *testing.T) {
	t.Parallel()
	text := "aaa bbbbbbb cc"

	fixed := CardRow{Plain: []string{text}, FontSize: 10}
	wrapping := CardRow{Plain: []string{text}, FontSize: 10, Wraps: true}

	line, lineW := fixed.Widest()
	word, wordW := wrapping.Widest()
	assert.Equal(t, text, line)
	assert.InDelta(t, 14*6.0, lineW, 1e-9)
	assert.Equal(t, "bbbbbbb", word)
	assert.InDelta(t, 7*6.0, wordW, 1e-9)
}
