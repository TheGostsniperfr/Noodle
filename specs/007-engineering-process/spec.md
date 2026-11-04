# Spec 007 · Engineering process: traceable specs, evaluated features, one command entry

- Status: **In progress**
- Mode: **full**
- Phase: dev, the engineering process, outside the product roadmap. Needed before the
  specs of #16, #18 and #93 to #107 are written (C11 does not apply: no product code).
- Issue: #108 · Backlog: B-59
- Decisions it builds on: ADR-0024
- Invariants it touches: C10, C12, C14

## Why

Seventeen issues are about to become specs, mostly implemented by agents. Agents
declare done when code compiles and grade their own work generously; the NoodleOpenCode
fork recorded both in `docs/agent-lessons.md`. Specs 001 to 006 have prose requirements
no tool traces to a test, and the commands an agent needs are spread over six scripts
called from `AGENTS.md`, the hooks, CI and the README.

## User scenarios

**An agent picks up a full-mode spec.** It reads `docs/COMMANDS.md`, writes requirements
in EARS, adds features with `covers`, drafts `contract.md` and gets it approved with
`make review-contract`. It implements, runs `make evaluate FEATURE=…`, and only flips
`passes` once the evidence file says PASS. `make check` tells it at each step what is
missing, with the file.

**Brian reviews a PR.** Each flipped feature links an evidence file with the verify
output and the evaluator's grading. The ADR index is current without anyone editing it.

## Requirements

- **FR-001** THE SYSTEM SHALL expose every developer command as a Makefile target with
  a `##` description, and `make spec-check` SHALL fail when `docs/COMMANDS.md` and the
  documented targets differ.
  *Scenario:* WHEN a target `foo` is added to the Makefile with no row in COMMANDS.md
  THEN `make spec-check` fails with `docs/COMMANDS.md: no row for make foo`.
- **FR-002** WHEN a spec, plan, tasks or contract file cites `Cnn` or `ADR-NNNN`, THE
  SYSTEM SHALL fail `make spec-check` unless the invariant is in `docs/CONSTITUTION.md`
  and the ADR file exists, a citation inside inline code being a quotation, not a
  citation; and WHEN `tasks.md` flips a feature, that feature SHALL exist
  in `docs/features.json`.
  *Scenario:* WHEN a spec cites an invariant number the constitution does not define THEN `make spec-check` fails naming the spec and that number.
- **FR-003** WHEN a spec declares `Mode: full`, THE SYSTEM SHALL fail `make spec-check`
  for each requirement without SHALL, without a scenario with WHEN and THEN, named by no
  task, or covered by no feature of the spec's phase, for each feature without a verify
  step or with a `covers` the spec does not declare, and for each feature with no
  criteria in `contract.md`.
  *Scenario:* WHEN FR-002 of a full-mode spec is named by no task THEN the check fails
  with `tasks.md: FR-002 is named by no task`.
- **FR-004** WHERE an ADR has YAML frontmatter, THE SYSTEM SHALL require a known status,
  a date, deciders, two considered options and a Confirmation section; and THE SYSTEM
  SHALL generate the index of `docs/adr/README.md` from every ADR, old or new, failing
  `make spec-check` when it is stale.
  *Scenario:* WHEN an ADR with frontmatter lists one option THEN the check fails; WHEN
  an ADR is added and `make adr-index` was not run THEN the check reports the stale index.
- **FR-005** WHEN `docs/features.json` sets `passes: true` on a feature that did not
  pass at the base ref, THE SYSTEM SHALL fail `make features` unless
  `<spec dir>/evidence/<id>.md` exists with `- Verdict: PASS`, and, for a full-mode
  spec, `- Evaluator: evaluator …`.
  *Scenario:* WHEN DEV-02 is flipped with no evidence file THEN `make features` fails
  naming the missing file and the command that writes it.
- **FR-006** WHEN `make evaluate FEATURE=<id>` runs, THE SYSTEM SHALL start the
  read-only evaluator agent in a fresh process, give it the feature, its spec and
  contract, and write its report to the feature's evidence file; `make review-contract
  SPEC=NNN` SHALL do the same for a contract, answering APPROVED or CHANGES.
  *Scenario:* WHEN `make evaluate FEATURE=DEV-04` runs THEN
  `specs/007-engineering-process/evidence/DEV-04.md` exists with a Verdict line.
- **FR-007** THE SYSTEM SHALL make every `run:` step of `ci.yml` call a `make` target, and
  no step of any workflow SHALL call a file under `scripts/`. A Claude Code hook that runs a check SHALL call
  `make`; a hook-only script (session context, stop guard, formatter) SHALL be listed
  under "Hooks" in `docs/COMMANDS.md`.
  *Scenario:* WHEN `.github/workflows/ci.yml` is searched for `scripts/` THEN nothing
  matches; WHEN a hook command names `scripts/stop-guard.sh` THEN COMMANDS.md lists it.
- **FR-008** IF a change removes a `features.json` entry or changes its description or
  verify step THEN THE SYSTEM SHALL fail `make features`, unless `ALLOW_FEATURE_EDIT=1`.
  *Scenario:* WHEN the description of P0-01 is edited THEN `make features` fails with
  `P0-01: description changed`.

## Out of scope

- Rewriting specs 001 to 006 or ADRs 0001 to 0023 in the new shape (C10).
- A Stop hook that loops on the evaluator until PASS: evaluation stays an explicit step.
- Vendoring a property-based testing library: done by the first spec that needs it.

## Clarifications

| # | Question | Answer | ADR |
|---|---|---|---|
| 1 | Scripts or Makefile as the entry point? | Makefile, plus one Markdown reference for agents (2026-11-04) | ADR-0024 |
| 2 | Adopt Spec Kit or OpenSpec wholesale? | No: keep our shape, add the best pieces, enforce with code (2026-11-04) | ADR-0024 |
| 4 | How many contract review rounds? | Three at most; after that only write-capability and always-pass or always-fail criteria are fixed (2026-11-04). Spec 007 ran five, before the cap existed | ADR-0024 |
| 3 | Must hooks call `make` too? (raised by the contract review) | Hooks that run a check do; hook-only scripts stay as hook implementations and are listed in COMMANDS.md, since a `make` target per hook would only add noise (2026-11-04) | ADR-0024 |

## Success criteria

Features `DEV-*` in `docs/features.json` pass their verify step, graded in `contract.md`.
