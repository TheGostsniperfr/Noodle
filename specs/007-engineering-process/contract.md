# Contract 007 · Engineering process

What "done" means for each feature, agreed before implementation (ADR-0024).
One failed criterion fails the feature. Every "Verified by" runs inside the evaluator's
allowlist (`EVALUATOR` in the Makefile); a command outside it is refused, and fails.

- Agreed: pending `make review-contract SPEC=007`. Bootstrap exception: this contract is
  reviewed after implementation, since the review command is part of what it grades.
- Round 1 (2026-11-04): CHANGES, 17 points, all applied; FR-007 amended (clarification 3).
- Round 2: CHANGES, 13 points, all applied; FR-002 declares the flip rule; tests assert
  exact finding lists.
- Round 3: CHANGES, 13 points, all applied; the allowlist covers every command below and
  no longer allows writing (`go run:*`, `make render:*` removed); FR-007 scoped to `ci.yml`.
- Evaluation order: DEV-01, DEV-02, DEV-03, DEV-05, then DEV-04, so that DEV-04 #5 is
  graded with flipped features. Each evidence file is committed alone, before its flip.

Fixture messages below are those of `tools/speccheck/speccheck_test.go`, where the demo
spec is `specs/001-demo` and its phase `d`.

## DEV-01 · Makefile is the only entry point

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The commands check reports both directions and undescribed targets | `go test ./tools/speccheck/ -run Commands -v` | exit 0; a missing row yields exactly `docs/COMMANDS.md: no row for make check`; an extra row exactly `Makefile: no documented target ghost (COMMANDS.md lists it; add it with a ## comment)`; a target without `##` exactly `Makefile: target hidden has no ## description`; matching sets 0 findings |
| 2 | Findings fail the command | `go test ./tools/speccheck/ -run TestRun_Exits -v` | exit 0: `run(-root <fixture> all)` returns `errFindings` and stderr contains the line `specs/001-demo/tasks.md: FR-002 is named by no task` |
| 3 | Every `run:` step of ci.yml calls make; no workflow calls a script | `grep -rc 'scripts/' .github/workflows/`; `grep -E '^\s+run:' .github/workflows/ci.yml \| grep -vc 'run: make '` | `0` for every file; `0` |
| 4 | Every script a hook names is listed under Hooks; other hook commands call make | `jq -r '.. \| .command? // empty' .claude/settings.json`, read against the Hooks table of `docs/COMMANDS.md` | each `scripts/*.sh` printed appears in that table; every other printed command contains `make ` |
| 5 | `check-features.sh` is gone and `init.sh` is a shim | `git grep --untracked -n 'check-features.sh\|scripts/init.sh' -- ':!docs/features.json' ':!docs/PROGRESS.md' ':!specs/00[1-7]-*'`; `grep -vc '^#' scripts/init.sh`; `cat scripts/init.sh` | `git grep` exits 1 with no output; the count is `1` and that line matches `^exec make .*init$` |
| 6 | `make help` lists one line per documented target | `make -s help \| wc -l`; `grep -cE '^[\|] .make ' docs/COMMANDS.md`; `go run ./tools/speccheck commands` | the two counts are equal; the last exits 0 with no output |
| 7 | The branch is green | `make init` | exit 0 |

## DEV-02 · Traceability of specs

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Each rule has a case whose finding list is exactly one message | `go test ./tools/speccheck/ -run Spec_ReportsEachBrokenRule -v` (each case asserts `Equal([]string{want}, findings)`) | exit 0, with these subtests and messages, each prefixed by the file named: unknown invariant `specs/001-demo/spec.md: cites C99, which is not in docs/CONSTITUTION.md`; in tasks `…/tasks.md: cites C77, …`; in plan `…/plan.md: cites C42, …`; unknown ADR `…/spec.md: cites ADR-0042, which does not exist`, in plan `…/plan.md: cites ADR-0077, …`, in contract `…/contract.md: cites ADR-0077, …`; `FR-001 is not EARS: no SHALL`; `FR-001 has no Scenario with WHEN and THEN` (no scenario, and scenario without THEN); `tasks.md: FR-002 is named by no task`; `docs/features.json: FR-002 of specs/001-demo/spec.md is covered by no feature of phase d` (no feature, and a feature of another phase only); `D-02 covers FR-009, which specs/001-demo/spec.md does not declare`; `D-02 has no verify step`; `contract.md: no criteria for D-02`; `contract.md: a full-mode spec needs a contract`; `tasks.md: flips D-07, which is not in docs/features.json`; `spec.md: no phase in docs/features.json points at this spec`; a citation quoted in inline code yields 0 findings; and the valid case with 0 findings |
| 2 | Legacy and quick specs keep only the citation rules | `go test ./tools/speccheck/ -run 'Spec_Legacy' -v` | exit 0: a legacy spec with prose requirements yields 0 findings; a legacy spec citing an unknown invariant yields exactly `specs/001-demo/spec.md: cites C99, which is not in docs/CONSTITUTION.md`; a quick spec with no contract and no scenarios yields 0 findings |
| 3 | The repository passes | `make spec-check` | exit 0 |

