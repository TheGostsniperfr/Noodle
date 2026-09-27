package model_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

func TestLoadSystem_ReadsModelViewsAndLayouts(t *testing.T) {
	t.Parallel()

	s, err := model.LoadSystem("testdata/minimal")

	require.NoError(t, err)
	assert.Len(t, s.Model.Elements, 4)
	assert.Equal(t, []string{"login", "request-path"}, s.ViewIDs())
	assert.Equal(t, "sequence", s.Views["login"].Type)
	route := s.Layouts["request-path"].Edges["c-api-db"]
	assert.Equal(t, "bus", route.Waypoints[0].Lane)
	assert.Equal(t, &model.Point{150, 400}, s.Layouts["request-path"].Edges["c-user-db"].Waypoints[0].Point)
}

// writeSystem copies testdata/minimal to a temp dir and overwrites one file.
func writeSystem(t *testing.T, rel, content string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.CopyFS(dir, os.DirFS("testdata/minimal")))
	require.NoError(t, os.WriteFile(filepath.Join(dir, rel), []byte(content), 0o644))
	return dir
}

func TestLoadSystem_Fails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, file, content, wantErr string
	}{
		{"on an unknown field, with file and line", "model.yaml",
			"apiVersion: noodle/v1alpha1\nkind: Model\nelements:\n  - {id: a, kind: backend, colour: red}\n",
			"model.yaml: yaml: unmarshal errors:\n  line 4: field colour not found"},
		{"on a wrong kind before any field error", "views/request-path.yaml",
			"apiVersion: noodle/v1alpha1\nkind: Layout\ntype: topology\n",
			`request-path.yaml: kind "Layout", want "View"`},
		{"on a missing apiVersion", "layouts/request-path.yaml",
			"kind: Layout\n",
			`apiVersion "", want "noodle/v1alpha1"`},
		{"on a waypoint that is neither a point nor a lane", "layouts/request-path.yaml",
			"apiVersion: noodle/v1alpha1\nkind: Layout\nedges:\n  c-api-db: {from: api.right, to: db.left, waypoints: [bus]}\n",
			`waypoint "bus" is neither [x, y] nor lane:<name>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := writeSystem(t, tt.file, tt.content)

			_, err := model.LoadSystem(dir)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestLoadSystem_Fails_WhenModelIsMissing(t *testing.T) {
	t.Parallel()

	_, err := model.LoadSystem(t.TempDir())

	assert.ErrorIs(t, err, os.ErrNotExist)
}
