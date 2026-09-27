package lint

import "github.com/TheGostsniperfr/Noodle/internal/house"

// checkZoneSpacing: sibling zones do not overlap and leave a MinZoneGap corridor.
func checkZoneSpacing(l *linter) {
	for i, a := range l.s.Zones {
		for _, b := range l.s.Zones[i+1:] {
			ra, rb := a.Rect(), b.Rect()
			if ra.Contains(rb) || rb.Contains(ra) {
				continue
			}
			if ra.Intersects(rb) {
				l.errf(a.ID, "overlaps zone %s", b.ID)
				continue
			}
			if ra.Inflate(house.MinZoneGap / 2).Intersects(rb.Inflate(house.MinZoneGap / 2)) {
				l.errf(a.ID, "closer than %.0fpx to zone %s", house.MinZoneGap, b.ID)
			}
		}
	}
}
