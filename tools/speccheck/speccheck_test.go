package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const constitution = "| **C01** | Model first. | Vision |\n| **C02** | Arrows. | ADR-0003 |\n"

const fullSpec = `# Spec 001 · Demo

- Mode: **full**
- Decisions it builds on: ADR-0001
- Invariants it touches: C01

## Requirements

- **FR-001** WHEN a thing happens THE SYSTEM SHALL do it.
  *Scenario:* WHEN input THEN output.
- **FR-002** THE SYSTEM SHALL always log.
  *Scenario:* WHEN started THEN a line is printed.

## Out of scope
`

const fullTasks = "- [ ] **T01** first (FR-001). Flip **D-01**.\n- [ ] **T02** second (FR-002). Flip **D-02**.\n"

const fullContract = "## D-01\n| 1 | x | y | z |\n## D-02\n| 1 | x | y | z |\n"

const featuresJSON = `{"phases": [{"id": "d", "spec": "specs/001-demo/spec.md", "features": [
  {"id": "D-01", "description": "one", "verify": "go test", "passes": false, "covers": ["FR-001"]},
  {"id": "D-02", "description": "two", "verify": "go test", "passes": false, "covers": ["FR-002"]}
]}], "backlog": []}`

// fixture writes a repository holding one full-mode spec that passes every rule; each
// test case breaks exactly one file.
func fixture(t *testing.T, override map[string]string) *Repo {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"docs/CONSTITUTION.md":       constitution,
		"docs/features.json":         featuresJSON,
		"docs/adr/0001-first.md":     "# ADR-0001 · First\n\n- Status: Accepted\n",
		"docs/adr/README.md":         "<!-- adr-index:start -->\n| ADR | Title | Status |\n|---|---|---|\n| [0001](0001-first.md) | First | Accepted |\n<!-- adr-index:end -->\n",
		"specs/001-demo/spec.md":     fullSpec,
		"specs/001-demo/tasks.md":    fullTasks,
		"specs/001-demo/contract.md": fullContract,
		"Makefile":                   "init: check ## Smoke test\ncheck: ## Every check\n\tgo test\nVAR := x ## not a target\n",
		"docs/COMMANDS.md":           "| `make init` | a |\n| `make check` | b |\n",
	}
	for k, v := range override {
		files[k] = v
	}
	for name, content := range files {
		if content == "" {
			continue
		}
		p := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	}
	return &Repo{Root: root}
}

func messages(fs []Finding) []string {
	var out []string
	for _, f := range fs {
		out = append(out, f.String())
	}
	return out
}

