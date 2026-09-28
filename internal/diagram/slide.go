package diagram

import "math"

// ForSlide keeps only the drawing: it drops the header, cards and notes, then moves
// everything so the content's bounding box starts at margin and the canvas ends margin
// after it. A slide shows the picture large and says the rest aloud; the full export
// keeps them.
func (s *Spec) ForSlide(margin float64) {
	s.Title, s.Subtitle, s.Meta = "", "", nil
	s.Cards, s.Notes = nil, nil
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	add := func(r Rect) {
		minX, minY = math.Min(minX, r.X), math.Min(minY, r.Y)
		maxX, maxY = math.Max(maxX, r.X+r.W), math.Max(maxY, r.Y+r.H)
	}
	for _, z := range s.Zones {
		add(z.Rect())
	}
	for _, n := range s.Nodes {
		r := n.Rect()
		if n.Shape == "actor" {
			r.H += 50 // the label sits under the figure
		}
		add(r)
	}
	for _, e := range s.Edges {
		for _, p := range e.Path {
			add(Rect{X: p.X(), Y: p.Y()})
		}
	}
	for _, a := range s.Arrows {
		add(Rect{X: a.From.X(), Y: a.From.Y()})
		add(Rect{X: a.To.X(), Y: a.To.Y()})
	}
	for _, o := range s.Offerings {
		add(o.Rect())
	}
	if math.IsInf(minX, 1) {
		return
	}
	dx, dy := margin-minX, margin-minY
	s.Width, s.Height = maxX-minX+2*margin, maxY-minY+2*margin
	for i := range s.Zones {
		s.Zones[i].X += dx
		s.Zones[i].Y += dy
	}
	for i := range s.Nodes {
		s.Nodes[i].X += dx
		s.Nodes[i].Y += dy
	}
	for i := range s.Edges {
		e := &s.Edges[i]
		for j := range e.Path {
			e.Path[j] = Point{e.Path[j].X() + dx, e.Path[j].Y() + dy}
		}
		if e.LabelAt != nil {
			p := Point{e.LabelAt.X() + dx, e.LabelAt.Y() + dy}
			e.LabelAt = &p
		}
	}
	for i := range s.Arrows {
		a := &s.Arrows[i]
		a.From = Point{a.From.X() + dx, a.From.Y() + dy}
		a.To = Point{a.To.X() + dx, a.To.Y() + dy}
	}
	for i := range s.Offerings {
		o := &s.Offerings[i]
		o.X += dx
		o.Y += dy
		o.Summary.Y += dy
		o.Request.Y += dy
		for j := range o.Provides {
			o.Provides[j].Y += dy
		}
		for j := range o.LabelsY {
			if o.LabelsY[j] != 0 {
				o.LabelsY[j] += dy
			}
		}
		for j := range o.Logos {
			o.Logos[j].X += dx
			o.Logos[j].Y += dy
		}
	}
}
