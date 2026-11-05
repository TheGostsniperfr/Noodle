package main

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var yamlFenceRe = regexp.MustCompile("(?s)```yaml\n(.*?)```")

func parseYAML(t *testing.T, text string) any {
	t.Helper()
	var out any
	require.NoError(t, yaml.Unmarshal([]byte(text), &out))
	return out
}

func TestTutorial_PageShowsTheSystemOfExamplesTutorial(t *testing.T) {
	t.Parallel()
	page, err := os.ReadFile("../../docs/guide/tutorial.md")
	require.NoError(t, err)
	fences := yamlFenceRe.FindAllStringSubmatch(string(page), -1)
	files := []string{"model.yaml", "views/request.yaml", "layouts/request.yaml"}
	require.Len(t, fences, len(files), "docs/guide/tutorial.md must hold the model, the view and the layout, in that order")

	for i, file := range files {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("../../examples/tutorial/" + file)
			require.NoError(t, err)
			// The page puts the system in architecture/demo; the example lives in examples/tutorial.
			onPage := strings.ReplaceAll(fences[i][1], "architecture/demo", "examples/tutorial")

			assert.Equal(t, parseYAML(t, string(source)), parseYAML(t, onPage), "docs/guide/tutorial.md differs from examples/tutorial/%s", file)
		})
	}
}
