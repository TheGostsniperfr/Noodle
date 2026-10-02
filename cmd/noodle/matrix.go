package main

import (
	"fmt"
	"html"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
)

// matrix draws an access matrix (ADR-0019): the fill of a cell is the level, its frame
// and tag are the change, so the two codes never share a colour.
func (r *renderer) matrix(m *diagram.Matrix) {
	for i, g := range m.ColumnGroups {
		color := r.zoneColor(g.Color)
		id := fmt.Sprintf("mx-group%d", i)
		r.vertex(id, "1", "", style("rounded=0", "fillColor="+color, "fillOpacity=8", "strokeColor=none", "movable=0"), g.X, g.Y, g.W, g.H)
		r.vertex(id+"-bar", "1", "", style("rounded=0", "fillColor="+color, "strokeColor=none", "movable=0"), g.X, g.Y, g.W, 4)
		v := fmt.Sprintf(`<b>%s</b>`, html.EscapeString(g.Title))
		if g.Sub != "" {
			v += fmt.Sprintf(`<br><font color="%s" style="font-size:%gpx">%s</font>`, r.th.Muted, house.SubFontSize, html.EscapeString(g.Sub))
		}
		r.vertex(id+"-text", "1", v, style("text", "html=1", "align=left", "verticalAlign=middle", "spacingLeft=6", font(house.ZoneFontSize, color)), g.X, g.Y+4, g.W, g.H-4)
	}
	for i, h := range m.Columns {
		r.vertex(fmt.Sprintf("mx-col%d", i), "1", r.headerText(h, house.MatrixHeaderFontSize, true),
			style("rounded=0", "html=1", "fillColor="+r.th.CardFill, "strokeColor=none", "align=center", "verticalAlign=middle", "whiteSpace=wrap",
				font(house.MatrixHeaderFontSize, r.th.Title)), h.X, h.Y, h.W, h.H)
	}
	r.vertex("mx-count-head", "1", r.headerText(m.CountHeader, house.MatrixHeaderFontSize, true),
		style("text", "html=1", "align=center", "verticalAlign=middle", font(house.MatrixHeaderFontSize, r.th.Title)),
		m.CountHeader.X, m.CountHeader.Y, m.CountHeader.W, m.CountHeader.H)
	for i, g := range m.RowGroups {
		v := fmt.Sprintf(`<b>%s</b>`, html.EscapeString(strings.ToUpper(g.Title)))
		if g.Sub != "" {
			v += fmt.Sprintf(`&nbsp;&nbsp;<font color="%s">%s</font>`, r.th.Muted, html.EscapeString(g.Sub))
		}
		r.vertex(fmt.Sprintf("mx-rowgroup%d", i), "1", v, style("text", "html=1", "align=left", "verticalAlign=bottom", "spacingLeft=2",
			font(house.SubFontSize, r.th.Text)), g.X, g.Y, g.W, g.H)
	}
	for i, h := range m.Rows {
		id := fmt.Sprintf("mx-row%d", i)
		r.vertex(id, "1", r.headerText(h, house.MatrixRowFontSize, false), style("rounded=0", "html=1", "fillColor="+r.th.CardFill,
			"strokeColor=none", "align=left", "verticalAlign=middle", "spacingLeft=14", "whiteSpace=wrap", font(house.MatrixRowFontSize, r.th.Title)),
			h.X, h.Y, h.W, h.H)
		r.vertex(id+"-bar", "1", "", style("rounded=0", "fillColor="+r.th.Muted, "strokeColor=none", "movable=0"), h.X, h.Y, 4, h.H)
	}
	for i, c := range m.Cells {
		r.matrixCell(fmt.Sprintf("mx-cell%d", i), c)
	}
	for i, h := range m.Counts {
		r.vertex(fmt.Sprintf("mx-count%d", i), "1", "<b>"+html.EscapeString(h.Lines[0])+"</b>",
			style("text", "html=1", "align=center", "verticalAlign=middle", font(14, r.th.Title)), h.X, h.Y, h.W, h.H)
	}
	for i, h := range m.LegendTitles {
		r.vertex(fmt.Sprintf("mx-legend-title%d", i), "1", r.headerText(h, house.MatrixRowFontSize, false),
			style("text", "html=1", "align=left", "verticalAlign=middle", font(house.MatrixRowFontSize, r.th.Title)), h.X, h.Y, h.W, h.H)
	}
	for i, c := range m.Legend {
		id := fmt.Sprintf("mx-legend%d", i)
		r.matrixCell(id, c)
		r.vertex(id+"-caption", "1", html.EscapeString(c.Caption), style("text", "html=1", "align=left", "verticalAlign=middle",
			font(house.MatrixRowFontSize, r.th.Text)), c.X+c.W+8, c.Y, house.MatrixLegendCaptionW-12, c.H)
	}
}

