# Spec 011 · Own SVG and PNG renderer

- Status: **Draft**
- Mode: **full**
- Phase: 2, the current phase once 6a closes (#113, #114). Every export depends today on
  draw.io desktop, and three field reports hit that limit (C11).
- Issue: #82 · Backlog: B-09, B-23, B-29 (#71), B-40 (#91)
- Decisions it builds on: ADR-0006 (house style), ADR-0011, ADR-0021, ADR-0022
- Invariants it touches: C04, C05, C06, C13

## Why

noodle computes all of the geometry: boxes, zones, routed paths, ports, pills, cards.
It then writes it as draw.io XML, and every image is exported by draw.io desktop, an
Electron app run under xvfb. The field reports show what that costs:

- **Speed and setup.** 20 to 40 s per export, and 24 exports took about 12 minutes
  (Arffornia, #82). A Claude cloud session cannot install draw.io at all, so the first
  PNG took about 15 tool calls (GenArchi, #91).
- **Silent corruption.** Headless draw.io crops and stretches PNGs wider than about
  4000 px (#71).
- **Measured against a guess.** The lint measures text at `ExportCharWidthEm = 0.602`,
  the widest fallback font draw.io has been seen to use. Spec 008 spent most of its
  convergence on the gap between that estimate and what draw.io drew.
- **No draft.** `render` refuses to emit while the lint fails, so fixing 52 findings is
  done blind (#82).

## User scenarios

**An engineer renders a diagram in CI.** `noodle render arch/platform -view runtime -o
runtime.svg` and `-o runtime.png` write the images in well under a second, with no
draw.io, X server or Electron installed. The PNG of a 3700 px wide diagram is whole.

**An agent in a Claude cloud session.** It downloads the release binary, runs `noodle
render … -o x.png` and shows the image. No other tool is needed.

**A README author.** The SVG embedded with `![](runtime.svg)` animates its flows, dashes
moving from the side that opens the connection to the side that listens (C02).

**Someone fixing a diagram that fails the lint.** `noodle render … -draft -o x.png`
writes the image anyway, with each finding drawn as a red marker and its rule id, and
exits non-zero.

**Someone who edits by hand.** `-o x.drawio` writes the same draw.io file as today.

## Requirements

- **FR-001** WHEN `-o` ends in `.svg`, THE SYSTEM SHALL write a standalone SVG of the
  view from the same resolved geometry as the draw.io output, with icons and the font
  embedded, and no external reference.
  *Scenario:* WHEN `examples/cnp-runtime -view runtime -o x.svg` is rendered THEN `x.svg`
  parses as XML, holds no `href` to a file or URL, and every node, zone, edge and card
  of the resolved diagram has an element with its id.
- **FR-002** WHEN `-o` ends in `.png`, THE SYSTEM SHALL rasterize that SVG at the scale
  given by `-scale` (default 2), whatever the canvas width.
  *Scenario:* WHEN a 3725 px wide view is rendered at `-scale 2` THEN the PNG is
  7450 px wide plus the border, and its right-most column is drawn.
- **FR-003** THE SYSTEM SHALL draw every view type the draw.io writer draws: topology,
  sequence, landscape, catalog, access matrix, and the diff and highlight states.
  *Scenario:* WHEN every example of `make examples` is rendered to SVG THEN each one
  succeeds and matches its committed golden SVG.
- **FR-004** THE SYSTEM SHALL measure text in the lint with the metrics of the embedded
  font, the one the SVG and PNG draw with.
  *Scenario:* WHEN a node title is one character too wide for its box at the embedded
  font's advance THEN the lint reports it; one character shorter, it does not.
- **FR-005** WHERE `-animate` is set, THE SYSTEM SHALL animate the dashes of traffic
  connections in the SVG with CSS only, from the opening side to the listening side,
  and SHALL leave references, denied and removed connections still.
  *Scenario:* WHEN `examples/cnp-runtime` is rendered with `-animate` THEN each traffic
  edge has the animation class, and no reference edge has it.
- **FR-006** WHERE `-draft` is set, THE SYSTEM SHALL write the image even when the lint
  fails, draw each finding as a red marker with its rule id at the place it names, and
  exit with status 1.
  *Scenario:* WHEN a view with one overlapping label is rendered with `-draft -o x.svg`
  THEN `x.svg` is written, holds one marker with the rule id, and the exit status is 1.
- **FR-007** WHEN `-o` ends in `.drawio`, THE SYSTEM SHALL write the same file as before
  this spec.
  *Scenario:* WHEN every example is rendered to `.drawio` THEN each file is byte-equal
  to its committed `.drawio` (`TestExamples_DrawioCurrent`).
- **FR-008** THE SYSTEM SHALL document the export without draw.io in the skill, the
  README and the tutorial, and the skill SHALL stop calling draw.io for images.
  *Scenario:* WHEN `grep -rn 'drawio -x' skills/ docs/guide/ README.md` runs THEN it
  prints nothing.

## Properties

- For every example and theme, rendering to SVG twice gives byte-identical files.
- For every resolved diagram, every element the lint measures lies inside the SVG's
  `viewBox`.

## Out of scope

- The web viewer (phase 3) and HTML tours (phase 4).
- Automatic edge routing (#77) and auto-layout: the geometry stays what layouts say.
- Pixel parity with draw.io's export: the SVG is the house style's reference image from
  now on, and draw.io stays the editable output (ARCHITECTURE: "draw.io stays an export
  target, not the engine").

## Clarifications

Five questions for Brian, recommended option first.

| # | Question | Options | Answer | ADR |
|---|---|---|---|---|
| 1 | How do the draw.io and SVG writers share the geometry? | a) one list of drawing primitives (rect, text, path, image, pill) built from `house`, written by two backends, so they cannot drift; b) a separate SVG writer next to `render.go`, which mirrors it; c) SVG only, and the draw.io writer frozen | | |
| 2 | Which font does the image use, and how does it get into the file? | a) JetBrains Mono (OFL), the house font, embedded as a subset `@font-face` in each SVG (tens of KB); b) system fonts by name, as today, smaller files but different on every machine; c) text converted to paths, exact but not selectable and much larger | | |
| 3 | How is the PNG made? | a) in the binary: the primitives of 1a drawn straight to pixels with `golang.org/x/image` (vector rasterizer and OpenType, one new vendored module), so `noodle` alone suffices (#91); b) `resvg` as an external tool, called the way `kustomize` is (ADR-0027), sharper but one more install; c) both: Go by default, `resvg` when it is on the PATH | | |
| 4 | Which view types does the first release draw? | a) all of them (FR-003), in waves: topology, then sequence and landscape, then catalog and matrix; b) topology only, the rest on draw.io until a later spec | | |
| 5 | What becomes the default output and the reference for review (C13)? | a) SVG and PNG from noodle; examples commit `.svg` and `.png`, `.drawio` stays for editing; b) keep draw.io's PNG as the reference until the SVG has been compared on every example | | |

## Success criteria

Features `P2-*` in `docs/features.json` pass their verify step, graded in `contract.md`.
