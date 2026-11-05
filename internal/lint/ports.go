package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// portBadgeClearance is how far each badge is inflated before two are compared, so two
// badges 4 px apart or less read as touching.
const portBadgeClearance = 2.0

type placed struct {
	edge  string
	badge house.PortBadge
}

// checkPorts: a port badge sits on the server's border, leaves its icon visible and does
// not touch another badge of the same node. Connections to the same port share a badge.
func checkPorts(l *linter) {
	var badges []placed
	seen := map[string]placed{}
	for _, e := range l.s.Edges {
		b, ok := l.ports[e.ID]
		if !ok {
			continue
		}
		if first, dup := seen[b.ID]; dup {
			if first.badge.Text != b.Text {
				l.errf(e.ID, "port %q lands on the same point of %s as %q of %s", b.Text, b.Node, first.badge.Text, first.edge)
			}
			continue
		}
		seen[b.ID] = placed{e.ID, b}
		if b.Side == "" {
			l.errf(e.ID, "port %q: last point is not on a border of %s", e.Port, e.To)
			continue
		}
		badges = append(badges, placed{e.ID, b})
		for _, n := range l.s.Nodes {
			if n.ID == b.Node && n.Icon != "" && b.Rect.Intersects(house.NodeIconRect(n).Inflate(2)) {
				l.errf(e.ID, "port badge %q covers the icon of %s", b.Text, n.ID)
			}
		}
	}
	for i, a := range badges {
		for _, b := range badges[i+1:] {
			if a.badge.Node == b.badge.Node && a.badge.Rect.Inflate(portBadgeClearance).Intersects(b.badge.Rect.Inflate(portBadgeClearance)) {
				l.errf(a.edge, "port badge %q overlaps the badge of %s", a.badge.Text, b.edge)
			}
		}
	}
}
