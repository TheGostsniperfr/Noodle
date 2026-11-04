---
status: accepted
date: 2026-11-04
deciders: [Brian Perret]
supersedes: []
extends: ["0024"]
specs: ["008", "009", "010"]
constitution: [C11, C12]
---

# ADR-0025 · The backlog is worked in thematic batches, in waves, all in full mode

## Context and problem

About sixty board items are open after the field reports of #71 to #92 and the ideas of
#93 to #107. Taken one by one through ADR-0024's process (spec, contract review,
evaluation), each costs a session; most also sit outside the current phase (6a), and C11
asks for an ADR to pull each one forward. At that pace the backlog never shrinks.

## Decision drivers

- Throughput: several items per spec, several specs at once.
- Safety stays as in ADR-0024: Brian wants every batch in full mode.
- One decision about order instead of one per item (C11).
- Brian answers one questionnaire per wave, not a question per ticket.

## Considered options

1. **One spec per ticket**, in board order, as today.
2. **Thematic batches in waves**: one full-mode spec per batch of related tickets,
   batches of a wave specified together and implemented in parallel worktrees.
3. **One big spec for the whole backlog.**

## Decision outcome

Chosen: **option 2**, because it keeps a spec small enough to review (one theme) while
cutting the fixed cost per ticket, and lets independent batches run side by side.

| Wave | Batch | Spec | Items |
|---|---|---|---|
| 1 | A · Lint robustness | 008 | #36, #44, #45, #72, #73, #78, #86, #87, #90 |
| 1 | B · Discovery, phase 6a completion | 009 | spec 003 T05 to T16, #75, #76, #79, #80, #81, #16, #25 |
| 1 | E · Project and distribution | 010 | #93, #94, #100, #106, #107, #88 |
| 2 | C · Check and analysis | next | #95, #96, #97, #98, #30, #74, umbrella #69 |
| 2 | D · Export pipeline (roadmap phase 2) | next | #82, #71, #91, #24, #34, #101 |
| 3 | F · Presentation | next | #92, #18, #62, #99, #85, #26, #28, #29, #27 |
| 3 | G · Layout and routing | next | #77, #33, #32, #31, #21, #22, #35 |
| later | H · Ideas, ADR first | none yet | #102, #103, #104, #105, #23, #84, #83, #43, #70, #15, #17 |

- This ADR is the C11 agreement to pull batches A, C, D, E, F and G forward. Batch B is
  the current phase.
- Every batch spec is in **full mode**; quick mode stays in PROCESS.md but is not used
  for these batches.
- A wave's specs are written together; their open questions go to Brian as one
  questionnaire. Each batch is then implemented on its own branch and worktree, by a
  Sonnet implementer when the tasks are clear; evaluations of a batch's features run as
  parallel `make evaluate` processes.
- Spec 003 stays as written (C10). Spec 009 takes over its open tasks; tasks of spec 003
  that spec 009 covers are ticked there with a pointer to spec 009.

### Consequences

- Good: a wave closes about twenty items with three specs and three contract reviews.
- Good: one ordering decision replaces about twenty pull-forward ADRs.
- Bad: a batch spec is larger than a ticket spec; the three-round review cap matters more.
- Bad: parallel branches touching the same package (lint for A, adapter for B) can
  conflict; batches are cut along package lines to keep that rare.

### Confirmation

- Specs 008, 009 and 010 declare `Mode: full`, cite ADR-0025, and pass `make spec-check`.
- Every item of a wave's table is named in its batch spec; a wave is done when each
  item is closed or explicitly moved by a later ADR.

## Pros and cons of the options

### One spec per ticket

- Good: smallest review unit.
- Bad: sixty sessions, sixty pull-forward decisions.

### Batches in waves

- Good: fixed costs shared, parallel work, one questionnaire per wave.
- Bad: bigger specs, possible merge conflicts between batches.

### One spec for everything

- Good: one review.
- Bad: unreviewable, and one blocking question stalls everything.
