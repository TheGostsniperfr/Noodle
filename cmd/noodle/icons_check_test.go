package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tutorialWithIcon copies examples/tutorial with its cluster zone's icon replaced.
func tutorialWithIcon(t *testing.T, icon string) string {
	t.Helper()
	dir := t.TempDir()
	src := "../../examples/tutorial"
	require.NoError(t, filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		raw = []byte(strings.Replace(string(raw), "icon: kubernetes,", "icon: "+icon+",", 1))
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, rel)), 0o755))
		return os.WriteFile(filepath.Join(dir, rel), raw, 0o600)
	}))
	return dir
}

func TestRender_ChecksIconsWithoutOutput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		icon    string
		wantErr string
	}{
		{"unknown-icon-fails", "no-such-icon", `icon "no-such-icon" not found`},
		{"known-icon-passes", "kubernetes", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := tutorialWithIcon(t, tt.icon)

			err := renderCommand([]string{dir})

			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Contains(t, err.Error(), "-icons")
		})
	}
}
