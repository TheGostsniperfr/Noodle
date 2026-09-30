// Package migrate turns a v0 single-file spec into a v1alpha1 system (ADR-0007). It is
// the inverse of resolve.Topology, and Verify proves it by resolving the result back.
package migrate

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/model"
	"github.com/TheGostsniperfr/Noodle/internal/resolve"
)

var (
	badgeSuffix = regexp.MustCompile(`^(.*) · !!⚠ (.+)!!$`)
	stepPrefix  = regexp.MustCompile(`^\[(\d+|[A-Z])\] (.*)$`)
	portText    = regexp.MustCompile(`^(\S+) (\d+)$`)
)

// FromV0 builds a system with one topology view named after the spec's id. Zones and
// nodes get the smallest zone that contains them as parent.
func FromV0(spec *diagram.Spec) (*model.System, error) {
	if spec.ID == "" {
		return nil, fmt.Errorf("migrate: the spec has no id, which names the view")
	}
	if len(spec.Arrows) > 0 {
		return nil, fmt.Errorf("migrate: arrows belong to landscape views, not to a v0 spec")
	}
	id := spec.ID
	m := &model.Model{APIVersion: model.APIVersion, Kind: "Model"}
	v := &model.View{APIVersion: model.APIVersion, Kind: "View", ID: id, Type: "topology",
		Title: spec.Title, Subtitle: spec.Subtitle, Meta: spec.Meta}
	l := &model.Layout{APIVersion: model.APIVersion, Kind: "Layout", ID: id,
		Canvas:   model.Canvas{Width: spec.Width, Height: spec.Height},
		Elements: map[string]model.Box{}, Edges: map[string]model.EdgeRoute{}}
	s := &model.System{Model: m, Views: map[string]*model.View{id: v}, Layouts: map[string]*model.Layout{id: l}}

	parents := parentsOf(spec)
	abs := map[string]diagram.Rect{}
	place := func(eid string, r diagram.Rect) {
		abs[eid] = r
		p := abs[parents[eid]]
		l.Elements[eid] = model.Box{X: r.X - p.X, Y: r.Y - p.Y, W: r.W, H: r.H}
	}
	ann := &annotations{targets: map[string][]string{}}
	for _, z := range spec.Zones {
		m.Elements = append(m.Elements, model.Element{ID: z.ID, Kind: z.Kind, Parent: parents[z.ID],
			Title: z.Label, Sub: z.Sub, Color: z.Color, Icon: z.Icon, Status: z.Status, Target: z.Target})
	}
	for _, n := range spec.Nodes {
		m.Elements = append(m.Elements, model.Element{ID: n.ID, Kind: n.Kind, Parent: parents[n.ID],
			Title: n.Title, Tech: n.Tech, Desc: n.Desc, Icon: n.Icon, Shape: n.Shape, Status: n.Status, Target: n.Target})
		ann.add(n.ID, strings.Fields(n.Badge))
	}
	// Parents are placed before children: zones enclose, so a larger zone never sits
	// inside a smaller one.
	for _, z := range byArea(spec.Zones) {
		place(z.ID, z.Rect())
	}
	for _, n := range spec.Nodes {
		place(n.ID, n.Rect())
	}

	if err := edges(spec, s, ann); err != nil {
		return nil, err
	}
	m.Annotations = ann.list()
	for _, n := range spec.Notes {
		v.Notes = append(v.Notes, model.Note{ID: n.ID, Text: n.Text})
		mapSet(&l.Notes)[n.ID] = model.Box{X: n.X, Y: n.Y, W: n.W, H: n.H}
	}
	for _, c := range spec.Cards {
		v.Cards = append(v.Cards, model.Card{ID: c.ID, Title: c.Title, Color: c.Color, Legend: c.Legend, Lines: c.Lines})
		mapSet(&l.Cards)[c.ID] = model.Box{X: c.X, Y: c.Y, W: c.W, H: c.H}
	}
	if err := routes(spec, s); err != nil {
		return nil, err
	}
	return s, nil
}

func mapSet(m *map[string]model.Box) map[string]model.Box {
	if *m == nil {
		*m = map[string]model.Box{}
	}
	return *m
}

// parentsOf gives each zone and node the smallest other zone that contains it.
func parentsOf(spec *diagram.Spec) map[string]string {
	out := map[string]string{}
	find := func(id string, r diagram.Rect) {
		best, bestArea := "", 0.0
		for _, z := range spec.Zones {
			a := z.W * z.H
			if z.ID == id || z.Rect() == r || !z.Rect().Contains(r) {
				continue
			}
			if best == "" || a < bestArea {
				best, bestArea = z.ID, a
			}
		}
		if best != "" {
			out[id] = best
		}
	}
	for _, z := range spec.Zones {
		find(z.ID, z.Rect())
	}
	for _, n := range spec.Nodes {
		find(n.ID, n.Rect())
	}
	return out
}