func TestSpec_ReportsEachBrokenRule(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		override map[string]string
		want     string
	}{
		{"valid full-mode spec", nil, ""},
		{"unknown invariant", map[string]string{"specs/001-demo/spec.md": strings.Replace(fullSpec, "C01", "C99", 1)},
			"specs/001-demo/spec.md: cites C99, which is not in docs/CONSTITUTION.md"},
		{"unknown ADR", map[string]string{"specs/001-demo/spec.md": strings.Replace(fullSpec, "ADR-0001", "ADR-0042", 1)},
			"specs/001-demo/spec.md: cites ADR-0042, which does not exist"},
		{"requirement without SHALL", map[string]string{"specs/001-demo/spec.md": strings.Replace(fullSpec, "SHALL do it", "does it", 1)},
			"specs/001-demo/spec.md: FR-001 is not EARS: no SHALL"},
		{"requirement without scenario", map[string]string{"specs/001-demo/spec.md": strings.Replace(fullSpec, "  *Scenario:* WHEN input THEN output.\n", "", 1)},
			"specs/001-demo/spec.md: FR-001 has no Scenario with WHEN and THEN"},
		{"requirement named by no task", map[string]string{"specs/001-demo/tasks.md": strings.Replace(fullTasks, "(FR-002)", "", 1)},
			"specs/001-demo/tasks.md: FR-002 is named by no task"},
		{"requirement covered by no feature", map[string]string{"docs/features.json": strings.Replace(featuresJSON, `["FR-002"]`, `[]`, 1)},
			"docs/features.json: FR-002 of specs/001-demo/spec.md is covered by no feature of phase d"},
		{"feature covers an undeclared requirement", map[string]string{"docs/features.json": strings.Replace(featuresJSON, `["FR-002"]`, `["FR-002", "FR-009"]`, 1)},
			"docs/features.json: D-02 covers FR-009, which specs/001-demo/spec.md does not declare"},
		{"feature without verify", map[string]string{"docs/features.json": strings.Replace(featuresJSON, `"verify": "go test", "passes": false, "covers": ["FR-002"]`, `"passes": false, "covers": ["FR-002"]`, 1)},
			"docs/features.json: D-02 has no verify step"},
		{"feature with an empty contract section", map[string]string{"specs/001-demo/contract.md": "## D-01\n| 1 | x | y | z |\n## D-02\n"},
			"specs/001-demo/contract.md: no criteria for D-02"},
		{"feature named only in contract prose", map[string]string{"specs/001-demo/contract.md": "Order: D-01, D-02.\n## D-01\n| 1 | x | y | z |\n"},
			"specs/001-demo/contract.md: no criteria for D-02"},
		{"requirement named on a task's continuation line", map[string]string{"specs/001-demo/tasks.md": strings.Replace(fullTasks, " (FR-002).", "\n  and more (FR-002).", 1)}, ""},
		{"criterion only inside a fenced example", map[string]string{"specs/001-demo/contract.md": "## D-01\n| 1 | x | y | z |\n```\n## D-02\n| 1 | x | y | z |\n```\n"},
			"specs/001-demo/contract.md: no criteria for D-02"},
		{"requirement named only in tasks prose", map[string]string{"specs/001-demo/tasks.md": strings.Replace(fullTasks, "(FR-002)", "", 1) + "Note: FR-002 comes later.\n"},
			"specs/001-demo/tasks.md: FR-002 is named by no task"},
		{"missing contract", map[string]string{"specs/001-demo/contract.md": ""},
			"specs/001-demo/contract.md: a full-mode spec needs a contract"},
		{"task flips an unknown feature", map[string]string{"specs/001-demo/tasks.md": fullTasks + "- [ ] **T03** Flip **D-07**.\n"},
			"specs/001-demo/tasks.md: flips D-07, which is not in docs/features.json"},
		{"unknown invariant in the tasks", map[string]string{"specs/001-demo/tasks.md": fullTasks + "Keeps C77.\n"},
			"specs/001-demo/tasks.md: cites C77, which is not in docs/CONSTITUTION.md"},
		{"unknown invariant in the plan", map[string]string{"specs/001-demo/plan.md": "Touches C42.\n"},
			"specs/001-demo/plan.md: cites C42, which is not in docs/CONSTITUTION.md"},
		{"unknown ADR in the contract", map[string]string{"specs/001-demo/contract.md": fullContract + "See ADR-0077.\n"},
			"specs/001-demo/contract.md: cites ADR-0077, which does not exist"},
		{"scenario without THEN", map[string]string{"specs/001-demo/spec.md": strings.Replace(fullSpec, "WHEN input THEN output.", "WHEN input, output.", 1)},
			"specs/001-demo/spec.md: FR-001 has no Scenario with WHEN and THEN"},
		{"requirement covered only by another phase", map[string]string{"docs/features.json": strings.Replace(
			strings.Replace(featuresJSON, `["FR-002"]`, `[]`, 1), `]}], "backlog"`,
			`]}, {"id": "e", "spec": "specs/009-x/spec.md", "features": [{"id": "E-01", "description": "e", "verify": "v", "passes": false, "covers": ["FR-002"]}]}], "backlog"`, 1)},
			"docs/features.json: FR-002 of specs/001-demo/spec.md is covered by no feature of phase d"},
		{"citation quoted in a fenced block", map[string]string{"specs/001-demo/plan.md": "```\ncites C99 and ADR-0099\n```\n"}, ""},
		{"citation quoted in inline code", map[string]string{"specs/001-demo/plan.md": "The message is `cites C99 and ADR-0099`.\n"}, ""},
		{"scenario without WHEN", map[string]string{"specs/001-demo/spec.md": strings.Replace(fullSpec, "WHEN input THEN output.", "input THEN output.", 1)},
			"specs/001-demo/spec.md: FR-001 has no Scenario with WHEN and THEN"},
		{"unknown invariant in the contract", map[string]string{"specs/001-demo/contract.md": fullContract + "Keeps C77.\n"},
			"specs/001-demo/contract.md: cites C77, which is not in docs/CONSTITUTION.md"},
		{"unknown ADR in the tasks", map[string]string{"specs/001-demo/tasks.md": fullTasks + "Uses ADR-0077.\n"},
			"specs/001-demo/tasks.md: cites ADR-0077, which does not exist"},
		{"unknown ADR in the plan", map[string]string{"specs/001-demo/plan.md": "Uses ADR-0077.\n"},
			"specs/001-demo/plan.md: cites ADR-0077, which does not exist"},
		{"spec with no phase", map[string]string{"docs/features.json": strings.Replace(featuresJSON, "001-demo", "002-other", 1)},
			"specs/001-demo/spec.md: no phase in docs/features.json points at this spec"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := fixture(t, tt.override)

			findings, err := repo.CheckSpecs(nil)

			require.NoError(t, err)
			if tt.want == "" {
				assert.Empty(t, messages(findings))
			} else {
				assert.Equal(t, []string{tt.want}, messages(findings))
			}
		})
	}
}

