package resolve

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/access"
	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// Access view metrics (plan 004). Port spacing keeps a 16 px label clear of the next
// line; lanes are 12 px apart so no two vertical runs come within the 6 px the
// collinear rule allows.
const (
	accessNodeMinW   = 220.0
	accessNodeMinH   = 72.0
	accessPortGap    = 26.0
	accessLaneGap    = 12.0
	accessLanePad    = 20.0
	accessLabelPad   = 20.0
	accessMinLabelW  = 60.0
	accessSlotH      = 26.0
	accessNudge      = 8.0
	accessCardW      = 380.0
	accessCardGap    = 40.0
	accessCardLineH  = 22.0
	accessMinCanvasW = 1200.0
)

// Access resolves an access view (ADR-0019): the hops of internal/access laid out in
// layers by longest path, sinks in the last layer, one pass-through slot per layer an
// edge skips, one lane per edge in each corridor. The system must pass model.Check.
func Access(s *model.System, viewID string) (*diagram.Spec, error) {
	v, ok := s.Views[viewID]
	if !ok {
		return nil, fmt.Errorf("resolve: no view %q", viewID)
	}
	if v.Type != "access" {
		return nil, fmt.Errorf("resolve: view %q is a %s view", viewID, v.Type)
	}
	state := v.State
	if state == "" {
		state = access.Current
	}
	var subject, resource string
	if v.Focus != nil {
		subject, resource = v.Focus.Subject, v.Focus.Resource
	}
	elements := map[string]model.Element{}
	for _, e := range s.Model.Elements {
		elements[e.ID] = e
	}
	g := access.New(s.Model, state)
	hops := g.Subgraph(subject, resource)
	if state != access.Diff {
		// Status says what the migration changes; only a diff shows the change.
		for i := range hops {
			hops[i].Status, hops[i].Target = "", ""
		}
	}
	if len(hops) == 0 {
		return nil, fmt.Errorf("resolve: access view %q has nothing to draw in state %s", viewID, state)
	}
	badges := badgesOf(s)
	l := &accessLayout{s: s, elements: elements, hops: hops, badges: badges}
	if err := l.layer(); err != nil {
		return nil, fmt.Errorf("resolve: access view %q: %w", viewID, err)
	}
	l.order()
	l.place()
	out := &diagram.Spec{ID: v.ID, Type: "access", Title: v.Title, Subtitle: v.Subtitle, Meta: v.Meta}
	out.Nodes = l.nodes(badges)
	out.Edges = l.edges()
	out.Width = math.Max(l.width, accessMinCanvasW)
	out.Cards = accessCards(s, g, state, subject, resource, out, l.bottom+house.MinZoneGap)
	out.Height = landscapeMargin
	for _, c := range out.Cards {
		out.Height = math.Max(out.Height, c.Y+c.H+landscapeMargin)
	}
	return out, nil
}

func badgesOf(s *model.System) map[string][]string {
	out := map[string][]string{}
	for _, a := range s.Model.Annotations {
		for _, t := range a.Targets {
			out[t] = append(out[t], a.ID)
		}
	}
	return out
}

// item is a node or a pass-through slot in one layer.
type item struct {
	id     string // element id, or "" for a slot
	hop    int    // slot only: the hop it carries
	layer  int
	pos    float64 // barycentre key while ordering
	y, h   float64
	in     []*segment
	out    []*segment
	inY    map[*segment]float64
	outY   map[*segment]float64
	center float64
}

// segment joins two items in consecutive layers; a hop is a chain of segments.
type segment struct {
	hop      int
	from, to *item
	lane     float64
	last     bool // ends on the hop's target, so it carries the label
}

type accessLayout struct {
	s        *model.System
	elements map[string]model.Element
	hops     []access.Hop
	badges   map[string][]string
	layerOf  map[string]int
	layers   [][]*item
	byID     map[string]*item
	chains   [][]*segment // per hop
	x        []float64    // left edge of each layer
	w        []float64    // width of each layer
	corridor []float64    // left edge of the corridor after each layer
	width    float64
	bottom   float64
}

