# Contract 007 · Engineering process

What "done" means for each feature, agreed before implementation (ADR-0024).
One failed criterion fails the feature. Every "Verified by" runs inside the evaluator's
allowlist (`EVALUATOR` in the Makefile); a command outside it is refused, and fails.
In a table cell, `\|` is the Markdown escape of a pipe: run every command with a plain `|`.

- Agreed: 2026-11-04, after five rounds (the three-round cap of PROCESS.md came after
  them); every point of every round applied, none pending. Bootstrap exception: this
  contract was reviewed after implementation, since the review command is part of what
  it grades.
- Round 1 (2026-11-04): CHANGES, 17 points, all applied; FR-007 amended (clarification 3).
- Round 2: CHANGES, 13 points, all applied; FR-002 declares the flip rule; tests assert
  exact finding lists.
- Round 3: CHANGES, 13 points, all applied; the allowlist covers every command below and
  no longer allows writing (`go run:*`, `make render:*` removed); FR-007 scoped to `ci.yml`.
- Round 4: CHANGES, 11 points, all applied; make patterns pinned (no variable overrides),
  the scenario rule reads the scenario itself, `ALLOW_FEATURE_EDIT` tested at the CLI.
- Evaluation order: DEV-01, DEV-02, DEV-03, DEV-04 (first run, expected NEEDS_WORK on #5),
  DEV-05 (graded on the committed `DEV-04.md`, which is its verify step's output), then
  DEV-04 again. Each evidence file is committed alone, before its flip.
- Round 5: CHANGES, 6 points, all applied; the pipe convention above, exact messages
  everywhere, the allowlist listed in full, `spec-check` tested on each of its checks.
- Accepted risks, not graded as "read-only": `Bash(go test:*)` allows `-o`, `-exec` and
  `-coverprofile`, and `Bash(git diff/log/show:*)` allow `--output`. They stay because the
  verify steps need them; ADR-0024 records the residual risk.

Fixture messages below are those of `tools/speccheck/speccheck_test.go`, where the demo
spec is `specs/001-demo` and its phase `d`.

## DEV-01 · Makefile is the only entry point

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The commands check reports both directions and undescribed targets | `go test ./tools/speccheck/ -run Commands -v` | exit 0; a missing row yields exactly `docs/COMMANDS.md: no row for make check`; an extra row exactly `Makefile: no documented target ghost (COMMANDS.md lists it; add it with a ## comment)`; a target without `##` exactly `Makefile: target hidden has no ## description`; matching sets 0 findings |
| 2 | `spec-check` fails on each of its three checks | `go test ./tools/speccheck/ -run TestRun_Exits -v` | exit 0: three cases of `run(-root <fixture> all)`, each returning `errFindings` with stderr exactly one line: `specs/001-demo/tasks.md: FR-002 is named by no task`, `docs/COMMANDS.md: no row for make check`, `docs/adr/README.md: index is stale: run make adr-index` |
| 3 | Every `run:` step of ci.yml calls make; no workflow calls a script | `grep -rc 'scripts/' .github/workflows/`; `grep -E '^\s+run:' .github/workflows/ci.yml \| grep -vc 'run: make '` | `0` for every file; `0` |
| 4 | Every script a hook names is listed under Hooks; other hook commands call make | `jq -r '.. \| .command? // empty' .claude/settings.json`, read against the Hooks table of `docs/COMMANDS.md` | each `scripts/*.sh` printed appears in that table; every other printed command contains `make ` |
| 5 | `check-features.sh` is gone and `init.sh` is a shim | `git grep --untracked -n -e check-features.sh -e scripts/init.sh -- ':!docs/features.json' ':!docs/PROGRESS.md' ':!specs/00[1-7]-*'`; `grep -vc '^#' scripts/init.sh`; `cat scripts/init.sh` | `git grep` exits 1 with no output; the count is `1` and that line matches `^exec make .*init$` |
| 6 | `make help` lists one line per documented target | `make -s help \| wc -l`; `grep -cE '^[\|] .make ' docs/COMMANDS.md`; `go run ./tools/speccheck commands` | the two counts are equal; the last exits 0 with no output |
| 7 | The branch is green | `make init` | exit 0 |

## DEV-02 · Traceability of specs

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Each rule has a case whose finding list is exactly one message | `go test ./tools/speccheck/ -run Spec_ReportsEachBrokenRule -v` (each case asserts `Equal([]string{want}, findings)`) | exit 0, with these exact messages: `specs/001-demo/spec.md: cites C99, which is not in docs/CONSTITUTION.md`; `specs/001-demo/tasks.md: cites C77, which is not in docs/CONSTITUTION.md`; `specs/001-demo/plan.md: cites C42, which is not in docs/CONSTITUTION.md`; `specs/001-demo/contract.md: cites C77, which is not in docs/CONSTITUTION.md`; `specs/001-demo/spec.md: cites ADR-0042, which does not exist`; `specs/001-demo/plan.md: cites ADR-0077, which does not exist`; `specs/001-demo/tasks.md: cites ADR-0077, which does not exist`; `specs/001-demo/contract.md: cites ADR-0077, which does not exist`; `specs/001-demo/spec.md: FR-001 is not EARS: no SHALL`; `specs/001-demo/spec.md: FR-001 has no Scenario with WHEN and THEN` (three cases: no scenario, no THEN, no WHEN); `specs/001-demo/tasks.md: FR-002 is named by no task` (two cases: no task line, and FR-002 only in prose); `docs/features.json: FR-002 of specs/001-demo/spec.md is covered by no feature of phase d` (two cases: no feature, another phase only); `docs/features.json: D-02 covers FR-009, which specs/001-demo/spec.md does not declare`; `docs/features.json: D-02 has no verify step`; `specs/001-demo/contract.md: no criteria for D-02` (two cases: an empty `## D-02` section, and D-02 named only in prose); `specs/001-demo/contract.md: a full-mode spec needs a contract`; `specs/001-demo/tasks.md: flips D-07, which is not in docs/features.json`; `specs/001-demo/spec.md: no phase in docs/features.json points at this spec`; and 0 findings for a citation quoted in inline code or a fenced block, for FR-002 on a task's continuation line, and for the valid spec |
| 2 | Legacy and quick specs keep only the citation rules | `go test ./tools/speccheck/ -run 'Spec_Legacy' -v` | exit 0: a legacy spec with prose requirements yields 0 findings; a legacy spec citing an unknown invariant yields exactly `specs/001-demo/spec.md: cites C99, which is not in docs/CONSTITUTION.md`; a quick spec with no contract and no scenarios yields 0 findings |
| 3 | The repository passes | `make spec-check` | exit 0 |

## DEV-03 · ADR shape and index

The verify step's `make adr-index` writes, so it is graded through #4: `go run
./tools/speccheck adr` exiting 0 with no output proves that `-write` would change nothing.

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Frontmatter ADRs are checked piece by piece | `go test ./tools/speccheck/ -run ADR_ChecksFrontmatter -v` (each case asserts `Equal` on the full message list) | exit 0; each subtest asserts exactly one finding, prefixed `docs/adr/0002-second.md: `: no status `status "" is not one of proposed, accepted, rejected, deprecated, superseded`; unknown status `status "maybe" is not one of proposed, accepted, rejected, deprecated, superseded`; no date and no deciders `frontmatter needs date and deciders`; one option `needs at least two considered options`; no and empty Confirmation `needs a Confirmation section: what proves the decision is kept`; the valid ADR 0 findings |
| 2 | Old ADRs parse from their status line | `go test ./tools/speccheck/ -run ADR_Legacy -v`; `make spec-check` | exit 0 both, with ADRs 0001 to 0023 in the repository |
| 3 | A stale index is reported, and `-write` fixes it | `go test ./tools/speccheck/ -run ADR_ReportsStaleIndex -v` | exit 0: the stale run yields exactly `docs/adr/README.md: index is stale: run make adr-index`; after `-write`, 0 findings |
| 4 | The repository index is current and complete | `go run ./tools/speccheck adr`; `grep -c '^| \[0' docs/adr/README.md`; `ls docs/adr/[0-9]*.md \| wc -l` | exit 0, no output; the two counts are equal |

## DEV-04 · Evidence gate and history guard

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | A new pass needs PASS evidence, named with its command | `go test ./tools/speccheck/ -run Features_Gates -v` | exit 0: no evidence yields exactly `docs/features.json: D-01 passes without evidence: no specs/001-demo/evidence/D-01.md (make evaluate FEATURE=D-01)`; NEEDS_WORK exactly `docs/features.json: D-01 passes without evidence: specs/001-demo/evidence/D-01.md has no '- Verdict: PASS' line`; evaluator PASS 0 findings |
| 2 | A feature that passed at base needs nothing | `go test ./tools/speccheck/ -run Features_AcceptsAFeatureThatAlreadyPassed -v` | exit 0, 0 findings |
| 3 | Full mode requires the evaluator | `go test ./tools/speccheck/ -run Features_Gates -v` | exit 0: `Evaluator: self` and no Evaluator line in full mode each yield exactly `docs/features.json: D-01 passes without evidence: specs/001-demo/evidence/D-01.md must come from the evaluator agent: the spec is in full mode`; `self` in quick mode 0 findings |
| 4 | History rules report exact messages, and the variable lifts them | `go test ./tools/speccheck/ -run 'Features_KeepsHistory\|TestRun_FeaturesAllows' -v` | exit 0: exactly `docs/features.json: D-02: removed`, `docs/features.json: D-01: description changed`, `docs/features.json: D-01: verify step changed`, each 0 findings when allowed; at the CLI, `errFindings` without `ALLOW_FEATURE_EDIT` and no error with `ALLOW_FEATURE_EDIT=1` |
| 5 | The CLI fails on findings, and the repository passes with real flips | `go test ./tools/speccheck/ -run TestRun_Features -v`; `make features BASE=origin/main`; `jq -r '.phases[] \| select(.id=="dev") \| .features[] \| select(.passes) \| .id' docs/features.json` | the test exits 0 (`errFindings`, `evidence/D-01.md` on stderr); the target exits 0; the jq list is exactly `DEV-01`, `DEV-02`, `DEV-03`, `DEV-05`, one per line in that order, each with `specs/007-engineering-process/evidence/<id>.md` holding `- Verdict: **PASS**` |
| 6 | Editing features.json runs the gate | `jq -r '.hooks.PostToolUse[] \| .hooks[].command' .claude/settings.json \| grep -c 'make .*features'` | `1` |

## DEV-05 · Evaluator and contract review

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The evaluator's tools are exactly the reviewed read-only set | `grep -A12 '^EVALUATOR' Makefile` | `--disallowedTools` is exactly `Edit Write NotebookEdit`; `--allowedTools` holds exactly these 37 entries, no more and no fewer: `Read`, `Glob`, `Grep`, `Bash(go test:*)`, `Bash(go vet:*)`, `Bash(go run ./tools/speccheck all)`, `Bash(go run ./tools/speccheck spec:*)`, `Bash(go run ./tools/speccheck adr)`, `Bash(go run ./tools/speccheck commands)`, `Bash(go run ./tools/speccheck features:*)`, `Bash(go run ./tools/speccheck locate:*)`, `Bash(git diff:*)`, `Bash(git log:*)`, `Bash(git show:*)`, `Bash(git status:*)`, `Bash(git grep:*)`, `Bash(git rev-parse:*)`, `Bash(grep:*)`, `Bash(ls:*)`, `Bash(cat:*)`, `Bash(head:*)`, `Bash(tail:*)`, `Bash(wc:*)`, `Bash(cut:*)`, `Bash(jq:*)`, `Bash(make check)`, `Bash(make init)`, `Bash(make spec-check)`, `Bash(make features)`, `Bash(make features BASE=origin/main)`, `Bash(make examples)`, `Bash(make test)`, `Bash(make vet)`, `Bash(make help)`, `Bash(make -s help)`, `Bash(make evaluate FEATURE=)`, `Bash(make review-contract SPEC=)` |
| 2 | It is a fresh process given the feature, spec and contract | `grep -A6 '^evaluate:' Makefile`; `go run ./tools/speccheck locate DEV-04` | the recipe runs `$(EVALUATOR)`, which is `claude -p --agent evaluator`; `dir` is assigned from `speccheck locate`; the prompt contains `$(FEATURE)`, `$$dir` and `contract.md`; locate prints exactly `specs/007-engineering-process` |
| 3 | The FR-006 scenario holds | `grep -cE '^- Verdict: \*\*(PASS\|NEEDS_WORK)\*\*$' specs/007-engineering-process/evidence/DEV-04.md`; `git status --short specs/007-engineering-process/evidence/DEV-04.md` | `1`; no output (the verify step is graded on this committed file, not rerun) |
| 4 | An evaluation commit holds only its evidence file | `git log -1 --format=%h -- specs/007-engineering-process/evidence/DEV-01.md`, then `git show --stat --format= <that sha>` | lists exactly that one file |
| 5 | The evidence file has the gate's shape | `grep -cE '^- Verdict: \*\*(PASS\|NEEDS_WORK)\*\*$' specs/007-engineering-process/evidence/DEV-01.md`; `grep -cE '^- Evaluator: evaluator ' specs/007-engineering-process/evidence/DEV-01.md` | `1` and `1` |
| 6 | The contract review answers in shape and is committed | `grep -cE '^- Review: \*\*(APPROVED\|CHANGES)\*\*$' specs/007-engineering-process/contract-review.md`; `git status --short specs/007-engineering-process/contract-review.md` | `1`; no output |
| 7 | Both targets refuse a missing argument | `make review-contract SPEC=`; `make evaluate FEATURE=` | each exits `2`, and its output contains the exact line `usage: make review-contract SPEC=<NNN>`, resp. `usage: make evaluate FEATURE=<id>` |
