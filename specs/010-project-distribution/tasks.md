# Tasks 010 · Project and distribution

Plan: [`plan.md`](plan.md). One task per commit; each ends with `make check` green.
Tasks in one wave do not depend on each other. Evaluate each feature with
`make evaluate FEATURE=<id>` after its tasks; commit the evidence alone, then the flip.

## Wave 1

- [x] **T01** `LICENSE`, `NOTICE` and the `license` fields of `plugin.json`, `flake.nix`
  and `SKILL.md` (FR-001).
- [x] **T02** `assets/icons/ATTRIBUTION.md` and `assets/attribution_test.go` (FR-002).
  Flip **PRJ-01** with T01.
- [x] **T03** `speccheck skills`, run by `all`; table-driven tests per rule (FR-004).
- [x] **T04** Skill wording: Claude Code-only names moved under "In Claude Code" (FR-004).
  Flip **PRJ-03** with T03.
- [x] **T05** `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, issue templates,
  README links (FR-008). Flip **PRJ-05**.
- [x] **T06** `noodle version` with the ldflags variable; the flake passes its revision
  (FR-012).

## Wave 2

- [ ] **T07** Fork triage record: commits, build and test result per clarification 2,
  one disposition per row, draft reply (FR-003). Only once T01 is on `main`. Brian
  posts the reply. Flip **PRJ-02**.
- [x] **T08** Read the Codex subagents page and write the agent TOML keys into
  `plan.md`; `content.go` embeds `skills` and `agents` (FR-005, FR-006).
- [x] **T09** `noodle setup opencode` and `noodle setup codex`, `-project` and `-user`,
  idempotent, with golden-file tests in temporary directories (FR-005, FR-006).
- [ ] **T10** The marker line, refusal to overwrite foreign files, and the README
  install section per host (FR-007). Flip **PRJ-04** with T08 and T09.
- [ ] **T11** `docs/guide/` tutorial (its system in `examples/tutorial`, linted by
  `make examples`), how-to pages, explanation, `make cli-doc`, `cli.md` and its test
  (FR-009).
- [ ] **T12** README positioning and the VISION paragraph (FR-010); the "Discovery"
  section of `reference.md` (FR-011). Flip **PRJ-06** with T11.

## Wave 3

- [ ] **T13** Release configuration, release workflow, changelog config, `make changelog`,
  `make release-snapshot` (in the evaluator allowlist, risk noted in ADR-0024), the
  `version` field of `plugin.json`, and the first `CHANGELOG.md` (FR-013).
- [ ] **T14** `bin/noodle` download fallback with checksum verification, and its Go test
  with a stub `curl` (FR-014). Flip **PRJ-07** with T06 and T13.
- [ ] **T15** The three process skills `issue-writer`, `release-notes`, `contrib-check`
  where clarification 4 says (FR-015). Flip **PRJ-08**.

## Convergence

Tasks the evaluator found missing, appended here; never edit the waves above.
