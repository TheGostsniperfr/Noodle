package main

import (
	"encoding/base64"
	"fmt"
	"html"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
)

type renderer struct {
	spec  *diagram.Spec
	th    *house.Theme
	icons *iconSet
	cache map[string]string
	ports map[string]house.PortBadge
	b     strings.Builder
}

func (r *renderer) iconURI(name string) (string, error) {
	if uri, ok := r.cache[name]; ok {
		return uri, nil
	}
	raw, err := r.icons.read(name, r.th.Name)
	if err != nil {
		return "", err
	}
	// draw.io splits styles on ';', so its data URIs omit the ";base64" marker.
	uri := "data:image/svg+xml," + base64.StdEncoding.EncodeToString(raw)
	r.cache[name] = uri
	return uri, nil
}

func xmlAttr(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch c {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\n':
			b.WriteString("&#10;")
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}

func style(parts ...string) string {
	kept := parts[:0:0]
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, ";") + ";"
}

func font(size float64, color string) string {
	return fmt.Sprintf("fontFamily=%s;fontSize=%g;fontColor=%s", house.FontFamily, size, color)
}

func (r *renderer) stepBadge(step string) string {
	if step[0] >= 'A' && step[0] <= 'Z' {
		return fmt.Sprintf(`<span style="border:1px solid %s;color:%s;border-radius:3px;padding:0px 3px;font-weight:bold;">%s</span>`, r.th.StepFill, r.th.Title, step)
	}
	return fmt.Sprintf(`<span style="background-color:%s;color:%s;border-radius:3px;padding:0px 4px;font-weight:bold;">%s</span>`, r.th.StepFill, r.th.StepText, step)
}

// markup: [1] data-plane step (filled badge), [A] control-plane step (outlined badge), **bold**, !!warning!!.
func (r *renderer) markup(s string) string {
	parts := strings.Split(s, "<br>")
	for i, p := range parts {
		p = html.EscapeString(p)
		p = house.StepRe.ReplaceAllStringFunc(p, func(m string) string { return r.stepBadge(m[1 : len(m)-1]) })
		p = house.BoldRe.ReplaceAllString(p, `<b><font color="`+r.th.Title+`">$1</font></b>`)
		p = house.WarnRe.ReplaceAllString(p, `<b><font color="`+r.th.Warn+`">$1</font></b>`)
		parts[i] = p
	}
	return strings.Join(parts, "<br>")
}

// noLigatures keeps code-like text literal: JetBrains Mono otherwise merges pairs such as
// ">-" or "->" into one glyph, so "<app>-frontend" renders as "<app>—frontend".
func noLigatures(value string) string {
	if value == "" {
		return value
	}
	return `<span style="font-variant-ligatures:none">` + value + `</span>`
}

func (r *renderer) vertex(id, parent, value, st string, x, y, w, h float64) {
	fmt.Fprintf(&r.b, `        <mxCell id="%s" value="%s" style="%s" vertex="1" parent="%s">
          <mxGeometry x="%g" y="%g" width="%g" height="%g" as="geometry" />
        </mxCell>
`, id, xmlAttr(noLigatures(value)), xmlAttr(st), parent, x, y, w, h)
}

func (r *renderer) image(id, parent, icon string, x, y, size float64) error {
	return r.imageStyled(id, parent, icon, x, y, size, "")
}

func (r *renderer) imageStyled(id, parent, icon string, x, y, size float64, extra string) error {
	uri, err := r.iconURI(icon)
	if err != nil {
		return err
	}
	r.vertex(id, parent, "", style("shape=image", "image="+uri, "imageAspect=1", "editable=0", "connectable=0", "movable=0", "resizable=0", extra), x, y, size, size)
	return nil
}

// hatch is draw.io's sketch fill for planned boxes (ADR-0008); dotted marks deprecated ones.
const (
	hatch  = "fillStyle=hatch;hachureGap=9;hachureAngle=-41;fillWeight=1;"
	dotted = "1 3"
)

