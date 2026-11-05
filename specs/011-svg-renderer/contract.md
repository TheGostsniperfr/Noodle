# Contract 011 · Own SVG and PNG renderer

What "done" means for each feature, agreed before implementation (ADR-0024).
Drafted by the implementer, reviewed with `make review-contract SPEC=011` until the
evaluator answers APPROVED. One failed criterion fails the feature. Every "Verified by"
runs inside the evaluator's allowlist (`EVALUATOR` in the Makefile). In a table cell,
`\|` is the Markdown escape of a pipe: run every command with a plain `|`.

- Agreed: 2026-11-05 after three rounds (cap reached) · implementer Opus 5.5 · evaluator Opus
- Round 1 (2026-11-05): CHANGES, 24 points, all applied: the font floor only when a
  `.drawio` is written (FR-004, ADR-0034), exact counts and fixtures, PNGs compared by
  pixel, a PATH without draw.io set by the test, the viewBox property, and a visual row
  for every feature that changes what noodle draws (C13).
- Round 2: CHANGES, 10 points, all applied: 11 topology views; `lint.Finding` gains
  `Rule` and `Rect` (P2-06 #0); exact banner and PNG marker; visual rows graded on
  noodle's images; the margin factors pinned; the CLI measures by output; ids checked
  against the committed `.drawio`; reference rule for every view type.
- Round 3, the last (PROCESS.md cap): CHANGES, 11 points, all applied as written below and
  re-checked by the evaluator at grading: P2-04 flips after T04 and `Finding.Rule` moves to
  T03; the 19 rule ids listed; test names compared with `git grep` on both refs; `Where`
  and `Msg` of existing tests kept; literal advances; absolute rectangles and edge points
  against the `.drawio`; embedded fonts parsed back; icon variants nested; `-draft` in the
  CLI doc under P2-06.
- **Every `go test -run` row**: the `-v` output shows `--- PASS` for every test and
  subtest the row names. `no tests to run` fails the row, and a test that enumerates
  zero inputs fails.
- **Counts** are those of the branch when the feature is graded. At the time of writing
  `examples/` holds 11 topology, 3 sequence, 1 landscape, 1 catalog and 1 access matrix
  views, and `assets/icons/` 45 SVG files.
- **PNG equality**: same dimensions, and no pixel differs by more than 2 in any channel.
  SVG equality is byte equality.
- **Visual rows** are graded by reading the PNG. Each check is yes or no; one "no"
  fails the row: every node title is legible and inside its box; every icon is drawn,
  with no empty or black box; nothing is cut at the canvas edge; every arrow points at
  the listening side; no tofu glyph.
- C04 holds throughout: no rule of `internal/lint` is removed, skipped or given a wider
  tolerance. P2-04 #5 checks it.

## Answers

| Clarification | Answer | Rows that depend on it |
|---|---|---|
| 1 · shared geometry | one scene, three writers (ADR-0033) | P2-07 all, P2-01 #3 |
| 2 · font | JetBrains Mono NL, embedded subsets, lint measures it; floor at `ExportCharWidthEm` when a `.drawio` is written (ADR-0034) | P2-01 #2, P2-04 all |
| 3 · PNG | in the binary, `golang.org/x/image` (ADR-0035) | P2-02 all |
| 4 · scope | every view type, three waves (ADR-0036) | P2-03 |
| 5 · reference | noodle's SVG and PNG; `-draft` (ADR-0036) | P2-03 #2, P2-06, P2-08 |

## Fixtures

| Path | Holds | Checked by |
|---|---|---|
| `internal/render/png/testdata/wide-71/` | one topology view whose resolved canvas is 3725 px wide, a node touching the right edge | P2-02 #1 |
| `cmd/noodle/testdata/draft/` | one topology view whose lint returns exactly one finding, rule `label-overlap` (from `label_overlaps.go`) on edge `e1` | P2-06 |
| `cmd/noodle/testdata/measure/` | one topology view whose node title fits its box at 0.600 em and not at 0.602 em | P2-04 #3b |
| `internal/render/png/testdata/filter.svg` | an icon with a `<filter>` element | P2-02 #3 |

## P2-01 · Topology views render to a standalone SVG with the same geometry as the draw.io output

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Every topology example renders to SVG that parses as XML | `go test ./cmd/noodle/ -run 'TestSVG_Topology' -v` | the count of `--- PASS: TestSVG_Topology/` lines equals the number of topology views in `examples/` × 2 themes (22 today) |
| 2 | No external reference, fonts embedded | same test | no `href`, `xlink:href` or `url(` naming a file or an `http` URL; one `@font-face` per weight the SVG's text uses (two on every example), each with a `data:font/ttf;base64,` source that decodes and parses with `sfnt`; every rune of the SVG's `<text>` has a glyph in the subset of its weight |
| 3 | Every element of the resolved diagram has an SVG element with its id, at its rectangle | `go test ./internal/render/svg/ -run 'TestSVG_ElementsMatchDrawio' -v` | every `mxCell id` of the committed `examples/cnp-runtime/runtime.drawio` except `0` and `1` has an SVG element; vertex rectangles equal the `mxGeometry` in absolute canvas coordinates (x and y of the parent chain added) within 0.5 px; every `mxPoint` of an edge's `Array as="points"` lies within the corner radius + 0.5 px of the SVG path |
| 4 | Diff and highlight states draw | `go test ./internal/render/svg/ -run 'TestSVG_States' -v` | PASS on `examples/topology-diff` views `proposal` (diff) and `focus` (highlight); the exact id lists drawn as added, removed and dimmed are asserted for each |
| 5 | Two renders are byte-identical | `go test ./internal/render/svg/ -run 'TestSVG_Deterministic' -v` | PASS on every topology view of every example, both themes |
| 6 | Every rectangle the lint measures lies inside the SVG's `viewBox` | `go test ./internal/render/svg/ -run 'TestSVG_InsideViewBox' -v` | PASS on every topology view, both themes |
| 7 | The tests exercise the writer | `git grep -n 'svg\.Write' -- 'cmd/noodle/*_test.go' 'internal/render/svg/*_test.go'` | a match in each file holding #1 to #6 |
| 8 | Visual, on noodle's images | `go test ./cmd/noodle/ -run 'TestExamples_ImagesCurrent/(cnp-runtime/runtime\|topology-diff/proposal)' -v`, then read `examples/cnp-runtime/runtime.light.png`, `runtime.dark.png` and `examples/topology-diff/proposal.light.png` | `--- PASS` for those three files under the PNG equality rule, so they are noodle's; every visual check yes. P2-01 therefore flips after T10 |

## P2-02 · PNG rasterized in the binary at any width

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The 3725 px view at `-scale 2` gives a whole PNG | `go test ./internal/render/png/ -run 'TestPNG_WideCanvas' -v` | the test asserts the scene width is 3725 and the PNG width exactly 7450; the right-most column holds at least one pixel that is not the background colour |
| 2 | The CLI defaults to `-scale 2` | `go test ./cmd/noodle/ -run 'TestRender_PNGDefaultScale' -v` | `render examples/cnp-runtime -view runtime -o x.png` with no `-scale` exits 0 and the PNG is exactly 2 × the scene width |
| 3 | Every curated icon renders, and an unknown SVG element fails | `go test ./internal/render/png/ -run 'TestIcons_AllRender' -v` | one subtest per file, named after it, with the light and dark backgrounds as nested subtests; the count of depth-one `--- PASS: TestIcons_AllRender/<name>` lines equals `ls assets/icons/*.svg \| wc -l` (45 today), and every nested subtest shows `--- PASS`; the `filter.svg` fixture fails with an error containing `filter` |
| 4 | No cgo and no external process | `go test ./internal/render/png/ -run 'TestPNG_NoCgoNoExec' -v` | PASS; the test lists the deps of `./cmd/noodle` and finds neither `runtime/cgo` nor an `os/exec` import in `internal/render/...` |
| 5 | Text is drawn with the embedded font | `go test ./internal/render/png/ -run 'TestPNG_TextWidth' -v` | the ink of a rendered 20-character title spans its measured width within one advance |
| 6 | Fast | `go test ./internal/render/png/ -run 'TestPNG_Fast' -v` | `examples/cnp-runtime` view `runtime` renders in under 1 s |
| 7 | Visual | as P2-01 #8 | as P2-01 #8; P2-02 flips after T10 |

## P2-03 · Every view type renders to SVG and PNG, and the examples' reference images are noodle's

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Every view type renders to SVG and PNG | `go test ./cmd/noodle/ -run 'TestSVG_AllViewTypes' -v` | subtest names carry the view type; at least one `--- PASS` per type of topology, sequence, landscape, catalog and matrix; the test fails when one type has zero views; each subtest parses the SVG as XML and asserts the rule of P2-01 #2, fonts included |
| 2 | Every example view commits `.svg`, `.light.png` and `.dark.png`, each equal to a fresh render | `go test ./cmd/noodle/ -run 'TestExamples_ImagesCurrent' -v` | PASS under the equality rules above; no view without its three files |
| 3 | No draw.io needed | `go test ./cmd/noodle/ -run 'TestExamples_NoDrawio' -v`; `grep -nE 'drawio\|draw\.io\|xvfb' Makefile scripts/*.sh` | the test sets `PATH` to an empty temp dir and renders every view; grep prints only the `OUT=<file.drawio>` usage lines of the `render` target (lines 66 and 67 today) |
| 4 | Deterministic and inside the canvas for every view | `go test ./internal/render/svg/ -run 'TestSVG_Deterministic\|TestSVG_InsideViewBox' -v` | PASS on every view of every example, both themes |
| 5 | Visual | `examples/cnp-runtime/runtime.light.png`, `examples/cnp-runtime/oidc-login.light.png`, `examples/landscape/stack.light.png`, `examples/catalog/catalog.light.png`, `examples/cnp-access/matrix.light.png` | every visual check yes |

## P2-04 · The lint measures text with the embedded font and reports runes it lacks

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Advances come from the font | `go test ./internal/font/ -run 'TestAdvance_JetBrainsMonoNL' -v` | unitsPerEm is `1000`, and the advances of `a`, `M` and `·` are exactly `0.6` em in Regular and in Bold |
| 2 | Without a `.drawio` output, width is the sum of advances, and the threshold is exact | `go test ./internal/house/ ./internal/lint/ -run 'TestTextWidth_FromFont\|TestNodeText_FontAdvance' -v` | with an injected advance table of 0.7 em, a node title one character too wide gives exactly one finding, `Rule: "node-text"`, `Where:` the node id, and the `Msg` the test states; one character shorter gives an empty list |
| 3 | With a `.drawio` output, no rune is measured narrower than `ExportCharWidthEm` | `go test ./internal/house/ -run 'TestTextWidth_Floor' -v` | PASS; the font's 0.6 em measured as 0.602 |
| 3b | The CLI measures by output | `go test ./cmd/noodle/ -run 'TestRender_MeasureByOutput' -v` | on `cmd/noodle/testdata/measure/`, `-o x.svg` exits 0; `-o x.drawio` exits 1 with the node-text finding on that node |
| 4 | Glyph coverage | `go test ./internal/lint/ -run 'TestGlyphCoverage' -v` | one passing subtest per rune of `·`, `⚠`, `→`, `−`, `•`, `●`, `◀`, `—` in Regular and in Bold (16); a label with U+4E2D gives exactly one finding, rule `glyph-coverage`, naming the element id and the rune |
| 5 | No lint rule removed or loosened | `git diff origin/main...HEAD -- internal/lint/ internal/house/`; `go test ./internal/lint/ -v` | `TextMargin`, `FitsWithMargin` and `MinAvailable` keep `ExportCharWidthEm`, `2` and `0.04` / `0.96` as their factors (each literal on a `-` line of `internal/house/text.go` is on a `+` line); no `const` of `internal/lint/` or `internal/house/` changes value; `git grep -h -oE 'func Test[A-Za-z0-9_]+' <ref> -- internal/lint/` on `HEAD` gives the `origin/main` set plus `TestGlyphCoverage`, `TestNodeText_FontAdvance` and `TestFinding_RuleAndRect`, each `--- PASS`; in `git diff origin/main...HEAD -- 'internal/lint/*_test.go'` every `Where:` and `Msg:` string on a `-` line is unchanged on a `+` line |
| 6 | The fonts ship with their licence, in the right cut | `ls assets/fonts/`; `grep -c 'SIL OPEN FONT LICENSE' assets/fonts/OFL.txt`; `go test ./internal/font/ -run 'TestFont_Version' -v` | exactly `JetBrainsMonoNL-Bold.ttf`, `JetBrainsMonoNL-Regular.ttf`, `OFL.txt`, `embed.go`; count at least 1; the name table says `JetBrains Mono NL`, version 2.304 |

## P2-05 · `-animate` moves traffic dashes from the opening side, with CSS only

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Exactly the traffic edges are animated | `go test ./internal/render/svg/ -run 'TestSVG_Animate' -v` | exact id lists asserted: every traffic edge of `cnp-runtime` view `runtime` has the class; its references and the denied `x-cross-tenant` do not; the removed edges of `topology-diff` view `proposal` do not |
| 2 | The dash moves from the opening side | same test | the first point of each animated path is on its source's border, and the keyframe goes from `0` to a negative `stroke-dashoffset` |
| 3 | CSS only | same test | no `<script>` in any SVG |
| 4 | Off by default | same test | without `-animate`, no animation class or keyframe |
| 5 | On the CLI and documented | `go test ./cmd/noodle/ -run 'TestRender_Animate\|TestCLIReference' -v` | `render … -animate -o x.svg` writes the class; `docs/guide/cli.md` lists `-animate` and `-scale` and is current |

## P2-06 · `-draft` draws a failing view with its findings and exits 1

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 0 | Every finding names its rule and, when it names an element, its rectangle | `go test ./internal/lint/ -run 'TestFinding_RuleAndRect' -v` | one subtest per rule id, each with a fixture that triggers it: `canvas-bounds`, `card-text`, `collinear`, `edge-crossings`, `edge-sides`, `header-overlap`, `ids`, `label-at`, `label-overlap`, `matrix-text`, `node-icon`, `node-in-zone`, `node-spacing`, `node-text`, `offering-text`, `ports`, `steps`, `zone-spacing`, `glyph-coverage` (19); each asserts `Rule` equals the id and, when `Where` names an element, a non-zero `Rect`; the count of `--- PASS: TestFinding_RuleAndRect/` lines is 19 |
| 1 | The fixture's lint returns exactly one finding | `go test ./cmd/noodle/ -run 'TestRender_Draft' -v` | `assert.Equal` on the finding list: one finding, rule `label-overlap`, on `e1` |
| 2 | `-draft -o x.svg` and `-draft -o x.png` write the image and exit 1 | same test | SVG: one marker element whose text is `label-overlap`, at `Finding.Rect`, and a text element equal to `DRAFT · 1 lint findings`. PNG: at least one pixel of the marker colour inside `Finding.Rect` × scale, and none outside the marker rectangle and the banner |
| 3 | `-draft -o x.drawio` writes nothing and exits 2 | same test | asserted |
| 4 | Without `-draft`, nothing is written and the exit is 1 (C04) | same test | asserted |
| 4b | `-draft` is documented | `go test ./cmd/noodle/ -run 'TestCLIReference' -v`; `grep -n -- '-draft' docs/guide/cli.md` | PASS; at least one line |
| 5 | C04 cites ADR-0036 | `grep -n 'ADR-0036' docs/CONSTITUTION.md` | one line, row C04 |

## P2-07 · The draw.io output is unchanged, written from the scene

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Every committed `.drawio` is unchanged and current | `git diff --stat origin/main...HEAD -- '*.drawio'`; `go test ./cmd/noodle/ -run 'TestExamples_DrawioCurrent' -v` | the diff prints nothing; PASS |
| 2 | The draw.io writer reads the scene, not `diagram.Spec` | `grep -rn 'diagram\.Spec' internal/render/drawio/*.go`; `grep -ln 'scene\.' internal/render/drawio/*.go` | no `diagram.Spec` outside `_test.go`; every non-test file writing an `mxCell` matches `scene.` |
| 3 | `cmd/noodle` no longer writes XML | `grep -ln 'mxCell' cmd/noodle/*.go` | only `_test.go` files |

## P2-08 · Docs and skill export without draw.io

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | No draw.io export command in the skill, the guide or the README | `grep -rniE '(drawio\|draw\.io).*(-x\|--export)\|xvfb' skills/ docs/guide/ README.md` | prints nothing |
| 2 | The skill renders PNG with noodle | `grep -nE 'noodle render .*-o [^ ]+\.png' skills/noodle-diagram/SKILL.md` | at least one line |
| 3 | README and tutorial show the export | `grep -nE 'noodle render .*-o [^ ]+\.(png\|svg)' README.md docs/guide/tutorial.md` | at least one line per file |
| 4 | `docs/diagrams/` images are noodle's | `go test ./cmd/noodle/ -run 'TestDocsDiagrams_ImagesCurrent' -v` | PASS under the equality rules above |
