# Evidence · DEV-04

- Feature: DEV-04
- Verdict: **NEEDS_WORK**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 0999af4
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | A new pass needs PASS evidence, named with its command | PASS | `go test ./tools/speccheck/ -run Features_Gates -v`: `--- PASS: .../no_evidence`, `--- PASS: .../needs_work`, `--- PASS: .../evaluator_pass`, `ok`. Each case uses `assert.Equal([]string{want}, ...)` and checks the exact message that the contract requires. |
| 2 | A feature that passed at base needs nothing | PASS | `--- PASS: TestFeatures_AcceptsAFeatureThatAlreadyPassedAtBase`, which asserts `Empty` findings. |
| 3 | Full mode requires the evaluator | PASS | `--- PASS: .../self_pass_in_full_mode`, `--- PASS: .../full_mode_without_an_Evaluator_line` (both check the exact `must come from the evaluator agent: the spec is in full mode` message) and `--- PASS: .../self_pass_in_quick_mode` (0 findings). |
| 4 | History rules report exact messages, and the variable lifts them | PASS | `go test ... -run 'Features_KeepsHistory\|TestRun_FeaturesAllows' -v`: `--- PASS: TestFeatures_KeepsHistory` (subtests removed, description changed, verify changed, each with an exact message and `Empty` when allowed) and `--- PASS: TestRun_FeaturesAllowsEditsWithALLOW_FEATURE_EDIT` (`ErrorIs errFindings` with the variable empty, `NoError` with `=1`). |
| 5 | The CLI fails on findings, and the repository passes with real flips | **FAIL** | The test passes: `--- PASS: TestRun_FeaturesFailsThroughTheCLI`. `make features BASE=origin/main` gives `exit=0`. The jq list is `DEV-01`, `DEV-02`, `DEV-03`, but the threshold requires `DEV-01`, `DEV-02`, `DEV-03`, `DEV-05`. `DEV-05` is missing, and `specs/007-engineering-process/evidence/DEV-05.md` does not exist: `grep: ... No such file or directory`. |
| 6 | Editing features.json runs the gate | PASS | `jq -r '.hooks.PostToolUse[] \| .hooks[].command' .claude/settings.json \| grep -c 'make .*features'` gives `1`. The hook runs `make -s -C "$CLAUDE_PROJECT_DIR" features >&2 \|\| exit 2`. |

## Verify step

    $ go test ./tools/speccheck/ -run Features -v
    --- PASS: TestRun_FeaturesAllowsEditsWithALLOW_FEATURE_EDIT (0.01s)
    --- PASS: TestFeatures_AcceptsAFeatureThatAlreadyPassedAtBase (0.02s)
    --- PASS: TestFeatures_KeepsHistory (0.00s)
        --- PASS: TestFeatures_KeepsHistory/removed (0.01s)
        --- PASS: TestFeatures_KeepsHistory/description_changed (0.01s)
        --- PASS: TestFeatures_KeepsHistory/verify_changed (0.02s)
    --- PASS: TestRun_FeaturesFailsThroughTheCLI (0.02s)
    --- PASS: TestFeatures_GatesNewPassesOnEvidence (0.00s)
        --- PASS: TestFeatures_GatesNewPassesOnEvidence/evaluator_pass (0.01s)
        --- PASS: TestFeatures_GatesNewPassesOnEvidence/self_pass_in_quick_mode (0.02s)
        --- PASS: TestFeatures_GatesNewPassesOnEvidence/self_pass_in_full_mode (0.02s)
        --- PASS: TestFeatures_GatesNewPassesOnEvidence/needs_work (0.02s)
        --- PASS: TestFeatures_GatesNewPassesOnEvidence/no_evidence (0.02s)
        --- PASS: TestFeatures_GatesNewPassesOnEvidence/full_mode_without_an_Evaluator_line (0.02s)
    PASS
    ok  	github.com/TheGostsniperfr/Noodle/tools/speccheck	0.044s

    $ make features BASE=origin/main
    == features (base origin/main)
    go run ./tools/speccheck features -base origin/main
    exit=0

    $ jq -r '.phases[] | select(.id=="dev") | .features[] | select(.passes) | .id' docs/features.json
    DEV-01
    DEV-02
    DEV-03

## Findings

- **Criterion 5 fails.** `docs/features.json` does not flip DEV-05, and `specs/007-engineering-process/evidence/DEV-05.md` does not exist. The contract's evaluation order expected this result on the first run. To fix it, evaluate DEV-05 on this committed `DEV-04.md`, flip DEV-05 on its PASS evidence, then evaluate DEV-04 again.
- **Tests per requirement.** The tests would catch a revert.
  - FR-005: `no_evidence` and `self_pass_in_full_mode` assert exact single-message finding lists. They would fail if `evidenceProblem` or the full-mode evaluator check were removed.
  - FR-008: `TestFeatures_KeepsHistory/removed` asserts `D-02: removed`. It would fail without the history loop. If the fixture `strings.Replace` matched nothing, the `Equal` assertion would fail too, so the test cannot pass vacuously.
- **Backlog flips look for the wrong file. This is outside the criteria.** At `tools/speccheck/repo.go:79-84`, `specDirOf` returns `docs/evidence` for a backlog entry. Then `tools/speccheck/features.go:69` reads `<dir>/evidence/<id>.md`, which gives `docs/evidence/evidence/B-NN.md`, and the error message names that doubled path. That path's `spec.md` never exists, so `specMode` is never `full` and backlog evidence never needs the evaluator. No test covers backlog flips. Fix: give `specDirOf` and `evidenceProblem` one evidence directory for entries without a spec, and add a test case that flips a backlog entry.
- **Weak verdict parsing. This is outside the criteria.** At `tools/speccheck/features.go:13`, `verdictRe` captures the first word after `- Verdict:`. An unfilled template line `- Verdict: **PASS** | **NEEDS_WORK**` therefore passes the gate. Fix: anchor the regex to the end of the line (`\**\s*$`), as criterion DEV-05 #3 already does.
- **A bad base ref fails quietly. This is outside the criteria.** At `tools/speccheck/features.go:94-98`, `baseFeatures` ignores every `git show` error, including a misspelled `BASE`. The history rules are then skipped without any message. Fix: return nil only when the error says the path does not exist at that ref, and return an error for an unknown ref.
- **Scope and Constitution.** The DEV-04 code is limited to `tools/speccheck/`, the `features` Makefile target and the hook. The DEV-04 entry in `docs/features.json` is unchanged and `passes: false` (C12). No lint rule was touched (C04).
