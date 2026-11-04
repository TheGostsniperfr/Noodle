package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// checkHeaderOverlap: nothing is drawn over the title, subtitle and meta lines.
func checkHeaderOverlap(l *linter) {
	header := house.HeaderBox(l.s)
	if header.W == 0 {
		return
	}
	bottom := header.Y + header.H
	for _, z := range l.s.Zones {
		if z.Rect().Intersects(header) {
			l.errf(z.ID, "overlaps the header, which ends at y %g", bottom)
		}
	}
	for _, n := range l.s.Nodes {
		if house.NodeFootprint(n).Intersects(header) {
			l.errf(n.ID, "overlaps the header, which ends at y %g", bottom)
		}
	}
	for _, c := range l.s.Cards {
		if c.Rect().Intersects(header) {
			l.errf(c.ID, "overlaps the header, which ends at y %g", bottom)
		}
	}
	for _, n := range l.s.Notes {
		if n.Rect().Intersects(header) {
			l.errf(n.ID, "overlaps the header, which ends at y %g", bottom)
		}
	}
}
