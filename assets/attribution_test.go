package assets_test

import (
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/assets"
)

func embeddedIconFiles(t *testing.T) []string {
	t.Helper()
	entries, err := fs.ReadDir(assets.Icons, "icons")
	require.NoError(t, err)
	var names []string
	for _, e := range entries {
		names = append(names, path.Base(e.Name()))
	}
	require.NotEmpty(t, names)
	return names
}

type attributionRow struct{ source, license string }

func attributionErrors(files []string, page string) []string {
	rows := map[string]attributionRow{}
	var errs []string
	for _, line := range strings.Split(page, "\n") {
		if !strings.HasPrefix(line, "| `") {
			continue
		}
		cells := strings.Split(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(line), "|"), "|"), "|")
		if len(cells) != 3 {
			errs = append(errs, "malformed attribution row: "+line)
			continue
		}
		name := strings.Trim(strings.TrimSpace(cells[0]), "`")
		rows[name] = attributionRow{source: strings.TrimSpace(cells[1]), license: strings.TrimSpace(cells[2])}
	}
	known := map[string]bool{}
	for _, f := range files {
		known[f] = true
		row, ok := rows[f]
		switch {
		case !ok:
			errs = append(errs, f+": no row in ATTRIBUTION.md")
		case !strings.HasPrefix(row.source, "https://"):
			errs = append(errs, f+": row has no https source URL")
		case row.license == "":
			errs = append(errs, f+": row has no license")
		}
	}
	for name := range rows {
		if !known[name] {
			errs = append(errs, name+": row for a file that is not in assets/icons")
		}
	}
	if !strings.Contains(strings.ToLower(page), "trademark") {
		errs = append(errs, "ATTRIBUTION.md does not say logos remain trademarks of their owners")
	}
	return errs
}

func replaceLine(page, prefix, replacement string) string {
	var out []string
	for _, line := range strings.Split(page, "\n") {
		if strings.HasPrefix(line, prefix) {
			if replacement == "" {
				continue
			}
			line = replacement
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func TestAttribution(t *testing.T) {
	raw, err := os.ReadFile("icons/ATTRIBUTION.md")
	require.NoError(t, err)
	page := string(raw)
	files := embeddedIconFiles(t)

	t.Run("repository", func(t *testing.T) {
		assert.Empty(t, attributionErrors(files, page))
	})

	t.Run("missing_row", func(t *testing.T) {
		errs := attributionErrors(files, replaceLine(page, "| `cilium.svg` |", ""))

		require.Len(t, errs, 1)
		assert.Contains(t, errs[0], "cilium.svg")
	})

	t.Run("row_without_url", func(t *testing.T) {
		broken := replaceLine(page, "| `cilium.svg` |", "| `cilium.svg` |  | Apache-2.0 |")

		errs := attributionErrors(files, broken)

		require.Len(t, errs, 1)
		assert.Contains(t, errs[0], "cilium.svg")
	})

	t.Run("row_without_license", func(t *testing.T) {
		broken := replaceLine(page, "| `cilium.svg` |", "| `cilium.svg` | https://example.org/cilium.svg |  |")

		errs := attributionErrors(files, broken)

		require.Len(t, errs, 1)
		assert.Contains(t, errs[0], "cilium.svg")
	})

	t.Run("row_for_unknown_file", func(t *testing.T) {
		errs := attributionErrors(files, page+"| `ghost.svg` | https://example.org | MIT |\n")

		require.Len(t, errs, 1)
		assert.Contains(t, errs[0], "ghost.svg")
	})

	t.Run("no_trademark_notice", func(t *testing.T) {
		errs := attributionErrors(files, strings.ReplaceAll(strings.ReplaceAll(page, "trademarks", "marks"), "trademark", "mark"))

		require.Len(t, errs, 1)
		assert.Contains(t, errs[0], "trademark")
	})
}