func (r *renderer) zoneColor(name string) string {
	if c, ok := r.th.Zones[name]; ok {
		return c
	}
	return r.th.Muted
}

// headerText is a matrix header: lines in bold, sub lines smaller and muted.
func (r *renderer) headerText(h diagram.MatrixHeader, size float64, bold bool) string {
	var parts []string
	for _, l := range h.Lines {
		l = html.EscapeString(l)
		if bold || len(parts) == 0 {
			l = "<b>" + l + "</b>"
		}
		parts = append(parts, l)
	}
	for _, l := range h.Sub {
		parts = append(parts, fmt.Sprintf(`<font color="%s" style="font-size:%gpx">%s</font>`, r.th.Muted, house.SubFontSize, html.EscapeString(l)))
	}
	return strings.Join(parts, "<br>")
}

// matrixCell draws one cell: the level's fill (hatched for break-glass and for a removed
// access), the change's frame (solid added, dashed changed, dotted removed) and a tag
// with the change symbol and the plan phase.
func (r *renderer) matrixCell(id string, c diagram.MatrixCell) {
	lv := r.th.Levels[c.Level]
	fill := style("fillColor=" + lv.Fill)
	text := lv.Text
	word := "<b>" + house.LevelWord(c.Level) + "</b>"
	switch {
	case c.Change == "removed":
		muted := r.th.Levels[""]
		fill = style("fillColor="+r.th.Hatch, "fillOpacity=30") + hatch
		text = muted.Text
		word = "<b><s>" + house.LevelWord(c.Before) + "</s></b>"
	case c.Level == "breakglass":
		fill = style("fillColor="+lv.Text, "fillOpacity=25") + hatch
	}
	frame := style("strokeColor=none")
	if col, ok := r.th.Changes[c.Change]; ok {
		frame = style("strokeColor="+col, "strokeWidth=3")
		switch c.Change {
		case "changed":
			frame += style("dashed=1", "dashPattern=4 2")
		case "removed":
			frame += style("dashed=1", "dashPattern=1 1.5")
		}
	}
	v := word
	if c.Note != "" {
		v += fmt.Sprintf(`<br><font style="font-size:%gpx">%s</font>`, house.MatrixNoteFontSize, html.EscapeString(c.Note))
	}
	if c.Level == "" && c.Change == "" {
		v = ""
	}
	r.vertex(id, "1", v, style("rounded=1", "absoluteArcSize=1", "arcSize=8", "html=1", "align=center", "verticalAlign=middle", "whiteSpace=wrap",
		"movable=0", font(house.MatrixLevelFontSize, text))+fill+frame, c.X, c.Y, c.W, c.H)
	if t := house.MatrixTagText(c.Change, c.Phase); t != "" {
		w := house.TextWidth(t, house.MatrixTagFontSize) + 10
		r.vertex(id+"-tag", "1", "<b>"+html.EscapeString(t)+"</b>", style("rounded=1", "arcSize=50", "html=1", "fillColor="+r.th.Changes[c.Change],
			"strokeColor="+r.th.Background, "strokeWidth=1.5", "align=center", "verticalAlign=middle", "movable=0", font(house.MatrixTagFontSize, r.th.PortText)),
			c.X+c.W-w+4, c.Y-house.MatrixTagH/2, w, house.MatrixTagH)
	}
}
