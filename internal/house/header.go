package house

import (
	"math"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
)

const (
	HeaderX            = 40.0
	HeaderTitleY       = 30.0
	HeaderTitleH       = 40.0
	HeaderTextX        = 64.0
	HeaderSubY         = 74.0
	HeaderSubH         = 80.0
	HeaderDotY         = 44.0
	HeaderDotSize      = 12.0
	HeaderTextInset    = 400.0
	HeaderTitleSize    = 24.0
	HeaderSubtitleSize = 13.0
	HeaderMetaSize     = 11.0
)

// HeaderLayout is where the renderer puts the header cells, and Box the area the header
// text covers once drawn: from the dot to the right edge of the title and sub cells,
// down to the last line of subtitle and meta.
type HeaderLayout struct {
	Dot, Title, Sub, Box diagram.Rect
}

// HeaderBox is the area the header covers, or the zero Rect when the spec has none.
func HeaderBox(s *diagram.Spec) diagram.Rect { return Header(s).Box }

func Header(s *diagram.Spec) HeaderLayout {
	if s.Title == "" && s.Subtitle == "" && len(s.Meta) == 0 {
		return HeaderLayout{}
	}
	textW := s.Width - HeaderTextInset
	h := HeaderLayout{
		Dot:   diagram.Rect{X: HeaderX, Y: HeaderDotY, W: HeaderDotSize, H: HeaderDotSize},
		Title: diagram.Rect{X: HeaderTextX, Y: HeaderTitleY, W: textW, H: HeaderTitleH},
		Sub:   diagram.Rect{X: HeaderTextX, Y: HeaderSubY, W: textW, H: HeaderSubH},
	}
	bottom := HeaderTitleY + HeaderTitleH
	if s.Subtitle != "" || len(s.Meta) > 0 {
		bottom = HeaderSubY
	}
	if s.Subtitle != "" {
		bottom += HeaderSubtitleSize * LineHeightEm
	}
	for _, m := range s.Meta {
		lines := math.Max(1, math.Ceil(TextWidth(PlainText(m), HeaderMetaSize)/textW))
		bottom += lines * HeaderMetaSize * LineHeightEm
	}
	h.Box = diagram.Rect{X: HeaderX, Y: HeaderTitleY, W: s.Width - 376, H: bottom - HeaderTitleY}
	return h
}
