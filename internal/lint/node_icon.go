package lint

// checkNodeIcon warns about a landscape item without a logo: a technology stack is read
// by its logos (ADR-0011), and a grid of text boxes reads as a list.
func checkNodeIcon(l *linter) {
	if l.s.Type != "landscape" {
		return
	}
	for _, n := range l.s.Nodes {
		if n.Icon == "" && n.Shape != "actor" {
			l.warnf(n.ID, "no icon")
		}
	}
}
