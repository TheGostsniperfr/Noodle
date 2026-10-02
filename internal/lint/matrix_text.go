package lint

import (
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
)

// checkMatrixText: every header, cell and group title of an access matrix fits its box
// (ADR-0006, ADR-0019). Long column titles wrap through the view's labels.
func checkMatrixText(l *linter) {
	m := l.s.Matrix
	if m == nil {
		return
	}
	fits := func(where, text string, size, width float64) {
		if w := house.TextWidth(text, size); w > width {
			l.errf(where, "%q overflows: %.0fpx text in %.0fpx", text, w, width)
		}
	}
	block := func(where string, h diagram.MatrixHeader, size float64) {
		for _, line := range h.Lines {
			fits(where, house.PlainText(line), size, h.W-2*house.MatrixPad)
		}
		for _, line := range h.Sub {
			fits(where, line, house.SubFontSize, h.W-2*house.MatrixPad)
		}
		need := float64(len(h.Lines))*size*house.LineHeightEm + float64(len(h.Sub))*house.SubFontSize*house.LineHeightEm + 2*house.MatrixPad
		if need > h.H {
			l.errf(where, "text needs %.0fpx height, header is %.0fpx", need, h.H)
		}
	}
	for _, g := range m.ColumnGroups {
		fits("matrix group "+g.Title, strings.ToUpper(g.Title), house.ZoneFontSize, g.W-2*house.MatrixPad)
		fits("matrix group "+g.Title, g.Sub, house.SubFontSize, g.W-2*house.MatrixPad)
	}
	for _, h := range m.Columns {
		block("matrix column "+firstLine(h), h, house.MatrixHeaderFontSize)
	}
	for _, h := range m.Rows {
		block("matrix row "+firstLine(h), h, house.MatrixRowFontSize)
	}
	for _, c := range append(append([]diagram.MatrixCell(nil), m.Cells...), m.Legend...) {
		inner := c.W - 2*house.MatrixPad
		fits("matrix cell", house.LevelWord(c.Level), house.MatrixLevelFontSize, inner)
		fits("matrix cell", c.Note, house.MatrixNoteFontSize, inner)
	}
}

func firstLine(h diagram.MatrixHeader) string {
	if len(h.Lines) == 0 {
		return ""
	}
	return h.Lines[0]
}
