package house_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/house"
)

func TestFitsWithMargin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		fontSize, avail, text float64
		expected              bool
	}{
		{"12px in 200, margin 14.448: 185 fits", 12, 200, 185, true},
		{"12px in 200, margin 14.448: 186 does not", 12, 200, 186, false},
		{"12px in 600, margin 24: 576 fits", 12, 600, 576, true},
		{"12px in 600, margin 24: 577 does not", 12, 600, 577, false},
		{"10px in 200, margin 12.04: 187 fits", 10, 200, 187, true},
		{"10px in 200, margin 12.04: 188 does not", 10, 200, 188, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := house.FitsWithMargin(tt.text, tt.avail, tt.fontSize)

			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMinAvailable_IsTheSmallestWidthThatFits(t *testing.T) {
	t.Parallel()
	for _, textW := range []float64{20, 168.56, 173.376, 400, 1500} {
		for _, fontSize := range []float64{10, 12} {
			min := house.MinAvailable(textW, fontSize)

			assert.True(t, house.FitsWithMargin(textW, min+1e-6, fontSize), "%v at %v", textW, fontSize)
			assert.False(t, house.FitsWithMargin(textW, min-1e-3, fontSize), "%v at %v", textW, fontSize)
		}
	}
}
