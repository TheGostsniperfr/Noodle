# Contract 008 · Lint robustness

What "done" means for each feature, agreed before implementation (ADR-0024).
One failed criterion fails the feature. Every "Verified by" runs inside the evaluator's
allowlist (`EVALUATOR` in the Makefile); a command outside it is refused, and fails.
In a table cell, `\|` is the Markdown escape of a pipe: run every command with a plain `|`.

- Agreed: pending `make review-contract SPEC=008`.
- Round 1 (2026-11-04): CHANGES, 23 points, all applied. Thresholds now assume the
  recommended answer (option a) of each clarification; see "Assumed answers" below.
- **Every `go test -run` row**: the `-v` output shows `--- PASS` for every test and
  subtest the row names. `no tests to run` fails the row. Every lint test asserts the
  exact finding list with `assert.Equal`.
- C04 holds throughout: no existing rule of `internal/lint` is removed, skipped or given a
  wider tolerance; LNT-08 #2 checks it.

## Assumed answers

The spec's clarifications are pending. These thresholds are written for option (a) of
each. If Brian picks another option, the affected rows are rewritten and re-reviewed
before the task that implements them starts.

| Clarification | Assumed | Rows that depend on it |
|---|---|---|
| 1 · export font | node text measured at `house.ExportCharWidthEm = 0.602` | LNT-03 #1 to #3 |
| 2 · undrawn annotation | a lint warning (never blocks) | LNT-05 #2, #3 |
| 3 · zone sub | wraps under the label in a column of at most 240 px | LNT-06 #1 to #3 |
| 4 · `""` and badges | `""` hides the text only, badges stay | LNT-05 #1, #5 |
| 5 · card layout | `house.CardLayout` shared by renderer and lint | LNT-01 #5, #6, #7 |

## Fixtures

| Path | Holds | Rendered by |
|---|---|---|
| `examples/regression-72/` | sequence view `participants` with the two participants quoted in #72 | `make examples` (added by T07); PNGs `participants.light.png`, `participants.dark.png` committed |
| `examples/regression-78/` | topology with `labels: {c1: ""}` and annotation G1 on `c1` | `make examples` (added by T05); PNGs `view.light.png`, `view.dark.png` |
| `examples/regression-87/` | topology whose zone `platform` has the sub of LNT-06 | `make examples` (added by T08); PNGs `view.light.png`, `view.dark.png` |
| `internal/resolve/testdata/landscape-90/` | one band of five flow sections, two bands of two and three sections, side column `Upkeep` with one item | `go test ./internal/resolve/` |
| `specs/008-lint-robustness/visual/` | `<fixture>-before.png` rendered from `main` for every fixture or example whose drawing changes | committed with the task that changes it |

The LNT-06 sub is the 90-character string `"abcdefghi "` repeated nine times with the final
space replaced by `.` and the last word `abcdefghi.`: nine words, greedy-wrapped at 40
characters into lines of 39, 39 and 10 characters.

## LNT-01 · Cards, notes and legend are measured

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | A legend row wider than its column fails | `go test ./internal/lint/ -run 'TestLint_CardText' -v` | subtest `legend_420px_ref_row`: a 420 px legend card using the `ref` edge kind yields exactly one finding, `Where` the card id, message containing `object reference, not traffic` and two widths in px |
| 2 | Legend rows taller than the card fail | same | subtest `legend_120px_six_kinds`: exactly one finding naming the card, the needed height and 120 |
| 3 | Listed lines taller than the card fail | same | subtest `listed_card_too_tall`: exactly one finding naming the card, the needed and the available height |
| 4 | Listed lines and notes wider than their space fail | same | subtests `listed_line_too_wide` and `note_too_wide`: exactly one finding each, naming the card or note and the line |
| 5 | Cards that fit pass | same | subtest `fitting_cards`: a fitting legend card, listed card and note yield 0 findings |
| 6 | Every drawn row is measured | `go test ./internal/lint/ -run 'TestLint_CardText_EveryLayoutRowIsMeasured' -v` | for each example spec's cards, shrinking the card so every `house.CardLayout` row overflows yields one finding per row |
| 7 | Renderer and lint share the layout | `grep -n 'CardLayout' cmd/noodle/render.go internal/lint/card_text.go` | both files call `house.CardLayout` |
| 8 | Moving the layout changed no drawing | `go test ./cmd/noodle/ -run 'TestRender_CardLayoutGolden' -v` | every example's cards and legend render byte-equal to golden `.drawio` fragments committed from `main` before T01 |

