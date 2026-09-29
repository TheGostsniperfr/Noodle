package resolve

import (
	"fmt"
	"math"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// Landscape grid metrics. Gaps are the house minimums (ADR-0006), so the computed grid
// passes the same spacing lint as a hand-placed topology.
const (
	landscapeMargin   = 40.0
	landscapeTop      = 170.0 // below the header
	landscapeWidth    = 2400.0
	tileHeight        = 60.0
	tileMinWidth      = 180.0
	sectionPadSide    = 20.0
	sectionPadTop     = 44.0 // clears the zone title row
	sectionPadBottom  = 20.0
	bandPadTop        = 48.0
	bandPadSide       = 24.0
	bandPadBottom     = 24.0
	sideSectionsWidth = 1 // tiles per row in the side column
)

// Landscape resolves a landscape view (ADR-0012): bands become region zones, sections
// group zones, items plain nodes, laid out in a computed grid. The system must pass
// model.Check first.
func Landscape(s *model.System, viewID string) (*diagram.Spec, error) {
	v, ok := s.Views[viewID]
	if !ok {
		return nil, fmt.Errorf("resolve: no view %q", viewID)
	}
	if v.Type != "landscape" {
		return nil, fmt.Errorf("resolve: view %q is a %s view", viewID, v.Type)
	}
	g := &grid{s: s, v: v, elements: map[string]model.Element{}}
	for _, e := range s.Model.Elements {
		g.elements[e.ID] = e
	}
	g.tileW = g.tileWidth()
	width := v.Width
	if width == 0 {
		width = g.naturalWidth()
	}
	out := &diagram.Spec{ID: v.ID, Type: "landscape", Title: v.Title, Subtitle: v.Subtitle, Meta: v.Meta, Width: width}
	g.out = out

	mainRight := width - landscapeMargin
	if len(v.Side) > 0 {
		sideW := 2*sectionPadSide + sideSectionsWidth*g.tileW
		sideX := width - landscapeMargin - sideW
		mainRight = sideX - house.MinZoneGap
		y := landscapeTop
		for i, sec := range v.Side {
			r := g.section(fmt.Sprintf("side-%d", i), "region", sec, sideX, y, sideW, sideSectionsWidth)
			y = r.Y + r.H + house.MinZoneGap
		}
	}

	y := landscapeTop
	for _, b := range v.Bands {
		y = g.band(b, landscapeMargin, y, mainRight-landscapeMargin) + house.MinZoneGap
	}
	bottom := y - house.MinZoneGap
	for _, z := range out.Zones {
		bottom = math.Max(bottom, z.Y+z.H)
	}
	out.Height = bottom + landscapeMargin
	return out, nil
}

// naturalWidth fits the widest band on one line, capped at landscapeWidth beyond which
// bands wrap.
func (g *grid) naturalWidth() float64 {
	widest := 0.0
	for _, b := range g.v.Bands {
		w := -house.MinZoneGap
		for _, sec := range b.Sections {
			w += house.MinZoneGap + g.sectionWidth(sec, len(sec.Items))
		}
		widest = math.Max(widest, w)
	}
	w := 2*landscapeMargin + 2*bandPadSide + widest
	if len(g.v.Side) > 0 {
		w += house.MinZoneGap + 2*sectionPadSide + sideSectionsWidth*g.tileW
	}
	return math.Min(math.Ceil(w), landscapeWidth)
}

type grid struct {
	s        *model.System
	v        *model.View
	elements map[string]model.Element
	tileW    float64
	out      *diagram.Spec
}

func (g *grid) title(id string) string {
	if l, ok := g.v.Labels[id]; ok {
		return l
	}
	return g.elements[id].Title
}

// tileWidth is one width for every tile of the view, wide enough for the longest
// title or tech line, so columns line up across bands.
func (g *grid) tileWidth() float64 {
	w := tileMinWidth
	each := func(secs []model.Section) {
		for _, sec := range secs {
			for _, id := range sec.Items {
				text := math.Max(house.TextWidth(g.title(id), house.TitleFontSize), house.TextWidth(g.elements[id].Tech, house.SubFontSize))
				w = math.Max(w, math.Ceil(house.TextPadLeft+text+8+2))
			}
		}
	}
	for _, b := range g.v.Bands {
		each(b.Sections)
	}
	each(g.v.Side)
	return w
}

// sectionWidth is the width a section needs with cols tiles per row, never narrower
// than its title.
func (g *grid) sectionWidth(sec model.Section, cols int) float64 {
	w := 2*sectionPadSide + float64(cols)*g.tileW + float64(cols-1)*house.MinNodeGap
	title := house.ZoneTitleBox(diagram.Zone{Label: sec.Title})
	return math.Max(w, title.W+2*sectionPadSide)
}

// section places a zone of width w (at least its natural width) and its tiles at (x, y),
// cols tiles per row. Extra width centres the tiles as one group.
func (g *grid) section(id, kind string, sec model.Section, x, y, w float64, cols int) diagram.Rect {
	cols = max(1, min(cols, len(sec.Items)))
	rows := (len(sec.Items) + cols - 1) / cols
	h := sectionPadTop + float64(rows)*tileHeight + float64(rows-1)*house.MinNodeGap + sectionPadBottom
	g.out.Zones = append(g.out.Zones, diagram.Zone{
		ID: id, Kind: kind, Label: sec.Title, Color: sec.Color,
		X: x, Y: y, W: w, H: h,
	})
	step := g.tileW + house.MinNodeGap
	offset := math.Floor((w - g.sectionWidth(sec, cols)) / 2)
	for i, item := range sec.Items {
		e := g.elements[item]
		st, tg := g.s.Status(item)
		col, row := i%cols, i/cols
		g.out.Nodes = append(g.out.Nodes, diagram.Node{
			ID: item, Kind: e.Kind, Icon: e.Icon, Title: g.title(item), Tech: e.Tech,
			Status: st, Target: tg,
			X: x + sectionPadSide + offset + float64(col)*step,
			Y: y + sectionPadTop + float64(row)*(tileHeight+house.MinNodeGap),
			W: g.tileW, H: tileHeight,
		})
	}
	return diagram.Rect{X: x, Y: y, W: w, H: h}
}

type placed struct {
	sec  model.Section
	cols int
	w    float64
}

// band packs its sections into lines, stretches each line to the band's width in
// proportion to the sections' natural widths, and returns the band's bottom edge.
func (g *grid) band(b model.Band, x, y, width float64) float64 {
	// The band is drawn before its sections, so reserve its slot now.
	slot := len(g.out.Zones)
	g.out.Zones = append(g.out.Zones, diagram.Zone{})
	avail := width - 2*bandPadSide
	var lines [][]placed
	used := 0.0
	for _, sec := range b.Sections {
		if sec.Color == "" {
			sec.Color = b.Color
		}
		cols := len(sec.Items)
		for cols > 1 && g.sectionWidth(sec, cols) > avail {
			cols--
		}
		p := placed{sec, cols, g.sectionWidth(sec, cols)}
		if len(lines) == 0 || used+house.MinZoneGap+p.w > avail {
			lines = append(lines, nil)
			used = -house.MinZoneGap
		}
		lines[len(lines)-1] = append(lines[len(lines)-1], p)
		used += house.MinZoneGap + p.w
	}
	top, n := y+bandPadTop, 0
	for _, line := range lines {
		natural := -house.MinZoneGap
		for _, p := range line {
			natural += house.MinZoneGap + p.w
		}
		stretch := (avail - natural) / (natural - house.MinZoneGap*float64(len(line)-1))
		cx, bottom := x+bandPadSide, top
		var prev *diagram.Rect
		for _, p := range line {
			w := math.Floor(p.w * (1 + stretch))
			r := g.section(fmt.Sprintf("%s-%d", b.ID, n), "group", p.sec, cx, top, w, p.cols)
			if b.Flow && prev != nil {
				mid := prev.Y + sectionPadTop + tileHeight/2
				g.out.Arrows = append(g.out.Arrows, diagram.Arrow{
					ID:   fmt.Sprintf("%s-flow-%d", b.ID, n),
					From: diagram.Point{prev.X + prev.W + 12, mid},
					To:   diagram.Point{r.X - 12, mid},
				})
			}
			prev = &r
			bottom = math.Max(bottom, r.Y+r.H)
			cx += w + house.MinZoneGap
			n++
		}
		top = bottom + house.MinZoneGap
	}
	bottom := top - house.MinZoneGap + bandPadBottom
	g.out.Zones[slot] = diagram.Zone{
		ID: b.ID, Kind: "region", Label: b.Title, Sub: b.Sub, Color: b.Color, Icon: b.Icon,
		X: x, Y: y, W: width, H: bottom - y,
	}
	return bottom
}
