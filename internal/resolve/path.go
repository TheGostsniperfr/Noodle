package resolve

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// anchor is a parsed "side[@NN%|@NNpx]". The axis is the one the endpoint slides on:
// y for left and right, x for top and bottom.
type anchor struct {
	side     string
	axis     int
	explicit bool
	percent  bool
	value    float64
}

func parseAnchor(edgeID, end string) (string, anchor, error) {
	id, rest := model.Endpoint(end)
	side, at, hasAt := strings.Cut(rest, "@")
	a := anchor{side: side}
	switch side {
	case "left", "right":
		a.axis = 1
	case "top", "bottom":
		a.axis = 0
	default:
		return "", a, fmt.Errorf("resolve: %s: endpoint %q: unknown side %q", edgeID, end, side)
	}
	if !hasAt {
		return id, a, nil
	}
	num, unit := at, ""
	switch {
	case strings.HasSuffix(at, "%"):
		num, unit = strings.TrimSuffix(at, "%"), "%"
	case strings.HasSuffix(at, "px"):
		num, unit = strings.TrimSuffix(at, "px"), "px"
	}
	v, err := strconv.ParseFloat(num, 64)
	if unit == "" || err != nil || v < 0 || (unit == "%" && v > 100) {
		return "", a, fmt.Errorf("resolve: %s: endpoint %q: want @0-100%% or @NNpx", edgeID, end)
	}
	a.explicit, a.percent, a.value = true, unit == "%", v
	return id, a, nil
}

// span returns where the side starts on the sliding axis and how long it is.
func span(r diagram.Rect, axis int) (start, length float64) {
	if axis == 1 {
		return r.Y, r.H
	}
	return r.X, r.W
}

// fixedAt places the endpoint on its side at the given coordinate of the sliding axis.
func fixedAt(r diagram.Rect, a anchor, along float64) diagram.Point {
	switch a.side {
	case "left":
		return diagram.Point{r.X, along}
	case "right":
		return diagram.Point{r.X + r.W, along}
	case "top":
		return diagram.Point{along, r.Y}
	default:
		return diagram.Point{along, r.Y + r.H}
	}
}

func explicitAlong(edgeID string, r diagram.Rect, a anchor) (float64, error) {
	start, length := span(r, a.axis)
	offset := a.value
	if a.percent {
		offset = length * a.value / 100
	}
	if offset > length {
		return 0, fmt.Errorf("resolve: %s: offset %gpx is past the %s side (%gpx long)", edgeID, offset, a.side, length)
	}
	return start + offset, nil
}

// onSide rejects an aligned endpoint that would land beyond its side.
func onSide(edgeID, end string, r diagram.Rect, a anchor, along float64) error {
	start, length := span(r, a.axis)
	if along < start || along > start+length {
		return fmt.Errorf("resolve: %s: endpoint %q lines up at %s=%g, outside its side (%g to %g): move a waypoint or give an @offset",
			edgeID, end, [2]string{"x", "y"}[a.axis], along, start, start+length)
	}
	return nil
}

func middle(r diagram.Rect, axis int) float64 {
	start, length := span(r, axis)
	return start + length/2
}

// lane is a resolved waypoint "lane:<name>": the path runs along coordinate value of
// axis (0 for an x lane, which is vertical; 1 for a y lane, which is horizontal).
type lane struct {
	axis  int
	value float64
}

// supplies reports the coordinate a waypoint fixes on the given axis, if any.
func (r *resolver) supplies(w model.Waypoint, axis int) (float64, bool) {
	if w.Point != nil {
		return w.Point[axis], true
	}
	if ln := r.lane(w.Lane); ln.axis == axis {
		return ln.value, true
	}
	return 0, false
}

func (r *resolver) lane(name string) lane {
	l := r.l.Lanes[name]
	if l.X != nil {
		return lane{axis: 0, value: *l.X}
	}
	return lane{axis: 1, value: *l.Y}
}

// onLane is the point where a path on lane ln meets coordinate p on the other axis.
func onLane(ln lane, p diagram.Point) diagram.Point {
	out := p
	out[ln.axis] = ln.value
	return out
}

// path resolves a route to absolute points. An endpoint without @ lines up with its
// neighbour so the first and last segments stay orthogonal: from with the first
// waypoint, to with the point before it. A lane waypoint adds the two bends where the
// path enters and leaves the lane.
func (r *resolver) path(id string, route model.EdgeRoute) ([]diagram.Point, error) {
	fromID, fa, err := parseAnchor(id, route.From)
	if err != nil {
		return nil, err
	}
	toID, ta, err := parseAnchor(id, route.To)
	if err != nil {
		return nil, err
	}
	fromRect, err := r.absolute(fromID)
	if err != nil {
		return nil, err
	}
	toRect, err := r.absolute(toID)
	if err != nil {
		return nil, err
	}
	wps := route.Waypoints

	var toPt *diagram.Point
	if ta.explicit {
		along, err := explicitAlong(id, toRect, ta)
		if err != nil {
			return nil, err
		}
		p := fixedAt(toRect, ta, along)
		toPt = &p
	}

	var fromAlong float64
	switch {
	case fa.explicit:
		if fromAlong, err = explicitAlong(id, fromRect, fa); err != nil {
			return nil, err
		}
	case len(wps) > 0:
		v, ok := r.supplies(wps[0], fa.axis)
		if !ok {
			v = middle(fromRect, fa.axis)
		}
		fromAlong = v
	case toPt != nil:
		fromAlong = toPt[fa.axis]
	default:
		fromAlong = middle(fromRect, fa.axis)
	}
	if err := onSide(id, route.From, fromRect, fa, fromAlong); err != nil {
		return nil, err
	}
	pts := []diagram.Point{fixedAt(fromRect, fa, fromAlong)}

	var open *lane
	for _, w := range wps {
		prev := pts[len(pts)-1]
		if w.Lane != "" {
			ln := r.lane(w.Lane)
			pts = append(pts, onLane(ln, prev))
			open = &ln
			continue
		}
		p := diagram.Point(*w.Point)
		if open != nil {
			pts = append(pts, onLane(*open, p))
			open = nil
		}
		pts = append(pts, p)
	}

	if toPt == nil {
		prev := pts[len(pts)-1]
		along := prev[ta.axis]
		if open != nil {
			along = middle(toRect, ta.axis)
			if open.axis == ta.axis {
				along = open.value
			}
		}
		if err := onSide(id, route.To, toRect, ta, along); err != nil {
			return nil, err
		}
		p := fixedAt(toRect, ta, along)
		toPt = &p
	}
	if open != nil {
		pts = append(pts, onLane(*open, *toPt))
	}
	pts = append(pts, *toPt)
	return dedupe(pts), nil
}

func dedupe(pts []diagram.Point) []diagram.Point {
	out := pts[:1]
	for _, p := range pts[1:] {
		if p != out[len(out)-1] {
			out = append(out, p)
		}
	}
	return out
}