func TestSpec_LegacySpecIsNotHeldToFullModeRules(t *testing.T) {
	t.Parallel()
	legacy := "# Spec 001 · Old\n\n- Status: Done\n\n## Requirements\n\n- **FR-001** prose without the keyword.\n"
	repo := fixture(t, map[string]string{"specs/001-demo/spec.md": legacy, "specs/001-demo/contract.md": ""})

	findings, err := repo.CheckSpecs(nil)

	require.NoError(t, err)
	assert.Empty(t, messages(findings))
}

const newADR = `---
status: accepted
date: 2026-11-04
deciders: [Brian Perret]
---

# ADR-0002 · Second

## Considered options

1. **A**, one.
2. **B**, two.

## Decision outcome

Chosen: A.

### Confirmation

A test.
`

func TestADR_ChecksFrontmatterADRs(t *testing.T) {
	t.Parallel()
	index := "<!-- adr-index:start -->\n| ADR | Title | Status |\n|---|---|---|\n| [0001](0001-first.md) | First | Accepted |\n| [0002](0002-second.md) | Second | Accepted |\n<!-- adr-index:end -->\n"
	tests := []struct {
		name string
		adr  string
		want string
	}{
		{"valid", newADR, ""},
		{"no status", strings.Replace(newADR, "status: accepted\n", "", 1), `status "" is not one of proposed, accepted, rejected, deprecated, superseded`},
		{"no date", strings.Replace(newADR, "date: 2026-11-04\n", "", 1), "frontmatter needs date and deciders"},
		{"unknown status", strings.Replace(newADR, "status: accepted", "status: maybe", 1), `status "maybe" is not one of proposed, accepted, rejected, deprecated, superseded`},
		{"no deciders", strings.Replace(newADR, "deciders: [Brian Perret]\n", "", 1), "frontmatter needs date and deciders"},
		{"one option", strings.Replace(newADR, "2. **B**, two.\n", "", 1), "needs at least two considered options"},
		{"no confirmation", strings.Replace(newADR, "### Confirmation\n\nA test.\n", "", 1), "needs a Confirmation section: what proves the decision is kept"},
		{"empty confirmation", strings.Replace(newADR, "A test.\n", "", 1), "needs a Confirmation section: what proves the decision is kept"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := fixture(t, map[string]string{"docs/adr/0002-second.md": tt.adr, "docs/adr/README.md": index})

			_, findings, err := repo.loadADRs()

			require.NoError(t, err)
			if tt.want == "" {
				assert.Empty(t, messages(findings))
				return
			}
			assert.Equal(t, []string{"docs/adr/0002-second.md: " + tt.want}, messages(findings))
		})
	}
}

func TestADR_ReportsStaleIndexAndRewritesIt(t *testing.T) {
	t.Parallel()
	repo := fixture(t, map[string]string{"docs/adr/0002-second.md": newADR})

	stale, err := repo.CheckADRs(false)
	require.NoError(t, err)
	_, err = repo.CheckADRs(true)
	require.NoError(t, err)
	fresh, err := repo.CheckADRs(false)
	require.NoError(t, err)

	assert.Equal(t, []string{"docs/adr/README.md: index is stale: run make adr-index"}, messages(stale))
	assert.Empty(t, messages(fresh))
}

func TestADR_LegacyADRParsesFromItsStatusLine(t *testing.T) {
	t.Parallel()
	repo := fixture(t, nil)

	findings, err := repo.CheckADRs(false)

	require.NoError(t, err)
	assert.Empty(t, messages(findings))
}

func TestCommands_ReportsBothDirections(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		doc  string
		want []string
	}{
		{"same targets", "| `make init` | a |\n| `make check` | b |\n", nil},
		{"target with no row", "| `make init` | a |\n", []string{"docs/COMMANDS.md: no row for make check"}},
		{"row with no target", "| `make init` | a |\n| `make check` | b |\n| `make ghost` | c |\n",
			[]string{"Makefile: no documented target ghost (COMMANDS.md lists it; add it with a ## comment)"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := fixture(t, map[string]string{"docs/COMMANDS.md": tt.doc})

			findings, err := repo.CheckCommands()

			require.NoError(t, err)
			assert.Equal(t, tt.want, messages(findings))
		})
	}
}