// layer assigns longest-path layers, moves sinks to the last layer and expands each hop
// into a chain of segments through pass-through slots.
func (l *accessLayout) layer() error {
	succ := map[string][]string{}
	nodes := map[string]bool{}
	for _, h := range l.hops {
		succ[h.From] = append(succ[h.From], h.To)
		nodes[h.From], nodes[h.To] = true, true
	}
	l.layerOf = map[string]int{}
	state := map[string]int{}
	var depth func(string) (int, error)
	// depth is the longest path from id to a sink; layers count from the sources.
	depth = func(id string) (int, error) {
		switch state[id] {
		case 1:
			return 0, fmt.Errorf("grants form a cycle through %s", id)
		case 2:
			return l.layerOf[id], nil
		}
		state[id] = 1
		d := 0
		for _, n := range succ[id] {
			dn, err := depth(n)
			if err != nil {
				return 0, err
			}
			d = max(d, dn+1)
		}
		state[id] = 2
		l.layerOf[id] = d
		return d, nil
	}
	maxDepth := 0
	for _, id := range sortedKeys(nodes) {
		d, err := depth(id)
		if err != nil {
			return err
		}
		maxDepth = max(maxDepth, d)
	}
	// Depth to a sink puts every sink in the last column; layers run left to right.
	for id, d := range l.layerOf {
		l.layerOf[id] = maxDepth - d
	}
	l.layers = make([][]*item, maxDepth+1)
	l.byID = map[string]*item{}
	for _, id := range sortedKeys(nodes) {
		it := &item{id: id, layer: l.layerOf[id], inY: map[*segment]float64{}, outY: map[*segment]float64{}}
		l.byID[id] = it
		l.layers[it.layer] = append(l.layers[it.layer], it)
	}
	l.chains = make([][]*segment, len(l.hops))
	for i, h := range l.hops {
		from := l.byID[h.From]
		for k := l.layerOf[h.From] + 1; k <= l.layerOf[h.To]; k++ {
			to := l.byID[h.To]
			if k < l.layerOf[h.To] {
				to = &item{hop: i, layer: k, inY: map[*segment]float64{}, outY: map[*segment]float64{}}
				l.layers[k] = append(l.layers[k], to)
			}
			sg := &segment{hop: i, from: from, to: to, last: to.id == h.To}
			from.out = append(from.out, sg)
			to.in = append(to.in, sg)
			l.chains[i] = append(l.chains[i], sg)
			from = to
		}
	}
	return nil
}

// order runs two barycentre sweeps, down then up, ties broken by id then hop, so the
// same model always gives the same drawing.
func (l *accessLayout) order() {
	key := func(it *item) string {
		if it.id != "" {
			return it.id
		}
		return fmt.Sprintf("~%04d", it.hop)
	}
	for _, layer := range l.layers {
		sort.SliceStable(layer, func(i, j int) bool { return key(layer[i]) < key(layer[j]) })
		for i, it := range layer {
			it.pos = float64(i)
		}
	}
	sweep := func(k int, neighbours func(*item) []*item) {
		layer := l.layers[k]
		for _, it := range layer {
			ns := neighbours(it)
			if len(ns) == 0 {
				continue
			}
			sum := 0.0
			for _, n := range ns {
				sum += n.pos
			}
			it.pos = sum / float64(len(ns))
		}
		sort.SliceStable(layer, func(i, j int) bool {
			if layer[i].pos != layer[j].pos {
				return layer[i].pos < layer[j].pos
			}
			return key(layer[i]) < key(layer[j])
		})
		for i, it := range layer {
			it.pos = float64(i)
		}
	}
	preds := func(it *item) []*item {
		var out []*item
		for _, sg := range it.in {
			out = append(out, sg.from)
		}
		return out
	}
	succs := func(it *item) []*item {
		var out []*item
		for _, sg := range it.out {
			out = append(out, sg.to)
		}
		return out
	}
	for range 2 {
		for k := 1; k < len(l.layers); k++ {
			sweep(k, preds)
		}
		for k := len(l.layers) - 2; k >= 0; k-- {
			sweep(k, succs)
		}
	}
}

