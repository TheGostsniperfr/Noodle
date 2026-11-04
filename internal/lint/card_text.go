package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// checkCardText: every text a card or a note draws fits the cell the renderer gives it,
// and a card is tall enough for its rows. The rows are house.CardLayout's, the same the
// renderer draws.
func checkCardText(l *linter) {
	for _, c := range l.s.Cards {
		layout := house.CardLayout(l.s, c)
		for _, row := range layout.Rows {
			if text, w := row.Widest(); w > row.Rect.W {
				l.errf(c.ID, "%q needs %.0fpx, its cell is %.0fpx wide", text, w, row.Rect.W)
			}
		}
		if layout.Needed > c.H {
			l.errf(c.ID, "text needs %.0fpx height, card is %.0fpx", layout.Needed, c.H)
		}
	}
	for _, n := range l.s.Notes {
		row := house.NoteRow(n)
		if text, w := row.Widest(); w > row.Rect.W {
			l.errf(n.ID, "%q needs %.0fpx, the note is %.0fpx wide", text, w, row.Rect.W)
		}
	}
}
