# Evidence · DEV-02

- Feature: DEV-02
- Verdict: **NEEDS_WORK**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 152dd84
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Each rule has a case whose finding list is exactly one message | PASS | `go test ./tools/speccheck/ -run Spec_ReportsEachBrokenRule -v`: `--- PASS: TestSpec_ReportsEachBrokenRule`, 23 subtests PASS, `ok`. I read `speccheck_test.go:86-131`. Every message in the contract is there word for word, including the three scenario cases (none, no THEN, no WHEN), the two coverage cases (no feature, another phase only), and 0 findings for the inline-code citation and for the valid spec. Lines 141-145 assert `Equal([]string{tt.want}, messages(findings))`. |
| 2 | Legacy and quick specs keep only the citation rules | PASS | `go test ./tools/speccheck/ -run 'Spec_Legacy' -v`: `--- PASS: TestSpec_LegacyAndQuickSpecs` (2 subtests) and `--- PASS: TestSpec_LegacySpecIsNotHeldToFullModeRules`. The test expects `[]string{"specs/001-demo/spec.md: cites C99, which is not in docs/CONSTITUTION.md"}` for the legacy citation case and `nil` for the quick spec. |
| 3 | The repository passes | PASS | `make spec-check` ran `go run ./tools/speccheck all` with no findings. `EXIT=0` |

## Verify step

    $ go test ./tools/speccheck/ -run Spec
    ok  	github.com/TheGostsniperfr/Noodle/tools/speccheck	0.007s
    $ make spec-check
    == spec-check
    go run ./tools/speccheck all
    EXIT=0

Other checks I ran: `make features BASE=origin/main` exits 0. `git diff origin/main...HEAD -- docs/features.json` only adds the `dev` phase and changes no existing entry (C12 holds). The diff touches nothing under `internal/` or `cmd/`.

## Findings

- `tools/speccheck/spec.go:171-174`: FR-003 requires a finding "for each feature with no criteria in `contract.md`". The code only checks `strings.Contains(contract, f.ID)`, so the feature ID appearing anywhere in the file is enough to pass. As a result the rule can never fire on this repository's own contract. `specs/007-engineering-process/contract.md:19-21` (the "Evaluation order" bullet) names DEV-01 to DEV-05 in prose. Deleting the whole `## DEV-02` section and its table would still give `make spec-check` exit 0. `grep -n 'DEV-0[1-5]' specs/007-engineering-process/contract.md` shows lines 19-20 holding every ID outside the sections. This is the same failure as the calibration case: the test is green but does not exercise the requirement. Fix: find the feature's section (a heading starting with `## <id>`) and require at least one criterion row (`| <n> |`) before the next heading. Report `no criteria for <id>` when the heading is missing or has no rows.
- `tools/speccheck/speccheck_test.go:103-104`: the "feature with no contract criteria" case sets the contract to `"## D-01\n"`, a heading with no rows. It then asserts that the only finding is for D-02, so the test locks in the bug above: a feature with an empty section passes. Fix: expect `no criteria for D-01` here too. Also add a case where the ID appears only in prose, for example `"Order: D-01, D-02.\n## D-01\n| 1 | x | y | z |\n"`, and assert `[]string{"specs/001-demo/contract.md: no criteria for D-02"}`.
- `tools/speccheck/spec.go:17,91` (minor, breaks no criterion): `codeSpanRe` only strips single-backtick spans. A `Cnn` or `ADR-NNNN` inside a fenced code block still counts as a citation, and FR-002 treats quoted code as quotation. Strip fenced blocks before the citation scan and add a test case for it.
- `tools/speccheck/spec.go:145` (minor, breaks no criterion): "named by a task" is `strings.Contains(tasks, req.id)`, so any mention anywhere in `tasks.md` counts, a prose note included. This is acceptable as it stands. Narrowing it to task lines (`- [ ] **Tnn**` / `- [x] **Tnn**`) would make it match the wording of FR-003.
- Tests vs. requirements: I read every case of `TestSpec_ReportsEachBrokenRule` and `TestSpec_LegacyAndQuickSpecs`. Each one would fail if its rule were removed from `spec.go`, because each asserts an exact non-empty list. The exception is the contract-criteria rule, which is only half tested, as described above.
