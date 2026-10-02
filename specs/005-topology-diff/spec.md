# Spec 005 · Topology diff

- Status: **Done** (`tasks.md`)
- Phase: 5a, the visual diff colours of phase 5 pulled forward for proposals (ADR-0021)
- Decisions it builds on: ADR-0006, ADR-0008, ADR-0014, ADR-0019, ADR-0021

## Why

A platform proposal is drawn on the platform as it is. The reader must see at once what
it adds and what it removes, with the live system dimmed behind.

## User scenario

**Platform engineer proposing an observability stack.** Adds Loki, Alloy and the
exporters to `model.yaml` as `planned` with `target: SP3`, adds their connections, copies
the platform view to `views/observability-proposal.yaml` with `state: diff`, and renders
it. The new parts stand out; the rest of the cluster is dimmed context.

## Requirements

- **FR-001** `status` and `target` on connections and references, checked like elements;
  a connection inherits `planned` or `deprecated` from its ends. Schemas updated.
- **FR-002** `state: diff` allowed on topology views; `current` and `target` rejected
  there with a message naming ADR-0021.
- **FR-003** The resolver marks every zone, node and edge of a diff view `added`,
  `removed` or `unchanged`, and prefixes edge labels with `+` or `−`.
- **FR-004** The renderer draws ADR-0021's code in both themes: pills, frames, opacity,
  drawing order, port badges, legend.
- **FR-005** An example in `examples/` renders in `init.sh` and CI.

## Out of scope

Model diff between two commits, per-PR preview (phase 5). `current` and `target` on
topology views.