## LNT-02 · Header overlap

The header bottom is `house.HeaderBox(spec)`: title from y 30 to 70, then from y 74 one
line of 13 px for the subtitle and one of 11 px per meta line, each times
`LineHeightEm` (1.25). With a subtitle and three meta lines the bottom is y 131.5.

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | A zone under the header fails | `go test ./internal/lint/ -run 'TestLint_HeaderOverlap' -v` | subtest `zone_at_40`: with a subtitle and three meta lines, a top zone at y 40 yields exactly one finding naming the zone and `131.5` |
| 2 | The boundary is exact | same | subtests `zone_at_bottom` (top at 131.5) yields 0 findings, `zone_1px_above` (top at 130.5) yields exactly 1 |
| 3 | Nodes, cards and notes are checked | same | subtests `node`, `card`, `note`: exactly one finding each, naming the element |
| 4 | No header, no rule | same | subtest `no_header`: a spec with no title, subtitle or meta yields 0 findings for an element at y 0 |
| 5 | Header geometry is shared | `grep -n 'HeaderBox' cmd/noodle/render.go internal/lint/header_overlap.go` | both use `house.HeaderBox` |

## LNT-03 · Text margin

Margin = max(2 × fontSize × 0.602, 0.04 × available width); a line fits iff
`textWidth + margin <= available`, with `textWidth` measured at 0.602 em.

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The margin formula, exactly | `go test ./internal/house/ -run 'TestFitsWithMargin' -v` | subtests at 12 px: available 200, margin 14.448: text 185 fits, 186 does not; available 600, margin 24: text 576 fits, 577 does not |
| 2 | Node text uses it on both sides | `go test ./internal/lint/ -run 'TestLint_NodeText_Margin' -v` | subtest `just_over`: exactly one finding `title overflows` on the node; subtest `just_under`: 0 findings |
| 3 | Sequence boxes keep the margin | `go test ./internal/resolve/ -run 'TestBoxWidth_KeepsTheMargin' -v` | for `Velocity proxy [3.4.0-SNAPSHOT · LB Services]` and `Website + API [Laravel 12 · nginx + php-fpm]`, box width minus text width at 0.602 em is at least the margin |
| 4 | The fixture renders and keeps one line | `make examples`; `ls examples/regression-72/`; read `participants.light.png` and `participants.dark.png` | exit 0; both PNGs listed; in each image both labels sit on one line inside their box |

## LNT-04 · Port badge overlap

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The #73 repro fails with the exact finding | `go test ./internal/lint/ -run 'TestLint_PortsOverlap' -v` | subtest `issue_73`: findings equal `[{Where: e1, Msg: port badge "TCP 443" overlaps the badge of e2}]` |
| 2 | The 2 px inflation is exact | same | subtests `3px_apart`: exactly 1 finding; `4px_apart`: 0 findings |
| 3 | A shared port is one badge | same | subtest `shared_port`: two connections to the same port of one node yield 0 findings |
| 4 | Property: finding iff inflated rects intersect | `go test ./internal/lint/ -run 'TestLint_PortsOverlap_Property' -v` | a `testing/quick` test over random badge pairs on one side asserts a finding exactly when `a.Inflate(2).Intersects(b.Inflate(2))` |