// gitFixture commits the fixture as the base, then applies changes on top, uncommitted.
func gitFixture(t *testing.T, base, changes map[string]string) *Repo {
	t.Helper()
	repo := fixture(t, base)
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"-c", "user.email=t@t", "-c", "user.name=t", "commit", "-qm", "base"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo.Root
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
	for name, content := range changes {
		p := filepath.Join(repo.Root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	}
	return repo
}

func TestFeatures_GatesNewPassesOnEvidence(t *testing.T) {
	t.Parallel()
	flipped := strings.Replace(featuresJSON, `"passes": false, "covers": ["FR-001"]`, `"passes": true, "covers": ["FR-001"]`, 1)
	quickSpec := strings.Replace(fullSpec, "**full**", "**quick**", 1)
	evidence := func(verdict, evaluator string) string {
		return "# Evidence · D-01\n\n- Feature: D-01\n- Verdict: **" + verdict + "**\n- Evaluator: " + evaluator + "\n"
	}
	tests := []struct {
		name    string
		changes map[string]string
		want    string
	}{
		{"no evidence", map[string]string{"docs/features.json": flipped},
			"docs/features.json: D-01 passes without evidence: no specs/001-demo/evidence/D-01.md (make evaluate FEATURE=D-01)"},
		{"needs work", map[string]string{"docs/features.json": flipped, "specs/001-demo/evidence/D-01.md": evidence("NEEDS_WORK", "evaluator (opus)")},
			"docs/features.json: D-01 passes without evidence: specs/001-demo/evidence/D-01.md has no '- Verdict: PASS' line"},
		{"evaluator pass", map[string]string{"docs/features.json": flipped, "specs/001-demo/evidence/D-01.md": evidence("PASS", "evaluator (opus)")}, ""},
		{"self pass in full mode", map[string]string{"docs/features.json": flipped, "specs/001-demo/evidence/D-01.md": evidence("PASS", "self")},
			"docs/features.json: D-01 passes without evidence: specs/001-demo/evidence/D-01.md must come from the evaluator agent: the spec is in full mode"},
		{"full mode without an Evaluator line", map[string]string{"docs/features.json": flipped, "specs/001-demo/evidence/D-01.md": "# Evidence · D-01\n\n- Verdict: **PASS**\n"},
			"docs/features.json: D-01 passes without evidence: specs/001-demo/evidence/D-01.md must come from the evaluator agent: the spec is in full mode"},
		{"self pass in quick mode", map[string]string{"docs/features.json": flipped, "specs/001-demo/spec.md": quickSpec, "specs/001-demo/evidence/D-01.md": evidence("PASS", "self")}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := gitFixture(t, nil, tt.changes)

			findings, err := repo.CheckFeatures("HEAD", false)

			require.NoError(t, err)
			if tt.want == "" {
				assert.Empty(t, messages(findings))
			} else {
				assert.Equal(t, []string{tt.want}, messages(findings))
			}
		})
	}
}

func TestFeatures_KeepsHistory(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		current string
		want    string
	}{
		{"removed", strings.Replace(featuresJSON, `,
  {"id": "D-02", "description": "two", "verify": "go test", "passes": false, "covers": ["FR-002"]}`, ``, 1), "docs/features.json: D-02: removed"},
		{"description changed", strings.Replace(featuresJSON, `"description": "one"`, `"description": "uno"`, 1), "docs/features.json: D-01: description changed"},
		{"verify changed", strings.Replace(featuresJSON, `"verify": "go test", "passes": false, "covers": ["FR-001"]`, `"verify": "go vet", "passes": false, "covers": ["FR-001"]`, 1), "docs/features.json: D-01: verify step changed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := gitFixture(t, nil, map[string]string{"docs/features.json": tt.current})

			findings, err := repo.CheckFeatures("HEAD", false)
			allowed, allowErr := repo.CheckFeatures("HEAD", true)

			require.NoError(t, err)
			require.NoError(t, allowErr)
			assert.Equal(t, []string{tt.want}, messages(findings))
			assert.Empty(t, messages(allowed))
		})
	}
}

