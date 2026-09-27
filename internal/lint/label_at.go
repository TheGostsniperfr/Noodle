package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// checkLabelAt: a pinned label sits on the path as drawn, port badges included.
func checkLabelAt(l *linter) {
	for _, e := range l.s.Edges {
		if len(e.Path) < 2 || e.LabelAt == nil {
			continue
		}
		path := house.DrawnPath(e, l.ports)
		on := false
		for i := 0; i+1 < len(path); i++ {
			on = on || house.OnSegment(*e.LabelAt, path[i], path[i+1])
		}
		if !on {
			l.errf(e.ID, "label_at %v is not on the drawn path", *e.LabelAt)
		}
	}
}
