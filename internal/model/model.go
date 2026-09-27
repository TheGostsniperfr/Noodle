// Package model holds the v1alpha1 contracts of ADR-0007: what exists (Model), what is
// shown (View) and where it is drawn (Layout), and loads them from a system directory.
package model

import (
	"fmt"
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
	Multiplicity string   `yaml:"multiplicity"`
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
	Denied   bool   `yaml:"denied"`
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
	Steps        []string          `yaml:"steps"`
	Background   []string          `yaml:"background"`
	Labels       map[string]string `yaml:"labels"`
	Participants []string          `yaml:"participants"`
	Cards        []Card            `yaml:"cards"`
	Notes        []Note            `yaml:"notes"`
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
