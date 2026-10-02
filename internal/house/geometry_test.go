package house_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
)

func TestPillText_ShowsTheTargetOfPlannedBoxesOnly(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, status, target, want string }{
		{"planned with target", "planned", "SP1", "SP1"},
		{"planned without target", "planned", "", "planned"},
		{"deprecated", "deprecated", "", ""},
		{"live", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := house.PillText(tt.status, tt.target)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNodePillRect_StraddlesTheTopBorderInsideTheBox(t *testing.T) {
	t.Parallel()
	n := diagram.Node{Status: "planned", Target: "SP2", X: 100, Y: 200, W: 200, H: 80}

	pill, ok := house.NodePillRect(n)

	assert.True(t, ok)
	assert.Equal(t, [3]bool{true, true, true}, [3]bool{
		pill.Y < n.Y && pill.Y+pill.H > n.Y,
		pill.X > n.X,
		pill.X+pill.W < n.X+n.W,
	})
}

func TestMatrixLevels_TextMeetsAAOnItsFill(t *testing.T) {
	t.Parallel()
	for _, th := range house.Themes {
		for level, st := range th.Levels {
			t.Run(th.Name+"/"+level, func(t *testing.T) {
				t.Parallel()

				ratio := house.Contrast(st.Text, st.Fill)

				assert.GreaterOrEqual(t, ratio, 4.5, "%s on %s", st.Text, st.Fill)
			})
		}
	}
}
