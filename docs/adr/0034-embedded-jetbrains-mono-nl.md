---
status: accepted
date: 2026-11-05
deciders: [Brian Perret]
supersedes: []
extends: ["0006"]
specs: ["011"]
constitution: [C04]
---

# ADR-0034 · SVG and PNG draw with JetBrains Mono NL, embedded, and the lint measures with its metrics

## Context and problem

Question 2 of spec 011. The house font is JetBrains Mono, named in every draw.io style
with three fallbacks. What gets drawn depends on the machine that exports, so the lint
measures at `ExportCharWidthEm = 0.602`, the widest fallback seen, and adds a margin
(spec 008). An SVG that names a font has the same problem in every browser that opens
it; a PNG made in the binary needs the font file itself.

## Decision drivers

- What the lint measures is what is drawn, on every machine.
- No ligatures: code-like text such as `->` or `!=` stays literal (today a
  `font-variant-ligatures:none` span in every label).
- Every non-ASCII glyph noodle draws (`·`, `⚠`, `→`, `−`, `•`, `●`, `◀`, `—`) exists.
- Files stay small enough to commit and embed in a README.
- The licence allows embedding and redistribution.

## Considered options

1. **JetBrains Mono NL Regular and Bold, OFL 1.1, compiled into the binary and embedded
   as a subset `@font-face` in each SVG.** The NL cut has no ligatures at all.
2. **System fonts by name**, as the draw.io output does today.
3. **Text converted to paths.**

## Decision outcome

Chosen: **option 1**, because the font drawn is the font measured, on every machine,
and both cuts carry every glyph listed above (checked against 2.304).

- The two TTF files and `OFL.txt` live in `assets/fonts/` and are compiled in.
- A subset keeps the glyph ids and blanks the outlines of glyphs the view does not use,
  so `cmap` and `hmtx` stay untouched; each SVG embeds one subset per cut it uses, as
  base64 TTF data.
- The lint sums each rune's advance from the font's `hmtx`. When the render writes a
  `.drawio`, which draw.io draws with system fonts, no rune is measured narrower than
  `ExportCharWidthEm`, so no rule gets looser for that file (C04). Examples write both
  and keep the floor; an SVG or PNG alone is measured at the font it embeds.
- A rune the font lacks is a lint error: it would be tofu in the PNG.

### Consequences

- Good: the SVG and PNG look the same on every machine, and text overflow becomes a
  measurement instead of an estimate.
- Good: the ligature span disappears from SVG output.
- Bad: each SVG is tens of KB larger, and the binary grows by about 540 KB.
- Bad: a label in a script the font lacks, such as CJK, now fails the lint.

### Confirmation

`go test ./internal/font/...` checks that every glyph of the list above has an outline
in both cuts and that a subset keeps advances unchanged. The lint rule `glyph_coverage`
fails a label holding U+4E2D.

## Pros and cons of the options

### Option 1

- Good: exact measurement; one look everywhere.
- Bad: larger files and binary.

### Option 2

- Good: smallest files.
- Bad: different on every machine, and the lint keeps guessing.

### Option 3

- Good: exact everywhere, no font in the file.
- Bad: text is no longer selectable or searchable, and files grow several times over.
