package resolve

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

const (
	seqTop       = 200.0 // below the header, as topologies place their first zone
	seqMargin    = 40.0
	seqBoxH      = 56.0
	seqMinBoxW   = 180.0
	seqGap       = 60.0 // between two participant boxes
	seqLabelPad  = 40.0 // around a message label, inside the span it crosses
	seqFirstRow  = 60.0 // from the boxes to the first step
	seqRow       = 48.0
	seqNoteH     = 28.0
	seqTail      = 40.0 // lifeline below the last step
	seqNoteExtra = 24.0
)

// Sequence lays out a sequence view: one column per participant, one row per step.
// Columns widen until every message label fits the span it crosses.
func Sequence(s *model.System, viewID string) (*diagram.Sequence, error) {
	v, ok := s.Views[viewID]
	if !ok {
		return nil, fmt.Errorf("resolve: no view %q", viewID)
	}
	if v.Type != "sequence" {
		return nil, fmt.Errorf("resolve: view %q is a %s view", viewID, v.Type)
	}
	elements := map[string]model.Element{}
	for _, e := range s.Model.Elements {
		elements[e.ID] = e
	}
	conns := map[string]model.Connection{}
	for _, c := range s.Model.Connections {
		conns[c.ID] = c
	}
	annotations := (&resolver{s: s}).badges()
	warn := func(over string) string {
		if g := annotations[over]; len(g) > 0 {
			return " · !!⚠ " + strings.Join(g, " ") + "!!"
		}
		return ""
	}

	order := append([]string(nil), v.Participants...)
	col := map[string]int{}
	for i, p := range order {
		col[p] = i
	}
	see := func(id string) {
		if _, ok := col[id]; !ok {
			col[id] = len(order)
			order = append(order, id)
		}
	}
	for _, st := range v.Steps {
		switch {
		case st.IsMessage():
			see(st.From)
			see(st.To)
		case st.Note != "":
			see(st.Note)
		}
	}

	type msg struct {
		step            model.Step
		from, to, label string
		kind            string
		reply           bool
	}
	byID := map[string]model.Step{}
	var rows []msg
	n := 0
	for _, st := range v.Steps {
		switch {
		case st.Note != "":
			rows = append(rows, msg{step: st, from: st.Note, label: st.Text})
		case st.Reply != "":
			req := byID[st.Reply]
			n++
			rows = append(rows, msg{step: st, from: req.To, to: req.From, kind: conns[req.Over].Kind, reply: true,
				label: strings.TrimSpace("[" + strconv.Itoa(n) + "] " + st.Text)})
		default:
			byID[st.ID] = st
			c := conns[st.Over]
			text := st.Text
			if text == "" {
				text = joinNonEmpty(" · ", c.Verb, c.Protocol)
			}
			n++
			rows = append(rows, msg{step: st, from: st.From, to: st.To, kind: c.Kind, label: "[" + strconv.Itoa(n) + "] " + text + warn(st.Over)})
		}
	}

	widths := make([]float64, len(order))
	for i, id := range order {
		widths[i] = boxWidth(elements[id], annotations[id])
	}
	// gaps[i] is the free space between box i and box i+1.
	gaps := make([]float64, max(len(order)-1, 0))
	for i := range gaps {
		gaps[i] = seqGap
	}
	centreDistance := func(a, b int) float64 {
		d := 0.0
		for i := a; i < b; i++ {
			d += widths[i]/2 + gaps[i] + widths[i+1]/2
		}
		return d
	}
	for _, m := range rows {
		if m.to == "" {
			continue
		}
		a, b := col[m.from], col[m.to]
		if a > b {
			a, b = b, a
		}
		need := labelWidth(m.label) + seqLabelPad
		if short := need - centreDistance(a, b); short > 0 {
			for i := a; i < b; i++ {
				gaps[i] += math.Ceil(short / float64(b-a))
			}
		}
	}

	out := &diagram.Sequence{}
	x := seqMargin
	centre := make([]float64, len(order))
	for i, id := range order {
		e := elements[id]
		centre[i] = x + widths[i]/2
		out.Frame.Nodes = append(out.Frame.Nodes, diagram.Node{
			ID: e.ID, Kind: e.Kind, Icon: e.Icon, Title: e.Title, Tech: e.Tech, Badge: strings.Join(annotations[e.ID], " "),
			X: x, Y: seqTop, W: widths[i], H: seqBoxH,
		})
		x += widths[i]
		if i < len(gaps) {
			x += gaps[i]
		}
	}

	y := seqTop + seqBoxH + seqFirstRow
	for _, m := range rows {
		if m.to == "" {
			w := labelWidth(m.label) + seqNoteExtra
			c := centre[col[m.from]]
			out.Notes = append(out.Notes, diagram.Note{ID: fmt.Sprintf("note-%d", len(out.Notes)+1), Text: m.label,
				X: c - w/2, Y: y - seqNoteH/2, W: w, H: seqNoteH})
		} else {
			id := m.step.ID
			if id == "" {
				id = fmt.Sprintf("reply-%s-%d", m.step.Reply, len(out.Messages)+1)
			}
			out.Messages = append(out.Messages, diagram.Message{ID: "msg-" + id, Label: m.label, Kind: m.kind, Reply: m.reply,
				FromX: centre[col[m.from]], ToX: centre[col[m.to]], Y: y})
		}
		y += seqRow
	}
	bottom := y - seqRow + seqTail
	for i, id := range order {
		out.Lifelines = append(out.Lifelines, diagram.Lifeline{ID: "life-" + id, X: centre[i], Top: seqTop + seqBoxH, Bottom: bottom})
	}
	out.Frame.ID, out.Frame.Title, out.Frame.Subtitle, out.Frame.Meta = v.ID, v.Title, v.Subtitle, v.Meta
	out.Frame.Width = math.Max(x+seqMargin, 900)
	out.Frame.Height = bottom + seqMargin
	return out, nil
}

// boxWidth fits a participant's title, gap badge and tech on one line each, icon
// included, measured as the node text lint measures them.
func boxWidth(e model.Element, badges []string) float64 {
	pad := 10.0 + 8
	if e.Icon != "" {
		pad = house.TextPadLeft + 8
	}
	title := e.Title
	if len(badges) > 0 {
		title += " ⚠ " + strings.Join(badges, " ")
	}
	w := math.Max(house.TextWidth(title, house.TitleFontSize), house.TextWidth(e.Tech, house.SubFontSize)) + pad
	return math.Max(seqMinBoxW, math.Ceil(w/10)*10)
}

func labelWidth(label string) float64 {
	return house.TextWidth(house.PlainText(label), house.EdgeFontSize) + 2*house.LabelPadX
}
