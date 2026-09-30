package migrate

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/TheGostsniperfr/Noodle/internal/model"
	"gopkg.in/yaml.v3"
)

// The model's types have no omitempty, so the files are written through these mirrors:
// one flow mapping per line with only the fields that are set, as the examples are.
type (
	element struct {
		ID     string       `yaml:"id"`
		Kind   string       `yaml:"kind"`
		Parent string       `yaml:"parent,omitempty"`
		Shape  string       `yaml:"shape,omitempty"`
		Color  string       `yaml:"color,omitempty"`
		Icon   string       `yaml:"icon,omitempty"`
		Title  string       `yaml:"title,omitempty"`
		Sub    string       `yaml:"sub,omitempty"`
		Tech   string       `yaml:"tech,omitempty"`
		Desc   string       `yaml:"desc,omitempty"`
		Ports  []model.Port `yaml:"ports,omitempty"`
		Status string       `yaml:"status,omitempty"`
		Target string       `yaml:"target,omitempty"`
	}
	connection struct {
		ID       string `yaml:"id"`
		From     string `yaml:"from"`
		To       string `yaml:"to"`
		Kind     string `yaml:"kind"`
		Port     string `yaml:"port,omitempty"`
		Protocol string `yaml:"protocol,omitempty"`
		Verb     string `yaml:"verb,omitempty"`
		Denied   bool   `yaml:"denied,omitempty"`
	}
	annotation struct {
		ID      string   `yaml:"id"`
		Targets []string `yaml:"targets"`
	}
	route struct {
		From        string        `yaml:"from"`
		To          string        `yaml:"to"`
		Waypoints   []model.Point `yaml:"waypoints,omitempty"`
		LabelAt     *model.Point  `yaml:"label_at,omitempty"`
		LabelOffset *model.Point  `yaml:"label_offset,omitempty"`
		AgainstFlow bool          `yaml:"against_flow,omitempty"`
	}
)

