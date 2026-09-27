package diagram

// Sequence is a resolved sequence view (ADR-0004, ADR-0011). Frame carries the header,
// the canvas and the participant boxes as nodes, so the renderers draw them as they
// draw a topology.
type Sequence struct {
	Frame     Spec
	Lifelines []Lifeline
	Messages  []Message
	Notes     []Note
}

type Lifeline struct {
	ID          string
	X           float64
	Top, Bottom float64
}

// Message is a horizontal arrow at Y from FromX to ToX. Kind is the connection's kind,
// so a message keeps its colour; a Reply is drawn dashed.
type Message struct {
	ID         string
	Label      string
	Kind       string
	Reply      bool
	FromX, ToX float64
	Y          float64
}
