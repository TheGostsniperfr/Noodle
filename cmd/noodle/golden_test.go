package main

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
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
