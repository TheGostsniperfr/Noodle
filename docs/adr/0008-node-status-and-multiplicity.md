# ADR-0008 · Node status and multiplicity are part of the model

- Status: Accepted
- Date: 2026-09-27

## Context

The first real project outside CNP (a second Kubernetes platform: platform overview,
application runtime view, repository map) needed two facts the spec cannot express:

- **A component is documented but not deployed.** A message broker, the observability
  stack and a clock service were in the docs, absent from GitOps and from the cluster.
  Readers must see at a glance that the box is not there yet.
- **A box stands for many identical instances.** One `applications/<app>` repository per
  team, one `bi/<tool>` repository per tool. Drawing seven boxes adds noise; one plain box
  hides the multiplicity.

Both were done by post-processing the generated `.drawio` (a script keyed on node IDs
listed in side files). That works but drifts from the lint: the lint does not know the
stack copies exist, and the legend does not explain either style.

## Decision

Both are facts about the system, so they live in the **model** (spec 001), not in views
or layouts. Every view of the element shows them.

- `status: planned | deprecated` on elements, zones included. Absent means live. A zone's
  status applies to every child that does not set its own.
- `multiplicity: <text>` on elements, e.g. `multiplicity: one per team`. Present means
  the element stands for many instances; the text says of what. One field, no boolean
  plus label pair.

Rendering:

- `planned`: diagonal hatch fill in the kind's muted colour, dashed border, text at 60 %,
  icon at 40 %.
- `deprecated`: dotted border, title struck through.
- `multiplicity`: two dashed copies of the box, offset 10 px up and right behind it, like
  the Kubernetes ReplicaSet icon (UML multi-object). The lint reserves their footprint for
  spacing and edge crossings.
- The legend lists each status used, and each multiplicity with its text.

Colour is not the only carrier: hatch, dash and the legend repeat the meaning (ADR-0006).

## Consequences

- The spec stays the single source of truth; no side files or post-processing.
- Both fields ship in the `v1alpha1` model schema (P1-02), so no version bump follows.
  Rendering and lint support are backlog B-05 and B-06.
- Status is declared, not observed. A later live overlay (phase 7) could derive `planned`
  from the absence of the resource in the cluster and flag disagreement with the model.
