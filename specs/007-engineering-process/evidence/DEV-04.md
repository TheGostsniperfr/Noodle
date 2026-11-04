# Evidence · DEV-04

- Feature: DEV-04
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: ced6a24
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | A new pass needs PASS evidence, named with its command | PASS | `go test ./tools/speccheck/ -run Features -v` (a superset of `Features_Gates`) passes these three subtests: `--- PASS: TestFeatures_GatesNewPassesOnEvidence/no_evidence`, `/needs_work` and `/evaluator_pass`. At `speccheck_test.go:306-310` the `want` strings match the contract word for word: `docs/features.json: D-01 passes without evidence: no specs/001-demo/evidence/D-01.md (make evaluate FEATURE=D-01)` and `... specs/001-demo/evidence/D-01.md has no '- Verdict: PASS' line`. Each one is asserted with `assert.Equal([]string{want}, ...)`, and `Empty` is asserted for evaluator PASS. |
| 2 | A feature that passed at base needs nothing | PASS | `go test ./tools/speccheck/ -run Features_AcceptsAFeatureThatAlreadyPassed -v` gives `--- PASS: TestFeatures_AcceptsAFeatureThatAlreadyPassedAtBase`. The base commit has D-01 `passes: true` and no evidence, and the test asserts `Empty`. |
| 3 | Full mode requires the evaluator | PASS | `--- PASS: .../self_pass_in_full_mode` and `--- PASS: .../full_mode_without_an_Evaluator_line` both expect exactly `... specs/001-demo/evidence/D-01.md must come from the evaluator agent: the spec is in full mode` (`speccheck_test.go:311-314`). `--- PASS: .../self_pass_in_quick_mode` asserts `Empty`. |
| 4 | History rules report exact messages, and the variable lifts them | PASS | `go test ./tools/speccheck/ -run 'Features_KeepsHistory\|TestRun_FeaturesAllows' -v` passes `/removed`, `/description_changed`, `/verify_changed` and `TestRun_FeaturesAllowsEditsWithALLOW_FEATURE_EDIT`. The expected strings are `docs/features.json: D-02: removed`, `D-01: description changed` and `D-01: verify step changed`, and `Empty` is asserted with `allowEdit=true`. At the CLI the test checks `ErrorIs errFindings` with the variable empty and `NoError` with `=1`. |
| 5 | The CLI fails on findings, and the repository passes with real flips | PASS | `go test ... -run TestRun_Features -v` passes `TestRun_FeaturesFailsThroughTheCLI`, which asserts `ErrorIs errFindings` and that stderr contains `evidence/D-01.md`. `make features BASE=origin/main` gives `exit=0`. The jq command prints `DEV-01`, `DEV-02`, `DEV-03`, `DEV-05`, one per line in that order. Each of those four evidence files has exactly one line `- Verdict: **PASS**` (`grep -c` gives `1`). |
| 6 | Editing features.json runs the gate | PASS | `jq -r '.hooks.PostToolUse[] \| .hooks[].command' .claude/settings.json \| grep -c 'make .*features'` gives `1`. The hook command is `... */docs/features.json) make -s -C "$CLAUDE_PROJECT_DIR" features >&2 \|\| exit 2;;`. |