## LNT-05 · Annotations

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | An empty label keeps the badges alone | `go test ./internal/resolve/ -run 'TestResolve_AnnotationLabel' -v` | subtest `empty_label`: label exactly `!!⚠ G1!!`; subtest `label`: exactly `<label> · !!⚠ G1!!` |
| 2 | An annotation with no drawn target warns | `go test ./internal/lint/ -run 'TestLint_UndrawnAnnotation' -v` | subtest `only_undrawn`: findings equal `[{Where: G1, Msg: targets nothing drawn in this view: c9, Warn: true}]` |
| 3 | One drawn target is enough | same | subtest `one_drawn_one_not`: 0 findings |
| 4 | The rendered label has no leading separator | `make examples`; `ls examples/regression-78/ specs/008-lint-robustness/visual/`; read `view.light.png` | exit 0; `78-before.png` and both after PNGs listed; the edge shows `⚠ G1` with no character before it |
| 5 | The lint measures the badge-only text | `go test ./internal/lint/ -run 'TestLint_AnnotationOnlyLabel' -v` | the edge's label box is the width of `⚠ G1`; a node placed 1 px into that box yields exactly one label-overlap finding |

## LNT-06 · Zone sub

For a zone at x 100, y 100, width 600, label `Platform` and the fixture sub: the label is
52.8 px wide, so the sub wraps in a 240 px column into three lines, the widest 234 px.

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The obstacle equals the drawn text | `go test ./internal/house/ -run 'TestZoneTitleBox_WrapsTheSub' -v` | `house.ZoneTitleBox` equals `Rect{X: 112, Y: 104, W: 234, H: 61.5}`; on `main` it is `Rect{X: 112, Y: 104, W: 618, H: 24}` |
| 2 | A lane right of the title is free | `go test ./internal/lint/ -run 'TestLint_ZoneSub' -v` | subtest `lane_free`: a vertical edge entering at x 366 yields 0 findings; subtest `through_title`: a vertical edge at x 150 yields exactly one `crosses title of zone platform` finding |
| 3 | The fixture renders as the obstacle says | `make examples`; `ls examples/regression-87/ specs/008-lint-robustness/visual/`; read `view.light.png` | exit 0; `87-before.png` and both after PNGs listed; the sub sits on three lines under the label, and nothing overlaps it |

## LNT-07 · Canvas bounds and landscape side column

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Each kind past the width fails | `go test ./internal/lint/ -run 'TestLint_CanvasBounds' -v` | subtests `zone`, `node`, `card`, `note`, `tile` (a node of a landscape spec), `side_column` (a zone of a landscape spec): exactly one finding each, naming the element |
| 2 | Past the height fails | same | subtest `past_height`: exactly one finding naming the node |
| 3 | Touching the edge passes | same | subtest `touching`: an element whose right edge equals the canvas width yields 0 findings |
| 4 | The #90 landscape fits | `go test ./internal/resolve/ -run 'TestLandscape_SideColumnFits' -v` | on `testdata/landscape-90/`, every side-column zone and tile lies inside the canvas, and `lint.Lint` returns 0 findings |
| 5 | The landscape example is unchanged, or shown | `git diff --exit-code origin/main -- examples/landscape`; else `ls specs/008-lint-robustness/visual/` and read `examples/landscape/stack.light.png` | exit 0; or `landscape-stack-before.png` is listed and the image shows the side column inside the canvas |

## LNT-08 · Examples stay green, rules are not relaxed

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Every example passes | `make examples` | exit 0 |
| 2 | No rule or test removed | `git diff origin/main -- internal/lint \| grep -cE '^-func (check\|Test)'`; `git diff origin/main -- internal/lint/lint.go` | `0`; the `rules` list loses no entry |
| 3 | No tolerance widened | `git diff origin/main -- internal/lint`, read | no existing comparison is changed to accept more (larger slack, smaller gap, skipped kind) |
| 4 | Every changed drawing is shown | `git diff --name-only origin/main -- examples docs/diagrams \| grep '\.png$'`; `ls specs/008-lint-robustness/visual/` | each changed PNG's fixture or example has a `<name>-before.png`; the evaluator reads each pair |
| 5 | The branch is green | `make check` | exit 0 |