// place sizes nodes to their text and ports, stacks each layer, spreads ports and
// assigns lanes, then widens each corridor to its lanes and its widest label.
func (l *accessLayout) place() {
	l.w = make([]float64, len(l.layers))
	for k, layer := range l.layers {
		for _, it := range layer {
			if it.id == "" {
				it.h = accessSlotH
				continue
			}
			w, h := accessNodeSize(l.elements[it.id], strings.Join(l.badges[it.id], " "))
			it.h = math.Max(h, float64(max(len(it.in), len(it.out))+1)*accessPortGap)
			l.w[k] = math.Max(l.w[k], w)
		}
	}
	l.stack()
	for _, layer := range l.layers {
		for _, it := range layer {
			l.bottom = math.Max(l.bottom, it.y+it.h)
		}
	}
	for _, layer := range l.layers {
		for _, it := range layer {
			spread(it, it.in, it.inY, func(sg *segment) float64 { return sg.from.center })
			spread(it, it.out, it.outY, func(sg *segment) float64 { return sg.to.center })
		}
	}
	l.x = make([]float64, len(l.layers))
	l.corridor = make([]float64, len(l.layers))
	x := landscapeMargin
	for k := range l.layers {
		l.x[k] = x
		x += l.w[k]
		if k == len(l.layers)-1 {
			break
		}
		l.corridor[k] = x
		x += l.corridorWidth(k)
	}
	l.width = x + landscapeMargin
}

// stack places each layer top to bottom in its order, then pulls every item towards
// the mean centre of its neighbours, forward then backward, never closer to the item
// above than the node gap. Straighter runs need fewer bends and cross less.
func (l *accessLayout) stack() {
	for _, layer := range l.layers {
		y := landscapeTop
		for _, it := range layer {
			it.y, it.center = y, y+it.h/2
			y += it.h + house.MinNodeGap
		}
	}
	align := func(k int, neighbours func(*item) []*item) {
		top := landscapeTop
		for _, it := range l.layers[k] {
			want := it.y
			if ns := neighbours(it); len(ns) > 0 {
				sum := 0.0
				for _, n := range ns {
					sum += n.center
				}
				want = sum/float64(len(ns)) - it.h/2
			}
			it.y = math.Max(math.Round(want), top)
			it.center = it.y + it.h/2
			top = math.Ceil(it.y + it.h + house.MinNodeGap)
		}
	}
	preds := func(it *item) []*item {
		var out []*item
		for _, sg := range it.in {
			out = append(out, sg.from)
		}
		return out
	}
	succs := func(it *item) []*item {
		var out []*item
		for _, sg := range it.out {
			out = append(out, sg.to)
		}
		return out
	}
	for k := len(l.layers) - 2; k >= 0; k-- {
		align(k, succs)
	}
	for k := 1; k < len(l.layers); k++ {
		align(k, preds)
	}
	// Alignment only pushes down; lift the drawing back under the header.
	top := math.Inf(1)
	for _, layer := range l.layers {
		for _, it := range layer {
			top = math.Min(top, it.y)
		}
	}
	for _, layer := range l.layers {
		for _, it := range layer {
			it.y -= top - landscapeTop
			it.center = it.y + it.h/2
		}
	}
}

func spread(it *item, segs []*segment, ys map[*segment]float64, other func(*segment) float64) {
	sort.SliceStable(segs, func(i, j int) bool { return other(segs[i]) < other(segs[j]) })
	for i, sg := range segs {
		if it.id == "" {
			ys[sg] = it.center
			continue
		}
		ys[sg] = it.y + float64(i+1)*it.h/float64(len(segs)+1)
	}
}

// corridorWidth assigns one lane per segment leaving layer k, nudges entries that
// would run on top of an exit, and returns the width the lanes and labels need.
func (l *accessLayout) corridorWidth(k int) float64 {
	var segs []*segment
	for _, it := range l.layers[k] {
		segs = append(segs, it.out...)
	}
	sort.SliceStable(segs, func(i, j int) bool { return segs[i].to.inY[segs[i]] < segs[j].to.inY[segs[j]] })
	exits := map[float64]bool{}
	for _, sg := range segs {
		exits[sg.from.outY[sg]] = true
	}
	labelW := accessMinLabelW
	for i, sg := range segs {
		sg.lane = accessLanePad + float64(i)*accessLaneGap
		for range 4 {
			y, ok := collides(sg.to.inY[sg], sg.from.outY[sg], exits)
			if !ok {
				break
			}
			nudge(sg, y)
		}
		if sg.last {
			if lb := accessLabel(l.hops[sg.hop], l.elements); lb != "" {
				labelW = math.Max(labelW, house.TextWidth(house.PlainText(lb), house.EdgeFontSize)+2*house.LabelPadX)
			}
		}
	}
	lanes := accessLanePad + float64(max(len(segs)-1, 0))*accessLaneGap
	return lanes + 2*accessLabelPad + labelW
}

