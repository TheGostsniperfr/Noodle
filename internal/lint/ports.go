package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// checkPorts: a port badge sits on the server's border and leaves its icon visible.
func checkPorts(l *linter) {
	seen := map[string]bool{}
	for _, e := range l.s.Edges {
		b, ok := l.ports[e.ID]
		if !ok || seen[b.ID] {
			continue
		}
		seen[b.ID] = true
		if b.Side == "" {
			l.errf(e.ID, "port %q: last point is not on a border of %s", e.Port, e.To)
			continue
		}
		for _, n := range l.s.Nodes {
			if n.ID == b.Node && n.Icon != "" && b.Rect.Intersects(house.NodeIconRect(n).Inflate(2)) {
				l.errf(e.ID, "port badge %q covers the icon of %s", b.Text, n.ID)
			}
		}
	}
}
