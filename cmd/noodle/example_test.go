package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleManifests = "examples/discovery-k8s/manifests"
	exampleGolden    = "examples/discovery-k8s/discovered/k8s-manifests.yaml"
)

var exampleArgs = []string{exampleManifests, "-observed-at", "2026-11-04T00:00:00Z"}

func TestDiscoverExample(t *testing.T) {
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	golden, err := os.ReadFile(filepath.Join(root, exampleGolden))
	require.NoError(t, err)
	list, err := os.ReadFile(filepath.Join(root, "examples/discovery-k8s/list.yaml"))
	require.NoError(t, err)

	t.Run("manifests-first", func(t *testing.T) {
		t.Chdir(root)

		stdout, _, err := runDiscover(t, "", exampleArgs...)

		require.NoError(t, err)
		assert.Equal(t, string(golden), stdout)
	})

	t.Run("manifests-second", func(t *testing.T) {
		t.Chdir(root)

		stdout, _, err := runDiscover(t, "", exampleArgs...)

		require.NoError(t, err)
		assert.Equal(t, string(golden), stdout)
	})

	t.Run("kubectl-list", func(t *testing.T) {
		staged := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(staged, exampleManifests), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(staged, exampleManifests, "platform.yaml"), list, 0o644))
		t.Chdir(staged)

		stdout, _, err := runDiscover(t, "", exampleArgs...)

		require.NoError(t, err)
		assert.Equal(t, string(golden), stdout)
	})
}

func TestDrift(t *testing.T) {
	want, err := os.ReadFile("../../examples/discovery-k8s/drift.txt")
	require.NoError(t, err)
	var stdout bytes.Buffer

	err = driftCommand([]string{"../../examples/discovery-k8s"}, &stdout)

	require.NoError(t, err, "drift is a report: it exits 0")
	assert.Equal(t, string(want), stdout.String())
	assert.Contains(t, string(want), "k8s:prod/deployment/worker  examples/discovery-k8s/manifests/platform.yaml:82")
}

func TestDrift_Fails(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"without a directory", nil, "usage: noodle drift DIR"},
		{"on a system without fragments", []string{"../../examples/landscape"}, "has no fragment"},
		{"on a directory that is no system", []string{t.TempDir()}, "model.yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := driftCommand(tt.args, &bytes.Buffer{})

			assert.ErrorContains(t, err, tt.want)
		})
	}
}
