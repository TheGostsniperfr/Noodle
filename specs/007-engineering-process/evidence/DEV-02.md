# Evidence · DEV-02

- Feature: DEV-02
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 674fac9
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Each rule has a case whose finding list is exactly one message | PASS | `go test ./tools/speccheck/ -run Spec_ReportsEachBrokenRule -v` gave `--- PASS: TestSpec_ReportsEachBrokenRule`, all 27 subtests PASS, `ok`, exit 0. I read `speccheck_test.go:79-154` and every message the contract lists is there word for word: C99, C77 and C42 in each of the four files, plus ADR-0042 and ADR-0077. It also has no SHALL, the three scenario cases (none, no THEN, no WHEN), the two task cases (no task line, prose only), the two phase cases, FR-009 not declared and no verify step. It has the two contract cases (empty `## D-02`, prose only), the missing contract, flips D-07 and no phase. The test expects 0 findings for inline code, a fenced block, a continuation line and the valid spec. Each case asserts `assert.Equal(t, []string{tt.want}, messages(findings))`, or `Empty` when no finding is expected. |
| 2 | Legacy and quick specs keep only the citation rules | PASS | `go test ./tools/speccheck/ -run 'Spec_Legacy' -v` gave `--- PASS: TestSpec_LegacySpecIsNotHeldToFullModeRules` (it asserts `Empty`) and `--- PASS: TestSpec_LegacyAndQuickSpecs`, with both subtests PASS. The test asserts `[]string{"specs/001-demo/spec.md: cites C99, which is not in docs/CONSTITUTION.md"}` for the legacy citation and `nil` for the quick spec. |
| 3 | The repository passes | PASS | `make spec-check` ran `go run ./tools/speccheck all` and printed no findings. `EXIT=0` |

## Verify step

    $ go test ./tools/speccheck/ -run Spec
    ok  	github.com/TheGostsniperfr/Noodle/tools/speccheck	0.010s
    $ make spec-check
    == spec-check
    go run ./tools/speccheck all
    EXIT=0

Other checks I ran myself:

- `go run ./tools/speccheck spec specs/007-engineering-process` exits 0, and the `dev` phase points at `specs/007-engineering-process/spec.md`. So this full-mode spec is checked against its own rules.
- `go vet ./tools/speccheck/` and `gofmt -l tools/speccheck` print nothing (C15).
- `git diff origin/main...HEAD -- docs/features.json` removes no line, so no entry is reworded or deleted (C12).

Previous findings (evidence at 152dd84). Each one is resolved in `tools/speccheck/spec.go` at 674fac9:

- **Contract criteria.** `contractCriteria` (`spec.go:197-210`) counts `^\| \d+ \|` rows under each `^## <id>` section, so an id in prose no longer counts.
- **Fenced blocks.** `fenceRe` strips fenced blocks before the citation scan (`spec.go:93`).
- **Task lines.** "Named by a task" now reads only task lines and their indented continuations (`taskLineRe`, `spec.go:131-132,163`).

Would the tests catch a revert? I checked one case per rule:

- **Contract rule.** "feature named only in contract prose" fails under the old `strings.Contains(contract, id)`.
- **Task rule.** "requirement named only in tasks prose" fails under the old `strings.Contains(tasks, id)`.
- **Fenced blocks.** "citation quoted in a fenced block" fails without `fenceRe`.
- **All other rules.** Every other rule has a case expecting exactly one non-empty message against a valid fixture that has 0 findings. Removing that rule fails its case.

## Findings

- `specs/007-engineering-process/contract.md:47` (minor, breaks no criterion): DEV-02 #1 was amended in 674fac9, after the contract was agreed and after the first evaluation. The change adds the prose-only, empty-section, fenced-block and continuation-line cases. It makes the criterion stricter, not weaker, but no new review round covers it and the contract's round log does not record it. Fix: add a line to the round log saying #1 was tightened after the DEV-02 NEEDS_WORK.
- `Makefile:75-78` (scope, breaks no criterion): the DEV-02 fix commit also changes the `evaluate` recipe, which now writes to a `mktemp` file instead of `$$out.tmp`. That recipe belongs to DEV-05. It is harmless, but it belongs in its own commit or under a DEV-05 task. On failure, the output now sits in `/tmp` instead of next to the evidence file.
- `tools/speccheck/spec.go:20-21` (minor, breaks no criterion): `contractCriteria` does not strip fenced blocks. A `## X` heading or a `| 1 |` row inside a fenced example in `contract.md` would count as a section or a criterion. Fix: apply `fenceRe` before scanning the sections, and add a test case.
- `tools/speccheck/spec.go:12` (minor, breaks no criterion): a requirement is found only when it is written `- **FR-NNN**`. A full-mode spec written in another shape has its requirements skipped silently, unless it has none at all, which is reported. This is acceptable because the template fixes the shape.
