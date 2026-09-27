package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// checkSteps keeps numbering checkable: one connection carries at most one step, and a
// step appears on one connection only. A flow that reuses connections belongs in a
// sequence diagram.
func checkSteps(l *linter) {
	owner := map[string]string{}
	for _, e := range l.s.Edges {
		steps := house.StepRe.FindAllStringSubmatch(e.Label, -1)
		if len(steps) > 1 {
			l.errf(e.ID, "carries %d steps, want at most one", len(steps))
		}
		for _, m := range steps {
			if prev, ok := owner[m[1]]; ok {
				l.errf(e.ID, "step [%s] already used on %s", m[1], prev)
				continue
			}
			owner[m[1]] = e.ID
		}
	}
}
