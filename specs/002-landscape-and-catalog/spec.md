# Spec 002 · Landscape and catalog views

- Status: **Planned** (`plan.md`, `tasks.md`)
- Phase: 1b, an extension agreed with Brian on 2026-09-28 before phase 1 closes
  (T13 to T17 of spec 001 stay open and follow this spec)
- Decisions it builds on: ADR-0006, ADR-0007, ADR-0008, ADR-0009, ADR-0012, ADR-0013,
  ADR-0014

## Why

Platform teams present two more pictures than topology and sequence: the **technology
stack** (a grid of logos by concern) and the **service catalogue** (what the platform
offers and how to get it). Today both are drawn by hand or by image generators, off the
house style and off the facts. Both are derived from what the model knows.

The `planned` status of ADR-0008 is part of both pictures ("coming soon") but is still
rendered by project post-processing (backlog B-05). It moves into the core first.

## User scenarios

1. **Tech lead presenting a sprint review.** Renders `tech-stack` and `service-catalog`
   from the platform's model, dark or light, and drops the PNGs in the deck. Items not
   deployed yet are hatched with their target sprint.
2. **Platform engineer retiring a tool.** Removes the element from `model.yaml`; lint
   fails on the landscape view that still lists it, so the stack cannot lie.
3. **AI agent.** Asked "how does my app get a database?", reads `offerings` in
   `model.yaml` and answers with the request line and the backing components.

## Requirements

- **FR-001** Elements and zones with `status: planned` render hatched, dashed, text at
  60 %, icon at 40 %; `deprecated` renders dotted with the title struck through; the
  legend lists the statuses used (ADR-0008, B-05).
- **FR-002** `target` on a planned element, zone or offering renders as a pill; `target`
  without `status: planned` is a check error (ADR-0014).
- **FR-003** Element kind `tool` exists in the model schema and renders with the
  `external` palette unless a view overrides the colour (ADR-0012).
- **FR-004** A `landscape` view with `bands` (sections of items), an optional `side`
  column and `labels` resolves, lints and renders with no layout file (ADR-0012).
- **FR-005** Landscape checks: unknown item, duplicate item in one view (errors), item
  without icon (warning).
- **FR-006** `offerings` in the model with `backed_by` references checked (ADR-0013).
- **FR-007** A `catalog` view with `include` and `columns` resolves, lints and renders
  one card per offering with no layout file.
- **FR-008** JSON Schemas in `schemas/v1alpha1/` cover every new field and type; every
  example validates in `go test`.
- **FR-009** DockAir's `postprocess-drawio.py` output equals the core rendering of
  `planned` on its three diagrams, so the script can be deleted.

## Success criteria

- Features P1b-01 to P1b-07 in `docs/features.json` pass their verify step.
- `examples/landscape` and `examples/catalog` render dark and light in `init.sh` and CI.
- The DockAir tech stack and service catalogue render from its model, reviewed by the
  `noodle:diagram-reviewer` agent with no blocking finding.

## Out of scope

Sequence views (spec 001, T14 and T15), auto-layout (phase 6), interactive filtering of
the landscape (phase 3 viewer), costs or owners as landscape overlays.

## Open questions

None left: model placement, `tool`, `offerings` and `target` were agreed with Brian on
2026-09-28 and are written as ADR-0012 to ADR-0014.
