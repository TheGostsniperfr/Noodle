package main

import (
	"fmt"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
)

// renderSequence draws participants as topology boxes, dashed lifelines under them,
// one arrow per message, and notes on the lifeline of the participant that acts.
func (r *renderer) renderSequence(seq *diagram.Sequence) (string, error) {
	r.spec = &seq.Frame
	r.open()
	r.background()
	r.header()
	for _, l := range seq.Lifelines {
		r.freeEdge(l.ID, "", style("html=1", "endArrow=none", "strokeColor="+r.th.Muted, "strokeWidth=1", "dashed=1", "dashPattern=4 4"),
			diagram.Point{l.X, l.Top}, diagram.Point{l.X, l.Bottom}, 0)
	}
	for _, m := range seq.Messages {
		k := r.th.Edges[m.Kind]
		parts := []string{"html=1", "strokeColor=" + k.Stroke, fmt.Sprintf("strokeWidth=%g", k.Width), "endArrow=block", "endFill=1", "endSize=7",
			"labelBackgroundColor=" + r.th.Background, font(house.EdgeFontSize, k.LabelColor)}
		if m.Reply {
			parts = append(parts, "dashed=1", "dashPattern=6 4", "endArrow=open", "endFill=0")
		}
		r.freeEdge(m.ID, r.markup(m.Label), style(parts...), diagram.Point{m.FromX, m.Y}, diagram.Point{m.ToX, m.Y}, -12)
	}
	for _, n := range seq.Frame.Nodes {
		if err := r.node(n); err != nil {
			return "", err
		}
	}
	for _, n := range seq.Notes {
		r.vertex(n.ID, "1", r.markup(n.Text), style("rounded=1", "absoluteArcSize=1", "arcSize=8", "html=1", "fillColor="+r.th.CardFill,
			"strokeColor="+r.th.Muted, "strokeWidth=1", "align=center", "verticalAlign=middle", font(house.EdgeFontSize, r.th.Text)), n.X, n.Y, n.W, n.H)
	}
	return r.close(), nil
}

// freeEdge is an edge between two points rather than two cells; labelDY lifts its label.
func (r *renderer) freeEdge(id, label, st string, from, to diagram.Point, labelDY float64) {
	fmt.Fprintf(&r.b, `        <mxCell id="%s" value="%s" style="%s" edge="1" parent="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="%g" y="%g" as="sourcePoint" />
            <mxPoint x="%g" y="%g" as="targetPoint" />
`, id, xmlAttr(noLigatures(label)), xmlAttr(st), from.X(), from.Y(), to.X(), to.Y())
	if labelDY != 0 {
		fmt.Fprintf(&r.b, "            <mxPoint x=\"0\" y=\"%g\" as=\"offset\" />\n", labelDY)
	}
	r.b.WriteString("          </mxGeometry>\n        </mxCell>\n")
}
