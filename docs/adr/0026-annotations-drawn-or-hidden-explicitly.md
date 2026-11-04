---
status: accepted
date: 2026-11-04
deciders: [Brian Perret]
supersedes: []
extends: ["0006", "0007"]
specs: ["008"]
constitution: [C04, C06]
---

# ADR-0026 · An annotation targets something drawn, or the view hides it explicitly

## Context and problem

Annotations (gaps such as CNP G1) can target any element or connection, but a view only
draws what it includes. Today an annotation whose targets are all outside the view
disappears with no trace (#86), and an edge whose label is emptied with `labels: {c1: ""}`
still shows ` · ⚠ G1` with a leading separator (#78). Spec 008 asked how both should
behave; Brian answered on 2026-11-04 (clarifications 2 and 4). A warning was rejected:
warnings that are always there become noise nobody reads.

## Decision drivers

- Nothing is skipped silently: a gap the reader cannot see is either a mistake or a choice.
- A choice is written down, with its reason, where the next reader finds it: in the view.
- No warning level that becomes normal noise.
- `""` keeps meaning "no text" (B-28), not "no annotations".

## Considered options

1. **Error unless the view acknowledges it**: an annotation with no drawn target fails
   the check; a view may list it in `hidden_annotations: {id: reason}`.
2. **A lint warning** naming the annotation, never blocking.
3. **A generated line in the Gaps card**: "not shown in this view: G1".

## Decision outcome

Chosen: **option 1**, because it keeps the lint binary (pass or fail) and turns every
hidden gap into a reviewed, explained line of the view.

- **`hidden_annotations: {<annotation id>: <reason>}`** on a view. The reason is a
  non-empty string. It is an error to list an id that is not an annotation of the model,
  an annotation that has a target drawn in the view (nothing to acknowledge), or an
  empty reason.
- An annotation of the model whose targets are all outside the view is an **error**
  unless the view lists it in `hidden_annotations`.
- **`hide_annotations: true`** on a view hides the annotation badges on everything it
  draws. Annotations it would have shown are not an error: the view states its choice for
  all of them at once. Annotations with no drawn target still need `hidden_annotations`.
- **`labels: {<id>: ""}`** hides the label text only. Badges stay, rendered alone with no
  separator.
- Checked in `internal/model` (view checks), from what the view includes, before
  rendering; both fields are added to `schemas/v1alpha1/view.schema.json` as optional,
  additive fields of the alpha contract (C06).

### Consequences

- Good: a view that leaves out an annotated connection says so, with a reason, in review.
- Good: no warning channel to maintain or ignore.
- Bad: existing views with undrawn annotations fail until they list them; spec 008's
  FR-010 fixes those views in the same batch.

### Confirmation

- `go test ./internal/model/...` has a failing case for an undrawn annotation, an empty
  reason, an unknown id and an acknowledged id that is drawn, and a passing case for each
  acknowledged one; spec 008 contract LNT-05.
- `make examples` is green with every example's undrawn annotations acknowledged.

## Pros and cons of the options

### Error unless acknowledged

- Good: binary, reviewed, explained.
- Bad: one more field for authors to learn.

### Warning

- Good: no new field.
- Bad: warnings that never block are read once and then ignored.

### Line in the Gaps card

- Good: the reader sees the gap exists.
- Bad: puts text on the diagram for something the view chose to leave out, and needs a Gaps card.
