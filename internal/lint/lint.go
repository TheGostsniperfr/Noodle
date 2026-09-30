// Package lint checks resolved diagrams against the house style (ADR-0006): readable
// text, clear spacing, orthogonal edges that cross no box, and checkable numbering
// (ADR-0003, ADR-0004). Each rule lives in its own file.
package lint

import (
	"fmt"
	"math"
	"sort"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
)

// Finding is an error unless Warn: warnings are printed but never block rendering.
type Finding struct {
	Where, Msg string
	Warn       bool
}

type rule func(*linter)

// rules run in this order; findings are then sorted by the id they name.
var rules = []rule{
	checkIDs,
	checkNodeText,
	checkNodeSpacing,
	checkNodeInZone,
	checkZoneSpacing,
	checkPorts,
	checkEdgeSides,
	checkLabelAt,
	checkEdgeCrossings,
	checkLabelOverlaps,
	checkCollinear,
	checkSteps,
	checkNodeIcon,
	checkOfferingText,
}

type linter struct {
	s        *diagram.Spec
	ports    map[string]house.PortBadge
	findings []Finding
}

func (l *linter) errf(where, format string, args ...any) {
	l.findings = append(l.findings, Finding{Where: where, Msg: fmt.Sprintf(format, args...)})
}

func (l *linter) warnf(where, format string, args ...any) {
	l.findings = append(l.findings, Finding{Where: where, Msg: fmt.Sprintf(format, args...), Warn: true})
}

func Lint(s *diagram.Spec) []Finding {
	l := &linter{s: s, ports: house.PortBadges(s)}
	for _, r := range rules {
		r(l)
	}
	sort.SliceStable(l.findings, func(i, j int) bool { return l.findings[i].Where < l.findings[j].Where })
	return l.findings
}

type obstacle struct {
	name  string
	rect  diagram.Rect
	owner string // edge id owning this label or port, node id for a node
	node  bool
}

func (l *linter) obstacles() []obstacle {
	var obs []obstacle
	for _, n := range l.s.Nodes {
		obs = append(obs, obstacle{name: "node " + n.ID, rect: house.NodeFootprint(n), owner: n.ID, node: true})
		if n.Shape == "actor" {
			obs = append(obs, obstacle{name: "label of " + n.ID, rect: house.ActorLabelBox(n), owner: n.ID})
		}
	}
	for _, z := range l.s.Zones {
		obs = append(obs, obstacle{name: "title of zone " + z.ID, rect: house.ZoneTitleBox(z)})
	}
	for _, nt := range l.s.Notes {
		obs = append(obs, obstacle{name: "note " + nt.ID, rect: nt.Rect()})
	}
	for _, c := range l.s.Cards {
		obs = append(obs, obstacle{name: "card " + c.ID, rect: c.Rect()})
	}
	seen := map[string]bool{}
	for _, e := range l.s.Edges {
		if b, ok := house.LabelBox(e, house.DrawnPath(e, l.ports)); ok {
			obs = append(obs, obstacle{name: "label of " + e.ID, rect: b, owner: e.ID})
		}
		if b, ok := l.ports[e.ID]; ok && !seen[b.ID] {
			seen[b.ID] = true
			obs = append(obs, obstacle{name: "port " + b.ID, rect: b.Rect, owner: b.ID})
		}
	}
	return obs
}

func segRect(a, b diagram.Point) diagram.Rect {
	return diagram.Rect{X: math.Min(a.X(), b.X()), Y: math.Min(a.Y(), b.Y()), W: math.Abs(b.X() - a.X()), H: math.Abs(b.Y() - a.Y())}
}

// segHits treats the segment as a zero-width rectangle so a line grazing a box counts as a hit.
func segHits(a, b diagram.Point, r diagram.Rect) bool {
	s := segRect(a, b)
	return s.X <= r.X+r.W && r.X <= s.X+s.W && s.Y <= r.Y+r.H && r.Y <= s.Y+s.H
}