func (r *renderer) arrow(a diagram.Arrow) {
	st := style("edgeStyle=none", "html=1", "strokeColor="+r.th.Muted, "strokeWidth=2.5", "endArrow=block", "endFill=1", "endSize=8")
	fmt.Fprintf(&r.b, `        <mxCell id="%s" value="" style="%s" edge="1" parent="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="%g" y="%g" as="sourcePoint" />
            <mxPoint x="%g" y="%g" as="targetPoint" />
          </mxGeometry>
        </mxCell>
`, a.ID, xmlAttr(st), a.From.X(), a.From.Y(), a.To.X(), a.To.Y())
}

// offering draws a catalog card (ADR-0013) from the blocks the resolver placed.
func (r *renderer) offering(o diagram.Offering) error {
	const pad = house.OfferingPad
	lineH := house.OfferingLineH(house.OfferingFontSize)
	fill, fade := "fillColor="+r.th.CardFill+";", ""
	if o.Status == "planned" {
		// Lighter than a planned node: a card carries paragraphs the hatch must not drown.
		fill = style("fillColor="+r.th.Hatch, "fillOpacity=45") + hatch + style("dashed=1", "dashPattern=6 4")
		fade = "textOpacity=75"
	}
	r.vertex(o.ID, "1", "", style("rounded=1", "absoluteArcSize=1", "arcSize=16")+fill+
		style("strokeColor="+r.th.Accent, "strokeWidth=1.5", "movable=0"), o.X, o.Y, o.W, o.H)
	text := func(id, value string, x, y, w, h, size float64, color string, extra ...string) {
		st := style(append([]string{"text", "html=1", "align=left", "verticalAlign=top", "spacing=0", "whiteSpace=nowrap",
			font(size, color), fade}, extra...)...)
		r.vertex(id, "1", value, st, x, y, w, h)
	}
	x, inner := o.X+pad, o.W-2*pad
	titleX := x
	if o.Icon != "" {
		iconStyle := ""
		if o.Status == "planned" {
			iconStyle = "opacity=40"
		}
		if err := r.imageStyled(o.ID+"__icon", "1", o.Icon, x, o.Y+pad+2, house.IconSize, iconStyle); err != nil {
			return err
		}
		titleX += house.IconSize + 10
	}
	text(o.ID+"__title", "<b>"+html.EscapeString(o.Title)+"</b>", titleX, o.Y+pad, inner-(titleX-x), house.OfferingHeaderH, 14, r.th.Title, "verticalAlign=middle")
	join := func(lines []string) string {
		esc := make([]string, len(lines))
		for i, l := range lines {
			esc[i] = html.EscapeString(l)
		}
		return strings.Join(esc, "<br>")
	}
	if len(o.Summary.Lines) > 0 {
		text(o.ID+"__summary", join(o.Summary.Lines), x, o.Summary.Y, inner, float64(len(o.Summary.Lines))*lineH, house.OfferingFontSize, r.th.Text)
	}
	for i, label := range []string{"YOU GET", "HOW TO REQUEST", "BACKED BY"} {
		if o.LabelsY[i] != 0 {
			text(fmt.Sprintf("%s__label%d", o.ID, i), "<b>"+label+"</b>", x, o.LabelsY[i], inner, house.OfferingLineH(house.OfferingLabelSize), house.OfferingLabelSize, r.th.Muted)
		}
	}
	for i, b := range o.Provides {
		h := float64(len(b.Lines)) * lineH
		text(fmt.Sprintf("%s__bullet%d", o.ID, i), "•", x, b.Y, house.OfferingBulletIndent, h, house.OfferingFontSize, r.th.Accent)
		text(fmt.Sprintf("%s__provides%d", o.ID, i), join(b.Lines), x+house.OfferingBulletIndent, b.Y, inner-house.OfferingBulletIndent, h, house.OfferingFontSize, r.th.Text)
	}
	if len(o.Request.Lines) > 0 {
		h := float64(len(o.Request.Lines))*lineH + 2*house.OfferingRequestPad
		r.vertex(o.ID+"__request-box", "1", "", style("rounded=1", "absoluteArcSize=1", "arcSize=8", "fillColor="+r.th.GridLine, "strokeColor=none", "movable=0"),
			x, o.Request.Y-house.OfferingRequestPad, inner, h)
		text(o.ID+"__request", join(o.Request.Lines), x+house.OfferingRequestPad, o.Request.Y, inner-2*house.OfferingRequestPad, h-2*house.OfferingRequestPad,
			house.OfferingFontSize, r.th.Title)
	}
	for i, l := range o.Logos {
		lx := l.X
		if l.Icon != "" {
			if err := r.image(fmt.Sprintf("%s__logo%d", o.ID, i), "1", l.Icon, l.X, l.Y, house.OfferingLogoSize); err != nil {
				return err
			}
			lx += house.OfferingLogoSize + 6
		}
		text(fmt.Sprintf("%s__logo%d-text", o.ID, i), html.EscapeString(l.Title), lx, l.Y, house.TextWidth(l.Title, 10)+4, house.OfferingLogoSize, 10, r.th.Text, "verticalAlign=middle")
	}
	if pill, ok := house.NodePillRect(diagram.Node{Status: o.Status, Target: o.Target, X: o.X, Y: o.Y, W: o.W, H: o.H}); ok {
		r.pill(o.ID+"__pill", house.PillText(o.Status, o.Target), pill)
	}
	return nil
}

