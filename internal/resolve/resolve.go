// Package resolve turns a model, one of its views and that view's layout into the
// absolute geometry the lint and the renderers consume (plan 001).
package resolve

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// View resolves topology and landscape views. Sequence views have their own geometry,
// see Sequence.
func View(s *model.System, viewID string) (*diagram.Spec, error) {
	v, ok := s.Views[viewID]
	if !ok {
		return nil, fmt.Errorf("resolve: no view %q", viewID)
	}
	switch v.Type {
	case "landscape":
		return Landscape(s, viewID)
	case "catalog":
		return Catalog(s, viewID)
	default:
		return Topology(s, viewID)
	}
}

// Topology resolves a topology view. The system must pass model.Check first: missing
// ids here are programming errors, reported but not explained.
func Topology(s *model.System, viewID string) (*diagram.Spec, error) {
	v, ok := s.Views[viewID]
	if !ok {
		return nil, fmt.Errorf("resolve: no view %q", viewID)
	}
	if v.Type != "topology" {
		return nil, fmt.Errorf("resolve: view %q is a %s view", viewID, v.Type)
	}
	l, ok := s.Layouts[viewID]
	if !ok {
		return nil, fmt.Errorf("resolve: view %q has no layout", viewID)
	}
	r := &resolver{s: s, v: v, l: l, elements: map[string]model.Element{}, abs: map[string]diagram.Rect{}}
	for _, e := range s.Model.Elements {
		r.elements[e.ID] = e
	}
	out := &diagram.Spec{
		ID: v.ID, Title: v.Title, Subtitle: v.Subtitle, Meta: v.Meta,
		Width: l.Canvas.Width, Height: l.Canvas.Height,
	}
	shown := map[string]bool{}
	for _, id := range s.Included(v) {
		shown[id] = true
	}
	badges := r.badges()
	for _, e := range s.Model.Elements {
		if !shown[e.ID] {
			continue
		}
		box, err := r.absolute(e.ID)
		if err != nil {
			return nil, err
		}
		status, target := s.Status(e.ID)
		if e.IsZone() {
			out.Zones = append(out.Zones, diagram.Zone{
				ID: e.ID, Kind: e.Kind, Label: e.Title, Sub: e.Sub, Color: e.Color, Icon: e.Icon,
				Status: status, Target: target,
				X: box.X, Y: box.Y, W: box.W, H: box.H,
			})
			continue
		}
		out.Nodes = append(out.Nodes, diagram.Node{
			ID: e.ID, Kind: e.Kind, Shape: e.Shape, Icon: e.Icon, Title: e.Title, Tech: e.Tech,
			Desc: e.Desc, Badge: strings.Join(badges[e.ID], " "),
			Status: status, Target: target, Multiplicity: e.Multiplicity,
			X: box.X, Y: box.Y, W: box.W, H: box.H,
		})
	}
	edges, err := r.edges(shown, badges)
	if err != nil {
		return nil, err
	}
	out.Edges = edges
	for _, n := range v.Notes {
		b := l.Notes[n.ID]
		out.Notes = append(out.Notes, diagram.Note{ID: n.ID, Text: n.Text, X: b.X, Y: b.Y, W: b.W, H: b.H})
	}
	for _, c := range v.Cards {
		b := l.Cards[c.ID]
		out.Cards = append(out.Cards, diagram.Card{
			ID: c.ID, Title: c.Title, Color: c.Color, Legend: c.Legend, Lines: c.Lines,
			X: b.X, Y: b.Y, W: b.W, H: b.H,
		})
	}
	return out, nil
}

type resolver struct {
	s        *model.System
	v        *model.View
	l        *model.Layout
	elements map[string]model.Element
	abs      map[string]diagram.Rect
}

// absolute adds the offsets of every enclosing zone to an element's layout box.
func (r *resolver) absolute(id string) (diagram.Rect, error) {
	if b, ok := r.abs[id]; ok {
		return b, nil
	}
	b, ok := r.l.Elements[id]
	if !ok {
		return diagram.Rect{}, fmt.Errorf("resolve: %s has no position", id)
	}
	rect := diagram.Rect{X: b.X, Y: b.Y, W: b.W, H: b.H}
	if parent := r.elements[id].Parent; parent != "" {
		p, err := r.absolute(parent)
		if err != nil {
			return diagram.Rect{}, err
		}
		rect.X += p.X
		rect.Y += p.Y
	}
	r.abs[id] = rect
	return rect, nil
}

// badges maps each annotated element or edge to its annotation ids, in model order.
func (r *resolver) badges() map[string][]string {
	out := map[string][]string{}
	for _, a := range r.s.Model.Annotations {
		for _, t := range a.Targets {
			out[t] = append(out[t], a.ID)
		}
	}
	return out
}

func (r *resolver) edges(shown map[string]bool, badges map[string][]string) ([]diagram.Edge, error) {
	steps := map[string]string{}
	for i, st := range r.v.Steps {
		steps[st.Connection] = strconv.Itoa(i + 1)
	}
	for i, id := range r.v.Background {
		steps[id] = string(rune('A' + i))
	}
	var out []diagram.Edge
	add := func(id, from, to, kind, label, port string) error {
		if !shown[from] || !shown[to] {
			return nil
		}
		route := r.l.Edges[id]
		path, err := r.path(id, route)
		if err != nil {
			return err
		}
		if override, ok := r.v.Labels[id]; ok {
			label = override
		}
		if step, ok := steps[id]; ok {
			label = "[" + step + "] " + label
		}
		if gaps := badges[id]; len(gaps) > 0 {
			label += " · !!⚠ " + strings.Join(gaps, " ") + "!!"
		}
		e := diagram.Edge{
			ID: id, From: from, To: to, Kind: kind, Label: label, Port: port,
			AgainstFlow: route.AgainstFlow, Path: path, LabelOffset: diagram.Point(route.LabelOffset),
		}
		if route.LabelAt != nil {
			p := diagram.Point(*route.LabelAt)
			e.LabelAt = &p
		}
		out = append(out, e)
		return nil
	}
	for _, c := range r.s.Model.Connections {
		kind := c.Kind
		if c.Denied {
			kind = "blocked"
		}
		if err := add(c.ID, c.From, c.To, kind, joinNonEmpty(" · ", c.Verb, c.Protocol), r.portText(c)); err != nil {
			return nil, err
		}
	}
	for _, ref := range r.s.Model.References {
		if err := add(ref.ID, ref.From, ref.To, "link", ref.Kind, ""); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// portText is the badge text of the port a connection lands on, e.g. "TCP 443".
func (r *resolver) portText(c model.Connection) string {
	for _, p := range r.elements[c.To].Ports {
		if p.Name == c.Port {
			return fmt.Sprintf("%s %d", p.Protocol, p.Number)
		}
	}
	return ""
}

func joinNonEmpty(sep string, parts ...string) string {
	var kept []string
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, sep)
}
