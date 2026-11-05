package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheGostsniperfr/Noodle/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// drawioIcons are the extra icon directories each committed diagram is rendered with,
// the same as the Makefile's examples target.
func drawioIcons(dir string) []string {
	switch {
	case strings.HasPrefix(dir, "../../examples/platform-regression"), strings.HasPrefix(dir, "../../examples/topology-diff"):
		return []string{"-icons", "../../examples/platform-regression/icons"}
	case strings.HasPrefix(dir, "../../docs/diagrams"):
		return []string{"-icons", "../../docs/diagrams/icons"}
	}
	return nil
}

// TestExamples_DrawioCurrent re-renders every committed .drawio of the examples and the
// docs diagrams, dark theme, and requires the committed bytes. A change to what noodle
// draws shows up as a diff of these files; regenerate them with -update.
func TestExamples_DrawioCurrent(t *testing.T) {
	var files []string
	for _, root := range []string{"../../examples", "../../docs/diagrams"} {
		require.NoError(t, filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(path, ".drawio") {
				files = append(files, path)
			}
			return err
		}))
	}
	require.NotEmpty(t, files)

	for _, file := range files {
		t.Run(strings.TrimPrefix(file, "../../"), func(t *testing.T) {
			dir := filepath.Dir(file)
			sys, err := model.LoadSystem(dir)
			require.NoError(t, err)
			view := strings.TrimSuffix(filepath.Base(file), ".drawio")
			if _, ok := sys.Views[view]; !ok {
				ids := sys.ViewIDs()
				require.Len(t, ids, 1, "%s: no view %q and not a single-view system", dir, view)
				view = ids[0]
			}
			out := filepath.Join(t.TempDir(), "out.drawio")

			err = renderCommand(append([]string{dir, "-view", view, "-o", out}, drawioIcons(dir)...))

			require.NoError(t, err)
			got, err := os.ReadFile(out)
			require.NoError(t, err)
			if *update {
				require.NoError(t, os.WriteFile(file, got, 0o644))
			}
			want, err := os.ReadFile(file)
			require.NoError(t, err)
			assert.True(t, string(want) == string(got), "%s is out of date: run go test ./cmd/noodle -run TestExamples_DrawioCurrent -update", file)
		})
	}
}