func (r *renderer) pill(id, text string, b diagram.Rect) {
	st := style("rounded=1", "arcSize=50", "html=1", "fillColor="+r.th.Muted, "strokeColor=none", "align=center",
		"verticalAlign=middle", "fontStyle=1", "movable=0", "connectable=0", font(house.PortFontSize, r.th.PortText))
	r.vertex(id, "1", html.EscapeString(text), st, b.X, b.Y, b.W, b.H)
}

func (r *renderer) open() {
	s := r.spec
	fmt.Fprintf(&r.b, `<mxfile host="archgen">
  <diagram id="%s-%s" name="%s (%s)">
    <mxGraphModel grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="0" pageScale="1" pageWidth="%g" pageHeight="%g" background="%s" math="0" shadow="0">
      <root>
        <mxCell id="0" />
        <mxCell id="1" parent="0" />
`, s.ID, r.th.Name, xmlAttr(s.Title), r.th.Name, s.Width, s.Height, r.th.Background)
}

func (r *renderer) close() string {
	r.b.WriteString("      </root>\n    </mxGraphModel>\n  </diagram>\n</mxfile>\n")
	return r.b.String()
}

func (r *renderer) render() (string, error) {
	s := r.spec
	r.open()
	r.background()
	r.header()
	for _, z := range s.Zones {
		if err := r.zone(z); err != nil {
			return "", err
		}
	}
	// Edges before nodes so boxes and port badges sit on top of line ends.
	for _, e := range s.Edges {
		r.edge(e)
	}
	for _, n := range s.Nodes {
		if err := r.node(n); err != nil {
			return "", err
		}
	}
	for _, a := range s.Arrows {
		r.arrow(a)
	}
	r.portBadges()
	for _, n := range s.Notes {
		r.vertex(n.ID, "1", r.markup(n.Text), style("text", "html=1", "whiteSpace=wrap", "align=left", "verticalAlign=top", font(house.SubFontSize, r.th.Muted)), n.X, n.Y, n.W, n.H)
	}
	for _, c := range s.Cards {
		r.card(c)
	}
	for _, o := range s.Offerings {
		if err := r.offering(o); err != nil {
			return "", err
		}
	}
	if s.Matrix != nil {
		r.matrix(s.Matrix)
	}
	return r.close(), nil
}

