# Contract 007 · Engineering process

What "done" means for each feature, agreed before implementation (ADR-0024).
One failed criterion fails the feature.

- Agreed: pending `make review-contract SPEC=007`. Bootstrap exception: this contract is
  reviewed after implementation, since the review command is part of what it grades.
- Round 1 (2026-11-04): CHANGES, 17 points, all applied; FR-007 amended (clarification 3).
- Round 2 (2026-11-04): CHANGES, 13 points, all applied; FR-002 now declares the flip rule;
  tests assert exact finding lists.

## DEV-01 · Makefile is the only entry point

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The commands check reports both directions and undescribed targets | `go test ./tools/speccheck/ -run Commands -v` | a target missing from COMMANDS.md yields exactly `docs/COMMANDS.md: no row for make check`; a row with no target yields exactly `Makefile: no documented target ghost (COMMANDS.md lists it; add it with a ## comment)`; a target without `##` yields exactly `Makefile: target hidden has no ## description`; matching sets yield 0 findings |
| 2 | Findings fail the command | `go test ./tools/speccheck/ -run TestRun -v` | `run(... all)` returns `errFindings` and stderr holds the exact finding line |
| 3 | Every CI step calls make, no workflow calls a script | `grep -rc 'scripts/' .github/workflows/` and `grep -E '^\s+run:' .github/workflows/ci.yml \| grep -vc 'run: make '` | `0` for every file, and `0` |
| 4 | Every script a hook names is listed under Hooks | `jq -r '.. \| .command? // empty' .claude/settings.json \| grep -o 'scripts/[a-z-]*\.sh' \| sort -u` and the Hooks table of `docs/COMMANDS.md` | each printed path appears in that table; and `jq -r '.. \| .command? // empty' .claude/settings.json \| grep -v 'scripts/[a-z-]*\.sh' \| grep -vc 'make '` prints `0` |
| 5 | `check-features.sh` is gone and `init.sh` is a shim | `git grep -n 'check-features.sh\|scripts/init.sh' -- ':!docs/features.json' ':!docs/PROGRESS.md' ':!specs/007-engineering-process'` and `grep -vc '^#' scripts/init.sh` | `git grep` exits 1 with no output; the count is `1` and that line matches `^exec make .*init$` |
| 6 | `make help` and COMMANDS.md list the same targets | `diff <(make -s help \| awk '{print $1}' \| sort) <(grep -oE '^[\|] .make [a-z-]+' docs/COMMANDS.md \| sed 's/.*make //' \| sort)` | exit 0, no output |
| 7 | The branch is green | `make init` | exit 0 |

## DEV-02 · Traceability of specs

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Each rule has a failing case whose finding list is exactly its message | `go test ./tools/speccheck/ -run Spec_ReportsEachBrokenRule -v` (cases assert `Equal([]string{want}, findings)`) | exit 0, with subtests for: unknown `Cnn`, missing ADR, no SHALL, no scenario, named by no task (exactly `specs/001-demo/tasks.md: FR-002 is named by no task`), covered by no feature, feature without verify, feature covering an undeclared FR, feature without contract criteria, missing contract, unknown flipped feature, spec with no phase, `Cnn` in `tasks.md`, `Cnn` in `plan.md`, ADR in `contract.md`; and a valid case with 0 findings |
| 2 | Legacy specs are not held to full-mode rules | `go test ./tools/speccheck/ -run Spec_LegacySpec -v` | PASS: prose requirements without a mode produce 0 findings |
| 3 | The repository passes | `make spec-check` | exit 0 |

