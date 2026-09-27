package lint

import (
	"math"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
)

// checkCollinear: two edges never run on top of each other.
func checkCollinear(l *linter) {
	type seg struct {
		edge string
		a, b diagram.Point
	}
	var segs []seg
	for _, e := range l.s.Edges {
		p := house.DrawnPath(e, l.ports)
		for i := 0; i+1 < len(p); i++ {
			segs = append(segs, seg{e.ID, p[i], p[i+1]})
		}
	}
	for i, s1 := range segs {
		for _, s2 := range segs[i+1:] {
			if s1.edge == s2.edge {
				continue
			}
			r1, r2 := segRect(s1.a, s1.b), segRect(s2.a, s2.b)
			sameH := r1.H == 0 && r2.H == 0 && math.Abs(r1.Y-r2.Y) < 6
			sameV := r1.W == 0 && r2.W == 0 && math.Abs(r1.X-r2.X) < 6
			if sameH && r1.X < r2.X+r2.W && r2.X < r1.X+r1.W {
				l.errf(s1.edge, "runs on top of %s (horizontal, y≈%.0f)", s2.edge, r1.Y)
			}
			if sameV && r1.Y < r2.Y+r2.H && r2.Y < r1.Y+r1.H {
				l.errf(s1.edge, "runs on top of %s (vertical, x≈%.0f)", s2.edge, r1.X)
			}
		}
	}
}
