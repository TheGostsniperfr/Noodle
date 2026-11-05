package lint_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
)

// scaffold is a clean diagram of nodes in two columns, row i holding nodes 2i and 2i+1,
// with one edge of kind edgeKinds[i] from the left node to the right one. Cards and
// notes are added by the case, so only the card rule has anything to say.
func scaffold(nodeKinds, edgeKinds []string) diagram.Spec {
	s := diagram.Spec{ID: "t", Width: 1600, Height: 900}
	for i, kind := range nodeKinds {
		s.Nodes = append(s.Nodes, diagram.Node{
			ID: fmt.Sprintf("n%d", i), Kind: kind, Title: "N", X: 40 + float64(i%2)*360, Y: 100 + float64(i/2)*100, W: 100, H: 60,
		})
	}
	for i, kind := range edgeKinds {
		y := 130 + float64(i)*100
		s.Edges = append(s.Edges, diagram.Edge{
			ID: fmt.Sprintf("e%d", i), From: fmt.Sprintf("n%d", 2*i), To: fmt.Sprintf("n%d", 2*i+1), Kind: kind,
			Path: []diagram.Point{{140, y}, {400, y}},
		})
	}
	return s
}

func withCards(s diagram.Spec, cards ...diagram.Card) diagram.Spec {
	s.Cards = cards
	return s
}

func TestLint_CardText(t *testing.T) {
	t.Parallel()
	lines := []string{"one", "two", "three"}
	tests := []struct {
		name     string
		spec     diagram.Spec
		expected []lint.Finding
	}{
		{"legend_420px_ref_row",
			withCards(scaffold([]string{"backend", "backend"}, []string{"link"}),
				diagram.Card{ID: "legend", Title: "Legend", Legend: true, X: 40, Y: 400, W: 420, H: 300}),
			[]lint.Finding{{Where: "legend", Msg: `"object reference, not traffic" needs 183px, its cell is 138px wide`}}},
		{"legend_120px_six_kinds",
			withCards(scaffold([]string{"frontend", "backend", "database", "cloud", "security", "bus"}, nil),
				diagram.Card{ID: "legend", Title: "Legend", Legend: true, X: 40, Y: 500, W: 700, H: 120}),
			[]lint.Finding{{Where: "legend", Msg: "text needs 318px height, card is 120px"}}},
		{"listed_card_too_tall",
			withCards(scaffold(nil, nil), diagram.Card{ID: "gaps", Title: "Gaps", Lines: lines, X: 40, Y: 400, W: 400, H: 80}),
			[]lint.Finding{{Where: "gaps", Msg: "text needs 111px height, card is 80px"}}},
		{"listed_line_too_wide",
			withCards(scaffold(nil, nil), diagram.Card{ID: "gaps", Title: "Gaps", Lines: []string{"short", strings.Repeat("a", 44)}, X: 40, Y: 400, W: 200, H: 300}),
			[]lint.Finding{{Where: "gaps", Msg: `"` + strings.Repeat("a", 44) + `" needs 290px, its cell is 164px wide`}}},
		{"note_too_wide",
			func() diagram.Spec {
				s := scaffold(nil, nil)
				s.Notes = []diagram.Note{{ID: "n", Text: "short<br>" + strings.Repeat("b", 40), X: 40, Y: 400, W: 100, H: 60}}
				return s
			}(),
			[]lint.Finding{{Where: "n", Msg: `"` + strings.Repeat("b", 40) + `" needs 240px, the note is 100px wide`}}},
		{"fitting_cards",
			func() diagram.Spec {
				s := withCards(scaffold([]string{"frontend", "backend", "database", "cloud"}, []string{"flow", "auth"}),
					diagram.Card{ID: "legend", Title: "Legend", Legend: true, X: 40, Y: 400, W: 700, H: 266},
					diagram.Card{ID: "gaps", Title: "Gaps", Lines: lines, X: 800, Y: 400, W: 400, H: 120})
				s.Notes = []diagram.Note{{ID: "n", Text: "a note that wraps onto several lines when narrow", X: 1300, Y: 400, W: 120, H: 100}}
				return s
			}(),
			nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings := lint.Lint(&tt.spec)

			assert.Equal(t, tt.expected, findings)
		})
	}
}

func frozenSpecs(t *testing.T) map[string]diagram.Spec {
	t.Helper()
	paths, err := filepath.Glob("../../cmd/noodle/testdata/cardlayout/*.spec.json")
	require.NoError(t, err)
	require.NotEmpty(t, paths)
	specs := map[string]diagram.Spec{}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		var spec diagram.Spec
		require.NoError(t, json.Unmarshal(raw, &spec))
		specs[filepath.Base(path)] = spec
	}
	return specs
}

// minWidthOf is the card width at which a row's cell is just wide enough for its text.
// Every cell width is affine in the card width, so two layouts give the slope.
func minWidthOf(t *testing.T, spec *diagram.Spec, card diagram.Card, index int) float64 {
	t.Helper()
	at := func(w float64) house.CardRow {
		c := card
		c.W = w
		return house.CardLayout(spec, c).Rows[index]
	}
	a, b := at(1000), at(1100)
	slope := (b.Room() - a.Room()) / 100
	require.NotZero(t, slope)
	_, need := a.Widest()
	return 1000 + (need-a.Room())/slope
}

func TestLint_CardText_EveryLayoutRowIsMeasured(t *testing.T) {
	t.Parallel()
	for name, spec := range frozenSpecs(t) {
		for _, card := range spec.Cards {
			t.Run(name+"/"+card.ID+"/width", func(t *testing.T) {
				t.Parallel()
				rows := house.CardLayout(&spec, card).Rows
				need := make([]float64, len(rows))
				widest := 0.0
				for i := range rows {
					need[i] = minWidthOf(t, &spec, card, i)
					widest = max(widest, need[i])
				}
				narrow := card
				narrow.W = widest - 1
				tooNarrow := 0
				for _, n := range need {
					if n > narrow.W {
						tooNarrow++
					}
				}
				narrow.H = 1e6
				probe := spec
				probe.Cards = []diagram.Card{narrow}

				var widthFindings int
				for _, f := range lint.Lint(&probe) {
					if f.Where == card.ID && strings.Contains(f.Msg, "cell is") {
						widthFindings++
					}
				}

				assert.Equal(t, tooNarrow, widthFindings)
			})
			t.Run(name+"/"+card.ID+"/height", func(t *testing.T) {
				t.Parallel()
				short := card
				short.H = house.CardLayout(&spec, card).Needed - 1
				probe := spec
				probe.Cards = []diagram.Card{short}

				var heightFindings int
				for _, f := range lint.Lint(&probe) {
					if f.Where == card.ID && strings.Contains(f.Msg, "height") {
						heightFindings++
					}
				}

				assert.Equal(t, 1, heightFindings)
			})
		}
	}
}

func TestLint_CardText_EveryNoteIsMeasured(t *testing.T) {
	t.Parallel()
	measured := 0
	for name, spec := range frozenSpecs(t) {
		for _, note := range spec.Notes {
			measured++
			t.Run(name+"/"+note.ID, func(t *testing.T) {
				t.Parallel()
				_, widest := house.NoteRow(note).Widest()
				narrow := note
				narrow.W = widest - 1
				probe := spec
				probe.Notes = []diagram.Note{narrow}

				var findings int
				for _, f := range lint.Lint(&probe) {
					if f.Where == note.ID && strings.Contains(f.Msg, "the note is") {
						findings++
					}
				}

				assert.Equal(t, 1, findings)
			})
		}
	}
	require.NotZero(t, measured, "the frozen specs hold at least one note")
}
