package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// checkEdgeCrossings: segments are orthogonal and cross no box, zone title, note,
// card, label or port badge other than their own ends.
func checkEdgeCrossings(l *linter) {
	obs := l.obstacles()
	for _, e := range l.s.Edges {
		if len(e.Path) < 2 {
			continue
		}
		path := house.DrawnPath(e, l.ports)
		own := map[string]bool{e.ID: true}
		if b, ok := l.ports[e.ID]; ok {
			own[b.ID] = true
		}
		last := len(path) - 2
		for i := 0; i <= last; i++ {
			a, b := path[i], path[i+1]
			if a.X() != b.X() && a.Y() != b.Y() {
				l.errf(e.ID, "segment %d is diagonal", i)
			}
			for _, o := range obs {
				if own[o.owner] {
					continue
				}
				r := o.rect
				if o.node {
					if (i == 0 && o.owner == e.From) || (i == last && o.owner == e.To) {
						continue
					}
					r = r.Inflate(house.ObstacleMargin)
				}
				if segHits(a, b, r) {
					l.errf(e.ID, "segment %d crosses %s", i, o.name)
				}
			}
		}
	}
}
