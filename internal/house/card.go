package house

import (
	"math"
	"strconv"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
)

// Kinds of row a card draws. A listed card has a title and lines; a legend card has a
// title, then a column of node rows with the step and port rows, a column of edge and
// status rows, and a notes block.
const (
	RowTitle            = "title"
	RowLine             = "line"
	RowLegendNode       = "legend-node"
	RowLegendPipe       = "legend-pipe"
	RowLegendSteps      = "legend-steps"
	RowLegendPort       = "legend-port"
	RowLegendEdge       = "legend-edge"
	RowLegendPlanned    = "legend-planned"
	RowLegendDeprecated = "legend-deprecated"
	RowLegendStack      = "legend-stack"
	RowLegendAdded      = "legend-added"
	RowLegendRemoved    = "legend-removed"
	RowLegendUnchanged  = "legend-unchanged"
	RowLegendNotes      = "legend-notes"
	RowNote             = "note"
)

const (
	CardPadLeft      = 20.0
	CardPadRight     = 16.0
	CardPadTop       = 14.0
	CardPadBottom    = 14.0
	CardTitleSize    = 13.0
	CardTextSize     = 11.0
	CardLineHeightEm = 1.6
	LegendTextSize   = 10.5
	LegendRowH       = 22.0
	LegendRowPitch   = 26.0
	LegendTop        = 52.0
	LegendNotesH     = 40.0
	LegendNotesBelow = 50.0
	LegendSample     = 40.0
	LegendTextGap    = 52.0
)

// CardRow is one text the renderer draws in a card, the box it gives it and, for a
// legend row, the sample drawn beside it. Plain is the text as drawn, one entry per
// source line, which is what the lint measures.
type CardRow struct {
	Kind string
	// ID is the cell id of a legend row; Key the node or edge kind it explains, or the
	// multiplicity of a stack.
	ID, Key string
	// Text is what the renderer is given: a title, a listed line (Bullet says it gets
	// one), a legend caption or, for the notes block, nothing, the renderer owns its markup.
	Text   string
	Bullet bool
	Plain  []string
	Rect   diagram.Rect
	Swatch diagram.Rect
	// FontSize and LineH size the text; Wraps says the cell wraps long text at spaces
	// (a listed card, a note), so that only a word wider than the cell overflows it.
	FontSize, LineH float64
	Wraps           bool
}

// Widest is the text that decides whether the row fits sideways, and its width: the widest
// line when the cell does not wrap, the widest word when it does.
func (r CardRow) Widest() (string, float64) {
	best, width := "", 0.0
	for _, line := range r.Plain {
		candidates := []string{line}
		if r.Wraps {
			candidates = strings.Fields(line)
		}
		for _, c := range candidates {
			if w := TextWidth(c, r.FontSize); w > width {
				best, width = c, w
			}
		}
	}
	return best, width
}

// Height is what the row needs once drawn: its lines, wrapped when the cell wraps.
func (r CardRow) Height() float64 {
	lines := 0
	for _, line := range r.Plain {
		if r.Wraps {
			lines += max(1, len(Wrap(line, r.FontSize, r.Rect.W)))
		} else {
			lines++
		}
	}
	return float64(lines) * r.LineH
}

// Layout is what a card draws, in drawing order, and the height it needs to hold it.
type Layout struct {
	Rows []CardRow
	// Needed is the height from the card's top to the bottom of the last row, plus the
	// padding or, for a legend, the notes block under the rows.
	Needed float64
}

// CardLayout lays out a card the way the renderer draws it. The renderer draws from the
// result and the lint measures it, so what is checked is what is drawn.
func CardLayout(s *diagram.Spec, c diagram.Card) Layout {
	textX, textW := c.X+CardPadLeft, c.W-CardPadLeft-CardPadRight
	title := CardRow{
		Kind: RowTitle, Text: c.Title, Plain: []string{"● " + PlainText(c.Title)},
		Rect:     diagram.Rect{X: textX, Y: c.Y + CardPadTop, W: textW, H: CardTitleSize * LineHeightEm},
		FontSize: CardTitleSize, LineH: CardTitleSize * LineHeightEm, Wraps: true,
	}
	l := Layout{Rows: []CardRow{title}}
	titleH := math.Max(title.Height(), title.LineH)
	y := c.Y + CardPadTop + titleH + CardTextSize*LineHeightEm
	if c.Legend {
		legendRows(&l, s, c)
		return l
	}
	for _, line := range c.Lines {
		bullet := !strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "!!")
		plain := PlainText(line)
		if bullet {
			plain = "• " + plain
		}
		row := CardRow{
			Kind: RowLine, Text: line, Bullet: bullet, Plain: []string{plain},
			Rect:     diagram.Rect{X: textX, Y: y, W: textW, H: CardTextSize * CardLineHeightEm},
			FontSize: CardTextSize, LineH: CardTextSize * CardLineHeightEm, Wraps: true,
		}
		y += math.Max(row.Height(), row.LineH)
		l.Rows = append(l.Rows, row)
	}
	l.Needed = y + CardPadBottom - c.Y
	return l
}

