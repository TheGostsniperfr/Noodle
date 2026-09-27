package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// checkIDs: unique ids, known kinds and colours, edges between existing boxes.
func checkIDs(l *linter) {
	seen := map[string]bool{}
	check := func(id string) {
		if seen[id] {
			l.errf(id, "duplicate id")
		}
		seen[id] = true
	}
	dark := house.Themes["dark"]
	for _, z := range l.s.Zones {
		check(z.ID)
		if _, ok := dark.Zones[z.Color]; !ok {
			l.errf(z.ID, "unknown zone color %q", z.Color)
		}
	}
	for _, n := range l.s.Nodes {
		check(n.ID)
		if _, ok := dark.Nodes[n.Kind]; !ok {
			l.errf(n.ID, "unknown node kind %q", n.Kind)
		}
	}
	for _, e := range l.s.Edges {
		check(e.ID)
		if _, ok := dark.Edges[e.Kind]; !ok {
			l.errf(e.ID, "unknown edge kind %q", e.Kind)
		}
		for _, end := range []string{e.From, e.To} {
			if _, ok := l.s.AnchorRect(end); !ok {
				l.errf(e.ID, "unknown endpoint %q", end)
			}
		}
		if len(e.Path) < 2 {
			l.errf(e.ID, "path needs at least 2 points")
		}
	}
}
