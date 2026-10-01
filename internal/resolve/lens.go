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
	zoneAt := map[string]int{}
	for i := range spec.Zones {
		zoneAt[spec.Zones[i].ID] = i
		if lv := levels[spec.Zones[i].ID]; lv != "" {
			spec.Zones[i].Level = lv
			reached++
		}
	}
	// A reached element the view does not show lights the nearest zone it does show,
	// so a coarse map still says where the subject reaches: SECRETS lit WRITE because
	// a project mount is written, though the mount has no box here.
	shown := map[string]bool{}
	for _, n := range spec.Nodes {
		shown[n.ID] = true
	}
	parents := map[string]string{}
	for _, e := range s.Model.Elements {
		parents[e.ID] = e.Parent
	}
	for _, id := range sortedKeys(levels) {
		if shown[id] {
			continue
		}
		if _, isZone := zoneAt[id]; isZone {
			continue
		}
		for p := parents[id]; p != ""; p = parents[p] {
			if i, ok := zoneAt[p]; ok {
				if z := &spec.Zones[i]; z.Level == "" || z.Level == "breakglass" || access.Stronger(levels[id], z.Level) {
					z.Level = levels[id]
				}
				break
			}
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
	// The key lives in the subtitle: a topology's legend card is sized by its layout,
	// which knows nothing of lenses.
	spec.Subtitle = fmt.Sprintf("What **%s** reaches, %s state: %d of the elements shown, %d removed by the migration. "+
		"Badge and border: level · grey: out of reach · dotted: reach the migration removes · a lit zone: something inside it, not drawn here.",
		title, state, reached, len(removedShown(spec)))
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
