package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

var connectionKinds = map[string]bool{"flow": true, "auth": true, "tunnel": true, "async": true, "blocked": true}

// checkEdgeSides: an edge starts and ends on borders; egress leaves right or bottom and
// ingress enters left or top, except against_flow and references (ADR-0003).
func checkEdgeSides(l *linter) {
	for _, e := range l.s.Edges {
		if len(e.Path) < 2 {
			continue
		}
		src, _ := l.s.AnchorRect(e.From)
		dst, _ := l.s.AnchorRect(e.To)
		out, in := house.BorderSide(e.Path[0], src), house.BorderSide(e.Path[len(e.Path)-1], dst)
		if out == "" {
			l.errf(e.ID, "first point %v is not on the border of %s", e.Path[0], e.From)
		}
		if in == "" {
			l.errf(e.ID, "last point %v is not on the border of %s", e.Path[len(e.Path)-1], e.To)
		}
		if connectionKinds[e.Kind] && !e.AgainstFlow {
			if out != "" && out != "right" && out != "bottom" {
				l.errf(e.ID, "egress leaves %s from the %s side, want right or bottom", e.From, out)
			}
			if in != "" && in != "left" && in != "top" {
				l.errf(e.ID, "ingress enters %s from the %s side, want left or top", e.To, in)
			}
		}
	}
}
