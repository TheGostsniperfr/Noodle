# Spec 008 · Lint robustness: what passes the lint is what the reader sees

- Status: **Done**
- Mode: **full**
- Phase: lint2, batch A of wave 1, pulled forward by ADR-0025 (C11)
- Issue: #22, #36, #44, #45, #72, #73, #78, #86, #87, #90 · Backlog: B-07, B-20, B-25, B-26,
  B-28 (follow-up), B-30, B-31
- Decisions it builds on: ADR-0003, ADR-0006, ADR-0011, ADR-0012, ADR-0024, ADR-0025, ADR-0026
- Invariants it touches: C04, C06, C12, C13

## Why

The lint's promise is C04: a diagram that breaks a readability rule is not emitted. Field
reports (Arffornia, agent-playbook, the skill's own minimal example) show nine places where
a diagram passed the lint and was still unreadable once draw.io drew it:

- cards, notes and the legend are never measured: a 420 px legend card truncated
  "object reference, not traffic" (#36), and a legend can list more rows than its card
  holds (#45);
- nothing checks the header: a zone at y 40 sits under a three-line meta (#44);
- text that fits by the lint's count still wraps in draw.io, mostly in sequence
  participants, because the lint leaves 0 to 9 px of slack and the export font is a few
  percent wider than `CharWidthEm` (#72);
- two port badges 30 px apart on one side overlap by 19 px with no finding (#73);
- an empty label with an annotation renders a leading ` · ⚠ G1`, which the lint then
  treats as a full label (#78);
- an annotation whose targets are not in the view disappears silently (#86);
- a long zone `sub` widens the title obstacle, so edges entering below it fail, and
  authors shorten real names to pass (#87, with #22 and B-20);
- a landscape `side:` column runs past the right edge of the canvas (#90).

Each one makes the lint trusted for something it does not check. The fix is always a rule
or a layout correction, never a relaxed rule (C04).

## User scenarios

**Agent rendering a platform diagram.** It writes the model and views, runs `make render`,
and gets every one of these problems as a finding naming the element and the measured
sizes, before draw.io export. It fixes the diagram, not the rule.

**Author of a sequence diagram.** Participant boxes sized by the resolver keep every
label on one line in the exported PNG, whatever the export machine's monospace fallback.

**Author of a landscape with a side column.** The side column is inside the canvas; the
bands make room for it.

## Requirements

- **FR-001** WHEN a line of a card, a note or a legend row is wider than the width the
  renderer gives it, THE SYSTEM SHALL report an error naming the card or note and the line,
  with the measured and available widths.
  *Scenario:* WHEN a topology has a 420 px legend card whose edge row reads "object
  reference, not traffic" THEN `make render` fails with a finding on that legend card
  naming that text.
- **FR-002** WHEN the title, the listed lines or the generated legend rows of a card need
  more height than the card has, THE SYSTEM SHALL report an error naming the card, the
  height needed and the height available.
  *Scenario:* WHEN a legend card 120 px high lists six node and edge kinds THEN the lint
  fails on that card with the needed height.
- **FR-003** WHEN a zone, node, card or note intersects the header block (title, subtitle
  and meta lines, measured as drawn), THE SYSTEM SHALL report an error naming the element
  and the header's bottom edge.
  *Scenario:* WHEN a system with a three-line meta has its top zone at y 40 THEN the lint
  fails naming that zone.
- **FR-004** THE SYSTEM SHALL accept a line of node text only when its measured width plus
  a safety margin of at least two characters, or 4 % of the available width if larger,
  fits the available width; and the sequence resolver SHALL size participant boxes with
  the same margin before rounding.
  *Scenario:* WHEN a sequence view has the two participants quoted in #72, `Velocity proxy
  [3.4.0-SNAPSHOT · LB Services]` and `Website + API [Laravel 12 · nginx + php-fpm]`, THEN
  their boxes leave that margin and the exported PNG shows each label on one line.
- **FR-005** WHEN two different port badges on the same node intersect once each is
  inflated by 2 px, THE SYSTEM SHALL report an error naming both edges; two connections
  to the same port SHALL still share one badge with no finding.
  *Scenario:* WHEN the report's repro puts edges to ports `p1` and `p2` at `b.top@100px`
  and `b.top@130px` THEN the lint fails naming both edges.
- **FR-006** WHEN an edge's label is empty and the edge carries annotation badges, THE
  SYSTEM SHALL render the badges alone, with no leading separator, and the lint SHALL
  measure exactly that text: `""` hides the text only (ADR-0026).
  *Scenario:* WHEN `labels: {c1: ""}` and an annotation G1 targets `c1` THEN the edge label
  is `!!⚠ G1!!` and renders as `⚠ G1`.
- **FR-007** IF an annotation targets only ids that the view does not draw, and the view
  does not list it in `hidden_annotations`, THEN THE SYSTEM SHALL fail the check with an
  error naming the annotation and its targets (ADR-0026).
  *Scenario:* WHEN annotation G1 targets connection `c9` and the view leaves `c9` out
  THEN `make render` fails naming G1 and `c9`; WHEN the view adds
  `hidden_annotations: {G1: "background job, drawn in jobs view"}` THEN it passes.
- **FR-008** WHERE a zone's label and sub do not fit on the title line within the zone's
  width, THE SYSTEM SHALL lay the sub out as clarification 3 decides, and the zone title
  obstacle SHALL cover exactly the text drawn.
  *Scenario:* WHEN a zone has a 90-character sub THEN a lane 20 px right of its title
  stays free for an edge, and the lint passes an edge entering there.
- **FR-009** WHEN a zone, node, card, note, tile or side column extends past the canvas,
  THE SYSTEM SHALL report an error naming it; and the landscape resolver SHALL size the
  canvas, or narrow the bands, so that the side column fits.
  *Scenario:* WHEN the #90 view (one band of five flow sections, two more bands, a side
  column "Upkeep") is resolved THEN its side column lies inside the canvas and the lint is
  green; WHEN a node is placed past the canvas width THEN the lint fails naming it.
- **FR-011** IF a view's `hidden_annotations` lists an id that is not an annotation of the
  model, an annotation that has a target drawn in the view, or an empty reason, THEN THE
  SYSTEM SHALL fail the check with an error naming the view and the id.
  *Scenario:* WHEN a view lists `hidden_annotations: {G1: ""}` THEN the check fails
  naming the view and G1; WHEN it lists G7, which the model does not define, THEN the
  check fails naming G7.
- **FR-012** WHERE a view sets `hide_annotations: true`, THE SYSTEM SHALL draw no annotation
  badge on its elements and edges, and the lint SHALL measure labels without them.
  *Scenario:* WHEN the view of `regression-78` sets `hide_annotations: true` THEN edge
  `c1` has no label and no badge.
- **FR-010** THE SYSTEM SHALL keep every example, regression fixture and docs diagram
  lint-green by fixing the diagrams that the new rules reject, never by relaxing a rule.
  *Scenario:* WHEN `make examples` runs on the branch THEN it exits 0, and the diff of
  `internal/lint` adds rules or tightens them, never loosens one.

## Properties

- For every resolved spec, every text the renderer draws in a card, note, legend or node
  has a lint measurement: no drawn text is unmeasured.
- For every pair of distinct port badges on a node, the lint reports them iff their
  2 px-inflated rectangles intersect.

## Out of scope

- Automatic edge routing and label placement (batch G, #77, #21).
- A new font pipeline beyond what clarification 1 decides.
- Structured gaps generating badges and a card (#29, B-14): FR-007 only reports.

## Clarifications

Answered by Brian on 2026-11-04.

| # | Question | Answer | ADR |
|---|---|---|---|
| 1 | How does the lint account for the export font (#72)? | (a) measure node text at 0.602 em, the widest monospace fallback, plus the FR-004 margin | ADR-0025 |
| 2 | What happens to an annotation with no drawn target (#86)? | Not a warning: warnings become noise. A lint error, unless the view acknowledges it with `hidden_annotations: {id: reason}`, a non-empty reason; acknowledging an id that is drawn or does not exist is itself an error | ADR-0026 |
| 3 | How is a long zone sub handled (#87, #22)? | (a) it wraps under the label in a column at most 240 px wide, and the obstacle covers exactly those lines. #22 is closed with this batch | ADR-0025 |
| 4 | Does `labels: {id: ""}` hide annotation badges (#78)? | (c) `""` hides the text only; a view-level `hide_annotations: true` hides the badges | ADR-0026 |
| 5 | Where does card and legend layout live (FR-001, FR-002)? | (a) a pure `house.CardLayout` shared by the renderer and the lint | ADR-0025 |

## Success criteria

Features `LNT-*` in `docs/features.json` pass their verify step, graded in `contract.md`.
