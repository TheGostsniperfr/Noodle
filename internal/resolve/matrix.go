package resolve

import (
	"fmt"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/access"
	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// Matrix resolves a matrix view (ADR-0019): one row per identity, one column per
// resource, both in the groups the view declares; each cell holds the identity's level
// on the resource and, in the diff state, what the plan changes. The system must pass
// model.Check first.
func Matrix(s *model.System, viewID string) (*diagram.Spec, error) {
	v, ok := s.Views[viewID]
	if !ok {
		return nil, fmt.Errorf("resolve: no view %q", viewID)
	}
	if v.Type != "matrix" {
		return nil, fmt.Errorf("resolve: view %q is a %s view", viewID, v.Type)
	}
	state := v.State
	if state == "" {
		state = access.Current
	}
	elements := map[string]model.Element{}
	for _, e := range s.Model.Elements {
		elements[e.ID] = e
	}
	before, after := access.New(s.Model, access.Current), access.New(s.Model, access.Target)
	if state == access.Current {
		after = before
	}
	if state == access.Target {
		before = after
	}
	diff := access.New(s.Model, access.Diff)

	m := &diagram.Matrix{State: state}
	gridX := landscapeMargin + house.MatrixRowHeaderW + house.MatrixGap
	y := landscapeTop

	// Column groups and headers, left to right with a gap between groups.
	type column struct {
		id string
		x  float64
	}
	var cols []column
	x := gridX
	for i, g := range v.Resources {
		if i > 0 {
			x += house.MatrixGroupGap
		}
		start := x
		for _, id := range g.Items {
			cols = append(cols, column{id, x})
			m.Columns = append(m.Columns, diagram.MatrixHeader{
				Lines: headerLines(v, elements[id]), X: x, Y: y + house.MatrixGroupH + house.MatrixGap,
				W: house.MatrixCellW, H: house.MatrixColumnH,
			})
			x += house.MatrixCellW + house.MatrixGap
		}
		m.ColumnGroups = append(m.ColumnGroups, diagram.MatrixGroup{
			Title: g.Title, Sub: g.Sub, Color: g.Color, X: start, Y: y, W: x - house.MatrixGap - start, H: house.MatrixGroupH,
		})
	}
	countX := x + house.MatrixGroupGap
	width := countX + house.MatrixCountW + landscapeMargin
	m.CountHeader = diagram.MatrixHeader{Lines: []string{"Can change"}, Sub: []string{countSub(state)},
		X: countX, Y: y + house.MatrixGroupH + house.MatrixGap, W: house.MatrixCountW, H: house.MatrixColumnH}
	y += house.MatrixGroupH + house.MatrixGap + house.MatrixColumnH + house.MatrixGap

	// Row groups: a title line, then one row per identity.
	for _, g := range v.Identities {
		y += house.MatrixGroupGap
		m.RowGroups = append(m.RowGroups, diagram.MatrixGroup{
			Title: g.Title, Sub: g.Sub, Color: g.Color, X: landscapeMargin, Y: y, W: width - 2*landscapeMargin, H: house.MatrixRowGroupH,
		})
		y += house.MatrixRowGroupH
		for _, id := range g.Items {
			e := elements[id]
			var sub []string
			if e.Desc != "" {
				sub = []string{e.Desc}
			}
			m.Rows = append(m.Rows, diagram.MatrixHeader{
				Lines: []string{titleOrLabel(v, e)}, Sub: sub, X: landscapeMargin, Y: y, W: house.MatrixRowHeaderW, H: house.MatrixCellH,
			})
			lb, la := before.Levels(id), after.Levels(id)
			gb, ga := before.GrantIDs(id), after.GrantIDs(id)
			changeable := [2]int{}
			for _, c := range cols {
				cell := diagram.MatrixCell{Level: la[c.id], X: c.x, Y: y, W: house.MatrixCellW, H: house.MatrixCellH}
				if state == access.Diff {
					cell.Before, cell.Change, cell.Note = change(lb[c.id], la[c.id], gb[c.id], ga[c.id])
					if cell.Change != "" {
						cell.Phase = diff.Phase(id, c.id)
					}
				}
				if access.CanChange(lb[c.id]) {
					changeable[0]++
				}
				if access.CanChange(la[c.id]) {
					changeable[1]++
				}
				m.Cells = append(m.Cells, cell)
			}
			count := fmt.Sprintf("%d", changeable[1])
			if state == access.Diff {
				count = fmt.Sprintf("%d → %d", changeable[0], changeable[1])
			}
			m.Counts = append(m.Counts, diagram.MatrixHeader{Lines: []string{count}, X: countX, Y: y, W: house.MatrixCountW, H: house.MatrixCellH})
			y += house.MatrixCellH + house.MatrixGap
		}
	}
	y = matrixLegend(m, state, y+house.MatrixLegendGap)
	return &diagram.Spec{
		ID: v.ID, Type: "matrix", Title: v.Title, Subtitle: v.Subtitle, Meta: v.Meta,
		Width: width, Height: y + landscapeMargin, Matrix: m,
	}, nil
}

// change compares a cell's level before and after the plan. Same level through other
// grants is a change too: the scope moved.
func change(before, after, grantsBefore, grantsAfter string) (was, kind, note string) {
	switch {
	case before == after && grantsBefore == grantsAfter:
		return "", "", ""
	case before == after:
		return before, "changed", "rescoped"
	case before == "":
		return "", "added", ""
	case after == "":
		return before, "removed", "removed"
	}
	return before, "changed", "was " + strings.ToLower(house.LevelWord(before))
}

func countSub(state string) string {
	if state == access.Diff {
		return "now → after"
	}
	return "write or admin"
}

func titleOrLabel(v *model.View, e model.Element) string {
	if l, ok := v.Labels[e.ID]; ok {
		return strings.ReplaceAll(l, "<br>", " ")
	}
	return e.Title
}

// headerLines is the column header text: the view's label for the element, split on
// <br>, or its title wrapped to the cell width.
func headerLines(v *model.View, e model.Element) []string {
	if l, ok := v.Labels[e.ID]; ok {
		return strings.Split(l, "<br>")
	}
	return house.Wrap(e.Title, house.MatrixHeaderFontSize, house.MatrixCellW-2*house.MatrixPad)
}

// matrixLegend lays out the two legend rows under the grid and returns the y below them.
func matrixLegend(m *diagram.Matrix, state string, y float64) float64 {
	row := func(title, sub string, items []diagram.MatrixCell) {
		m.LegendTitles = append(m.LegendTitles, diagram.MatrixHeader{Lines: []string{title}, Sub: []string{sub}, X: landscapeMargin, Y: y, W: house.MatrixRowHeaderW, H: house.MatrixCellH})
		x := landscapeMargin + house.MatrixRowHeaderW + house.MatrixGap
		for _, it := range items {
			it.X, it.Y, it.W, it.H = x, y, house.MatrixCellW, house.MatrixCellH
			m.Legend = append(m.Legend, it)
			x += house.MatrixCellW + house.MatrixLegendCaptionW
		}
		y += house.MatrixCellH + house.MatrixLegendGap/2
	}
	row("Level", "fill", []diagram.MatrixCell{
		{Level: "admin", Caption: "admin"}, {Level: "write", Caption: "write"}, {Level: "read", Caption: "read"},
		{Level: "breakglass", Caption: "break-glass"}, {Caption: "none"},
	})
	if state == access.Diff {
		row("Change", "tag: plan phase", []diagram.MatrixCell{
			{Level: "read", Caption: "unchanged"},
			{Level: "write", Change: "added", Phase: "P1", Caption: "added"},
			{Level: "write", Before: "admin", Change: "changed", Phase: "P1", Note: "was admin", Caption: "changed"},
			{Before: "read", Change: "removed", Phase: "P1", Note: "removed", Caption: "removed"},
		})
	}
	return y
}
