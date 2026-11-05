package resolve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
)

func TestBoxWidth_KeepsTheMargin(t *testing.T) {
	t.Parallel()
	seq := loadSequence(t, "../../examples/regression-72", "participants")
	require.Len(t, seq.Frame.Nodes, 2)
	for _, n := range seq.Frame.Nodes {
		t.Run(n.ID, func(t *testing.T) {
			t.Parallel()
			avail := n.W - 18

			titleFits := house.FitsWithMargin(house.ExportTextWidth(n.Title, house.TitleFontSize), avail, house.TitleFontSize)
			techFits := house.FitsWithMargin(house.ExportTextWidth(n.Tech, house.SubFontSize), avail, house.SubFontSize)

			assert.True(t, titleFits)
			assert.True(t, techFits)
		})
	}
	t.Run("main_box_fails", func(t *testing.T) {
		t.Parallel()
		tech := seq.Frame.Nodes[0].Tech

		fits := house.FitsWithMargin(house.ExportTextWidth(tech, house.SubFontSize), 190-18, house.SubFontSize)

		assert.False(t, fits, "the 190px box the old measure gave")
	})
	t.Run("the resolved frame passes the lint", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, lint.Lint(&seq.Frame))
	})
}
