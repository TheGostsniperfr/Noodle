package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// checkNodeSpacing: boxes, stacks included, keep MinNodeGap between them.
func checkNodeSpacing(l *linter) {
	for i, n := range l.s.Nodes {
		for _, o := range l.s.Nodes[i+1:] {
			if house.NodeFootprint(n).Inflate(house.MinNodeGap / 2).Intersects(house.NodeFootprint(o).Inflate(house.MinNodeGap / 2)) {
				l.errf(n.ID, "closer than %.0fpx to %s", house.MinNodeGap, o.ID)
			}
		}
	}
}
