package house

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
)

func BorderSide(p diagram.Point, r diagram.Rect) string {
	const eps = 1.0
	inX := p.X() >= r.X-eps && p.X() <= r.X+r.W+eps
	inY := p.Y() >= r.Y-eps && p.Y() <= r.Y+r.H+eps
	switch {
	case math.Abs(p.X()-r.X) < eps && inY:
		return "left"
	case math.Abs(p.X()-r.X-r.W) < eps && inY:
		return "right"
	case math.Abs(p.Y()-r.Y) < eps && inX:
		return "top"
	case math.Abs(p.Y()-r.Y-r.H) < eps && inX:
		return "bottom"
	}
	return ""
}

type PortBadge struct {
	ID, Node, Text, Side string
	Rect                 diagram.Rect
	Outer                diagram.Point
}

func PortWidth(text string) float64 { return TextWidth(text, PortFontSize) + 2*PortPadX }

// ports returns the badge each edge lands on. Edges reaching the same point of the same
// node with the same port share one badge.
func PortBadges(s *diagram.Spec) map[string]PortBadge {
	byEdge := map[string]PortBadge{}
	for _, e := range s.Edges {
		if e.Port == "" || len(e.Path) < 2 {
			continue
		}
		r, ok := s.AnchorRect(e.To)
		if !ok {
			continue
		}
		p := e.Path[len(e.Path)-1]
		side := BorderSide(p, r)
		w, h := PortWidth(e.Port), PortHeight
		b := PortBadge{
			ID:   fmt.Sprintf("%s__port_%s_%.0f_%.0f", e.To, side, p.X(), p.Y()),
			Node: e.To, Text: e.Port, Side: side,
			Rect: diagram.Rect{X: p.X() - w/2, Y: p.Y() - h/2, W: w, H: h},
		}
		switch side {
		case "left":
			b.Outer = diagram.Point{p.X() - w/2, p.Y()}
		case "right":
			b.Outer = diagram.Point{p.X() + w/2, p.Y()}
		case "top":
			b.Outer = diagram.Point{p.X(), p.Y() - h/2}
		default:
			b.Outer = diagram.Point{p.X(), p.Y() + h/2}
		}
		byEdge[e.ID] = b
	}
	return byEdge
}

// drawnPath is the path as rendered: it stops on the outer side of the port badge.
func DrawnPath(e diagram.Edge, ports map[string]PortBadge) []diagram.Point {
	p := append([]diagram.Point(nil), e.Path...)
	if b, ok := ports[e.ID]; ok {
		p[len(p)-1] = b.Outer
	}
	return p
}

func SegLen(a, b diagram.Point) float64 { return math.Abs(b.X()-a.X()) + math.Abs(b.Y()-a.Y()) }

func OnSegment(p, a, b diagram.Point) bool {
	const eps = 0.5
	return p.X() >= math.Min(a.X(), b.X())-eps && p.X() <= math.Max(a.X(), b.X())+eps &&
		p.Y() >= math.Min(a.Y(), b.Y())-eps && p.Y() <= math.Max(a.Y(), b.Y())+eps &&
		(math.Abs(a.X()-b.X()) < eps || math.Abs(a.Y()-b.Y()) < eps)
}

func LabelAnchor(e diagram.Edge, path []diagram.Point) diagram.Point {
	if e.LabelAt != nil {
		return *e.LabelAt
	}
	best, bestLen := 0, -1.0
	for i := 0; i+1 < len(path); i++ {
		if l := SegLen(path[i], path[i+1]); l > bestLen {
			best, bestLen = i, l
		}
	}
	a, b := path[best], path[best+1]
	return diagram.Point{(a.X() + b.X()) / 2, (a.Y() + b.Y()) / 2}
}

// labelFraction is where the anchor sits along the path, which is how draw.io positions edge labels.
func LabelFraction(anchor diagram.Point, path []diagram.Point) float64 {
	total, before := 0.0, -1.0
	for i := 0; i+1 < len(path); i++ {
		a, b := path[i], path[i+1]
		if before < 0 && OnSegment(anchor, a, b) {
			before = total + SegLen(a, anchor)
		}
		total += SegLen(a, b)
	}
	if total == 0 || before < 0 {
		return 0.5
	}
	return before / total
}

var (
	StepRe = regexp.MustCompile(`\[([0-9]{1,2}|[A-Z])\]`)
	BoldRe = regexp.MustCompile(`\*\*(.+?)\*\*`)
	WarnRe = regexp.MustCompile(`!!(.+?)!!`)
)

// plainText approximates the rendered width: a step badge is its text plus padding.
func PlainText(s string) string {
	s = StepRe.ReplaceAllString(s, " $1 ")
	s = BoldRe.ReplaceAllString(s, "$1")
	return WarnRe.ReplaceAllString(s, "$1")
}

