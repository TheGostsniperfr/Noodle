package resolve

import (
	"fmt"

	"github.com/TheGostsniperfr/Noodle/internal/access"
	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// Lens resolves a topology view lit for one subject (ADR-0019, ADR-0020): the same
// geometry, with every element the subject reaches marked with its level, the subject
// in focus, and everything else out of focus. In the target state, an element the
// subject reaches today and no longer will is marked removed.
func Lens(s *model.System, viewID, lensID string) (*diagram.Spec, error) {
	v, ok := s.Views[viewID]
	if !ok {
		return nil, fmt.Errorf("resolve: no view %q", viewID)
	}
	var lens *model.Lens
	for i := range v.Lenses {
		if v.Lenses[i].ID == lensID {
			lens = &v.Lenses[i]
		}
	}
	if lens == nil {
		return nil, fmt.Errorf("resolve: view %q has no lens %q", viewID, lensID)
	}
	spec, err := Topology(s, viewID)
	if err != nil {
		return nil, err
	}
	state := lens.State
	if state == "" {
		state = access.Current
	}
	levels := access.New(s.Model, state).Levels(lens.Subject)
	removed := map[string]bool{}
	if state == access.Target {
		for id := range access.New(s.Model, access.Current).Levels(lens.Subject) {
			if _, kept := levels[id]; !kept {
				removed[id] = true
			}
		}
	}
	reached := 0
	for i := range spec.Nodes {
		n := &spec.Nodes[i]
		switch {
		case n.ID == lens.Subject:
			n.Focus = true
		case levels[n.ID] != "":
			n.Level = levels[n.ID]
			reached++
		case removed[n.ID]:
			n.Removed, n.Dim = true, true
		default:
			n.Dim = true
		}
	}
	for i := range spec.Zones {
		if lv := levels[spec.Zones[i].ID]; lv != "" {
			spec.Zones[i].Level = lv
			reached++
		}
	}
	for i := range spec.Edges {
		spec.Edges[i].Dim = true
	}
	title := lens.Subject
	for _, e := range s.Model.Elements {
		if e.ID == lens.Subject {
			title = e.Title
		}
	}
	spec.Lens = lens.Subject
	spec.Title += " · lens " + title
	spec.Subtitle = fmt.Sprintf("What **%s** reaches, %s state: %d of the elements shown, %d removed by the migration.", title, state, reached, len(removedShown(spec)))
	return spec, nil
}

func removedShown(spec *diagram.Spec) []string {
	var out []string
	for _, n := range spec.Nodes {
		if n.Removed {
			out = append(out, n.ID)
		}
	}
	return out
}
