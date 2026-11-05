package lint_test

import (
	"bytes"
	"math/rand"
	"os"
	"strings"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
)

// twoBadges is node b with two edges from a landing on its top border, the second
// gap px to the right of the first badge's right edge.
func twoBadges(t *testing.T, gap float64) diagram.Spec {
	t.Helper()
	text := "TCP 443"
	x2 := 400 + house.PortWidth(text) + gap
	s := diagram.Spec{ID: "t", Width: 800, Height: 400,
		Nodes: []diagram.Node{
			{ID: "a", Kind: "backend", Title: "A", X: 300, Y: 40, W: 300, H: 80},
			{ID: "b", Kind: "backend", Title: "B", X: 300, Y: 200, W: 300, H: 80},
		},
		Edges: []diagram.Edge{
			{ID: "e1", From: "a", To: "b", Kind: "flow", Port: text, Path: []diagram.Point{{400, 120}, {400, 200}}},
			{ID: "e2", From: "a", To: "b", Kind: "flow", Port: text, Path: []diagram.Point{{x2, 120}, {x2, 200}}},
		},
	}
	return s
}

func TestLint_PortsOverlap(t *testing.T) {
	t.Parallel()

	t.Run("issue_73", func(t *testing.T) {
		t.Parallel()
		raw, err := os.ReadFile("testdata/ports-73.yaml")
		require.NoError(t, err)
		var spec diagram.Spec
		dec := yaml.NewDecoder(bytes.NewReader(raw))
		dec.KnownFields(true)
		require.NoError(t, dec.Decode(&spec))

		findings := lint.Lint(&spec)

		assert.Equal(t, []lint.Finding{{Where: "e1", Msg: `port badge "TCP 443" overlaps the badge of e2`}}, findings)
	})
	t.Run("3px_apart", func(t *testing.T) {
		t.Parallel()
		spec := twoBadges(t, 3)

		findings := lint.Lint(&spec)

		assert.Equal(t, []lint.Finding{{Where: "e1", Msg: `port badge "TCP 443" overlaps the badge of e2`}}, findings)
	})
	t.Run("4px_apart", func(t *testing.T) {
		t.Parallel()
		spec := twoBadges(t, 4)

		findings := lint.Lint(&spec)

		assert.Empty(t, findings)
	})
	t.Run("shared_port", func(t *testing.T) {
		t.Parallel()
		spec := diagram.Spec{ID: "t", Width: 900, Height: 400,
			Nodes: []diagram.Node{
				{ID: "a", Kind: "backend", Title: "A", X: 300, Y: 40, W: 200, H: 80},
				{ID: "c", Kind: "backend", Title: "C", X: 650, Y: 40, W: 100, H: 80},
				{ID: "b", Kind: "backend", Title: "B", X: 300, Y: 220, W: 300, H: 80},
			},
			Edges: []diagram.Edge{
				{ID: "e1", From: "a", To: "b", Kind: "flow", Port: "TCP 443", Path: []diagram.Point{{400, 120}, {400, 220}}},
				{ID: "e2", From: "c", To: "b", Kind: "flow", Port: "TCP 443", Path: []diagram.Point{{700, 120}, {700, 212}, {400, 212}, {400, 220}}},
			}}

		findings := lint.Lint(&spec)

		assert.Empty(t, findings)
	})
	t.Run("two_ports_on_one_point", func(t *testing.T) {
		t.Parallel()
		spec := diagram.Spec{ID: "t", Width: 900, Height: 400,
			Nodes: []diagram.Node{
				{ID: "a", Kind: "backend", Title: "A", X: 300, Y: 40, W: 200, H: 80},
				{ID: "c", Kind: "backend", Title: "C", X: 650, Y: 40, W: 100, H: 80},
				{ID: "b", Kind: "backend", Title: "B", X: 300, Y: 220, W: 300, H: 80},
			},
			Edges: []diagram.Edge{
				{ID: "e1", From: "a", To: "b", Kind: "flow", Port: "TCP 443", Path: []diagram.Point{{400, 120}, {400, 220}}},
				{ID: "e2", From: "c", To: "b", Kind: "flow", Port: "TCP 8443", Path: []diagram.Point{{700, 120}, {700, 212}, {400, 212}, {400, 220}}},
			}}

		findings := lint.Lint(&spec)

		assert.Equal(t, []lint.Finding{{Where: "e2", Msg: `port "TCP 8443" lands on the same point of b as "TCP 443" of e1`}}, findings)
	})
}

// TestLint_PortsOverlap_Property: for badges placed anywhere on the border of one node,
// the lint reports an overlap exactly when the inflated badges intersect.
func TestLint_PortsOverlap_Property(t *testing.T) {
	t.Parallel()
	node := diagram.Node{ID: "b", Kind: "backend", Title: "B", X: 300, Y: 200, W: 200, H: 120}
	border := func(side, offset uint8, texts uint8) (diagram.Point, diagram.Point, string) {
		text := "TCP " + strings.Repeat("9", 1+int(texts%5))
		switch side % 4 {
		case 0:
			x := node.X + float64(int(offset)%int(node.W))
			return diagram.Point{x, node.Y - 40}, diagram.Point{x, node.Y}, text
		case 1:
			x := node.X + float64(int(offset)%int(node.W))
			return diagram.Point{x, node.Y + node.H + 40}, diagram.Point{x, node.Y + node.H}, text
		case 2:
			y := node.Y + float64(int(offset)%int(node.H))
			return diagram.Point{node.X - 40, y}, diagram.Point{node.X, y}, text
		}
		y := node.Y + float64(int(offset)%int(node.H))
		return diagram.Point{node.X + node.W + 40, y}, diagram.Point{node.X + node.W, y}, text
	}
	property := func(sideA, offA, lenA, sideB, offB, lenB uint8) bool {
		fromA, toA, textA := border(sideA, offA, lenA)
		fromB, toB, textB := border(sideB, offB, lenB)
		if toA == toB {
			return true
		}
		spec := diagram.Spec{ID: "t", Width: 900, Height: 600,
			Nodes: []diagram.Node{node, {ID: "a", Kind: "backend", Title: "A", X: 0, Y: 0, W: 10, H: 10}},
			Edges: []diagram.Edge{
				{ID: "e1", From: "a", To: "b", Kind: "link", Port: textA, Path: []diagram.Point{fromA, toA}},
				{ID: "e2", From: "a", To: "b", Kind: "link", Port: textB, Path: []diagram.Point{fromB, toB}},
			}}
		badges := house.PortBadges(&spec)
		if badges["e1"].ID == badges["e2"].ID {
			return true
		}
		want := badges["e1"].Rect.Inflate(2).Intersects(badges["e2"].Rect.Inflate(2))

		got := false
		for _, f := range lint.Lint(&spec) {
			got = got || strings.Contains(f.Msg, "overlaps the badge of")
		}

		return got == want
	}

	err := quick.Check(property, &quick.Config{MaxCount: 5000, Rand: rand.New(rand.NewSource(73))})

	assert.NoError(t, err)
}
