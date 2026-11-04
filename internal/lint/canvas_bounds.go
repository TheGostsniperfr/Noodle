package lint

import (
	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
)

// checkCanvasBounds: every zone, node, card and note lies inside the canvas, so no export
// clips it. A stacked node counts with its copies.
func checkCanvasBounds(l *linter) {
	for _, z := range l.s.Zones {
		l.checkInside(z.ID, z.Rect())
	}
	for _, n := range l.s.Nodes {
		l.checkInside(n.ID, house.NodeFootprint(n))
	}
	for _, c := range l.s.Cards {
		l.checkInside(c.ID, c.Rect())
	}
	for _, n := range l.s.Notes {
		l.checkInside(n.ID, n.Rect())
	}
}

func (l *linter) checkInside(id string, r diagram.Rect) {
	if r.X < 0 {
		l.errf(id, "extends past the canvas: left edge at %g", r.X)
	}
	if r.Y < 0 {
		l.errf(id, "extends past the canvas: top edge at %g", r.Y)
	}
	if right := r.X + r.W; right > l.s.Width {
		l.errf(id, "extends past the canvas: right edge at %g, canvas is %g wide", right, l.s.Width)
	}
	if bottom := r.Y + r.H; bottom > l.s.Height {
		l.errf(id, "extends past the canvas: bottom edge at %g, canvas is %g high", bottom, l.s.Height)
	}
}
