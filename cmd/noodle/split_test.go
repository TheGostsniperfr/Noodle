package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderCommand_WritesTheSplitCNPExample_WithFlagsAfterTheDirectory(t *testing.T) {
	out := filepath.Join(t.TempDir(), "runtime.drawio")

	err := renderCommand([]string{"../../examples/cnp-runtime", "--view", "runtime", "-o", out})

	require.NoError(t, err)
	raw, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `id="e-3-connector-envoy"`)
}

func TestRenderCommand_Fails_OnAnUnknownView(t *testing.T) {
	err := renderCommand([]string{"../../examples/cnp-runtime", "-view", "nope"})

	assert.ErrorContains(t, err, `no view "nope"`)
}
