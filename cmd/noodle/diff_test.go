package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
)

func renderDiff(t *testing.T) string {
	t.Helper()
	icons, err := newIconSet(nil)
	require.NoError(t, err)
	spec := &diagram.Spec{ID: "d", Width: 1000, Height: 600, Diff: true,
		Nodes: []diagram.Node{
			{ID: "live", Kind: "backend", Title: "Live", Change: diagram.Unchanged, X: 40, Y: 40, W: 200, H: 80},
			{ID: "new", Kind: "backend", Title: "New", Status: "planned", Target: "SP3", Change: diagram.Added, X: 400, Y: 40, W: 200, H: 80},
			{ID: "old", Kind: "backend", Title: "Old", Status: "deprecated", Change: diagram.Removed, X: 40, Y: 300, W: 200, H: 80},
		},
		Edges: []diagram.Edge{
			{ID: "e-new", From: "new", To: "live", Kind: "flow", Change: diagram.Added, Path: []diagram.Point{{400, 80}, {240, 80}}},
			{ID: "e-live", From: "live", To: "old", Kind: "flow", Change: diagram.Unchanged, Path: []diagram.Point{{140, 120}, {140, 300}}},
		},
	}
	r := &renderer{spec: spec, th: house.Themes["dark"], icons: icons, cache: map[string]string{}, ports: house.PortBadges(spec)}
	xml, err := r.render()
	require.NoError(t, err)
	return xml
}

func cellStyle(t *testing.T, xml, id string) string {
	t.Helper()
	m := regexp.MustCompile(`<mxCell id="` + id + `" value="[^"]*" style="([^"]*)"`).FindStringSubmatch(xml)
	require.NotNil(t, m, id)
	return m[1]
}

func TestRender_DimsUnchangedItems_InADiff(t *testing.T) {
	t.Parallel()

	xml := renderDiff(t)

	assert.Contains(t, cellStyle(t, xml, "live"), "opacity=30;textOpacity=30")
	assert.Contains(t, cellStyle(t, xml, "e-live"), "opacity=30")
	assert.NotContains(t, cellStyle(t, xml, "new"), "opacity=30")
}

func TestRender_FramesAddedAndRemovedItems_WithTheirChangeColour(t *testing.T) {
	t.Parallel()
	th := house.Themes["dark"]

	xml := renderDiff(t)

	assert.Contains(t, cellStyle(t, xml, "new"), "strokeColor="+th.Diff[diagram.Added]+";strokeWidth=3")
	assert.Contains(t, cellStyle(t, xml, "old"), "strokeColor="+th.Diff[diagram.Removed]+";strokeWidth=3")
	assert.Contains(t, cellStyle(t, xml, "e-new"), "strokeColor="+th.Diff[diagram.Added])
	assert.Contains(t, cellStyle(t, xml, "new__pill"), "fillColor="+th.Diff[diagram.Added])
}

func TestRender_DrawsChangedEdgesAfterUnchangedOnes_InADiff(t *testing.T) {
	t.Parallel()

	xml := renderDiff(t)

	assert.Less(t, regexp.MustCompile(`id="e-live"`).FindStringIndex(xml)[0], regexp.MustCompile(`id="e-new"`).FindStringIndex(xml)[0])
}
