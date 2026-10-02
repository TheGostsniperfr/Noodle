# ADR-0021 · A topology view can show a proposal as a diff

- Status: Accepted
- Date: 2026-10-02
- Extends: ADR-0008, ADR-0014, ADR-0019
- Pulls forward: the "visual diff colours" of phase 5

## Context

A proposal for a platform (adding an observability stack to the PAE cluster, removing a
tunnel) is reviewed on a diagram of the platform as it is. ADR-0008 already marks what is
not deployed yet: `planned` boxes are hatched and faded. That is right for a reference
diagram, where the live system is the subject. In a proposal the subject is the change:
the reader must find in seconds what the plan adds and removes, and the live system is
context. Today the change is the faintest thing on the page.

ADR-0019 solved the same question for the access matrix with `state: diff`: `planned`
is what a plan adds, `deprecated` what it removes, and the frame shows the change.
Connections carry no status, so the new arrows of a proposal cannot be told apart.

## Decision

- **`status` and `target` on connections and references**, with the values and rules of
  ADR-0008 and ADR-0014. A connection without its own status takes it from its ends:
  `planned` when either end is planned, else `deprecated` when either end is deprecated.
  A connection cannot exist before its ends do, nor outlive them.
- **`state: diff` on topology views.** Absent keeps the ADR-0008 rendering, so existing
  diagrams do not change. `current` and `target` stay matrix-only until a topology needs
  them: dropping elements also drops their layout, which is a separate problem.
- **In `diff`, the change is the subject:**
  - *added* (`planned`): drawn in full, no hatch, frame in the theme's added colour
    (the matrix's), pill `+ <target>` or `+ added`;
  - *removed* (`deprecated`): frame in the theme's removed colour, dotted, title struck,
    pill `− removed`;
  - *unchanged*: the whole element, icon, port badge and arrow at 30 % opacity.
  Arrows follow the same code: added arrows take the added colour, removed ones the
  removed colour and a dotted line; their label starts with `+` or `−`. Changed
  elements and arrows are drawn last, on top of the dimmed ones.
- **The removed colour is red** (`#f87171` dark, `#dc2626` light), not the matrix's grey:
  in a topology diff grey is what unchanged looks like.
- **The legend** lists added, removed and unchanged in place of planned and deprecated.
  Colour is never alone (ADR-0006): the pill symbols, the dotted line, the struck title
  and the opacity repeat it.

## Consequences

- One model serves the reference diagram and the proposal: the same view file with
  `state: diff` added is the proposal. When the plan ships, the statuses go and both
  views show the new platform.
- `v1alpha1` additions, no version bump.
- Phase 5 keeps the model diff between two commits and the per-PR preview. It will draw
  its result with this code.
