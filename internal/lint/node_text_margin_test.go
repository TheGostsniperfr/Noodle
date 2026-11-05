package lint_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
)

func TestLint_NodeText_Margin(t *testing.T) {
	t.Parallel()
	tech := strings.Repeat("t", 28)
	longerTech := strings.Repeat("t", 29)
	tests := []struct {
		name     string
		title    string
		tech     string
		expected []lint.Finding
	}{
		{"just_under", strings.Repeat("T", 23), "", nil},
		{"just_over", strings.Repeat("T", 24), "",
			[]lint.Finding{{Where: "a", Msg: "title overflows: 173px text + 14px margin in 182px"}}},
		{"line_just_under", "A", tech, nil},
		{"line_just_over", "A", longerTech,
			[]lint.Finding{{Where: "a", Msg: `"` + longerTech + `" overflows: 175px text + 12px margin in 182px`}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := clean()
			s.Nodes[0].Title, s.Nodes[0].Tech = tt.title, tt.tech

			findings := lint.Lint(s)

			assert.Equal(t, tt.expected, findings)
		})
	}
}

func TestLint_NodeText_MarginHoldsOnBadgeAndIconBoxes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		mutate   func(n *diagram.Node)
		expected []lint.Finding
	}{
		{"a badge lengthens the title", func(n *diagram.Node) { n.Title, n.Badge = strings.Repeat("T", 19), "G1" },
			[]lint.Finding{{Where: "a", Msg: "title overflows: 173px text + 14px margin in 182px"}}},
		{"an icon takes the left padding", func(n *diagram.Node) { n.Icon, n.Title = "k8s-svc", strings.Repeat("T", 19) },
			[]lint.Finding{{Where: "a", Msg: "title overflows: 137px text + 14px margin in 142px"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := clean()
			tt.mutate(&s.Nodes[0])

			findings := lint.Lint(s)

			assert.Equal(t, tt.expected, findings)
		})
	}
}