## DEV-03 · ADR shape and index

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Frontmatter ADRs are checked piece by piece | `go test ./tools/speccheck/ -run ADR_ChecksFrontmatter -v` | exit 0; seven subtests (no status, unknown status, no date, no deciders, one option, no Confirmation, empty Confirmation) each assert exactly one finding with its message; the valid ADR yields 0 findings |
| 2 | Old ADRs parse from their status line | `go test ./tools/speccheck/ -run ADR_Legacy -v` and `make spec-check` | PASS, and exit 0 with ADRs 0001 to 0023 in the repository |
| 3 | A stale index is reported, and `-write` fixes it | `go test ./tools/speccheck/ -run ADR_ReportsStaleIndex -v` | the stale run yields exactly `docs/adr/README.md: index is stale: run make adr-index`; after `-write`, 0 findings |
| 4 | The repository index is current and complete | `make adr-index && git diff --exit-code docs/adr/README.md`; `grep -c '^| \[0' docs/adr/README.md` and `ls docs/adr/[0-9]*.md \| wc -l` | exit 0; the two counts are equal |

## DEV-04 · Evidence gate and history guard

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | A new pass needs PASS evidence, named with its command | `go test ./tools/speccheck/ -run Features_Gates -v` | no evidence: exactly one finding containing `evidence/D-01.md` and `make evaluate FEATURE=D-01`; `NEEDS_WORK` evidence rejected; evaluator PASS accepted |
| 2 | A feature that passed at base needs nothing | `go test ./tools/speccheck/ -run Features_AcceptsAFeatureThatAlreadyPassed -v` | 0 findings |
| 3 | Full mode requires the evaluator | `go test ./tools/speccheck/ -run Features_Gates -v` | `Evaluator: self` and no Evaluator line are rejected in full mode; `self` is accepted in quick mode |
| 4 | History rules report exact messages | `go test ./tools/speccheck/ -run Features_KeepsHistory -v` | exactly `D-02: removed`, `D-01: description changed`, `D-01: verify step changed` (prefixed `docs/features.json: `), and 0 findings for each with `ALLOW_FEATURE_EDIT` |
| 5 | The CLI and the target fail on findings | `go test ./tools/speccheck/ -run TestRun_Features -v`; `make features BASE=origin/main` | `run(... features -base HEAD)` returns `errFindings` with `evidence/D-01.md` on stderr; the target exits 0 on the branch |
| 6 | Editing features.json runs the gate | `jq -r '.hooks.PostToolUse[] \| .hooks[].command' .claude/settings.json \| grep -c 'make .*features'` | `1` |

## DEV-05 · Evaluator and contract review

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The evaluator has no Write or Edit tool, and its run changes nothing else | `make -n evaluate FEATURE=DEV-04 \| grep -o -- '--disallowedTools "[^"]*"'`; on a clean commit where `DEV-04.md` is committed, `make evaluate FEATURE=DEV-04`, then `git diff --name-only` and `git status --porcelain`; restore with `git checkout` afterwards | lists Edit, Write and NotebookEdit; both commands list only `specs/007-engineering-process/evidence/DEV-04.md` |
| 2 | It is a fresh process given the feature and its spec | `make -n evaluate FEATURE=DEV-04` | the recipe runs `claude -p --agent evaluator`; the prompt contains `DEV-04`, `specs/007-engineering-process` and `contract.md`; `go run ./tools/speccheck locate DEV-04` prints exactly `specs/007-engineering-process` |
| 3 | The evidence file has the gate's shape | `grep -cE '^- Verdict: \*\*(PASS\|NEEDS_WORK)\*\*$' specs/007-engineering-process/evidence/DEV-04.md`; `grep -cE '^- Evaluator: evaluator ' specs/007-engineering-process/evidence/DEV-04.md` | `1` and `1` |
| 4 | Contract review answers in shape, and both targets refuse a missing argument | `make review-contract SPEC=001`, then `rm specs/001-model-view-layout-split/contract-review.md`; `make review-contract SPEC= 2>&1 >/dev/null`; `make evaluate FEATURE= 2>&1 >/dev/null` | exit 0 and `grep -cE '^- Review: \*\*(APPROVED\|CHANGES)\*\*$'` on the written file = `1`; the two others exit non-zero and print exactly `usage: make review-contract SPEC=<NNN>` and `usage: make evaluate FEATURE=<id>` |
