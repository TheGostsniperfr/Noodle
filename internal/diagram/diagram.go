// Package diagram is the resolved geometry the renderer and the lint consume: absolute
// positions and paths, as the v0 single-file spec expressed them.
package diagram

type Point [2]float64

func (p Point) X() float64 { return p[0] }
func (p Point) Y() float64 { return p[1] }

type Rect struct{ X, Y, W, H float64 }

func (r Rect) Inflate(d float64) Rect { return Rect{r.X - d, r.Y - d, r.W + 2*d, r.H + 2*d} }

func (r Rect) Intersects(o Rect) bool {
	return r.X < o.X+o.W && o.X < r.X+r.W && r.Y < o.Y+o.H && o.Y < r.Y+r.H
}

func (r Rect) Contains(o Rect) bool {
	return o.X >= r.X && o.Y >= r.Y && o.X+o.W <= r.X+r.W && o.Y+o.H <= r.Y+r.H
}

type Spec struct {
	ID string `yaml:"id"`
	// Type is the view type it was resolved from; empty for topology and v0 files.
	Type     string   `yaml:"-"`
	Title    string   `yaml:"title"`
	Subtitle string   `yaml:"subtitle"`
	Meta     []string `yaml:"meta"`
	Width    float64  `yaml:"width"`
	Height   float64  `yaml:"height"`
	Zones    []Zone   `yaml:"zones"`
	Nodes    []Node   `yaml:"nodes"`
	Edges    []Edge   `yaml:"edges"`
	Notes    []Note   `yaml:"notes"`
	Cards    []Card   `yaml:"cards"`
	// Arrows are reading aids between groups (a landscape's flow), not connections:
	// they carry no semantics and the edge rules ignore them.
	Arrows []Arrow `yaml:"arrows"`
	// Offerings are catalog cards (ADR-0012), laid out by the catalog resolver.
	Offerings []Offering `yaml:"-"`
}

// Offering is a resolved catalog card. Text is already wrapped to the card width; each
// block's Y is absolute, so the renderer and the lint agree on where text sits.
type Offering struct {
	ID, Title, Icon string
	Status, Target  string
	Summary         Block
	Provides        []Block // one per bullet
	Request         Block
	Logos           []Logo
	LabelsY         [3]float64 // "you get", "request", "backed by" label rows; 0 when absent
	X, Y, W, H      float64
}

type Block struct {
	Lines []string
	Y     float64
}

type Logo struct {
	Icon, Title string
	X, Y        float64
}

func (o Offering) Rect() Rect { return Rect{o.X, o.Y, o.W, o.H} }

type Arrow struct {
	ID       string `yaml:"id"`
	From, To Point
}

func (s *Spec) AnchorRect(id string) (Rect, bool) {
	for _, n := range s.Nodes {
		if n.ID == id {
			return n.Rect(), true
		}
	}
	for _, z := range s.Zones {
		if z.ID == id {
			return z.Rect(), true
		}
	}
	return Rect{}, false
}

// Zone is a dashed boundary: "region" for an infra or trust perimeter, "group" for a functional category inside one.
type Zone struct {
	ID    string `yaml:"id"`
	Kind  string `yaml:"kind"`
	Label string `yaml:"label"`
	Sub   string `yaml:"sub"`
	Color string `yaml:"color"`
	Icon  string `yaml:"icon"`
	// Status and Target follow ADR-0008 and ADR-0013; a zone's apply to its children.
	Status string  `yaml:"status"`
	Target string  `yaml:"target"`
	X      float64 `yaml:"x"`
	Y      float64 `yaml:"y"`
	W      float64 `yaml:"w"`
	H      float64 `yaml:"h"`
}

func (z Zone) Rect() Rect { return Rect{z.X, z.Y, z.W, z.H} }

// Node text follows the C4 triptych: name, [technology], one-line responsibility.
type Node struct {
	ID    string `yaml:"id"`
	Kind  string `yaml:"kind"`
	Shape string `yaml:"shape"` // box (default), cylinder, actor
	Icon  string `yaml:"icon"`
	Title string `yaml:"title"`
	Tech  string `yaml:"tech"`
	Desc  string `yaml:"desc"`
	Badge string `yaml:"badge"`
	// Status and Target are already inherited from enclosing zones.
	Status string  `yaml:"status"`
	Target string  `yaml:"target"`
	X      float64 `yaml:"x"`
	Y      float64 `yaml:"y"`
	W      float64 `yaml:"w"`
	H      float64 `yaml:"h"`
}

func (n Node) Rect() Rect { return Rect{n.X, n.Y, n.W, n.H} }

func (n Node) Lines() []string {
	var l []string
	if n.Tech != "" {
		l = append(l, n.Tech)
	}
	if n.Desc != "" {
		l = append(l, n.Desc)
	}
	return l
}

// Edge is drawn from the side that opens the connection to the side that listens.
// Port names the listening port on the target and becomes a badge on its border.
type Edge struct {
	ID          string  `yaml:"id"`
	From        string  `yaml:"from"`
	To          string  `yaml:"to"`
	Kind        string  `yaml:"kind"`
	Label       string  `yaml:"label"`
	Port        string  `yaml:"port"`
	AgainstFlow bool    `yaml:"against_flow"` // outbound connection drawn against the reading direction, e.g. a tunnel
	Path        []Point `yaml:"path"`
	LabelAt     *Point  `yaml:"label_at"`
	LabelOffset Point   `yaml:"label_offset"`
}

type Note struct {
	ID   string  `yaml:"id"`
	Text string  `yaml:"text"`
	X    float64 `yaml:"x"`
	Y    float64 `yaml:"y"`
	W    float64 `yaml:"w"`
	H    float64 `yaml:"h"`
}

func (n Note) Rect() Rect { return Rect{n.X, n.Y, n.W, n.H} }

type Card struct {
	ID     string   `yaml:"id"`
	Title  string   `yaml:"title"`
	Color  string   `yaml:"color"`
	Legend bool     `yaml:"legend"`
	Lines  []string `yaml:"lines"`
	X      float64  `yaml:"x"`
	Y      float64  `yaml:"y"`
	W      float64  `yaml:"w"`
	H      float64  `yaml:"h"`
}

func (c Card) Rect() Rect { return Rect{c.X, c.Y, c.W, c.H} }