func TestLocate_FindsTheSpecOfAFeature(t *testing.T) {
	t.Parallel()
	repo := fixture(t, nil)

	dir, err := repo.Locate("D-02")
	_, missing := repo.Locate("D-99")

	require.NoError(t, err)
	assert.Equal(t, filepath.Join("specs", "001-demo"), dir)
	assert.Error(t, missing)
}

func TestCommands_ReportsATargetWithoutDescription(t *testing.T) {
	t.Parallel()
	repo := fixture(t, map[string]string{"Makefile": "init: check ## Smoke test\ncheck: ## Every check\nhidden:\n\techo\n"})

	findings, err := repo.CheckCommands()

	require.NoError(t, err)
	assert.Equal(t, []string{"Makefile: target hidden has no ## description"}, messages(findings))
}

func TestFeatures_AcceptsAFeatureThatAlreadyPassedAtBase(t *testing.T) {
	t.Parallel()
	passing := strings.Replace(featuresJSON, `"passes": false, "covers": ["FR-001"]`, `"passes": true, "covers": ["FR-001"]`, 1)
	repo := gitFixture(t, map[string]string{"docs/features.json": passing}, nil)

	findings, err := repo.CheckFeatures("HEAD", false)

	require.NoError(t, err)
	assert.Empty(t, messages(findings))
}

func TestRun_ExitsWithFindingsThroughTheCLI(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		override map[string]string
		want     string
	}{
		{"spec finding", map[string]string{"specs/001-demo/tasks.md": strings.Replace(fullTasks, "(FR-002)", "", 1)},
			"specs/001-demo/tasks.md: FR-002 is named by no task"},
		{"commands finding", map[string]string{"docs/COMMANDS.md": "| `make init` | a |\n"},
			"docs/COMMANDS.md: no row for make check"},
		{"stale ADR index", map[string]string{"docs/adr/0002-second.md": newADR},
			"docs/adr/README.md: index is stale: run make adr-index"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := fixture(t, tt.override)
			var stdout, stderr strings.Builder

			err := run([]string{"-root", repo.Root, "all"}, &stdout, &stderr)

			assert.ErrorIs(t, err, errFindings)
			assert.Equal(t, tt.want+"\n", stderr.String())
		})
	}
}

func TestRun_FeaturesFailsThroughTheCLI(t *testing.T) {
	t.Parallel()
	flipped := strings.Replace(featuresJSON, `"passes": false, "covers": ["FR-001"]`, `"passes": true, "covers": ["FR-001"]`, 1)
	repo := gitFixture(t, nil, map[string]string{"docs/features.json": flipped})
	var stdout, stderr strings.Builder

	err := run([]string{"-root", repo.Root, "features", "-base", "HEAD"}, &stdout, &stderr)

	assert.ErrorIs(t, err, errFindings)
	assert.Contains(t, stderr.String(), "evidence/D-01.md")
}

func TestSpec_LegacyAndQuickSpecs(t *testing.T) {
	t.Parallel()
	legacy := "# Spec 001 · Old\n\n- Status: Done\n\n## Requirements\n\n- **FR-001** prose without the keyword.\n"
	quick := "# Spec 001 · Bug\n\n- Mode: **quick**\n\n- **FR-001** prose, no scenario.\n"
	tests := []struct {
		name string
		spec string
		want []string
	}{
		{"legacy spec citing an unknown invariant", legacy + "Touches C99.\n", []string{"specs/001-demo/spec.md: cites C99, which is not in docs/CONSTITUTION.md"}},
		{"quick spec without contract or scenarios", quick, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := fixture(t, map[string]string{"specs/001-demo/spec.md": tt.spec, "specs/001-demo/contract.md": ""})

			findings, err := repo.CheckSpecs(nil)

			require.NoError(t, err)
			assert.Equal(t, tt.want, messages(findings))
		})
	}
}

func TestRun_FeaturesAllowsEditsWithALLOW_FEATURE_EDIT(t *testing.T) {
	repo := gitFixture(t, nil, map[string]string{"docs/features.json": strings.Replace(featuresJSON, `"description": "one"`, `"description": "uno"`, 1)})
	var stdout, stderr strings.Builder

	t.Setenv("ALLOW_FEATURE_EDIT", "")
	refused := run([]string{"-root", repo.Root, "features", "-base", "HEAD"}, &stdout, &stderr)
	t.Setenv("ALLOW_FEATURE_EDIT", "1")
	allowed := run([]string{"-root", repo.Root, "features", "-base", "HEAD"}, &stdout, &stderr)

	assert.ErrorIs(t, refused, errFindings)
	assert.NoError(t, allowed)
}
