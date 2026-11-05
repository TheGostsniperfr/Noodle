package main

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIReference(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../docs/guide/cli.md")
	require.NoError(t, err)

	t.Run("current", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, checkCLIReference(string(raw)))
	})

	t.Run("stale", func(t *testing.T) {
		t.Parallel()
		altered := strings.Replace(string(raw), "-namespace", "-ns", 1)
		require.NotEqual(t, string(raw), altered)

		err := checkCLIReference(altered)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "docs/guide/cli.md")
	})

	t.Run("lists_every_flag_of_every_subcommand", func(t *testing.T) {
		t.Parallel()
		sets := []*flag.FlagSet{newRenderFlags().set, newDiscoverFlags().set, newSetupFlags().set}
		listIcons, _, _ := newListIconsFlags()
		sets = append(sets, listIcons)
		var names []string
		for _, set := range sets {
			set.VisitAll(func(f *flag.Flag) { names = append(names, f.Name) })
		}
		require.GreaterOrEqual(t, len(names), 12)

		for _, name := range names {
			assert.Contains(t, string(raw), "  -"+name, "flag -%s is missing from docs/guide/cli.md", name)
		}
	})
}

func TestHelp_PrintsTheUsageAndTheMarkdownReference(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"plain", nil, fullUsage()},
		{"markdown", []string{"-markdown"}, cliReference()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var out strings.Builder

			err := helpCommand(tt.args, &out)

			require.NoError(t, err)
			assert.Equal(t, tt.want, out.String())
		})
	}

	t.Run("unknown argument", func(t *testing.T) {
		t.Parallel()

		err := helpCommand([]string{"render"}, &strings.Builder{})

		assert.Error(t, err)
	})
}
