package lint

import (
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/house"
)

// checkLabelOverlaps: an edge label covers nothing but its own edge, cards aside.
func checkLabelOverlaps(l *linter) {
	obs := l.obstacles()
	for _, e := range l.s.Edges {
		if len(e.Path) < 2 {
			continue
		}
		lb, ok := house.LabelBox(e, house.DrawnPath(e, l.ports))
		if !ok {
			continue
		}
		for _, o := range obs {
			if o.owner == e.ID || strings.HasPrefix(o.name, "card ") {
				continue
			}
			if lb.Intersects(o.rect) {
				l.errf(e.ID, "label overlaps %s", o.name)
			}
		}
	}
}
