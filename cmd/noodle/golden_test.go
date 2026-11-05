package main

import (
	"encoding/json"
	"flag"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "rewrite golden files and committed .drawio files")

// cardFragment is the header and every card of a frozen spec, as one theme draws them.
func cardFragment(t *testing.T, spec *diagram.Spec, theme string) string {
	t.Helper()
	th, err := house.ThemeByName(theme)
	require.NoError(t, err)
	set, err := newIconSet(nil)
	require.NoError(t, err)
	r := &renderer{spec: spec, th: th, icons: set, cache: map[string]string{}, ports: house.PortBadges(spec)}
	r.header()
	for _, c := range spec.Cards {
		r.card(c)
	}
	return r.b.String()
}

func TestRender_CardLayoutGolden(t *testing.T) {
	specs, err := filepath.Glob("testdata/cardlayout/*.spec.json")
	require.NoError(t, err)
	require.NotEmpty(t, specs)
	for _, path := range specs {
		name := strings.TrimSuffix(filepath.Base(path), ".spec.json")
		for _, theme := range []string{"dark", "light"} {
			t.Run(name+"/"+theme, func(t *testing.T) {
				raw, err := os.ReadFile(path)
				require.NoError(t, err)
				var spec diagram.Spec
				require.NoError(t, json.Unmarshal(raw, &spec))
				golden := filepath.Join("testdata/cardlayout", name+"."+theme+".golden.xml")

				got := cardFragment(t, &spec, theme)

				if *update {
					require.NoError(t, os.WriteFile(golden, []byte(got), 0o644))
				}
				want, err := os.ReadFile(golden)
				require.NoError(t, err)
				assert.Equal(t, string(want), got)
			})
		}
	}
}

func TestRender_EveryDrawnLegendCellIsALayoutRow(t *testing.T) {
	specs, err := filepath.Glob("testdata/cardlayout/*.spec.json")
	require.NoError(t, err)
	idRe := regexp.MustCompile(`<mxCell id="(legend-[^"]*)"`)
	for _, path := range specs {
		t.Run(filepath.Base(path), func(t *testing.T) {
			raw, err := os.ReadFile(path)
			require.NoError(t, err)
			var spec diagram.Spec
			require.NoError(t, json.Unmarshal(raw, &spec))
			allowed := map[string]bool{}
			for _, c := range spec.Cards {
				for _, row := range house.CardLayout(&spec, c).Rows {
					if row.ID == "" {
						continue
					}
					for _, id := range []string{row.ID, row.ID + "-text", row.ID + "-0", row.ID + "-1", row.ID + "-2"} {
						allowed[id] = true
					}
				}
			}

			drawn := idRe.FindAllStringSubmatch(cardFragment(t, &spec, "dark"), -1)

			for _, m := range drawn {
				assert.True(t, allowed[m[1]], "%s is drawn but is not a CardLayout row", m[1])
			}
		})
	}
}

func TestRender_EveryCardRowIsDrawn(t *testing.T) {
	specs, err := filepath.Glob("testdata/cardlayout/*.spec.json")
	require.NoError(t, err)
	for _, path := range specs {
		t.Run(filepath.Base(path), func(t *testing.T) {
			raw, err := os.ReadFile(path)
			require.NoError(t, err)
			var spec diagram.Spec
			require.NoError(t, json.Unmarshal(raw, &spec))
			th, err := house.ThemeByName("dark")
			require.NoError(t, err)
			set, err := newIconSet(nil)
			require.NoError(t, err)
			r := &renderer{spec: &spec, th: th, icons: set, cache: map[string]string{}, ports: house.PortBadges(&spec)}
			for _, c := range spec.Cards {
				r.b.Reset()
				r.card(c)
				drawn := r.b.String()
				for _, row := range house.CardLayout(&spec, c).Rows {
					switch row.Kind {
					case house.RowTitle:
						assert.Contains(t, drawn, xmlAttr(html.EscapeString(row.Text)), "%s title", c.ID)
					case house.RowLine:
						assert.Contains(t, drawn, xmlAttr(r.markup(row.Text)), "%s line %q", c.ID, row.Text)
					}
				}
			}
		})
	}
}