func byArea(zones []diagram.Zone) []diagram.Zone {
	out := append([]diagram.Zone(nil), zones...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].W*out[j].H > out[j-1].W*out[j-1].H; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// annotations collects gap badges in order of first appearance, so each target lists
// its badges in the order the spec wrote them.
type annotations struct {
	order   []string
	targets map[string][]string
}

func (a *annotations) add(target string, ids []string) {
	for _, id := range ids {
		if _, ok := a.targets[id]; !ok {
			a.order = append(a.order, id)
		}
		a.targets[id] = append(a.targets[id], target)
	}
}

func (a *annotations) list() []model.Annotation {
	var out []model.Annotation
	for _, id := range a.order {
		out = append(out, model.Annotation{ID: id, Targets: a.targets[id]})
	}
	return out
}

// edges splits each edge's label back into step, verb, protocol and gap badges. Steps
// that are not a gapless 1..n and A.. sequence stay as label overrides.
func edges(spec *diagram.Spec, s *model.System, ann *annotations) error {
	m, v := s.Model, s.Views[spec.ID]
	numbered := map[int]string{}
	lettered := map[int]string{}
	full := map[string]string{}
	ports := map[string][]model.Port{}
	for _, e := range spec.Edges {
		label := e.Label
		if g := badgeSuffix.FindStringSubmatch(label); g != nil {
			label = g[1]
			ann.add(e.ID, strings.Fields(g[2]))
		}
		full[e.ID] = label
		if st := stepPrefix.FindStringSubmatch(label); st != nil && e.Kind != "link" {
			label = st[2]
			if n, err := strconv.Atoi(st[1]); err == nil {
				numbered[n] = e.ID
			} else {
				lettered[int(st[1][0]-'A')] = e.ID
			}
		}
		if e.Kind == "link" {
			m.References = append(m.References, model.Reference{ID: e.ID, From: e.From, To: e.To, Kind: label})
			continue
		}
		c := model.Connection{ID: e.ID, From: e.From, To: e.To, Kind: e.Kind}
		if e.Kind == "blocked" {
			c.Kind, c.Denied = "flow", true
		}
		if c.Kind == "" {
			return fmt.Errorf("migrate: edge %s has no kind", e.ID)
		}
		if i := strings.LastIndex(label, " · "); i >= 0 {
			c.Verb, c.Protocol = label[:i], label[i+len(" · "):]
		} else {
			c.Verb = label
		}
		if e.Port != "" {
			p := portText.FindStringSubmatch(e.Port)
			if p == nil {
				return fmt.Errorf("migrate: edge %s: port %q, want \"PROTOCOL NUMBER\"", e.ID, e.Port)
			}
			num, _ := strconv.Atoi(p[2])
			c.Port = strings.ToLower(p[1]) + "-" + p[2]
			if !hasPort(ports[e.To], c.Port) {
				ports[e.To] = append(ports[e.To], model.Port{Name: c.Port, Protocol: p[1], Number: num})
			}
		}
		m.Connections = append(m.Connections, c)
	}
	for i := range m.Elements {
		m.Elements[i].Ports = ports[m.Elements[i].ID]
	}
	v.Steps = sequence(numbered, 1, v, full, func(id string) model.Step { return model.Step{Connection: id} })
	v.Background = sequence(lettered, 0, v, full, func(id string) string { return id })
	return nil
}

// sequence returns the edges in step order when the keys run from first without a gap;
// otherwise it keeps every label verbatim as an override and returns no steps.
func sequence[T any](steps map[int]string, first int, v *model.View, full map[string]string, as func(string) T) []T {
	var out []T
	for i := first; i < first+len(steps); i++ {
		id, ok := steps[i]
		if !ok {
			for _, id := range steps {
				setLabel(v, id, full[id])
			}
			return nil
		}
		out = append(out, as(id))
	}
	return out
}

func setLabel(v *model.View, id, text string) {
	if v.Labels == nil {
		v.Labels = map[string]string{}
	}
	v.Labels[id] = text
}

func hasPort(ps []model.Port, name string) bool {
	for _, p := range ps {
		if p.Name == name {
			return true
		}
	}
	return false
}

// routes writes each drawn path as symbolic endpoints and waypoints, preferring a bare
// side over an explicit offset whenever it resolves to the same points.
func routes(spec *diagram.Spec, s *model.System) error {
	l := s.Layouts[spec.ID]
	for _, e := range spec.Edges {
		if len(e.Path) < 2 {
			return fmt.Errorf("migrate: edge %s has no path", e.ID)
		}
		fromRect, _ := spec.AnchorRect(e.From)
		toRect, _ := spec.AnchorRect(e.To)
		first, last := e.Path[0], e.Path[len(e.Path)-1]
		var wps []model.Waypoint
		for _, p := range e.Path[1 : len(e.Path)-1] {
			pt := model.Point(p)
			wps = append(wps, model.Waypoint{Point: &pt})
		}
		base := model.EdgeRoute{Waypoints: wps, LabelOffset: model.Point(e.LabelOffset), AgainstFlow: e.AgainstFlow}
		if e.LabelAt != nil {
			p := model.Point(*e.LabelAt)
			base.LabelAt = &p
		}
		found := false
		for _, from := range endpoints(e.From, fromRect, first) {
			for _, to := range endpoints(e.To, toRect, last) {
				r := base
				r.From, r.To = from, to
				if got, err := resolve.EdgePath(s, spec.ID, e.ID, r); err == nil && reflect.DeepEqual(got, e.Path) {
					l.Edges[e.ID] = r
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return fmt.Errorf("migrate: edge %s: no route reproduces its path %v", e.ID, e.Path)
		}
	}
	return nil
}

// endpoints lists the ways to write point p on the border of r: each side p lies on,
// bare first, then with its offset from the start of the side.
func endpoints(id string, r diagram.Rect, p diagram.Point) []string {
	var out []string
	add := func(side string, onLine bool, along, start, length float64) {
		if !onLine || along < start || along > start+length {
			return
		}
		off := strconv.FormatFloat(along-start, 'f', -1, 64)
		out = append(out, id+"."+side, id+"."+side+"@"+off+"px")
	}
	add("left", p.X() == r.X, p.Y(), r.Y, r.H)
	add("right", p.X() == r.X+r.W, p.Y(), r.Y, r.H)
	add("top", p.Y() == r.Y, p.X(), r.X, r.W)
	add("bottom", p.Y() == r.Y+r.H, p.X(), r.X, r.W)
	return out
}

// Verify resolves the migrated system and compares it with the v0 spec. The only
// difference allowed is draw order: resolve draws references after connections.
func Verify(spec *diagram.Spec, s *model.System) error {
	if findings := model.Check(s); len(findings) > 0 {
		msgs := make([]string, len(findings))
		for i, f := range findings {
			msgs[i] = f.String()
		}
		return fmt.Errorf("migrate: the migrated system fails its checks:\n  %s", strings.Join(msgs, "\n  "))
	}
	got, err := resolve.Topology(s, spec.ID)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	want := *spec
	want.Edges = linksLast(spec.Edges)
	emptyLinesAsNil(&want)
	emptyLinesAsNil(got)
	if diff := firstDiff(&want, got); diff != "" {
		return fmt.Errorf("migrate: the migrated system does not draw the same diagram: %s", diff)
	}
	return nil
}

// emptyLinesAsNil equates "lines: []" with no lines: both draw an empty card.
func emptyLinesAsNil(s *diagram.Spec) {
	s.Cards = append([]diagram.Card(nil), s.Cards...)
	for i := range s.Cards {
		if len(s.Cards[i].Lines) == 0 {
			s.Cards[i].Lines = nil
		}
	}
}

func linksLast(edges []diagram.Edge) []diagram.Edge {
	var conns, links []diagram.Edge
	for _, e := range edges {
		if e.Kind == "link" {
			links = append(links, e)
		} else {
			conns = append(conns, e)
		}
	}
	return append(conns, links...)
}

func firstDiff(want, got *diagram.Spec) string {
	if reflect.DeepEqual(want, got) {
		return ""
	}
	for i := 0; i < len(want.Zones) && i < len(got.Zones); i++ {
		if !reflect.DeepEqual(want.Zones[i], got.Zones[i]) {
			return fmt.Sprintf("zone %s: want %+v, got %+v", want.Zones[i].ID, want.Zones[i], got.Zones[i])
		}
	}
	for i := 0; i < len(want.Nodes) && i < len(got.Nodes); i++ {
		if !reflect.DeepEqual(want.Nodes[i], got.Nodes[i]) {
			return fmt.Sprintf("node %s: want %+v, got %+v", want.Nodes[i].ID, want.Nodes[i], got.Nodes[i])
		}
	}
	for i := 0; i < len(want.Edges) && i < len(got.Edges); i++ {
		if !reflect.DeepEqual(want.Edges[i], got.Edges[i]) {
			return fmt.Sprintf("edge %s: want %+v, got %+v", want.Edges[i].ID, want.Edges[i], got.Edges[i])
		}
	}
	for i := 0; i < len(want.Notes) && i < len(got.Notes); i++ {
		if !reflect.DeepEqual(want.Notes[i], got.Notes[i]) {
			return fmt.Sprintf("note %s: want %+v, got %+v", want.Notes[i].ID, want.Notes[i], got.Notes[i])
		}
	}
	for i := 0; i < len(want.Cards) && i < len(got.Cards); i++ {
		if !reflect.DeepEqual(want.Cards[i], got.Cards[i]) {
			return fmt.Sprintf("card %s: want %+v, got %+v", want.Cards[i].ID, want.Cards[i], got.Cards[i])
		}
	}
	wantHead, gotHead := *want, *got
	wantHead.Zones, wantHead.Nodes, wantHead.Edges, wantHead.Notes, wantHead.Cards = nil, nil, nil, nil, nil
	gotHead.Zones, gotHead.Nodes, gotHead.Edges, gotHead.Notes, gotHead.Cards = nil, nil, nil, nil, nil
	if !reflect.DeepEqual(wantHead, gotHead) {
		return fmt.Sprintf("header: want %+v, got %+v", wantHead, gotHead)
	}
	return fmt.Sprintf("counts differ: want %d zones %d nodes %d edges, got %d %d %d",
		len(want.Zones), len(want.Nodes), len(want.Edges), len(got.Zones), len(got.Nodes), len(got.Edges))
}
