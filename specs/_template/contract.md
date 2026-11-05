# Contract NNN · Title

What "done" means for each feature, agreed before implementation (ADR-0024).
Drafted by the implementer, reviewed with `make review-contract SPEC=NNN` until the
evaluator answers APPROVED. One failed criterion fails the feature.

- Agreed: <date> · implementer <model> · evaluator <model>

## PNN-01 · <feature description>

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | <observable behaviour> | `<command>` | <exact expectation, e.g. exit 0 and output equals golden> |
| 2 | Tests fail when the behaviour is removed | revert the change, run `make test` | at least one test fails |
| 3 | No change outside the plan's packages | `git diff --stat origin/main...` | only the listed paths |