func (r *renderer) background() {
	s := r.spec
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%g" height="%g"><defs><pattern id="g" width="40" height="40" patternUnits="userSpaceOnUse"><path d="M 40 0 L 0 0 0 40" fill="none" stroke="%s" stroke-width="0.5"/></pattern></defs><rect width="100%%" height="100%%" fill="%s"/><rect width="100%%" height="100%%" fill="url(#g)"/></svg>`,
		s.Width, s.Height, r.th.GridLine, r.th.Background)
	uri := "data:image/svg+xml," + base64.StdEncoding.EncodeToString([]byte(svg))
	r.vertex("bg", "1", "", style("shape=image", "image="+uri, "imageAspect=0", "locked=1", "editable=0", "movable=0", "resizable=0", "deletable=0"), 0, 0, s.Width, s.Height)
}

func (r *renderer) header() {
	s := r.spec
	if s.Title == "" && s.Subtitle == "" && len(s.Meta) == 0 {
		return
	}
	r.vertex("header-dot", "1", "", style("ellipse", "fillColor="+r.th.Accent, "strokeColor=none"), 40, 44, 12, 12)
	title := fmt.Sprintf(`<b>%s</b>`, html.EscapeString(s.Title))
	r.vertex("header-title", "1", title, style("text", "html=1", "align=left", "verticalAlign=middle", font(24, r.th.Title)), 64, 30, s.Width-400, 40)
	var sub []string
	if s.Subtitle != "" {
		sub = append(sub, fmt.Sprintf(`<font color="%s" style="font-size:13px">%s</font>`, r.th.Text, r.markup(s.Subtitle)))
	}
	for _, m := range s.Meta {
		sub = append(sub, r.markup(m))
	}
	r.vertex("header-sub", "1", strings.Join(sub, "<br>"), style("text", "html=1", "align=left", "verticalAlign=top", font(11, r.th.Muted)), 64, 74, s.Width-400, 80)
}

func (r *renderer) zone(z diagram.Zone) error {
	color := r.th.Zones[z.Color]
	dash, width, opacity, arc := "4 4", 1.0, 5, 8
	if z.Kind == "region" {
		dash, width, opacity, arc = "8 4", 1.2, 3, 12
	}
	pad := 12.0
	if z.Icon != "" {
		pad += house.ZoneIconSize + 8
	}
	v := fmt.Sprintf(`<b>%s</b>`, html.EscapeString(z.Label))
	if z.Sub != "" {
		v += fmt.Sprintf(`&nbsp;&nbsp;<font color="%s" style="font-size:%gpx">%s</font>`, r.th.Muted, house.SubFontSize, html.EscapeString(z.Sub))
	}
	fill := style("fillColor="+color, fmt.Sprintf("fillOpacity=%d", opacity))
	titleOpacity := ""
	switch z.Status {
	case "planned":
		fill = style("fillColor="+r.th.Hatch, "fillOpacity=35") + hatch
		titleOpacity = "textOpacity=60"
	case "deprecated":
		dash = dotted
		v = strings.Replace(v, "<b>", "<b><s>", 1)
		v = strings.Replace(v, "</b>", "</s></b>", 1)
	}
	st := style("rounded=1", "absoluteArcSize=1", fmt.Sprintf("arcSize=%d", arc), "html=1") + fill +
		style("strokeColor="+color, fmt.Sprintf("strokeWidth=%g", width), "dashed=1", "dashPattern="+dash, "container=0", "collapsible=0")
	r.vertex(z.ID, "1", "", st, z.X, z.Y, z.W, z.H)
	// Title and icon share one row and are both centred on it; draw.io's own label
	// padding would otherwise leave the icon a few pixels above the text.
	title := house.ZoneTitleBox(z)
	rowY := title.Y - z.Y
	r.vertex(z.ID+"__title", z.ID, v, style("text", "html=1", "align=left", "verticalAlign=middle", "spacing=0",
		fmt.Sprintf("spacingLeft=%g", pad-12), font(house.ZoneFontSize, color), "movable=0", "resizable=0", "connectable=0", titleOpacity),
		12, rowY, title.W, title.H)
	if pill, ok := house.ZonePillRect(z); ok {
		r.pill(z.ID+"__pill", house.PillText(z.Status, z.Target), pill)
	}
	if z.Icon != "" {
		iconStyle := ""
		if z.Status == "planned" {
			iconStyle = "opacity=40"
		}
		return r.imageStyled(z.ID+"__icon", z.ID, z.Icon, 12, rowY+(title.H-house.ZoneIconSize)/2, house.ZoneIconSize, iconStyle)
	}
	return nil
}

