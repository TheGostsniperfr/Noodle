package house

import "strings"

// Access matrix metrics (ADR-0019), shared by the matrix resolver, the lint and the
// renderer.
const (
	MatrixRowHeaderW     = 260.0
	MatrixCellW          = 104.0
	MatrixCellH          = 58.0
	MatrixGap            = 4.0
	MatrixGroupGap       = 16.0
	MatrixGroupH         = 46.0
	MatrixColumnH        = 46.0
	MatrixRowGroupH      = 26.0
	MatrixCountW         = 150.0
	MatrixPad            = 6.0
	MatrixLegendGap      = 36.0
	MatrixLegendCaptionW = 120.0
	MatrixHeaderFontSize = 10.5
	MatrixLevelFontSize  = 11.0
	MatrixNoteFontSize   = 9.5
	MatrixRowFontSize    = 12.0
	MatrixTagFontSize    = 9.0
	MatrixTagH           = 15.0
)

// LevelWord is what a cell says for a level.
func LevelWord(level string) string {
	if level == "breakglass" {
		return "BREAK-GLASS"
	}
	return strings.ToUpper(level)
}

// ChangeSymbol starts a cell's tag: + added, → changed, − removed.
var ChangeSymbol = map[string]string{"added": "+", "changed": "→", "removed": "−"}

// MatrixTagText is a cell's tag: the change symbol and the plan phase.
func MatrixTagText(change, phase string) string {
	if change == "" {
		return ""
	}
	return strings.TrimSpace(ChangeSymbol[change] + " " + phase)
}
