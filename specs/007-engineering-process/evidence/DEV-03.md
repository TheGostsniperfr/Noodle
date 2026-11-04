# Evidence · DEV-03

- Feature: DEV-03
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: d63ce3a
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Frontmatter ADRs are checked piece by piece | PASS | `go test ./tools/speccheck/ -run ADR_ChecksFrontmatter -v` exits 0 with 8 subtests passing: valid, no_status, unknown_status, no_date, no_deciders, one_option, no_confirmation, empty_confirmation. Each failing case asserts `Equal([]string{"docs/adr/0002-second.md: " + want}, …)`. The `want` strings match the contract's messages character for character (speccheck_test.go:200-206). |
| 2 | Old ADRs parse from their status line | PASS | `go test … -run ADR_Legacy -v` gives `--- PASS: TestADR_LegacyADRParsesFromItsStatusLine`, exit 0. `make spec-check` exits 0. `ls docs/adr/` lists 0001 to 0023 in the legacy shape, plus 0024. |
| 3 | A stale index is reported, and `-write` fixes it | PASS | `go test … -run ADR_ReportsStaleIndex -v` gives `--- PASS: TestADR_ReportsStaleIndexAndRewritesIt`, exit 0. It asserts `Equal([]string{"docs/adr/README.md: index is stale: run make adr-index"}, stale)` and `Empty(fresh)` after `CheckADRs(true)`. |
| 4 | The repository index is current and complete | PASS | `go run ./tools/speccheck adr` exits 0 with no output. `grep -c '^| \[0' docs/adr/README.md` gives `24`. `ls docs/adr/[0-9]*.md \| wc -l` gives `24`. |

## Verify step

    $ go test ./tools/speccheck/ -run ADR -v
    --- PASS: TestADR_LegacyADRParsesFromItsStatusLine (0.00s)
    --- PASS: TestADR_ReportsStaleIndexAndRewritesIt (0.00s)
    --- PASS: TestADR_ChecksFrontmatterADRs (0.00s)
        --- PASS: TestADR_ChecksFrontmatterADRs/empty_confirmation (0.00s)
        --- PASS: TestADR_ChecksFrontmatterADRs/unknown_status (0.00s)
        --- PASS: TestADR_ChecksFrontmatterADRs/no_date (0.00s)
        --- PASS: TestADR_ChecksFrontmatterADRs/valid (0.00s)
        --- PASS: TestADR_ChecksFrontmatterADRs/one_option (0.00s)
        --- PASS: TestADR_ChecksFrontmatterADRs/no_status (0.00s)
        --- PASS: TestADR_ChecksFrontmatterADRs/no_deciders (0.00s)
        --- PASS: TestADR_ChecksFrontmatterADRs/no_confirmation (0.00s)
    PASS
    ok  	github.com/TheGostsniperfr/Noodle/tools/speccheck	0.006s

    $ go run ./tools/speccheck adr        # stands in for "make adr-index leaves README.md unchanged"
    (no output, exit 0)
    $ git status --short docs/adr
    (no output)

The second half of the verify step was not run as written. `make adr-index` writes to disk, so it is outside the evaluator's allowlist. The contract says to grade it through criterion #4. If `CheckADRs(false)` returns no stale finding, the generated block equals the committed one. In that case `CheckADRs(true)` returns before it writes anything (adr.go, `if current == want { return findings, nil }`).

## Findings

These findings break no contract criterion.

- **The two-option check can be fooled.** tools/speccheck/adr.go:24 and adr.go:118. Under "Considered options", `optionItemRe` counts every line that starts with `- `, `### ` or `N. `. An ADR with one numbered option and one unindented bullet in that section, such as a note or a pro/con, counts as two options and passes FR-004's "two considered options". The repo's template uses a numbered list (`_template.md:24-25`). A fix would count only the numbered items, or only the `###` option headings, depending on the template's shape. A test case would cover "one option followed by a bullet".
- **The tests would catch a revert.** If the frontmatter checks were removed, each failing subtest of `TestADR_ChecksFrontmatterADRs` would get an empty finding list instead of its single expected finding, so it would fail. If the index comparison were removed, `TestADR_ReportsStaleIndexAndRewritesIt` would fail its `Equal` on the stale message. `TestADR_LegacyADRParsesFromItsStatusLine` asserts no findings on a fixture whose index says `Accepted`. If the legacy status parsing broke, the status would come out empty and the generated index would no longer match, so this test would fail too. That covers more than its name suggests.
- **Scope and constitution.** The ADR code is limited to `tools/speccheck/adr.go`, its tests, the Makefile targets `adr-index` and `spec-check`, `docs/adr/_template.md`, `docs/adr/README.md` and ADR-0024. All of these are in the plan. The legacy ADRs 0001 to 0023 are not rewritten (C10 holds). The DEV-03 entry in `docs/features.json` is unchanged and still `passes: false` (C12 holds).