func legendRows(l *Layout, s *diagram.Spec, c diagram.Card) {
	row := func(kind, id, key, text string, plain string, text_ diagram.Rect, swatch diagram.Rect) {
		l.Rows = append(l.Rows, CardRow{
			Kind: kind, ID: id, Key: key, Text: text, Plain: []string{plain}, Rect: text_, Swatch: swatch,
			FontSize: LegendTextSize, LineH: LegendTextSize * LineHeightEm,
		})
	}
	cell := func(x, y, w float64) diagram.Rect { return diagram.Rect{X: x, Y: y, W: w, H: LegendRowH} }
	usedNodes, usedEdges := map[string]bool{}, map[string]bool{}
	usesPipe := false
	for _, n := range s.Nodes {
		if n.Shape != "actor" {
			usedNodes[n.Kind] = true
		}
		usesPipe = usesPipe || n.Shape == "pipe"
	}
	for _, e := range s.Edges {
		usedEdges[e.Kind] = true
	}
	half := c.W / 2
	x, y := c.X+20, c.Y+LegendTop
	for _, kind := range NodeKindOrder {
		if !usedNodes[kind] {
			continue
		}
		row(RowLegendNode, "legend-node-"+kind, kind, LegendNodes[kind], LegendNodes[kind],
			cell(x+32, y, half-60), diagram.Rect{X: x, Y: y + 4, W: 22, H: 14})
		y += LegendRowPitch
	}
	if usesPipe {
		const text = "queue · messages in transit"
		row(RowLegendPipe, "legend-pipe", "bus", text, text, cell(x+32, y, half-60), diagram.Rect{X: x, Y: y + 4, W: 22, H: 14})
		y += LegendRowPitch
	}
	y += 8
	const steps = "[1] request step · [A] background step"
	row(RowLegendSteps, "legend-steps", "", steps, PlainText(steps), cell(x, y, half-20), diagram.Rect{})
	y += LegendRowPitch
	const port = "listening port on the server"
	row(RowLegendPort, "legend-port", "", port, port, cell(x+LegendTextGap, y, half-70),
		diagram.Rect{X: x, Y: y + 3, W: PortWidth("TCP 80"), H: PortHeight})
	leftBottom := y + LegendRowPitch

	x, y = c.X+half+10, c.Y+LegendTop
	for _, kind := range EdgeKindOrder {
		if !usedEdges[kind] {
			continue
		}
		row(RowLegendEdge, "legend-edge-"+kind, kind, LegendEdges[kind], LegendEdges[kind],
			cell(x+LegendTextGap, y, half-72), diagram.Rect{X: x, Y: y, W: LegendSample, H: LegendRowH})
		y += LegendRowPitch
	}
	y = statusRows(l, s, x, y, half-72)
	l.Rows = append(l.Rows, CardRow{
		Kind: RowLegendNotes, ID: c.ID + "-notes",
		Plain:    []string{"Large dashed frame: infra perimeter · small: functional group", "⚠ Gx: docs and code disagree, see \"Gaps\""},
		Rect:     diagram.Rect{X: c.X + 20, Y: c.Y + c.H - LegendNotesBelow, W: c.W - 40, H: LegendNotesH},
		FontSize: LegendTextSize, LineH: LegendTextSize * LineHeightEm, Wraps: true,
	})
	l.Needed = math.Max(leftBottom, y) - c.Y + LegendNotesBelow
}

// statusRows explains the planned and deprecated styles, a diff's code or the dimmed
// boxes, and each stacked box used in the diagram, so the hatch, the dots and the
// stacks never carry meaning alone (ADR-0006, ADR-0008, ADR-0021, ADR-0022).
func statusRows(l *Layout, s *diagram.Spec, x, y, w float64) float64 {
	row := func(kind, id, key, text string) {
		l.Rows = append(l.Rows, CardRow{
			Kind: kind, ID: id, Key: key, Text: text, Plain: []string{text},
			Rect:     diagram.Rect{X: x + LegendTextGap, Y: y, W: w, H: LegendRowH},
			Swatch:   diagram.Rect{X: x, Y: y + 2, W: LegendSample, H: 18},
			FontSize: LegendTextSize, LineH: LegendTextSize * LineHeightEm,
		})
		y += LegendRowPitch
	}
	dimmed := func() {
		text := "unchanged, dimmed"
		if s.Highlight {
			text = "outside the highlight, dimmed"
		}
		row(RowLegendUnchanged, "legend-unchanged", "", text)
	}
	used := map[string]bool{}
	for _, n := range s.Nodes {
		used[n.Status] = true
	}
	for _, z := range s.Zones {
		used[z.Status] = true
	}
	switch {
	case s.Diff:
		row(RowLegendAdded, "legend-added", "", "added by the proposal · pill: target")
		row(RowLegendRemoved, "legend-removed", "", "removed by the proposal")
		dimmed()
		used = nil
	case s.Highlight:
		dimmed()
	}
	if used["planned"] {
		row(RowLegendPlanned, "legend-planned", "", "planned · pill: target")
	}
	if used["deprecated"] {
		row(RowLegendDeprecated, "legend-deprecated", "", "deprecated")
	}
	shown := map[string]bool{}
	for i, n := range s.Nodes {
		if n.Multiplicity == "" || n.Shape == "actor" || shown[n.Multiplicity] {
			continue
		}
		shown[n.Multiplicity] = true
		row(RowLegendStack, "legend-stack-"+strconv.Itoa(i), n.Multiplicity, n.Multiplicity)
	}
	return y
}

// NoteRow is the text of a note: one source line per <br>, wrapped by the cell at the
// note's width.
func NoteRow(n diagram.Note) CardRow {
	row := CardRow{Kind: RowNote, ID: n.ID, Text: n.Text, Rect: n.Rect(), FontSize: SubFontSize, LineH: SubFontSize * LineHeightEm, Wraps: true}
	for _, line := range strings.Split(n.Text, "<br>") {
		row.Plain = append(row.Plain, PlainText(line))
	}
	return row
}