func TextBox(cx, cy float64, lines []string, fontSize float64) diagram.Rect {
	w := 0.0
	for _, l := range lines {
		w = math.Max(w, TextWidth(PlainText(l), fontSize))
	}
	h := float64(len(lines))*fontSize*LineHeightEm + 2*LabelPadY
	w += 2 * LabelPadX
	return diagram.Rect{X: cx - w/2, Y: cy - h/2, W: w, H: h}
}

func LabelBox(e diagram.Edge, path []diagram.Point) (diagram.Rect, bool) {
	if e.Label == "" {
		return diagram.Rect{}, false
	}
	a := LabelAnchor(e, path)
	return TextBox(a.X()+e.LabelOffset.X(), a.Y()+e.LabelOffset.Y(), strings.Split(e.Label, "<br>"), EdgeFontSize), true
}

func ActorLabelBox(n diagram.Node) diagram.Rect {
	lines := append([]string{n.Title}, n.Lines()...)
	b := TextBox(n.X+n.W/2, 0, lines, TitleFontSize)
	b.Y = n.Y + n.H + ActorLabelGap - LabelPadY
	return b
}

func NodeIconRect(n diagram.Node) diagram.Rect {
	return diagram.Rect{X: n.X + IconInset, Y: n.Y + IconInset, W: IconSize, H: IconSize}
}

func ZoneTitleBox(z diagram.Zone) diagram.Rect {
	x := z.X + 12
	w := TextWidth(z.Label, ZoneFontSize) + 12
	if z.Icon != "" {
		w += ZoneIconSize + 8
	}
	if z.Sub != "" {
		w += 2*ZoneFontSize*CharWidthEm + TextWidth(z.Sub, SubFontSize)
	}
	return diagram.Rect{X: x, Y: z.Y + 4, W: w, H: math.Max(ZoneFontSize*LineHeightEm, ZoneIconSize) + 6}
}

// PillText is what a planned box's pill says: its target, or "planned" without one
// (ADR-0014).
func PillText(status, target string) string {
	if status != "planned" {
		return ""
	}
	if target != "" {
		return target
	}
	return "planned"
}

// DimOpacity is the opacity, in percent, of what a topology diff leaves unchanged
// (ADR-0021).
const DimOpacity = 30

// BoxPillText is the pill of a node or zone: in a diff its change, otherwise its target
// when planned.
func BoxPillText(status, target, change string) string {
	switch change {
	case diagram.Added:
		if target == "" {
			target = "added"
		}
		return "+ " + target
	case diagram.Removed:
		return "− removed"
	case diagram.Unchanged:
		return ""
	}
	return PillText(status, target)
}

// StackOffset is how far each of the two copies behind a stacked node sits up and to
// the right (ADR-0008).
const StackOffset = 10.0

// NodeFootprint is the area a node covers, its stack copies included, which the lint
// keeps clear.
func NodeFootprint(n diagram.Node) diagram.Rect {
	r := n.Rect()
	if n.Multiplicity == "" || n.Shape == "actor" {
		return r
	}
	return diagram.Rect{X: r.X, Y: r.Y - 2*StackOffset, W: r.W + 2*StackOffset, H: r.H + 2*StackOffset}
}

// NodePillRect straddles the top border of a planned node, flush with its right end,
// like a tag clipped on the box.
func NodePillRect(n diagram.Node) (diagram.Rect, bool) {
	text := BoxPillText(n.Status, n.Target, n.Change)
	if text == "" {
		return diagram.Rect{}, false
	}
	w := PortWidth(text)
	return diagram.Rect{X: n.X + n.W - w - 10, Y: n.Y - PortHeight/2, W: w, H: PortHeight}, true
}

// ZonePillRect sits on the zone's title row, after the title.
func ZonePillRect(z diagram.Zone) (diagram.Rect, bool) {
	text := BoxPillText(z.Status, z.Target, z.Change)
	if text == "" {
		return diagram.Rect{}, false
	}
	t := ZoneTitleBox(z)
	return diagram.Rect{X: t.X + t.W + 6, Y: t.Y + (t.H-PortHeight)/2, W: PortWidth(text), H: PortHeight}, true
}

// Wrap breaks text at spaces into lines no wider than width at fontSize. A word wider
// than width stays whole on its own line; the lint then reports it.
func Wrap(text string, fontSize, width float64) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(text) {
		next := word
		if line != "" {
			next = line + " " + word
		}
		if line != "" && TextWidth(next, fontSize) > width {
			lines = append(lines, line)
			next = word
		}
		line = next
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// Offering card metrics (ADR-0013), shared by the catalog resolver, the lint and the
// renderer.
const (
	OfferingPad          = 20.0
	OfferingHeaderH      = 32.0
	OfferingFontSize     = 11.0
	OfferingLabelSize    = 9.0
	OfferingGap          = 12.0
	OfferingLogoSize     = 20.0
	OfferingBulletIndent = 14.0
	OfferingRequestPad   = 8.0
)

// OfferingLineH is the height of one text line in an offering card.
func OfferingLineH(size float64) float64 { return math.Ceil(size * LineHeightEm) }
