package house_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
)

// zoneSub is nine words, greedy-wrapped at 40 characters into lines of 39, 39 and 10.
var zoneSub = strings.TrimSuffix(strings.Repeat("abcdefghi ", 9), " ")[:89] + "."

func TestZoneTitleBox_WrapsTheSub(t *testing.T) {
	t.Parallel()
	zone := diagram.Zone{ID: "platform", Label: "Platform", Sub: zoneSub, X: 100, Y: 100, W: 600, H: 300}

	got := house.ZoneTitleBox(zone)

	assert.Equal(t, diagram.Rect{X: 112, Y: 104, W: 234, H: 61.5}, got)
	assert.Equal(t, []string{
		"abcdefghi abcdefghi abcdefghi abcdefghi",
		"abcdefghi abcdefghi abcdefghi abcdefghi",
		"abcdefghi.",
	}, house.ZoneSubLines(zone))
}

func TestZoneTitleBox_KeepsTheSubOnTheTitleLine_WhenItFitsTheZone(t *testing.T) {
	t.Parallel()
	zone := diagram.Zone{ID: "platform", Label: "Platform", Sub: "short sub", X: 100, Y: 100, W: 600, H: 300}

	got := house.ZoneTitleBox(zone)

	assert.Equal(t, diagram.Rect{X: 112, Y: 104, W: 52.8 + 12 + 13.2 + 54, H: 24}, got)
	assert.Empty(t, house.ZoneSubLines(zone))
}

func TestZonePillRect_SitsOnTheLabelRow_WhenTheSubWraps(t *testing.T) {
	t.Parallel()
	zone := diagram.Zone{ID: "platform", Label: "Platform", Sub: zoneSub, Status: "planned", X: 100, Y: 100, W: 600, H: 300}

	pill, ok := house.ZonePillRect(zone)

	assert.True(t, ok)
	assert.Equal(t, 112+64.8+6, pill.X)
	assert.Equal(t, 104+(24-house.PortHeight)/2, pill.Y)
}

func TestZoneSubLines_WrapAtTheColumnWidth(t *testing.T) {
	t.Parallel()
	zone := diagram.Zone{Label: "L", Sub: strings.Repeat("word ", 60), X: 0, Y: 0, W: 300}

	lines := house.ZoneSubLines(zone)

	assert.NotEmpty(t, lines)
	for _, l := range lines {
		assert.LessOrEqual(t, house.TextWidth(l, house.SubFontSize), house.ZoneSubColumn)
	}
}
