package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// checkNodeInZone: a box is inside a zone or outside it, and never on its title.
func checkNodeInZone(l *linter) {
	for _, n := range l.s.Nodes {
		for _, z := range l.s.Zones {
			if n.Rect().Intersects(house.ZoneTitleBox(z)) {
				l.errf(n.ID, "covers title of zone %s", z.ID)
			}
			if n.Rect().Intersects(z.Rect()) && !z.Rect().Contains(n.Rect()) {
				l.errf(n.ID, "straddles border of zone %s", z.ID)
			}
		}
	}
}
