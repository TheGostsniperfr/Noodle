package resolve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
	"github.com/TheGostsniperfr/Noodle/internal/model"
	"github.com/TheGostsniperfr/Noodle/internal/resolve"
)

func smallMatrix(t *testing.T) *diagram.Spec {
	t.Helper()
	s, err := model.LoadSystem("testdata/matrix-small")
	require.NoError(t, err)
	require.Empty(t, model.Check(s))
	spec, err := resolve.View(s, "diff")
	require.NoError(t, err)
	return spec
}

func TestMatrix_CellsCarryLevelAndChange(t *testing.T) {
	t.Parallel()
	spec := smallMatrix(t)
	// rows: dev, svc; columns: vault-a, apps, dns
	got := map[string]string{}
	rows := []string{"dev", "svc"}
	cols := []string{"vault-a", "apps", "dns"}
	for i, c := range spec.Matrix.Cells {
		got[rows[i/3]+"/"+cols[i%3]] = c.Level + "|" + c.Change + "|" + c.Before + "|" + c.Phase + "|" + c.Note
	}

	assert.Equal(t, map[string]string{
		"dev/vault-a": "read|changed|write|P3|was write",
		"dev/apps":    "|removed|read|P3|removed",
		"dev/dns":     "||||",
		"svc/vault-a": "|removed|admin|P2|removed",
		"svc/apps":    "write|added||P2|",
		"svc/dns":     "write|changed|write|P2|rescoped",
	}, got)
}

func TestMatrix_CountsWhatEachRowCanChange(t *testing.T) {
	t.Parallel()
	spec := smallMatrix(t)

	var counts []string
	for _, c := range spec.Matrix.Counts {
		counts = append(counts, c.Lines[0])
	}

	assert.Equal(t, []string{"1 → 0", "2 → 2"}, counts)
}

func TestMatrix_PassesTheLint(t *testing.T) {
	t.Parallel()
	spec := smallMatrix(t)

	findings := lint.Lint(spec)

	assert.Empty(t, findings)
}

func TestMatrix_LegendExplainsChangesOnlyInADiff(t *testing.T) {
	t.Parallel()
	spec := smallMatrix(t)

	assert.Len(t, spec.Matrix.LegendTitles, 2)
	assert.Len(t, spec.Matrix.Legend, 9)
}
