// Package model holds the v1alpha1 contracts of ADR-0007: what exists (Model), what is
// shown (View) and where it is drawn (Layout), and loads them from a system directory.
package model

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const APIVersion = "noodle/v1alpha1"

type Model struct {
	APIVersion  string       `yaml:"apiVersion"`
	Kind        string       `yaml:"kind"`
	Elements    []Element    `yaml:"elements"`
	Connections []Connection `yaml:"connections"`
	References  []Reference  `yaml:"references"`
	Annotations []Annotation `yaml:"annotations"`
	Offerings   []Offering   `yaml:"offerings"`
}

// Offering is what a platform promises its users and how to get it (ADR-0013).
// BackedBy lists the elements that deliver it.
type Offering struct {
	ID       string   `yaml:"id"`
	Title    string   `yaml:"title"`
	Icon     string   `yaml:"icon"`
	Summary  string   `yaml:"summary"`
	Provides []string `yaml:"provides"`
	Request  string   `yaml:"request"`
	BackedBy []string `yaml:"backed_by"`
	Owner    string   `yaml:"owner"`
	Status   string   `yaml:"status"`
	Target   string   `yaml:"target"`
}

// Element is a component or a zone. Zones have kind region or group and may carry
// color and sub.
type Element struct {
	ID           string   `yaml:"id"`
	Kind         string   `yaml:"kind"`
	Parent       string   `yaml:"parent"`
	Title        string   `yaml:"title"`
	Tech         string   `yaml:"tech"`
	Desc         string   `yaml:"desc"`
	Icon         string   `yaml:"icon"`
	Shape        string   `yaml:"shape"`
	Color        string   `yaml:"color"`
	Sub          string   `yaml:"sub"`
	Ports        []Port   `yaml:"ports"`
	Tags         []string `yaml:"tags"`
	Status       string   `yaml:"status"`
	Target       string   `yaml:"target"`
	Multiplicity string   `yaml:"multiplicity"`
	// Matches lists the discovered ids this element stands for (ADR-0015).
	Matches []string `yaml:"matches"`
}

func (e Element) IsZone() bool { return e.Kind == "region" || e.Kind == "group" }

type Port struct {
	Name     string `yaml:"name"`
	Protocol string `yaml:"protocol"`
	Number   int    `yaml:"port"`
}

// Connection is opened by From and listened to by To (ADR-0003). Port names a port of To.
type Connection struct {
	ID       string `yaml:"id"`
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	Port     string `yaml:"port"`
	Protocol string `yaml:"protocol"`
	Verb     string `yaml:"verb"`
	Kind     string `yaml:"kind"`
	// Denied is the intent that this connection must not happen; EnforcedBy names the
	// elements that make it so (ADR-0010). Denied without EnforcedBy is intent only.
	Denied     bool     `yaml:"denied"`
	EnforcedBy []string `yaml:"enforced_by"`
}

type Reference struct {
	ID   string `yaml:"id"`
	From string `yaml:"from"`
	To   string `yaml:"to"`
	Kind string `yaml:"kind"`
}

type Annotation struct {
	ID       string   `yaml:"id"`
	Severity string   `yaml:"severity"`
	Title    string   `yaml:"title"`
	Text     string   `yaml:"text"`
	Targets  []string `yaml:"targets"`
}

// Fragment is what an adapter discovers from one source (ADR-0015). Its ids are
// discovered ids, never curated ones; the curated model links to them with matches.
type Fragment struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	// ID is the file stem, <adapter>-<source>.
	ID          string                 `yaml:"-"`
	Provenance  Provenance             `yaml:"provenance"`
	Elements    []DiscoveredElement    `yaml:"elements"`
	Connections []DiscoveredConnection `yaml:"connections"`
	References  []DiscoveredReference  `yaml:"references"`
}

type Provenance struct {
	Adapter    string `yaml:"adapter"`
	Source     string `yaml:"source"`
	Ref        string `yaml:"ref"`
	ObservedAt string `yaml:"observedAt"`
}

// Src is where one discovered item comes from: a file and line, or an API object.
type Src struct {
	File   string `yaml:"file"`
	Line   int    `yaml:"line"`
	Object string `yaml:"object"`
}

type DiscoveredElement struct {
	Element `yaml:",inline"`
	Src     Src `yaml:"src"`
}

// DiscoveredConnection is Inferred when a heuristic found it, such as a Service DNS
// name in an env value, rather than a manifest that declares it.
type DiscoveredConnection struct {
	Connection `yaml:",inline"`
	Inferred   bool `yaml:"inferred"`
	Src        Src  `yaml:"src"`
}

type DiscoveredReference struct {
	Reference `yaml:",inline"`
	Src       Src `yaml:"src"`
}

// discoveredID is <adapter>:<path>, e.g. k8s:prod/deployment/api (ADR-0015).
var discoveredID = regexp.MustCompile(`^[a-z][a-z0-9]*:\S+$`)

func IsDiscoveredID(id string) bool { return discoveredID.MatchString(id) }

