# Evidence · DEV-05

- Feature: DEV-05
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 9dc1743
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The evaluator's tools are exactly the reviewed read-only set | PASS | `grep -A12 '^EVALUATOR' Makefile` (Makefile:15-25). `--disallowedTools "Edit Write NotebookEdit"`. `--allowedTools` has 37 entries, counted line by line: 3+2+3+3+6+8+5+5+2. The set is the same as the contract's list, with nothing extra and nothing missing. |
| 2 | It is a fresh process given the feature, spec and contract | PASS | `EVALUATOR := claude -p --agent evaluator`. The `evaluate:` recipe sets `dir="$$($(SPECCHECK) locate $(FEATURE))"` and pipes `Evaluate feature $(FEATURE). Its spec, contract and tasks are in $$dir (spec.md, contract.md, tasks.md)…` into `$(EVALUATOR)`. `go run ./tools/speccheck locate DEV-04` printed `specs/007-engineering-process` and exited 0. |
| 3 | The FR-006 scenario holds | PASS | The `grep -cE '^- Verdict: …'` count on `evidence/DEV-04.md` is `1`. `git status --short` on that file prints nothing. The file has `- Verdict: **NEEDS_WORK**` and was committed in e4157a4. |
| 4 | An evaluation commit holds only its evidence file | PASS | `git log -1 --format=%h -- …/DEV-01.md` returned `5d16338`. `git show --stat --format= 5d16338` lists only `specs/007-engineering-process/evidence/DEV-01.md \| 50 +++`, so 1 file changed. |
| 5 | The evidence file has the gate's shape | PASS | On `DEV-01.md`, the Verdict grep counts `1` and the `^- Evaluator: evaluator ` grep counts `1`. |
| 6 | The contract review answers in shape and is committed | PASS | The `^- Review: \*\*(APPROVED\|CHANGES)\*\*$` grep on `contract-review.md` counts `1`, and the line is `- Review: **CHANGES**`. `git status --short` on that file prints nothing. |
| 7 | Both targets refuse a missing argument | PASS | `make review-contract SPEC=` printed `usage: make review-contract SPEC=<NNN>` and exited 2. `make evaluate FEATURE=` printed `usage: make evaluate FEATURE=<id>` and exited 2. |

## Verify step

The contract says the verify step is graded on the committed `DEV-04.md`, which is that step's output. This evaluation also checks the step live, because it is itself running inside `make evaluate FEATURE=DEV-05`.

    $ grep -cE '^- Verdict: \*\*(PASS|NEEDS_WORK)\*\*$' specs/007-engineering-process/evidence/DEV-04.md
    1
    $ git show --stat --format= e4157a4
     specs/007-engineering-process/evidence/DEV-04.md | 59 ++++++++++++++++++++++++
     1 file changed, 59 insertions(+)
    $ go run ./tools/speccheck locate DEV-04
    specs/007-engineering-process
    $ make spec-check; make features BASE=origin/main
    exit=0; exit=0

## Findings

None of these breaks a contract criterion. They are recorded for convergence.

- **Unknown feature id still runs the evaluator.** `Makefile:75` joins its steps with `;`. If `speccheck locate` fails (I checked `locate NOPE-99`: it prints `feature NOPE-99 is not in docs/features.json` and exits 1), `dir` is empty. The recipe then runs `mkdir -p "/evidence"`, starts a full evaluator run for a feature that does not exist, and fails only at the `mv` to `/evidence/NOPE-99.md`. Fix: write `dir="$$(…)" || exit 2` before the `mkdir`.
- **`review-contract` can overwrite a good review or write to a wrong path.**
  - It writes straight to `$$dir/contract-review.md` (`Makefile:83`) and does not use `mktemp` the way `evaluate` does (T09). If a run is interrupted or produces no output, the committed review is replaced. The Stop hook can also fire in the middle of the run.
  - If `SPEC` matches no directory, `ls -d specs/$(SPEC)-*` leaves `dir` empty and the output path becomes `/contract-review.md`. If it matches two directories, `dir` holds both.
  - Fix: use the same tmp-then-`mv` pattern as `evaluate`, and add a guard that requires exactly one match.
- **T04 is checked off but DEV-05 is not flipped.** `specs/007-engineering-process/tasks.md` T04 is `[x]` and says "Flip **DEV-05**", but `docs/features.json:101` still has `"passes": false` for DEV-05. The checkbox claims more than has happened. This evidence is the condition for that flip. Fix: flip DEV-05 now that this evidence exists, or reword T04.
- **No automated test covers FR-006.** Criteria 1, 2 and 7 are proven by grepping the recipe and by running the argument guards. Nothing in `go test` would catch a regression in the `evaluate` recipe, such as the prompt losing `contract.md`. That is acceptable under this contract. A Makefile change is only caught by rerunning these greps.