func (r *renderer) nodeLabel(n diagram.Node) string {
	title := html.EscapeString(n.Title)
	if n.Status == "deprecated" {
		title = "<s>" + title + "</s>"
	}
	v := fmt.Sprintf(`<b><font color="%s" style="font-size:%gpx">%s</font></b>`, r.th.Title, house.TitleFontSize, title)
	if n.Badge != "" {
		v += fmt.Sprintf(` <b><font color="%s" style="font-size:%gpx">⚠ %s</font></b>`, r.th.Warn, house.SubFontSize, html.EscapeString(n.Badge))
	}
	if n.Tech != "" {
		v += fmt.Sprintf(`<br><i><font color="%s" style="font-size:%gpx">%s</font></i>`, r.th.Text, house.SubFontSize, html.EscapeString(n.Tech))
	}
	if n.Desc != "" {
		v += fmt.Sprintf(`<br><font color="%s" style="font-size:%gpx">%s</font>`, r.th.Muted, house.SubFontSize, r.markup(n.Desc))
	}
	return v
}

func (r *renderer) node(n diagram.Node) error {
	k := r.th.Nodes[n.Kind]
	if n.Shape == "actor" {
		uri, err := r.iconURI(n.Icon)
		if err != nil {
			return err
		}
		v := fmt.Sprintf(`<b>%s</b>`, html.EscapeString(n.Title))
		for _, l := range n.Lines() {
			v += fmt.Sprintf(`<br><font color="%s" style="font-size:%gpx">%s</font>`, r.th.Muted, house.SubFontSize, html.EscapeString(l))
		}
		st := style("shape=image", "image="+uri, "imageAspect=1", "html=1", "verticalLabelPosition=bottom", "verticalAlign=top",
			"labelPosition=center", "align=center", fmt.Sprintf("spacingTop=%g", house.ActorLabelGap), font(house.TitleFontSize, r.th.Title))
		r.vertex(n.ID, "1", v, st, n.X, n.Y, n.W, n.H)
		return nil
	}
	shape := []string{"rounded=1", "absoluteArcSize=1", "arcSize=12"}
	top := 10.0
	if n.Shape == "cylinder" {
		shape = []string{"shape=cylinder3", "size=8", "boundedLbl=1", "backgroundOutline=1"}
		top = 18
	}
	pad := 10.0
	if n.Icon != "" {
		pad = house.TextPadLeft
	}
	fill, iconStyle := "fillColor="+k.Fill+";", ""
	switch n.Status {
	case "planned":
		// Hatch at half strength and text at 75 %: at slide scale a full hatch drowns the title.
		fill = style("fillColor="+r.th.Hatch, "fillOpacity=50") + hatch + style("dashed=1", "dashPattern=6 4", "textOpacity=75")
		iconStyle = "opacity=40"
	case "deprecated":
		fill += style("dashed=1", "dashPattern="+dotted)
	}
	// Drawn before the box so they sit behind it, the farthest first (ADR-0008).
	if n.Multiplicity != "" {
		for _, depth := range []float64{2, 1} {
			d := depth * house.StackOffset
			r.vertex(fmt.Sprintf("%s__stack%g", n.ID, depth), "1", "", style(shape...)+style("fillColor="+k.Fill, "strokeColor="+k.Stroke,
				"strokeWidth=1.5", "dashed=1", "dashPattern=4 3", "editable=0", "movable=0", "connectable=0"), n.X+d, n.Y-d, n.W, n.H)
		}
	}
	st := style(shape...) + fill + style("html=1", "whiteSpace=wrap", "strokeColor="+k.Stroke, "strokeWidth=1.5",
		"align=left", "verticalAlign=top", fmt.Sprintf("spacingLeft=%g", pad), fmt.Sprintf("spacingTop=%g", top-4), "spacingRight=8",
		font(house.TitleFontSize, r.th.Text))
	r.vertex(n.ID, "1", r.nodeLabel(n), st, n.X, n.Y, n.W, n.H)
	if n.Icon != "" {
		iconY := house.IconInset
		if n.Shape == "cylinder" {
			iconY += 8
		}
		// Child of the node so it moves with it when edited by hand in draw.io.
		if err := r.imageStyled(n.ID+"__icon", n.ID, n.Icon, house.IconInset, iconY, house.IconSize, iconStyle); err != nil {
			return err
		}
	}
	if pill, ok := house.NodePillRect(n); ok {
		r.pill(n.ID+"__pill", house.PillText(n.Status, n.Target), pill)
	}
	return nil
}

