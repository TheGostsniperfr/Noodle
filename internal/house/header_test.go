package house

import (
	"strings"
	"testing"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/stretchr/testify/assert"
)

func TestHeaderBox(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("x", 190)
	tests := []struct {
		name     string
		spec     diagram.Spec
		expected diagram.Rect
	}{
		{"a subtitle and three one-line meta entries end at 131.5",
			diagram.Spec{Width: 1600, Title: "T", Subtitle: "s", Meta: []string{"a", "b", "c"}}, diagram.Rect{X: 40, Y: 30, W: 1224, H: 101.5}},
		{"a meta line wider than the text column counts as two lines",
			diagram.Spec{Width: 1600, Title: "T", Subtitle: "s", Meta: []string{long}}, diagram.Rect{X: 40, Y: 30, W: 1224, H: 87.75}},
		{"a title alone ends with the title cell",
			diagram.Spec{Width: 1600, Title: "T"}, diagram.Rect{X: 40, Y: 30, W: 1224, H: 40}},
		{"no title, subtitle or meta draws no header", diagram.Spec{Width: 1600}, diagram.Rect{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, HeaderBox(&tt.spec))
		})
	}
}
