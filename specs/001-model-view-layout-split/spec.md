# Spec 001 · Split model, views and layout (contracts v1alpha1)

- Status: **Draft**, ready for planning
- Phase: 1 (see `docs/ROADMAP.md`, acceptance in `docs/features.json` P1-*)
- Decisions it builds on: ADR-0002, ADR-0003, ADR-0004, ADR-0008

## Why

The v0 spec mixes what exists, what is shown and where it is drawn. As a result, one
diagram can show only one path. It also blocks every later level: AI agents read layout
noise, discovery cannot produce coordinates, and PR diffs compare drawings instead of
facts.

## User scenarios

1. **Tech lead.** Looks at the CNP runtime topology, then at the OIDC login sequence. Both
   come from the same model, so a component renamed once is renamed in both.
2. **AI agent.** Asked "which components talk to Keycloak, on which port?", it reads
   `model.yaml` only and answers without opening a layout or a picture.
3. **Author.** Adds a new view of an existing model, the security perspective, without
   copying a single element or connection.
4. **Future discovery adapter.** Emits a model fragment with no coordinates; layout is
   computed or pinned afterwards.

## Requirements

- **FR-001** A `Model` file describes elements (including zones), connections, references,
  ports and annotations, with no coordinates and no step numbers. Elements carry the
  optional `status` and `multiplicity` fields of ADR-0008.
- **FR-002** A `View` file selects part of a model and adds presentation: kind
  (`topology` or `sequence`), numbered scenarios, notes, cards, redaction.
- **FR-003** A `Layout` file gives, for one view, positions, sizes, edge paths, label
  anchors. It references ids only.
- **FR-004** Every file carries `apiVersion: noodle/v1alpha1` and `kind`, and validates
  against a JSON Schema in `schemas/v1alpha1/`.
- **FR-005** Referencing an id that does not exist in the model is a lint error with the
  file and the id.
- **FR-006** All v0 lint rules keep working on the split format, with the same messages.
- **FR-007** The CNP runtime example renders identically from the split format.
- **FR-008** A second view of the same model renders the OIDC login as a sequence
  (ADR-0004), numbered 1..n, responses allowed as dashed returns in sequence views only.
- **FR-009** The v0 single-file format stays readable until FR-007 passes, then is removed.

## Success criteria

- Features P1-01 to P1-06 in `docs/features.json` pass their verify step.
- Zero duplicated facts between the two CNP views: every element and connection is
  defined once, in the model.
- `scripts/init.sh` is green.

## Out of scope

Auto-layout (phase 6), the SVG renderer (phase 2), the web viewer (phase 3), tours and
IcePanel step types beyond plain numbered messages (phase 4).

## Open questions

- [NEEDS CLARIFICATION] File layout: one directory per system (`model.yaml`, `views/`,
  `layouts/`), or a single multi-document YAML? To settle in ADR-0007.
- [NEEDS CLARIFICATION] Do layouts live next to views, or in a separate tree so they can
  be regenerated without touching reviewed files?
- [NEEDS CLARIFICATION] Sequence layout: computed from step order (no coordinates at all)
  or pinned like topologies? Computed looks simpler and removes a whole class of lint work.
- [NEEDS CLARIFICATION] Where the edge `kind` lives: `auth` or `tunnel` is a property of
  the connection (model), but `blocked` is a claim a view makes. Proposal: model keeps
  connection nature, views may add `blocked` overlays.
- [NEEDS CLARIFICATION] How layouts address geometry. Field use showed absolute edge
  points are the main cost of every layout change. Settle in ADR-0007 whether `v1alpha1`
  layouts support symbolic endpoints (`node.side@ratio`, B-10), named lanes (B-11) and
  zone-relative positions with row/column layout (B-12), or reserve room for them.

## Next steps

1. Resolve the open questions with Brian, then write ADR-0007.
2. `plan.md`: technical plan (packages, schema generation, migration steps).
3. `tasks.md`: small, verifiable tasks mapped to P1-* features.
