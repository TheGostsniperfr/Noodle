package resolve

import (
	"fmt"
	"math"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

const (
	catalogCardWidth = 560.0
	catalogGap       = 40.0
	catalogColumns   = 2
	logoGap          = 16.0
	logoTextGap      = 6.0
	logoFontSize     = 10.0
)

// Catalog resolves a catalog view (ADR-0013): one card per offering, in include order
// (all offerings when include is empty), columns per row, each row as tall as its
// tallest card. The system must pass model.Check first.
func Catalog(s *model.System, viewID string) (*diagram.Spec, error) {
	v, ok := s.Views[viewID]
	if !ok {
		return nil, fmt.Errorf("resolve: no view %q", viewID)
	}
	if v.Type != "catalog" {
		return nil, fmt.Errorf("resolve: view %q is a %s view", viewID, v.Type)
	}
	byID := map[string]model.Offering{}
	var order []string
	for _, o := range s.Model.Offerings {
		byID[o.ID] = o
		order = append(order, o.ID)
	}
	if len(v.Include) > 0 {
		order = v.Include
	}
	elements := map[string]model.Element{}
	for _, e := range s.Model.Elements {
		elements[e.ID] = e
	}
	cols := v.Columns
	if cols == 0 {
		cols = catalogColumns
	}
	cols = min(cols, max(1, len(order)))
	width := 2*landscapeMargin + float64(cols)*catalogCardWidth + float64(cols-1)*catalogGap
	out := &diagram.Spec{ID: v.ID, Type: "catalog", Title: v.Title, Subtitle: v.Subtitle, Meta: v.Meta, Width: width}

	y := landscapeTop
	for start := 0; start < len(order); start += cols {
		row := order[start:min(start+cols, len(order))]
		first := len(out.Offerings)
		rowH := 0.0
		for i, id := range row {
			x := landscapeMargin + float64(i)*(catalogCardWidth+catalogGap)
			card := offeringCard(byID[id], elements, x, y, catalogCardWidth)
			rowH = math.Max(rowH, card.H)
			out.Offerings = append(out.Offerings, card)
		}
		for i := first; i < len(out.Offerings); i++ {
			out.Offerings[i].H = rowH
		}
		y += rowH + catalogGap
	}
	out.Height = y - catalogGap + landscapeMargin
	return out, nil
}

// offeringCard stacks the card's blocks top to bottom and returns it at its natural height.
func offeringCard(o model.Offering, elements map[string]model.Element, x, y, w float64) diagram.Offering {
	const pad, gap = house.OfferingPad, house.OfferingGap
	lineH := house.OfferingLineH(house.OfferingFontSize)
	labelH := house.OfferingLineH(house.OfferingLabelSize) + 4
	inner := w - 2*pad
	card := diagram.Offering{ID: o.ID, Title: o.Title, Icon: o.Icon, Status: o.Status, Target: o.Target, X: x, Y: y, W: w}
	cy := y + pad + house.OfferingHeaderH + gap
	if o.Summary != "" {
		card.Summary = diagram.Block{Lines: house.Wrap(o.Summary, house.OfferingFontSize, inner), Y: cy}
		cy += float64(len(card.Summary.Lines))*lineH + gap
	}
	if len(o.Provides) > 0 {
		card.LabelsY[0] = cy
		cy += labelH
		for _, p := range o.Provides {
			b := diagram.Block{Lines: house.Wrap(p, house.OfferingFontSize, inner-house.OfferingBulletIndent), Y: cy}
			card.Provides = append(card.Provides, b)
			cy += float64(len(b.Lines))*lineH + 2
		}
		cy += gap - 2
	}
	if o.Request != "" {
		card.LabelsY[1] = cy
		cy += labelH
		card.Request = diagram.Block{Lines: house.Wrap(o.Request, house.OfferingFontSize, inner-2*house.OfferingRequestPad), Y: cy + house.OfferingRequestPad}
		cy += float64(len(card.Request.Lines))*lineH + 2*house.OfferingRequestPad + gap
	}
	if len(o.BackedBy) > 0 {
		card.LabelsY[2] = cy
		cy += labelH
		lx := x + pad
		for _, id := range o.BackedBy {
			e := elements[id]
			lw := house.OfferingLogoSize + logoTextGap + house.TextWidth(e.Title, logoFontSize)
			if lx > x+pad && lx+lw > x+w-pad {
				lx, cy = x+pad, cy+house.OfferingLogoSize+8
			}
			card.Logos = append(card.Logos, diagram.Logo{Icon: e.Icon, Title: e.Title, X: lx, Y: cy})
			lx += lw + logoGap
		}
		cy += house.OfferingLogoSize + gap
	}
	card.H = cy - gap + pad - y
	return card
}