func (r *renderer) kindStrokeOf(id string) string {
	for _, n := range r.spec.Nodes {
		if n.ID == id {
			return r.th.Nodes[n.Kind].Stroke
		}
	}
	for _, z := range r.spec.Zones {
		if z.ID == id {
			return r.th.Zones[z.Color]
		}
	}
	return r.th.Muted
}

func (r *renderer) portBadges() {
	seen := map[string]bool{}
	for _, e := range r.spec.Edges {
		b, ok := r.ports[e.ID]
		if !ok || seen[b.ID] {
			continue
		}
		seen[b.ID] = true
		host, _ := r.spec.AnchorRect(b.Node)
		st := style("rounded=1", "absoluteArcSize=1", "arcSize=6", "html=1", "fillColor="+r.kindStrokeOf(b.Node), "strokeColor="+r.th.Background,
			"strokeWidth=1", "align=center", "verticalAlign=middle", "fontStyle=1", font(house.PortFontSize, r.th.PortText), "movable=0", "resizable=0")
		r.vertex(b.ID, b.Node, html.EscapeString(b.Text), st, b.Rect.X-host.X, b.Rect.Y-host.Y, b.Rect.W, b.Rect.H)
	}
}

var sideEntry = map[string][2]float64{"left": {0, 0.5}, "right": {1, 0.5}, "top": {0.5, 0}, "bottom": {0.5, 1}}

func (r *renderer) edge(e diagram.Edge) {
	k := r.th.Edges[e.Kind]
	src, _ := r.spec.AnchorRect(e.From)
	path := house.DrawnPath(e, r.ports)
	p0 := path[0]
	target := e.To
	var entry [2]float64
	if b, ok := r.ports[e.ID]; ok {
		target = b.ID
		entry = sideEntry[b.Side]
	} else {
		dst, _ := r.spec.AnchorRect(e.To)
		pn := path[len(path)-1]
		entry = [2]float64{(pn.X() - dst.X) / dst.W, (pn.Y() - dst.Y) / dst.H}
	}
	endFill := "1"
	if k.EndArrow == "cross" || k.EndArrow == "open" {
		endFill = "0"
	}
	parts := []string{"edgeStyle=orthogonalEdgeStyle", "rounded=1", "orthogonalLoop=1", "html=1", "jumpStyle=arc", "jumpSize=10",
		"strokeColor=" + k.Stroke, fmt.Sprintf("strokeWidth=%g", k.Width),
		"endArrow=" + k.EndArrow, "endFill=" + endFill, "endSize=7",
		fmt.Sprintf("exitX=%.4f", (p0.X()-src.X)/src.W), fmt.Sprintf("exitY=%.4f", (p0.Y()-src.Y)/src.H), "exitPerimeter=0",
		fmt.Sprintf("entryX=%.4f", entry[0]), fmt.Sprintf("entryY=%.4f", entry[1]), "entryPerimeter=0",
		"labelBackgroundColor=" + r.th.Background, font(house.EdgeFontSize, k.LabelColor)}
	if k.Dash != "" {
		parts = append(parts, "dashed=1", "dashPattern="+k.Dash)
	}
	if k.EndArrow == "cross" {
		parts = append(parts, "endSize=10")
	}
	frac := house.LabelFraction(house.LabelAnchor(e, path), path)
	fmt.Fprintf(&r.b, `        <mxCell id="%s" value="%s" style="%s" edge="1" parent="1" source="%s" target="%s">
          <mxGeometry x="%.4f" relative="1" as="geometry">
`, e.ID, xmlAttr(noLigatures(r.markup(e.Label))), xmlAttr(style(parts...)), e.From, target, frac*2-1)
	if e.LabelOffset != (diagram.Point{}) {
		fmt.Fprintf(&r.b, "            <mxPoint x=\"%g\" y=\"%g\" as=\"offset\" />\n", e.LabelOffset.X(), e.LabelOffset.Y())
	}
	if len(path) > 2 {
		r.b.WriteString("            <Array as=\"points\">\n")
		for _, p := range path[1 : len(path)-1] {
			fmt.Fprintf(&r.b, "              <mxPoint x=\"%g\" y=\"%g\" />\n", p.X(), p.Y())
		}
		r.b.WriteString("            </Array>\n")
	}
	r.b.WriteString("          </mxGeometry>\n        </mxCell>\n")
}

