// Package house is the house style of ADR-0006, shared by the lint and the renderers:
// palettes, font metrics, and the boxes that text, labels and port badges occupy.
package house

import "fmt"

// Palettes are Tailwind CSS v3 (the one Cocoon-AI uses). Dark: 400 strokes on slate-950,
// fills pre-composited so arrows never show through a box. Light: 600 strokes, 50 fills,
// 700 for coloured text so every text pair clears WCAG AA (4.5:1).
type Theme struct {
	Name                             string
	Background, GridLine             string
	CardFill, CardStroke             string
	Title, Text, Muted, Warn, Accent string
	StepFill, StepText               string
	PortText                         string
	// Hatch fills planned boxes (ADR-0008): the same grey whatever the kind, so a
	// planned box reads as absent before its colour is read.
	Hatch string
	Nodes map[string]NodeKind
	Zones map[string]string
	Edges map[string]EdgeKind
}

type NodeKind struct{ Stroke, Fill, Legend string }

type EdgeKind struct {
	Stroke, LabelColor, Dash, EndArrow, Legend string
	Width                                      float64
}

const (
	FontFamily    = "JetBrains Mono,Noto Sans Mono,DejaVu Sans Mono,monospace"
	TitleFontSize = 12.0
	SubFontSize   = 10.0
	EdgeFontSize  = 10.0
	ZoneFontSize  = 11.0
	PortFontSize  = 9.0
	CharWidthEm   = 0.6
	LineHeightEm  = 1.25

	IconSize      = 28.0
	IconInset     = 12.0
	TextPadLeft   = IconInset + IconSize + 10
	ZoneIconSize  = 18.0
	ActorLabelGap = 6.0
	LabelPadX     = 4.0
	LabelPadY     = 2.0
	PortPadX      = 4.0
	PortHeight    = 16.0

	MinNodeGap     = 24.0
	MinZoneGap     = 60.0
	ObstacleMargin = 4.0
)

var NodeKindOrder = []string{"frontend", "backend", "database", "cloud", "security", "bus", "external", "tool"}
var EdgeKindOrder = []string{"flow", "auth", "tunnel", "async", "blocked", "link"}

var LegendNodes = map[string]string{
	"frontend": "client / frontend",
	"backend":  "service / routing",
	"database": "data / secrets",
	"cloud":    "edge / tunnel / cloud",
	"security": "security / identity",
	"bus":      "bus / messaging",
	"external": "external / out of scope",
	"tool":     "tool, no runtime traffic",
}

var LegendEdges = map[string]string{
	"flow":    "connection, client → server",
	"auth":    "authentication connection",
	"tunnel":  "outbound tunnel, opened in advance",
	"async":   "background sync",
	"blocked": "must not happen",
	"link":    "object reference, not traffic",
}

func nodes(pairs ...string) map[string]NodeKind {
	m := map[string]NodeKind{}
	for i := 0; i < len(pairs); i += 3 {
		m[pairs[i]] = NodeKind{Stroke: pairs[i+1], Fill: pairs[i+2], Legend: LegendNodes[pairs[i]]}
	}
	return m
}

func edges(pairs map[string][2]string, blockedEnd string) map[string]EdgeKind {
	dash := map[string]string{"auth": "6 4", "tunnel": "3 3", "async": "6 4", "blocked": "6 4", "link": "2 3"}
	width := map[string]float64{"blocked": 1.8, "link": 1.2}
	m := map[string]EdgeKind{}
	for k, c := range pairs {
		w := width[k]
		if w == 0 {
			w = 1.5
		}
		end := "blockThin"
		if k == "blocked" {
			end = blockedEnd
		}
		if k == "link" {
			end = "open"
		}
		m[k] = EdgeKind{Stroke: c[0], LabelColor: c[1], Dash: dash[k], EndArrow: end, Legend: LegendEdges[k], Width: w}
	}
	return m
}

var Themes = map[string]*Theme{
	"dark": {
		Name: "dark", Background: "#020617", GridLine: "#1e293b",
		CardFill: "#0b1222", CardStroke: "#1e293b",
		Title: "#f8fafc", Text: "#e2e8f0", Muted: "#94a3b8", Warn: "#fb923c", Accent: "#34d399",
		StepFill: "#f8fafc", StepText: "#020617", PortText: "#020617", Hatch: "#64748b",
		Nodes: nodes(
			"frontend", "#22d3ee", "#0c2234",
			"backend", "#34d399", "#0b2d31",
			"database", "#a78bfa", "#271955",
			"cloud", "#fbbf24", "#2f2022",
			"security", "#fb7185", "#3f152f",
			"bus", "#fb923c", "#563c2f",
			"external", "#94a3b8", "#172033",
			"tool", "#94a3b8", "#172033",
		),
		Zones: map[string]string{"cyan": "#22d3ee", "emerald": "#34d399", "violet": "#a78bfa", "amber": "#fbbf24",
			"rose": "#fb7185", "orange": "#fb923c", "slate": "#94a3b8", "indigo": "#818cf8", "sky": "#38bdf8"},
		Edges: edges(map[string][2]string{
			"flow": {"#94a3b8", "#e2e8f0"}, "auth": {"#fb7185", "#fda4af"}, "tunnel": {"#fbbf24", "#fcd34d"},
			"async": {"#a78bfa", "#c4b5fd"}, "blocked": {"#f87171", "#fca5a5"}, "link": {"#64748b", "#94a3b8"},
		}, "cross"),
	},
	"light": {
		Name: "light", Background: "#f8fafc", GridLine: "#e2e8f0",
		CardFill: "#ffffff", CardStroke: "#e2e8f0",
		Title: "#0f172a", Text: "#1e293b", Muted: "#475569", Warn: "#c2410c", Accent: "#059669",
		StepFill: "#0f172a", StepText: "#ffffff", PortText: "#ffffff", Hatch: "#94a3b8",
		Nodes: nodes(
			"frontend", "#0891b2", "#ecfeff",
			"backend", "#059669", "#ecfdf5",
			"database", "#7c3aed", "#f5f3ff",
			"cloud", "#d97706", "#fffbeb",
			"security", "#e11d48", "#fff1f2",
			"bus", "#ea580c", "#fff7ed",
			"external", "#64748b", "#f1f5f9",
			"tool", "#64748b", "#f1f5f9",
		),
		Zones: map[string]string{"cyan": "#0891b2", "emerald": "#059669", "violet": "#7c3aed", "amber": "#d97706",
			"rose": "#e11d48", "orange": "#ea580c", "slate": "#64748b", "indigo": "#4f46e5", "sky": "#0284c7"},
		Edges: edges(map[string][2]string{
			"flow": {"#64748b", "#1e293b"}, "auth": {"#e11d48", "#be123c"}, "tunnel": {"#d97706", "#b45309"},
			"async": {"#7c3aed", "#6d28d9"}, "blocked": {"#dc2626", "#b91c1c"}, "link": {"#94a3b8", "#475569"},
		}, "cross"),
	},
}

func ThemeByName(name string) (*Theme, error) {
	t, ok := Themes[name]
	if !ok {
		return nil, fmt.Errorf("unknown theme %q", name)
	}
	return t, nil
}

func TextWidth(s string, fontSize float64) float64 {
	return float64(len([]rune(s))) * fontSize * CharWidthEm
}
