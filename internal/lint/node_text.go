package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// checkNodeText: every line of a box fits its width with a margin for the export font,
// and the box is tall enough.
func checkNodeText(l *linter) {
	for _, n := range l.s.Nodes {
		if n.Shape == "actor" {
			continue
		}
		avail := n.W - 10 - 8
		if n.Icon != "" {
			avail = n.W - house.TextPadLeft - 8
		}
		if n.Shape == "pipe" {
			avail -= house.PipeCap + house.PipeEndRatio*n.W
		}
		title := n.Title
		if n.Badge != "" {
			title += " ⚠ " + n.Badge
		}
		if w := house.ExportTextWidth(title, house.TitleFontSize); !house.FitsWithMargin(w, avail, house.TitleFontSize) {
			l.errf(n.ID, "title overflows: %.0fpx text + %.0fpx margin in %.0fpx", w, house.TextMargin(avail, house.TitleFontSize), avail)
		}
		for _, line := range n.Lines() {
			if w := house.ExportTextWidth(house.PlainText(line), house.SubFontSize); !house.FitsWithMargin(w, avail, house.SubFontSize) {
				l.errf(n.ID, "%q overflows: %.0fpx text + %.0fpx margin in %.0fpx", line, w, house.TextMargin(avail, house.SubFontSize), avail)
			}
		}
		need := 10 + house.TitleFontSize*house.LineHeightEm + float64(len(n.Lines()))*house.SubFontSize*house.LineHeightEm + 10
		if n.Shape == "cylinder" {
			need += 16
		}
		if need > n.H {
			l.errf(n.ID, "text needs %.0fpx height, box is %.0fpx", need, n.H)
		}
	}
}
