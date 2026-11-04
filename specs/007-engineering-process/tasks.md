# Tasks 007 · Engineering process

Plan: [`plan.md`](plan.md). Each task ends with `make check` green.

## Wave 1

- [x] **T01** `tools/speccheck` with table-driven tests: `spec` (FR-002, FR-003), `adr`
  (FR-004), `commands` (FR-001), `features` (FR-005, FR-008), `locate`.
- [x] **T02** `docs/CONSTITUTION.md`, `docs/PROCESS.md`, templates, ADR-0024 in the new
  format, generated ADR index.

## Wave 2

- [x] **T03** `Makefile` and `docs/COMMANDS.md`; CI, hooks, README, AGENTS.md and the
  session context call `make`; `scripts/check-features.sh` removed, `scripts/init.sh` a shim to `make init`;
  `gnumake` in the flake (FR-001, FR-007). Flip **DEV-01**.
- [x] **T04** Evaluator agent, `make evaluate` and `make review-contract` (FR-006).
  Flip **DEV-05**.

## Wave 3

- [ ] **T05** Contract reviewed, every feature evaluated; flip **DEV-02**, **DEV-03**,
  **DEV-04** on their evidence.

## Convergence

- [x] **T06** Drop the `target without description` case of `TestCommands_ReportsBothDirections`:
  same input as `same targets`, so it tested nothing (DEV-01 evidence, finding 1).
- [x] **T07** Contract criteria are the `| n |` rows under a `## <id>` section, not any
  mention of the id; tests for an empty section and for an id only in prose (DEV-02, 1-2).
- [x] **T08** Citations inside fenced code blocks are quotations too; a requirement is
  named by a task line, not by prose in `tasks.md` (DEV-02, findings 3 and 4).
- [x] **T09** `make evaluate` writes its draft with `mktemp`, outside the tree, so the Stop
  hook does not fire mid-evaluation; on failure the draft path is printed (DEV-05; landed
  in `674fac9`, noted by the DEV-02 re-evaluation).
- [x] **T10** Contract criteria ignore fenced examples (DEV-02 re-evaluation, finding 3).
- [x] **T11** Considered options are the numbered items only, so a bullet note no longer
  counts as a second option (DEV-03, finding 1).
- [x] **T12** The gate reads only a filled verdict line; backlog evidence lives in
  `docs/evidence/`; an unknown `BASE` is an error, not a skipped history check, except
  GitHub's forty-zero "before" of a first push (DEV-04 first run, findings 3 to 5).
