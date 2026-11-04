package house

import "math"

// ExportCharWidthEm is the widest monospace fallback draw.io has been seen to export
// with, about 0.3 % over CharWidthEm. Node text is measured at this width, plus a margin.
const ExportCharWidthEm = 0.602

// ExportTextWidth is the width of s at fontSize in the widest export font.
func ExportTextWidth(s string, fontSize float64) float64 {
	return float64(len([]rune(s))) * fontSize * ExportCharWidthEm
}

// TextMargin is the slack text keeps inside availW: two characters, or 4 % of the
// available width when that is more.
func TextMargin(availW, fontSize float64) float64 {
	return math.Max(2*fontSize*ExportCharWidthEm, 0.04*availW)
}

// FitsWithMargin says whether text textW wide, measured at ExportCharWidthEm, fits availW
// with its margin.
func FitsWithMargin(textW, availW, fontSize float64) bool {
	return textW+TextMargin(availW, fontSize) <= availW
}

// MinAvailable is the smallest width in which FitsWithMargin holds for text textW wide.
func MinAvailable(textW, fontSize float64) float64 {
	return math.Max(textW+2*fontSize*ExportCharWidthEm, textW/0.96)
}
