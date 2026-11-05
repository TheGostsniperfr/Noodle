---
status: accepted
date: 2026-11-05
deciders: [Brian Perret]
supersedes: []
extends: ["0006"]
specs: ["011"]
constitution: [C04, C05, C13]
---

# ADR-0036 · The SVG and PNG noodle draws are the reference image; `-draft` draws a failing view with its findings

## Context and problem

Questions 4 and 5 of spec 011, and FR-006. Reviews compare PNGs exported by draw.io
(C13). Once noodle draws its own, one of the two has to be the image a PR shows and a
golden test holds. Separately, C04 says a diagram that fails the lint is not emitted,
so fixing 52 findings is done without seeing them (#82).

## Decision drivers

- One reference image, reproducible byte for byte from git (C05).
- Review stays visual (C13), and works in CI where draw.io is absent.
- A failing diagram is never mistaken for a finished one (C04).

## Considered options

1. **noodle's SVG and PNG are the reference** for examples, goldens and PR
   before/after, for every view type, in waves. `.drawio` stays committed for hand
   edits. `-draft` writes a failing view with each finding drawn as a red marker and
   exits 1.
2. **draw.io's PNG stays the reference** until the SVG has been compared on every
   example; no `-draft`.

## Decision outcome

Chosen: **option 1**, because draw.io's PNG cannot be reproduced in CI and differs by
machine, which is what this phase removes.

- Examples commit `.svg`, `.light.png` and `.dark.png` from noodle; `make examples`
  fails when one is not current. The `.drawio` files stay and stay checked.
- Topology views come first, then sequence and landscape, then catalog and access
  matrix. A view type not yet drawn falls back to the draw.io export path until its
  wave lands.
- C04 is amended: a diagram that fails the lint is not emitted **as a deliverable**.
  `-draft` writes the image with every finding drawn on it, a red outline and its rule
  id at the rectangle the finding names plus a banner `DRAFT · N lint findings`, and
  exits 1. It refuses `.drawio`, so a draft never becomes the editable source.

### Consequences

- Good: PR before/after images can be made in CI, and goldens are exact.
- Good: a failing diagram can be fixed while looking at it.
- Bad: every committed example PNG changes once, in one reviewed PR per wave.
- Bad: someone can commit a draft image; the banner and the non-zero exit make it
  visible.

### Confirmation

`TestExamples_ImagesCurrent` re-renders every example and compares bytes. A test of
`-draft` checks the exit status, the banner, one marker per finding, and that
`-draft -o x.drawio` is refused. `docs/CONSTITUTION.md` C04 cites this ADR.

## Pros and cons of the options

### Option 1

- Good: reproducible, CI-friendly reference; fixing with eyes open.
- Bad: one-off churn of every example image.

### Option 2

- Good: no visible change until the SVG is trusted.
- Bad: reviews keep depending on a desktop app, and the goldens stay inexact.