// collides returns an exit of another segment that runs within 6 px of entry, the
// distance under which the collinear rule reports two parallel runs.
func collides(entry, ownExit float64, exits map[float64]bool) (float64, bool) {
	for y := range exits {
		if y != ownExit && math.Abs(entry-y) < 6 {
			return y, true
		}
	}
	return 0, false
}

// nudge moves an entry off the exit it would run on, below it or above it, whichever
// keeps it clear of the item's other entries so their labels do not touch. A
// pass-through slot moves as a whole, so the edge stays straight through its layer.
func nudge(sg *segment, exit float64) {
	y := exit + accessNudge
	if sg.to.id != "" {
		for _, cand := range []float64{exit + accessNudge, exit - accessNudge} {
			if cand > sg.to.y+accessNudge && cand < sg.to.y+sg.to.h-accessNudge && clearOfEntries(sg, cand) {
				y = cand
				break
			}
		}
		sg.to.inY[sg] = y
		return
	}
	sg.to.center = y
	sg.to.inY[sg] = y
	for _, next := range sg.to.out {
		sg.to.outY[next] = y
	}
}

// clearOfEntries is true when y keeps a label's height plus a margin from every other
// entry of the same item.
func clearOfEntries(sg *segment, y float64) bool {
	for other, oy := range sg.to.inY {
		if other != sg && math.Abs(oy-y) < accessPortGap-accessNudge+2 {
			return false
		}
	}
	return true
}

func accessNodeSize(e model.Element, badge string) (w, h float64) {
	lines := []string{}
	if e.IsZone() {
		lines = append(lines, "[zone "+e.Sub+"]")
	} else if e.Tech != "" {
		lines = append(lines, e.Tech)
	}
	if e.Desc != "" {
		lines = append(lines, e.Desc)
	}
	title := e.Title
	if badge != "" {
		title += " ⚠ " + badge
	}
	text := house.TextWidth(title, house.TitleFontSize)
	for _, ln := range lines {
		text = math.Max(text, house.TextWidth(house.PlainText(ln), house.SubFontSize))
	}
	pad := 10.0 + 8
	if e.Icon != "" {
		pad = house.TextPadLeft + 8
	}
	w = math.Max(accessNodeMinW, math.Ceil(text+pad+4))
	h = math.Max(accessNodeMinH, 10+house.TitleFontSize*house.LineHeightEm+float64(len(lines))*house.SubFontSize*house.LineHeightEm+10)
	if e.Shape == "cylinder" {
		h += 16
	}
	return w, h
}

