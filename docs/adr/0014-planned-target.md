# ADR-0014 · A planned element names when it is expected

- Status: Accepted
- Date: 2026-09-28
- Extends: ADR-0008 (`status: planned`)

## Context

ADR-0008 marks a documented but undeployed component as `planned`. Readers then ask
"when?". On the DockAir platform the answer differs per item: the message broker is due
in sprint 1, the observability, FinOps and backup stacks later. Diagrams so far wrote it
by hand in the title ("RabbitMQ · soon"), which the lint cannot check and a landscape
cannot align.

## Decision

- **`target: <text>`** on elements and offerings, e.g. `target: SP1`, `target: Q1 2027`.
  Free text: teams plan in sprints, quarters or releases.
- Valid only with `status: planned`; the check rejects it otherwise. A zone's target
  applies to children that set none, like its status.
- Rendering: a small pill with the target text in the corner of the hatched box or tile.
  A planned item without a target shows "planned".

## Consequences

- One source for "what is coming and when", shown the same way in every view type.
- An addition to `v1alpha1`: no version bump.
