---
status: accepted
date: 2026-11-05
deciders: [Brian Perret]
supersedes: []
extends: ["0006"]
specs: ["011"]
constitution: [C04, C06]
---

# ADR-0033 · draw.io, SVG and PNG are three writers of one scene built from the house geometry

## Context and problem

Question 1 of spec 011. `cmd/noodle/render.go` computes styles and writes draw.io XML in
the same functions: the colour of a diff edge, the opacity of a dimmed node and the
rows of a card are decided while the `mxCell` is printed. A second writer for SVG would
have to decide them again. Spec 008 showed what two places computing one thing cost:
most of its convergence went to the gap between what the lint measured and what the
renderer drew.

## Decision drivers

- The three outputs cannot drift: a style or position is decided once.
- The lint measures what is drawn (spec 008, ADR-0026).
- `.drawio` stays editable: edges stay connectors with a source and a target.
- No behaviour change for `.drawio` while the code moves (FR-007).

## Considered options

1. **One scene, three writers.** `internal/scene` turns a resolved `diagram.Spec`, its
   theme and the house geometry into an ordered list of elements, each with its id,
   rectangle, resolved style and text runs. `internal/render/drawio`, `svg` and `png`
   only translate that list.
2. **A separate SVG writer** next to `render.go`, which mirrors its decisions.
3. **SVG only**, and the draw.io writer frozen as it is.

## Decision outcome

Chosen: **option 1**, because a style decided in the scene reaches every output, and the
lint can read the same scene the writers draw.

The scene is semantic, not a bag of shapes: an edge is an `Edge` element with its
source, its target and its full path, so the draw.io writer keeps writing a connector
and the SVG writer writes a `<path>`. Text is a list of lines, each a list of runs with
weight, colour and size, parsed once from the label markup.

### Consequences

- Good: diff, highlight and status styles are decided in one place for all outputs.
- Good: the web viewer of phase 3 reads the same scene.
- Bad: `render.go`, `sequence.go` and `matrix.go` are rewritten before anything new is
  drawn, and that refactor must keep every `.drawio` byte-identical.
- Bad: the scene carries a few draw.io-only fields, such as the label fraction, that the
  other writers ignore.

### Confirmation

`TestExamples_DrawioCurrent` keeps every committed `.drawio` byte-equal through the
refactor (FR-007). The SVG writer has no import of `internal/house` beyond fonts and
constants, checked by a test listing its imports.

## Pros and cons of the options

### Option 1

- Good: one decision per style; the lint and three writers share it.
- Bad: a large refactor first.

### Option 2

- Good: SVG output sooner.
- Bad: every later style change is made twice, and the two drift silently.

### Option 3

- Good: least code.
- Bad: diff and highlight states would stop working in the file people edit by hand.