func (r *renderer) card(c diagram.Card) {
	color := r.th.Zones[c.Color]
	if color == "" {
		color = r.th.Muted
	}
	st := style("rounded=1", "absoluteArcSize=1", "arcSize=16", "html=1", "whiteSpace=wrap", "fillColor="+r.th.CardFill,
		"strokeColor="+r.th.CardStroke, "align=left", "verticalAlign=top", "spacingLeft=20", "spacingTop=14", "spacingRight=16",
		font(11, r.th.Text))
	head := fmt.Sprintf(`<font color="%s">●</font>&nbsp;<b><font color="%s" style="font-size:13px">%s</font></b><br><br>`, color, r.th.Title, html.EscapeString(c.Title))
	var lines []string
	for _, l := range c.Lines {
		if strings.HasPrefix(l, "[") || strings.HasPrefix(l, "!!") {
			lines = append(lines, r.markup(l))
		} else {
			lines = append(lines, "• "+r.markup(l))
		}
	}
	body := ""
	if !c.Legend {
		body = `<div style="line-height:1.6">` + strings.Join(lines, "<br>") + `</div>`
	}
	r.vertex(c.ID, "1", head+body, st, c.X, c.Y, c.W, c.H)
	if c.Legend {
		r.legend(c)
	}
}

func (r *renderer) legendText(id, text string, x, y, w float64) {
	r.vertex(id, "1", text, style("text", "html=1", "align=left", "verticalAlign=middle", font(10.5, r.th.Text)), x, y, w, 22)
}

// legendStatuses explains the planned and deprecated styles and each stacked box used in
// the diagram, so the hatch, the dots and the stacks never carry meaning alone (ADR-0006,
// ADR-0008).
func (r *renderer) legendStatuses(x, y, w float64) {
	used := map[string]bool{}
	for _, n := range r.spec.Nodes {
		used[n.Status] = true
	}
	for _, z := range r.spec.Zones {
		used[z.Status] = true
	}
	k := r.th.Nodes["external"]
	if used["planned"] {
		r.vertex("legend-planned", "1", "", style("rounded=1", "absoluteArcSize=1", "arcSize=6", "fillColor="+r.th.Hatch)+hatch+
			style("dashed=1", "dashPattern=6 4", "strokeColor="+k.Stroke, "strokeWidth=1.5"), x, y+2, 40, 18)
		r.legendText("legend-planned-text", "planned · pill: target", x+52, y, w)
		y += 26
	}
	if used["deprecated"] {
		r.vertex("legend-deprecated", "1", "", style("rounded=1", "absoluteArcSize=1", "arcSize=6", "fillColor="+k.Fill,
			"dashed=1", "dashPattern="+dotted, "strokeColor="+k.Stroke, "strokeWidth=1.5"), x, y+2, 40, 18)
		r.legendText("legend-deprecated-text", "deprecated", x+52, y, w)
		y += 26
	}
	shown := map[string]bool{}
	for i, n := range r.spec.Nodes {
		if n.Multiplicity == "" || n.Shape == "actor" || shown[n.Multiplicity] {
			continue
		}
		shown[n.Multiplicity] = true
		for depth := 2.0; depth >= 0; depth-- {
			st := style("rounded=1", "absoluteArcSize=1", "arcSize=6", "fillColor="+k.Fill, "strokeColor="+k.Stroke, "strokeWidth=1.5")
			if depth > 0 {
				st += style("dashed=1", "dashPattern=4 3")
			}
			r.vertex(fmt.Sprintf("legend-stack-%d-%g", i, depth), "1", "", st, x+3*depth, y+4-3*depth, 34, 14)
		}
		r.legendText(fmt.Sprintf("legend-stack-%d-text", i), html.EscapeString(n.Multiplicity), x+52, y, w)
		y += 26
	}
}

