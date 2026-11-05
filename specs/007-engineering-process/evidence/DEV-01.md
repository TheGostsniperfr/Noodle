# Evidence · DEV-01

- Feature: DEV-01
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 13a6638
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The commands check reports both directions and undescribed targets | PASS | `go test ./tools/speccheck/ -run Commands -v` gives `ok`, with 5 subtests PASS. `speccheck_test.go:250` asserts `Equal([]string{"docs/COMMANDS.md: no row for make check"})`. `:253` asserts `Equal` on `Makefile: no documented target ghost (COMMANDS.md lists it; add it with a ## comment)`. `:373` asserts `Equal` on `Makefile: target hidden has no ## description`. The `same targets` case asserts nil. |
| 2 | `spec-check` fails on each of its three checks | PASS | `go test ./tools/speccheck/ -run TestRun_Exits -v` gives `ok`, with 3 subtests PASS. `:407-410` calls `run([-root <fixture> all])`, then checks `ErrorIs(err, errFindings)` and `Equal(want+"\n", stderr)` for the three exact messages in the contract. |
| 3 | Every `run:` step of ci.yml calls make; no workflow calls a script | PASS | `grep -rc 'scripts/' .github/workflows/` prints `dependabot-automerge.yml:0`, `ci.yml:0` and `security.yml:0`. The ci.yml `run:` steps piped to `grep -vc 'run: make '` print `0`. The five steps are `make vet test`, `make examples`, `make features BASE=…`, `make spec-check` and `make vendor-check`. |
| 4 | Every script a hook names is listed under Hooks; other hook commands call make | PASS | The `jq` command prints `scripts/session-context.sh`, `scripts/stop-guard.sh` and `scripts/fmt-hook.sh`. All three are rows of the Hooks table in COMMANDS.md. The only other command contains `make -s -C "$CLAUDE_PROJECT_DIR" features`. |
| 5 | `check-features.sh` is gone and `init.sh` is a shim | PASS | `git grep --untracked …` printed nothing and exited 1. `grep -vc '^#' scripts/init.sh` gives `1`. That line is `exec make -C "$(dirname "$0")/.." init`, which matches `^exec make .*init$`. |
| 6 | `make help` lists one line per documented target | PASS | `make -s help \| wc -l` gives `17`. `grep -cE '^[\|] .make ' docs/COMMANDS.md` gives `17`. `go run ./tools/speccheck commands` exits 0 with no output. |
| 7 | The branch is green | PASS | `make init` exits 0. The last sections show `== features (base HEAD)`, `== spec-check`, `✔ Validation passed with warnings` and then `== next features`. |

## Verify step

    $ go test ./tools/speccheck/ -run Commands
    --- PASS: TestCommands_ReportsBothDirections (0.00s)
    --- PASS: TestCommands_ReportsATargetWithoutDescription (0.00s)
    ok  	github.com/TheGostsniperfr/Noodle/tools/speccheck	0.004s

    $ grep -rc 'scripts/' .github/workflows/
    .github/workflows/dependabot-automerge.yml:0
    .github/workflows/ci.yml:0
    .github/workflows/security.yml:0

    $ make init; echo "exit=$?"
    == features (base HEAD)
    go run ./tools/speccheck features -base HEAD
    == spec-check
    go run ./tools/speccheck all
    == plugin manifest
    ✔ Validation passed with warnings
    == next features (docs/features.json)
    exit=0

## Findings

None of these breaks a contract criterion:

- `tools/speccheck/speccheck_test.go:251`: the table case `target without description` has the same input and the same expected result as `same targets` (nil). It never builds a target without `##`, so it would pass whatever `CheckCommands` does with such a target. The behaviour is actually tested by `TestCommands_ReportsATargetWithoutDescription` (`:366`). Fix: remove the case, or give it a Makefile override with a target that has no `##` and expect the finding.
- Revert check: if `CheckCommands` were stubbed to return nil, the `target with no row` and `row with no target` cases, the hidden-target test and the `commands finding` CLI case would all fail. So the FR-001 tests would catch the feature being removed. FR-007 has no Go test. The contract checks it with criteria 3 and 4 instead, by grep and jq on the real files, and those catch a regression.
- Scope: the diff stays within the plan: Makefile, COMMANDS.md, CI, hooks, `tools/speccheck`, process docs and templates, and `gnumake`/`jq` in `flake.nix`. `make examples` renders the same list as the old `init.sh` and the old ci.yml steps, so no lint coverage was lost. The only change to `features.json` is the new `dev` phase, which is pure addition (C12). `gofmt -l tools/` is clean. No lint rule was touched (C04).
- Stop hook: the evaluator is read-only and does not edit `docs/PROGRESS.md`. The only changed file is `specs/007-engineering-process/evidence/DEV-01.md.tmp`, which `make evaluate` writes. The implementer commits the evidence, flips DEV-01 and adds the PROGRESS line (AGENTS.md).