// Write stores the system as model.yaml, views/<id>.yaml and layouts/<id>.yaml under dir.
// It refuses to overwrite an existing model.
func Write(s *model.System, dir string) error {
	if _, err := os.Stat(filepath.Join(dir, "model.yaml")); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("migrate: %s already holds a model.yaml, or cannot be read", dir)
	}
	for _, sub := range []string{"views", "layouts"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	e := &enc{}
	files := map[string]*yaml.Node{"model.yaml": e.modelDoc(s.Model)}
	for id, v := range s.Views {
		files[filepath.Join("views", id+".yaml")] = e.viewDoc(v)
	}
	for id, l := range s.Layouts {
		files[filepath.Join("layouts", id+".yaml")] = e.layoutDoc(s.Model, l)
	}
	if e.err != nil {
		return e.err
	}
	for name, doc := range files {
		var buf bytes.Buffer
		yenc := yaml.NewEncoder(&buf)
		yenc.SetIndent(2)
		if err := yenc.Encode(doc); err != nil {
			return fmt.Errorf("migrate: %s: %w", name, err)
		}
		// yaml.v3 quotes the key y as a YAML 1.1 boolean; noodle reads it as a string.
		raw := bytes.ReplaceAll(buf.Bytes(), []byte(`, "y": `), []byte(", y: "))
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0o644); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

func (e *enc) modelDoc(m *model.Model) *yaml.Node {
	doc := e.header("Model")
	var elements, conns, refs, anns []any
	for _, el := range m.Elements {
		elements = append(elements, element{ID: el.ID, Kind: el.Kind, Parent: el.Parent, Shape: el.Shape, Color: el.Color,
			Icon: el.Icon, Title: el.Title, Sub: el.Sub, Tech: el.Tech, Desc: el.Desc, Ports: el.Ports, Status: el.Status, Target: el.Target})
	}
	for _, c := range m.Connections {
		conns = append(conns, connection{ID: c.ID, From: c.From, To: c.To, Kind: c.Kind, Port: c.Port,
			Protocol: c.Protocol, Verb: c.Verb, Denied: c.Denied})
	}
	for _, r := range m.References {
		refs = append(refs, r)
	}
	for _, a := range m.Annotations {
		anns = append(anns, annotation{ID: a.ID, Targets: a.Targets})
	}
	e.addFlowList(doc, "elements", elements)
	e.addFlowList(doc, "connections", conns)
	e.addFlowList(doc, "references", refs)
	e.addFlowList(doc, "annotations", anns)
	return doc
}

func (e *enc) viewDoc(v *model.View) *yaml.Node {
	doc := e.header("View")
	e.add(doc, "type", v.Type)
	e.add(doc, "title", v.Title)
	if v.Subtitle != "" {
		e.add(doc, "subtitle", v.Subtitle)
	}
	if len(v.Meta) > 0 {
		e.add(doc, "meta", v.Meta)
	}
	if len(v.Steps) > 0 {
		steps := make([]string, len(v.Steps))
		for i, st := range v.Steps {
			steps[i] = st.Connection
		}
		e.add(doc, "steps", steps)
	}
	if len(v.Background) > 0 {
		e.add(doc, "background", v.Background)
	}
	if len(v.Labels) > 0 {
		e.add(doc, "labels", v.Labels)
	}
	if len(v.Notes) > 0 {
		e.add(doc, "notes", v.Notes)
	}
	if len(v.Cards) > 0 {
		e.add(doc, "cards", v.Cards)
	}
	return doc
}

// layoutDoc lists element boxes in model order, so a zone comes before what it holds.
func (e *enc) layoutDoc(m *model.Model, l *model.Layout) *yaml.Node {
	doc := e.header("Layout")
	e.add(doc, "canvas", e.flow(l.Canvas))
	elements := &yaml.Node{Kind: yaml.MappingNode}
	for _, el := range m.Elements {
		if b, ok := l.Elements[el.ID]; ok {
			addNode(elements, el.ID, e.flow(b))
		}
	}
	addNode(doc, "elements", elements)
	edges := &yaml.Node{Kind: yaml.MappingNode}
	for _, id := range edgeOrder(m, l) {
		r := l.Edges[id]
		out := route{From: r.From, To: r.To, LabelAt: r.LabelAt, AgainstFlow: r.AgainstFlow}
		for _, w := range r.Waypoints {
			out.Waypoints = append(out.Waypoints, *w.Point)
		}
		if r.LabelOffset != (model.Point{}) {
			p := r.LabelOffset
			out.LabelOffset = &p
		}
		addNode(edges, id, e.flow(out))
	}
	addNode(doc, "edges", edges)
	for name, boxes := range map[string]map[string]model.Box{"notes": l.Notes, "cards": l.Cards} {
		if len(boxes) == 0 {
			continue
		}
		n := &yaml.Node{Kind: yaml.MappingNode}
		for _, id := range sortedKeys(boxes) {
			addNode(n, id, e.flow(boxes[id]))
		}
		addNode(doc, name, n)
	}
	return doc
}

func edgeOrder(m *model.Model, l *model.Layout) []string {
	var ids []string
	for _, c := range m.Connections {
		if _, ok := l.Edges[c.ID]; ok {
			ids = append(ids, c.ID)
		}
	}
	for _, r := range m.References {
		if _, ok := l.Edges[r.ID]; ok {
			ids = append(ids, r.ID)
		}
	}
	return ids
}

func (e *enc) header(kind string) *yaml.Node {
	doc := &yaml.Node{Kind: yaml.MappingNode}
	e.add(doc, "apiVersion", model.APIVersion)
	e.add(doc, "kind", kind)
	return doc
}

// enc builds YAML nodes and keeps the first encoding error, so the document builders
// read as plain sequences of fields.
type enc struct{ err error }

func (e *enc) add(n *yaml.Node, key string, v any) {
	if node, ok := v.(*yaml.Node); ok {
		addNode(n, key, node)
		return
	}
	var val yaml.Node
	if err := val.Encode(v); err != nil && e.err == nil {
		e.err = fmt.Errorf("migrate: encode %s: %w", key, err)
	}
	addNode(n, key, &val)
}

func addNode(n *yaml.Node, key string, val *yaml.Node) {
	n.Content = append(n.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, val)
}

func (e *enc) addFlowList(doc *yaml.Node, key string, items []any) {
	if len(items) == 0 {
		return
	}
	seq := &yaml.Node{Kind: yaml.SequenceNode}
	for _, it := range items {
		seq.Content = append(seq.Content, e.flow(it))
	}
	addNode(doc, key, seq)
}

func (e *enc) flow(v any) *yaml.Node {
	var n yaml.Node
	if err := n.Encode(v); err != nil && e.err == nil {
		e.err = fmt.Errorf("migrate: encode %T: %w", v, err)
	}
	n.Style = yaml.FlowStyle
	return &n
}

func sortedKeys(m map[string]model.Box) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