func (r *renderer) legend(c diagram.Card) {
	usedNodes, usedEdges := map[string]bool{}, map[string]bool{}
	for _, n := range r.spec.Nodes {
		if n.Shape != "actor" {
			usedNodes[n.Kind] = true
		}
	}
	for _, e := range r.spec.Edges {
		usedEdges[e.Kind] = true
	}
	x, y := c.X+20, c.Y+52
	for _, kind := range house.NodeKindOrder {
		if !usedNodes[kind] {
			continue
		}
		k := r.th.Nodes[kind]
		r.vertex("legend-node-"+kind, "1", "", style("rounded=1", "absoluteArcSize=1", "arcSize=6", "fillColor="+k.Fill, "strokeColor="+k.Stroke, "strokeWidth=1.5"), x, y+4, 22, 14)
		r.legendText("legend-node-"+kind+"-text", k.Legend, x+32, y, c.W/2-60)
		y += 26
	}
	y += 8
	r.legendText("legend-steps", r.markup("[1] request step · [A] background step"), x, y, c.W/2-20)
	y += 26
	st := style("rounded=1", "absoluteArcSize=1", "arcSize=6", "html=1", "fillColor="+r.th.Nodes["backend"].Stroke, "strokeColor=none",
		"align=center", "verticalAlign=middle", "fontStyle=1", font(house.PortFontSize, r.th.PortText))
	r.vertex("legend-port", "1", "TCP 80", st, x, y+3, house.PortWidth("TCP 80"), house.PortHeight)
	r.legendText("legend-port-text", "listening port on the server", x+52, y, c.W/2-70)

	x, y = c.X+c.W/2+10, c.Y+52
	for _, kind := range house.EdgeKindOrder {
		if !usedEdges[kind] {
			continue
		}
		k := r.th.Edges[kind]
		endFill := "1"
		if k.EndArrow == "cross" || k.EndArrow == "open" {
			endFill = "0"
		}
		st := style("edgeStyle=none", "html=1", "strokeColor="+k.Stroke, fmt.Sprintf("strokeWidth=%g", k.Width), "endArrow="+k.EndArrow, "endFill="+endFill, "endSize=6")
		if k.Dash != "" {
			st += "dashed=1;dashPattern=" + k.Dash + ";"
		}
		fmt.Fprintf(&r.b, `        <mxCell id="legend-edge-%s" value="" style="%s" edge="1" parent="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="%g" y="%g" as="sourcePoint" />
            <mxPoint x="%g" y="%g" as="targetPoint" />
          </mxGeometry>
        </mxCell>
`, kind, xmlAttr(st), x, y+11, x+40, y+11)
		r.legendText("legend-edge-"+kind+"-text", k.Legend, x+52, y, c.W/2-72)
		y += 26
	}
	r.legendStatuses(x, y, c.W/2-72)
	notes := fmt.Sprintf(`Large dashed frame: infra perimeter · small: functional group<br><font color="%s"><b>⚠ Gx</b></font>: docs and code disagree, see "Gaps"`, r.th.Warn)
	r.vertex(c.ID+"-notes", "1", notes, style("text", "html=1", "align=left", "verticalAlign=top", "whiteSpace=wrap", font(10.5, r.th.Muted)), c.X+20, c.Y+c.H-50, c.W-40, 40)
}