func (l *accessLayout) nodes(badges map[string][]string) []diagram.Node {
	var out []diagram.Node
	for k, layer := range l.layers {
		for _, it := range layer {
			if it.id == "" {
				continue
			}
			e := l.elements[it.id]
			status, target := l.s.Status(e.ID)
			kind, tech := e.Kind, e.Tech
			if e.IsZone() {
				// A grant may target a zone, such as a tenant namespace; in an access view
				// it is a box like any resource, in the neutral palette.
				kind, tech = "external", "["+strings.TrimSpace("zone "+e.Sub)+"]"
			}
			shape := e.Shape
			if shape == "actor" {
				// A box keeps every port on a border; an actor's label hangs below it.
				shape = ""
			}
			out = append(out, diagram.Node{
				ID: e.ID, Kind: kind, Shape: shape, Icon: e.Icon, Title: e.Title, Tech: tech, Desc: e.Desc,
				Badge: strings.Join(badges[e.ID], " "), Status: status, Target: target,
				X: l.x[k], Y: it.y, W: l.w[k], H: it.h,
			})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (l *accessLayout) nodeW(id string) float64 { return l.w[l.layerOf[id]] }

func (l *accessLayout) edges() []diagram.Edge {
	var out []diagram.Edge
	for i, h := range l.hops {
		chain := l.chains[i]
		first := chain[0]
		src := l.byID[h.From]
		path := []diagram.Point{{l.x[src.layer] + l.nodeW(h.From), src.outY[first]}}
		var labelAt *diagram.Point
		for _, sg := range chain {
			k := sg.from.layer
			laneX := l.corridor[k] + sg.lane
			y0, y1 := path[len(path)-1].Y(), sg.to.inY[sg]
			path = append(path, diagram.Point{laneX, y0}, diagram.Point{laneX, y1})
			end := l.x[k+1]
			if sg.to.id == "" {
				end += l.w[k+1]
			}
			path = append(path, diagram.Point{end, y1})
			if sg.last {
				labelAt = &diagram.Point{(lanesEndOf(l, k) + end) / 2, y1}
			}
		}
		e := diagram.Edge{
			ID: fmt.Sprintf("a-%02d-%s-%s", i+1, h.From, h.To), From: h.From, To: h.To,
			Kind: accessKind(h), Label: accessLabel(h, l.elements), Status: h.Status,
			Path: simplify(path),
		}
		if e.Label != "" {
			e.LabelAt = labelAt
		}
		out = append(out, e)
	}
	return out
}

func (l *accessLayout) segmentsLeaving(k int) []*segment {
	var segs []*segment
	for _, it := range l.layers[k] {
		segs = append(segs, it.out...)
	}
	return segs
}

// lanesEndOf is the x after the last lane of corridor k, where labels may start.
func lanesEndOf(l *accessLayout, k int) float64 {
	n := len(l.segmentsLeaving(k))
	return l.corridor[k] + accessLanePad + float64(max(n-1, 0))*accessLaneGap + accessLabelPad
}

// simplify drops repeated points and points in the middle of a straight run.
func simplify(p []diagram.Point) []diagram.Point {
	var out []diagram.Point
	for _, pt := range p {
		if n := len(out); n > 0 && out[n-1] == pt {
			continue
		}
		if n := len(out); n >= 2 {
			a, b := out[n-2], out[n-1]
			if (a.X() == b.X() && b.X() == pt.X()) || (a.Y() == b.Y() && b.Y() == pt.Y()) {
				out[n-1] = pt
				continue
			}
		}
		out = append(out, pt)
	}
	return out
}

func accessKind(h access.Hop) string {
	switch h.Kind {
	case "member", "escalation":
		return h.Kind
	}
	return "grant-" + h.Level
}

// accessLabel is the text a hop carries: auth on the hop where an identity proves
// itself, level and scope where the grant lands, the status where it changes.
func accessLabel(h access.Hop, elements map[string]model.Element) string {
	var parts []string
	switch {
	case h.Kind == "escalation":
		parts = append(parts, "!!escalation via "+titleOf(elements, h.Through)+"!!")
	case h.Kind == "member":
		parts = append(parts, "member")
	case h.Leg == "grant" || h.Leg == "direct":
		parts = append(parts, h.Level)
		if h.Scope != "" {
			parts = append(parts, h.Scope)
		}
	}
	if h.Leg != "grant" {
		if a := authText(h.Auth); a != "" {
			parts = append(parts, a)
		}
	}
	switch h.Status {
	case "deprecated":
		parts = append(parts, "removed")
	case "planned":
		parts = append(parts, house.PillText(h.Status, h.Target))
	}
	return strings.Join(parts, " · ")
}

// authText flags a credential that never expires: the first thing a reviewer looks for.
func authText(a *model.Auth) string {
	if a == nil {
		return ""
	}
	t := a.Method
	if a.MFA {
		t += " + MFA"
	}
	if a.Lifetime != "" {
		return t + " " + a.Lifetime
	}
	switch a.Method {
	case "static-token", "access-key", "password":
		return "!!" + t + ", no expiry!!"
	}
	return t
}

func titleOf(elements map[string]model.Element, id string) string {
	if t := elements[id].Title; t != "" {
		return t
	}
	return id
}

// accessCards lays out the legend, the reach summary and the escalation paths in one
// row under the graph.
func accessCards(s *model.System, g *access.Graph, state, subject, resource string, spec *diagram.Spec, y float64) []diagram.Card {
	elements := map[string]model.Element{}
	for _, e := range s.Model.Elements {
		elements[e.ID] = e
	}
	reach := reachLines(s, state, subject, resource, elements)
	if state == access.Diff {
		// A diff shows where the migration leads, so its escalations are the target's.
		g = access.New(s.Model, access.Target)
	}
	esc := escalationLines(s, g, subject, resource, elements)
	if len(esc) == 0 {
		esc = []string{"None in this state."}
	}
	w := math.Max(accessCardW, (spec.Width-2*landscapeMargin-2*accessCardGap)/3)
	if 3*w+2*accessCardGap+2*landscapeMargin > spec.Width {
		spec.Width = 3*w + 2*accessCardGap + 2*landscapeMargin
	}
	nodeKinds, edgeKinds := map[string]bool{}, map[string]bool{}
	for _, n := range spec.Nodes {
		nodeKinds[n.Kind] = true
	}
	statuses := map[string]bool{}
	for _, e := range spec.Edges {
		edgeKinds[e.Kind] = true
		if e.Status != "" {
			statuses[e.Status] = true
		}
	}
	legendH := 52 + 26*float64(max(len(nodeKinds), len(edgeKinds)+len(statuses))) + 24
	textH := func(lines []string) float64 {
		perLine := (w - 36) / (11 * house.CharWidthEm)
		n := 0
		for _, ln := range lines {
			n += int(math.Ceil(float64(len([]rune(house.PlainText(ln)))+2) / perLine))
		}
		return 60 + float64(n)*accessCardLineH
	}
	h := math.Max(legendH, math.Max(textH(reach), textH(esc)))
	x := landscapeMargin
	var cards []diagram.Card
	for _, c := range []diagram.Card{
		{ID: "card-legend", Title: "Legend", Color: "slate", Legend: true},
		{ID: "card-reach", Title: "Reach", Color: "sky", Lines: reach},
		{ID: "card-escalation", Title: "Escalation paths", Color: "orange", Lines: esc},
	} {
		c.X, c.Y, c.W, c.H = x, y, w, h
		cards = append(cards, c)
		x += w + accessCardGap
	}
	return cards
}

func reachLines(s *model.System, state, subject, resource string, elements map[string]model.Element) []string {
	summary := func(levels map[string]string) string {
		count := map[string]int{}
		for _, lv := range levels {
			count[lv]++
		}
		var parts []string
		for _, lv := range []string{"admin", "write", "read", "breakglass"} {
			if count[lv] > 0 {
				parts = append(parts, fmt.Sprintf("%d %s", count[lv], lv))
			}
		}
		if len(parts) == 0 {
			return "nothing"
		}
		return fmt.Sprintf("%d resources: %s", len(levels), strings.Join(parts, ", "))
	}
	if subject != "" {
		who := "**" + titleOf(elements, subject) + "**"
		if state == access.Diff {
			return []string{
				who + " today reaches " + summary(access.New(s.Model, access.Current).Levels(subject)),
				who + " in the target reaches " + summary(access.New(s.Model, access.Target).Levels(subject)),
			}
		}
		return []string{who + " reaches " + summary(access.New(s.Model, state).Levels(subject))}
	}
	if resource != "" {
		g := access.New(s.Model, state)
		_, subjects := g.Reachers(resource)
		var lines []string
		for _, id := range subjects {
			if lv := g.Levels(id)[resource]; lv != "" {
				lines = append(lines, fmt.Sprintf("**%s** · %s", titleOf(elements, id), lv))
			}
		}
		return lines
	}
	return []string{fmt.Sprintf("%d grants, %d memberships in the model", len(s.Model.Grants), len(s.Model.Memberships))}
}

func escalationLines(s *model.System, g *access.Graph, subject, resource string, elements map[string]model.Element) []string {
	var subjects []string
	if subject != "" {
		subjects = []string{subject}
	} else {
		seen := map[string]bool{}
		for _, gr := range s.Model.Grants {
			seen[gr.Subject] = true
		}
		for _, ms := range s.Model.Memberships {
			seen[ms.Subject] = true
		}
		subjects = sortedKeys(seen)
	}
	var lines []string
	for _, id := range subjects {
		for _, e := range g.Escalations(id) {
			if resource != "" && e.Grant.Resource != resource {
				continue
			}
			lines = append(lines, fmt.Sprintf("**%s** can change **%s**, which gives %s on **%s**",
				titleOf(elements, id), titleOf(elements, e.Through), e.Grant.Level, titleOf(elements, e.Grant.Resource)))
		}
	}
	return lines
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