## Verify step

    $ go test ./tools/speccheck/ -run Features -v
    --- PASS: TestRun_FeaturesAllowsEditsWithALLOW_FEATURE_EDIT (0.02s)
    --- PASS: TestFeatures_TreatsGitHubsNullBeforeAsNoBase (0.02s)
    --- PASS: TestRun_FeaturesFailsThroughTheCLI (0.02s)
    --- PASS: TestFeatures_RejectsAnUnknownBaseRef (0.02s)
    --- PASS: TestFeatures_GatesNewPassesOnEvidence (0.00s)
        --- PASS: TestFeatures_GatesNewPassesOnEvidence/full_mode_without_an_Evaluator_line (0.02s)
        --- PASS: TestFeatures_GatesNewPassesOnEvidence/evaluator_pass (0.02s)
        --- PASS: TestFeatures_GatesNewPassesOnEvidence/self_pass_in_quick_mode (0.02s)
        --- PASS: TestFeatures_GatesNewPassesOnEvidence/self_pass_in_full_mode (0.02s)
        --- PASS: TestFeatures_GatesNewPassesOnEvidence/no_evidence (0.02s)
        --- PASS: TestFeatures_GatesNewPassesOnEvidence/unfilled_template_verdict (0.03s)
        --- PASS: TestFeatures_GatesNewPassesOnEvidence/needs_work (0.03s)
    --- PASS: TestFeatures_KeepsHistory (0.00s)
        --- PASS: TestFeatures_KeepsHistory/verify_changed (0.02s)
        --- PASS: TestFeatures_KeepsHistory/removed (0.03s)
        --- PASS: TestFeatures_KeepsHistory/description_changed (0.03s)
    --- PASS: TestFeatures_GatesBacklogFlipsOnDocsEvidence (0.00s)
        --- PASS: TestFeatures_GatesBacklogFlipsOnDocsEvidence/self_evidence (0.02s)
        --- PASS: TestFeatures_GatesBacklogFlipsOnDocsEvidence/no_evidence (0.03s)
    --- PASS: TestFeatures_AcceptsAFeatureThatAlreadyPassedAtBase (0.03s)
    PASS
    ok  	github.com/TheGostsniperfr/Noodle/tools/speccheck	0.059s
    exit=0

    $ make features BASE=origin/main
    == features (base origin/main)
    go run ./tools/speccheck features -base origin/main
    exit=0

    $ jq -r '.phases[] | select(.id=="dev") | .features[] | select(.passes) | .id' docs/features.json
    DEV-01
    DEV-02
    DEV-03
    DEV-05

    $ go run ./tools/speccheck features -base nope
    base ref "nope" is not a commit: exit status 1
    exit=1

## Findings

These are minor findings outside the criteria. None of them breaks a contract criterion.

- **Each requirement has a test that would catch a revert.**
  - FR-005: `no_evidence` and `self_pass_in_full_mode` (`speccheck_test.go:306-314`) assert exact single-message lists. Removing `evidenceProblem` (`features.go:59-61`) or the full-mode branch (`features.go:82-87`) turns them red. `self_pass_in_quick_mode` would catch a gate that ignored the mode.
  - FR-008: `TestFeatures_KeepsHistory/removed` asserts `D-02: removed`. If the fixture `strings.Replace` matched nothing, the `Equal` would fail, so the test cannot pass vacuously.
- **The first-run findings are fixed.** Commit `9dc1743` fixes the three side findings of the first run, each with a test:
  - unfilled template verdict: `features.go:13` is anchored to end of line, and `unfilled_template_verdict` tests it;
  - backlog path: `repo.go:82` returns `docs`, so the gate reads `docs/evidence/B-NN.md`, tested by `TestFeatures_GatesBacklogFlipsOnDocsEvidence`;
  - unknown base ref: `features.go:96-99` rejects it, tested by `TestFeatures_RejectsAnUnknownBaseRef`. I confirmed this at the CLI with exit 1.
- **The evaluator check is loose (`tools/speccheck/features.go:84`).** `strings.HasPrefix(..., "evaluator")` also accepts `- Evaluator: evaluatorish` or `- Evaluator: evaluator-self`. Fix: match `^evaluator( |$)`.
- **An unreadable spec.md drops the evaluator requirement (`tools/speccheck/features.go:81`).** `spec, _ := r.read(dir, "spec.md")` discards the read error. If a phase points at a spec.md that is missing or unreadable, the mode is not `full`, so self evidence is accepted. Fix: for a phase with a `spec` field, return the read error as a finding.
- **The gate trusts the Evaluator line as text.** Anyone can hand-write `- Evaluator: evaluator (…)`. This is inherent to a file-based gate, and the reviewed-commit rule (DEV-05 #4) is the mitigation. ADR-0024 could state this as an accepted risk, next to the allowlist risks it already records.
- **Scope and Constitution.** The DEV-04 code is limited to `tools/speccheck/features.go`, `repo.go` and their tests, the `features` Makefile target and the PostToolUse hook. Against `origin/main`, `docs/features.json` only adds the `dev` phase: no existing entry is edited or removed (C12), and the DEV-04 entry is still `passes: false`. No lint rule is touched (C04).