type View struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	// ID is the file stem, not a field: renaming the file renames the view.
	ID           string            `yaml:"-"`
	Type         string            `yaml:"type"`
	Title        string            `yaml:"title"`
	Subtitle     string            `yaml:"subtitle"`
	Meta         []string          `yaml:"meta"`
	Include      []string          `yaml:"include"`
	Steps        []Step            `yaml:"steps"`
	Background   []string          `yaml:"background"`
	Labels       map[string]string `yaml:"labels"`
	Participants []string          `yaml:"participants"`
	Cards        []Card            `yaml:"cards"`
	Notes        []Note            `yaml:"notes"`
	// Landscape views (ADR-0012): rows of sections, an optional side column, and the
	// canvas width the grid wraps to.
	Bands []Band    `yaml:"bands"`
	Side  []Section `yaml:"side"`
	Width float64   `yaml:"width"`
	// Catalog views (ADR-0013): cards per row; Include lists offering ids.
	Columns int `yaml:"columns"`
}

// Band is one row of a landscape. Flow draws an arrow between consecutive sections.
type Band struct {
	ID       string    `yaml:"id"`
	Title    string    `yaml:"title"`
	Sub      string    `yaml:"sub"`
	Color    string    `yaml:"color"`
	Icon     string    `yaml:"icon"`
	Flow     bool      `yaml:"flow"`
	Sections []Section `yaml:"sections"`
}

// Section groups landscape items; Items are element ids.
type Section struct {
	Title string   `yaml:"title"`
	Color string   `yaml:"color"`
	Items []string `yaml:"items"`
}

// Step is a connection id in a topology view, and a message, reply or note in a
// sequence view (ADR-0011).
type Step struct {
	Connection string
	ID         string
	From       string
	To         string
	Over       string
	Text       string
	Reply      string
	Note       string
}

func (s Step) IsMessage() bool { return s.From != "" || s.To != "" || s.Over != "" }

var stepKeys = map[string]bool{"id": true, "from": true, "to": true, "over": true, "text": true, "reply": true, "note": true}

func (s *Step) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		s.Connection = n.Value
		return nil
	}
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: a step is a connection id or a mapping", n.Line)
	}
	fields := map[string]*string{"id": &s.ID, "from": &s.From, "to": &s.To, "over": &s.Over, "text": &s.Text, "reply": &s.Reply, "note": &s.Note}
	for i := 0; i+1 < len(n.Content); i += 2 {
		k, v := n.Content[i], n.Content[i+1]
		if !stepKeys[k.Value] {
			return fmt.Errorf("line %d: field %s not found in type model.Step", k.Line, k.Value)
		}
		if v.Kind != yaml.ScalarNode {
			return fmt.Errorf("line %d: step field %s must be a string", v.Line, k.Value)
		}
		*fields[k.Value] = v.Value
	}
	return nil
}

type Card struct {
	ID     string   `yaml:"id"`
	Title  string   `yaml:"title"`
	Color  string   `yaml:"color"`
	Legend bool     `yaml:"legend"`
	Lines  []string `yaml:"lines"`
}

type Note struct {
	ID   string `yaml:"id"`
	Text string `yaml:"text"`
}

type Layout struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	// ID is the file stem and names the view this layout draws.
	ID       string               `yaml:"-"`
	Canvas   Canvas               `yaml:"canvas"`
	Lanes    map[string]Lane      `yaml:"lanes"`
	Elements map[string]Box       `yaml:"elements"`
	Edges    map[string]EdgeRoute `yaml:"edges"`
	Cards    map[string]Box       `yaml:"cards"`
	Notes    map[string]Box       `yaml:"notes"`
}

type Canvas struct {
	Width  float64 `yaml:"width"`
	Height float64 `yaml:"height"`
}

// Lane is a named vertical (X) or horizontal (Y) line that edge routes can share.
type Lane struct {
	X *float64 `yaml:"x"`
	Y *float64 `yaml:"y"`
}

// Box is relative to the parent zone's top-left corner, or to the canvas at top level.
type Box struct {
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
	W float64 `yaml:"w"`
	H float64 `yaml:"h"`
}

type Point [2]float64

// EdgeRoute draws one connection or reference. From and To are "id.side",
// "id.side@NN%" or "id.side@NNpx". Waypoints and LabelAt are canvas coordinates.
type EdgeRoute struct {
	From        string     `yaml:"from"`
	To          string     `yaml:"to"`
	Waypoints   []Waypoint `yaml:"waypoints"`
	LabelAt     *Point     `yaml:"label_at"`
	LabelOffset Point      `yaml:"label_offset"`
	AgainstFlow bool       `yaml:"against_flow"`
}

// Waypoint is either an absolute point [x, y] or "lane:<name>".
type Waypoint struct {
	Point *Point
	Lane  string
}

func (w *Waypoint) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		lane, ok := strings.CutPrefix(n.Value, "lane:")
		if !ok || lane == "" {
			return fmt.Errorf("line %d: waypoint %q is neither [x, y] nor lane:<name>", n.Line, n.Value)
		}
		w.Lane = lane
		return nil
	}
	var p Point
	if err := n.Decode(&p); err != nil {
		return fmt.Errorf("line %d: waypoint: %w", n.Line, err)
	}
	w.Point = &p
	return nil
}

// Endpoint splits "id.side@NN%" into its element id and the rest.
func Endpoint(s string) (id, anchor string) {
	id, anchor, _ = strings.Cut(s, ".")
	return id, anchor
}
