package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// checkNodeSpacing: boxes keep MinNodeGap between them.
func checkNodeSpacing(l *linter) {
	for i, n := range l.s.Nodes {
		for _, o := range l.s.Nodes[i+1:] {
			if n.Rect().Inflate(house.MinNodeGap / 2).Intersects(o.Rect().Inflate(house.MinNodeGap / 2)) {
				l.errf(n.ID, "closer than %.0fpx to %s", house.MinNodeGap, o.ID)
			}
		}
	}
}
