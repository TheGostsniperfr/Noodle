package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoLigatures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, input, expected string
	}{
		{"empty value stays empty so blank cells keep no markup", "", ""},
		{"text is wrapped so pairs like >- stay two glyphs", "&lt;app&gt;-frontend", `<span style="font-variant-ligatures:none">&lt;app&gt;-frontend</span>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, noLigatures(tt.input))
		})
	}
}
