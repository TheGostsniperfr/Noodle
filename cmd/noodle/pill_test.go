package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tutorialWith copies examples/tutorial, applying each old→new replacement to its model.
func tutorialWith(t *testing.T, replacements ...string) string {
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
		text := string(raw)
		if filepath.Base(path) == "model.yaml" {
			text = strings.NewReplacer(replacements...).Replace(text)
		}
		rel, _ := filepath.Rel(src, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, rel)), 0o755))
		return os.WriteFile(filepath.Join(dir, rel), []byte(text), 0o600)
	}))
	return dir
}

func TestRender_InheritedStatusPill(t *testing.T) {
	t.Parallel()
	const plannedZone = "title: CLUSTER}"
	tests := []struct {
		name  string
		repl  []string
		pills []string
	}{
		{"zone-pill-once", []string{plannedZone, "title: CLUSTER, status: planned, target: Q1}"}, []string{"z-cluster__pill"}},
		{"child-own-target-keeps-pill", []string{plannedZone, "title: CLUSTER, status: planned, target: Q1}",
			"title: Postgres,", "title: Postgres, target: Q2,"}, []string{"db__pill", "z-cluster__pill"}},
		{"planned-node-alone", []string{"title: Postgres,", "title: Postgres, status: planned,"}, []string{"db__pill"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := tutorialWith(t, tt.repl...)
			out := filepath.Join(t.TempDir(), "x.drawio")

			require.NoError(t, renderCommand([]string{dir, "-o", out}))

			raw, err := os.ReadFile(out)
			require.NoError(t, err)
			var got []string
			for _, field := range strings.Split(string(raw), `id="`)[1:] {
				if id, _, _ := strings.Cut(field, `"`); strings.HasSuffix(id, "__pill") {
					got = append(got, id)
				}
			}
			assert.ElementsMatch(t, tt.pills, got)
		})
	}
}
