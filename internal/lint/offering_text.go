package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// checkOfferingText: every line of a catalog card fits its width, and cards do not
// overlap (ADR-0012). The resolver wraps at spaces, so this catches single long words.
func checkOfferingText(l *linter) {
	for i, o := range l.s.Offerings {
		inner := o.W - 2*house.OfferingPad
		fits := func(what string, lines []string, width float64) {
			for _, line := range lines {
				if w := house.TextWidth(line, house.OfferingFontSize); w > width {
					l.errf(o.ID, "%s %q overflows: %.0fpx text in %.0fpx", what, line, w, width)
				}
			}
		}
		titleW := inner
		if o.Icon != "" {
			titleW -= house.IconSize + 10
		}
		if w := house.TextWidth(o.Title, 14); w > titleW {
			l.errf(o.ID, "title overflows: %.0fpx text in %.0fpx", w, titleW)
		}
		fits("summary", o.Summary.Lines, inner)
		for _, b := range o.Provides {
			fits("item", b.Lines, inner-house.OfferingBulletIndent)
		}
		fits("request", o.Request.Lines, inner-2*house.OfferingRequestPad)
		for _, p := range l.s.Offerings[i+1:] {
			if o.Rect().Intersects(p.Rect()) {
				l.errf(o.ID, "overlaps card %s", p.ID)
			}
		}
	}
}
