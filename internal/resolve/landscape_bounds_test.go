package resolve_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
	"github.com/TheGostsniperfr/Noodle/internal/model"
	"github.com/TheGostsniperfr/Noodle/internal/resolve"
)

func TestLandscape_SideColumnFits(t *testing.T) {
	t.Parallel()
	s, err := model.LoadSystem("testdata/landscape-90")
	require.NoError(t, err)
	require.Empty(t, model.Check(s))

	spec, err := resolve.View(s, "stack")

	require.NoError(t, err)
	canvas := diagram.Rect{X: 0, Y: 0, W: spec.Width, H: spec.Height}
	var side []diagram.Rect
	for _, z := range spec.Zones {
		if strings.HasPrefix(z.ID, "side-") {
			side = append(side, z.Rect())
		}
	}
	for _, n := range spec.Nodes {
		if n.ID == "backup" {
			side = append(side, n.Rect())
		}
	}
	require.Len(t, side, 2, "the side zone and its tile")
	for _, r := range side {
		assert.True(t, canvas.Contains(r), "%v is outside the canvas %v", r, canvas)
	}
	assert.Equal(t, 2400.0, spec.Width, "the fixture reaches the width cap")
	assert.Empty(t, lint.Lint(spec))
}