## DEV-03 · ADR shape and index

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Frontmatter ADRs are checked piece by piece | `go test ./tools/speccheck/ -run ADR_ChecksFrontmatter -v` | exit 0; each subtest asserts exactly one finding, prefixed `docs/adr/0002-second.md: `: no status and unknown status `status "<value>" is not one of proposed, accepted, rejected, deprecated, superseded`; no date and no deciders `frontmatter needs date and deciders`; one option `needs at least two considered options`; no and empty Confirmation `needs a Confirmation section: what proves the decision is kept`; the valid ADR 0 findings |
| 2 | Old ADRs parse from their status line | `go test ./tools/speccheck/ -run ADR_Legacy -v`; `make spec-check` | exit 0 both, with ADRs 0001 to 0023 in the repository |
| 3 | A stale index is reported, and `-write` fixes it | `go test ./tools/speccheck/ -run ADR_ReportsStaleIndex -v` | exit 0: the stale run yields exactly `docs/adr/README.md: index is stale: run make adr-index`; after `-write`, 0 findings |
| 4 | The repository index is current and complete | `go run ./tools/speccheck adr`; `grep -c '^| \[0' docs/adr/README.md`; `ls docs/adr/[0-9]*.md \| wc -l` | exit 0, no output; the two counts are equal |

## DEV-04 · Evidence gate and history guard

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | A new pass needs PASS evidence, named with its command | `go test ./tools/speccheck/ -run Features_Gates -v` | exit 0: no evidence yields exactly `docs/features.json: D-01 passes without evidence: no specs/001-demo/evidence/D-01.md (make evaluate FEATURE=D-01)`; NEEDS_WORK exactly `docs/features.json: D-01 passes without evidence: specs/001-demo/evidence/D-01.md has no '- Verdict: PASS' line`; evaluator PASS 0 findings |
| 2 | A feature that passed at base needs nothing | `go test ./tools/speccheck/ -run Features_AcceptsAFeatureThatAlreadyPassed -v` | exit 0, 0 findings |
| 3 | Full mode requires the evaluator | `go test ./tools/speccheck/ -run Features_Gates -v` | exit 0: `Evaluator: self` and no Evaluator line in full mode each yield exactly `docs/features.json: D-01 passes without evidence: specs/001-demo/evidence/D-01.md must come from the evaluator agent: the spec is in full mode`; `self` in quick mode 0 findings |
| 4 | History rules report exact messages | `go test ./tools/speccheck/ -run Features_KeepsHistory -v` | exit 0: exactly `docs/features.json: D-02: removed`, `docs/features.json: D-01: description changed`, `docs/features.json: D-01: verify step changed`; each 0 findings with `ALLOW_FEATURE_EDIT` |
| 5 | The CLI fails on findings, and the repository passes with real flips | `go test ./tools/speccheck/ -run TestRun_Features -v`; `make features BASE=origin/main`; `jq -r '.phases[] \| select(.id=="dev") \| .features[] \| select(.passes) \| .id' docs/features.json` | the test exits 0 (`errFindings`, `evidence/D-01.md` on stderr); the target exits 0; the jq list holds DEV-01, DEV-02, DEV-03 and DEV-05, each with `specs/007-engineering-process/evidence/<id>.md` holding `- Verdict: **PASS**` |
| 6 | Editing features.json runs the gate | `jq -r '.hooks.PostToolUse[] \| .hooks[].command' .claude/settings.json \| grep -c 'make .*features'` | `1` |

## DEV-05 · Evaluator and contract review

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The evaluator has no write tool and no write-capable pattern | `make -n evaluate FEATURE=DEV-04 \| grep -o -- '--disallowedTools "[^"]*"'`; `make -n evaluate FEATURE=DEV-04 \| grep -cE 'Bash\((go run\|make render):\*\)'` | lists Edit, Write and NotebookEdit; `0` |
| 2 | Its allowlist covers every command of this contract | `make -n evaluate FEATURE=DEV-04 \| grep -o -- '--allowedTools "[^"]*"'` | contains `Bash(jq:*)`, `Bash(git grep:*)`, `Bash(wc:*)`, `Bash(make init)`, `Bash(make -n evaluate:*)`, `Bash(make evaluate FEATURE=)`, `Bash(make review-contract SPEC=)`, `Bash(go run ./tools/speccheck adr)` |
| 3 | It is a fresh process given the feature, spec and contract | `make -n evaluate FEATURE=DEV-04`; `go run ./tools/speccheck locate DEV-04` | the recipe runs `claude -p --agent evaluator`; `dir` is assigned from `speccheck locate`; the prompt contains `DEV-04`, `$dir` and `contract.md`; locate prints exactly `specs/007-engineering-process` |
| 4 | An evaluation commit holds only its evidence file | `git show --stat --format= $(git log -1 --format=%h -- specs/007-engineering-process/evidence/DEV-01.md)` | lists exactly that one file |
| 5 | The evidence file has the gate's shape | `grep -cE '^- Verdict: \*\*(PASS\|NEEDS_WORK)\*\*$' specs/007-engineering-process/evidence/DEV-01.md`; `grep -cE '^- Evaluator: evaluator ' specs/007-engineering-process/evidence/DEV-01.md` | `1` and `1` |
| 6 | The contract review answers in shape | `grep -cE '^- Review: \*\*(APPROVED\|CHANGES)\*\*$' specs/007-engineering-process/contract-review.md` | `1`, in a file committed from `make review-contract SPEC=007` |
| 7 | Both targets refuse a missing argument | `make review-contract SPEC=`; `make evaluate FEATURE=` | each exits non-zero and prints exactly `usage: make review-contract SPEC=<NNN>` and `usage: make evaluate FEATURE=<id>` |
