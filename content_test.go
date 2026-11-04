package noodle_test

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	noodle "github.com/TheGostsniperfr/Noodle"
)

func TestContent_CarriesTheSkillAndTheAgents(t *testing.T) {
	t.Parallel()
	tests := []struct {
		dir  string
		want []string
	}{
		{"skills/noodle-diagram", []string{"SKILL.md", "reference.md"}},
		{"agents", []string{"diagram-reviewer.md", "fact-finder.md", "icon-curator.md"}},
	}
	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			t.Parallel()

			entries, err := fs.ReadDir(noodle.Content, tt.dir)

			require.NoError(t, err)
			var got []string
			for _, e := range entries {
				got = append(got, e.Name())
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
