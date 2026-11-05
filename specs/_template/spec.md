# Spec NNN · Title

- Status: **Draft** | **Clarified** | **Planned** | **In progress** | **Done**
- Mode: **full**
- Phase: <roadmap phase id>, and why now (C11)
- Issue: #NN · Backlog: B-NN
- Decisions it builds on: ADR-NNNN
- Invariants it touches: CNN

## Why

The problem, who has it, and the evidence (a diagram, a repository, a field report).
No solution here.

## User scenarios

**Role doing something.** What they do, step by step, and what they see at the end.
One scenario per distinct use; they become the acceptance tests.

## Requirements

One behaviour each, in EARS, each with a scenario a test can run.

- **FR-001** WHEN <event> THE SYSTEM SHALL <behaviour>.
  *Scenario:* WHEN <concrete input> THEN <observable output>.
- **FR-002** IF <unwanted condition> THEN THE SYSTEM SHALL <behaviour>.
  *Scenario:* WHEN <concrete input> THEN <observable output>.

## Properties

Statements true for every input, each tested property-based. Delete the section if none.

- For every <input class>, <property>.

## Out of scope

What a reader could expect here and will not get, with where it goes instead.

## Clarifications

Questions asked before planning, five at most per pass, and Brian's answers. An answer
that sets direction is also an ADR.

| # | Question | Answer | ADR |
|---|---|---|---|

## Success criteria

Features `PNN-*` in `docs/features.json` pass their verify step, graded in `contract.md`.
