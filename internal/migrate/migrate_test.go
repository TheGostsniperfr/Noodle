package migrate

import (
	"bytes"
	"os"
	"testing"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func loadV0(t *testing.T, path string) *diagram.Spec {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var s diagram.Spec
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	require.NoError(t, dec.Decode(&s))
	return &s
}

func TestMigrate_WrittenSystemDrawsTheV0Diagram(t *testing.T) {
	t.Parallel()
	specs := []string{
		"testdata/cnp-runtime.yaml",
		"testdata/platform-overview.yaml",
		"testdata/app-runtime-view.yaml",
		"testdata/repo-map.yaml",
		"testdata/contract-pipeline.yaml",
	}
	for _, path := range specs {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			spec := loadV0(t, path)
			dir := t.TempDir()

			migrated, err := FromV0(spec)
			require.NoError(t, err)
			require.NoError(t, Write(migrated, dir))
			reloaded, err := model.LoadSystem(dir)
			require.NoError(t, err)

			assert.NoError(t, Verify(loadV0(t, path), reloaded))
		})
	}
}

func TestMigrate_Fails_WhenTheSpecHasNoID(t *testing.T) {
	_, err := FromV0(&diagram.Spec{})

	assert.ErrorContains(t, err, "no id")
}

func TestWrite_Refuses_ToOverwriteAModel(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(dir+"/model.yaml", nil, 0o644))

	err := Write(&model.System{Model: &model.Model{}}, dir)

	assert.ErrorContains(t, err, "already holds a model.yaml")
}
