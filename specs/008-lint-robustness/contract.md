# Contract 008 · Lint robustness

What "done" means for each feature, agreed before implementation (ADR-0024).
One failed criterion fails the feature. Every "Verified by" runs inside the evaluator's
allowlist (`EVALUATOR` in the Makefile); a command outside it is refused, and fails.
In a table cell, `\|` is the Markdown escape of a pipe: run every command with a plain `|`.

- Agreed: pending `make review-contract SPEC=008`.
- Round 1 (2026-11-04): CHANGES, 23 points, all applied. Thresholds now assume the
  recommended answer (option a) of each clarification; see "Assumed answers" below.
- Round 2: CHANGES, 12 points, all applied: the margin takes the font size, the sequence
  check subtracts the padding, frozen golden copies, committed `.drawio` checked current,
  one naming rule for before PNGs, exact inputs and messages throughout.
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

| Path | Holds | Checked by |
|---|---|---|
| `examples/regression-72/` | sequence view `participants` with the two participants quoted in #72 | linted by `make examples` (added by T07); `participants.drawio`, `participants.light.png`, `participants.dark.png` exported by the implementer and committed |
| `examples/regression-78/` | topology view `view` with `labels: {c1: ""}` and annotation G1 on `c1` | linted by `make examples` (added by T05); `view.drawio` and both PNGs committed |
| `examples/regression-87/` | topology view `view` whose zone `platform` has the sub of LNT-06 | linted by `make examples` (added by T08); `view.drawio` and both PNGs committed |
| `internal/lint/testdata/ports-73.yaml` | node `b` with ports `p1` (`TCP 443`) and `p2` (`TCP 8443`), edges `e1` at `b.top@100px` and `e2` at `b.top@130px` | `go test ./internal/lint/` |
| `internal/resolve/testdata/landscape-90/` | one band of five flow sections, two bands of two and three sections, side column `Upkeep` with one item | `go test ./internal/resolve/` |
| `cmd/noodle/testdata/cardlayout/` | frozen copies of every example spec from `main`, with golden `.drawio` card and legend fragments | `go test ./cmd/noodle/ -run TestRender_CardLayoutGolden`; added in T01's commit, before `render.go` changes |
| `specs/008-lint-robustness/visual/` | `<example-dir>-<view>-before.png`, rendered from `main`'s code, for every example or fixture whose drawing changes | committed with the task that changes the drawing |

The evaluator cannot render. A committed PNG is trusted as the export of its committed
`.drawio`, and `go test ./cmd/noodle/ -run 'TestExamples_DrawioCurrent' -v` re-renders every
fixture above and every example with a committed `.drawio`, byte-equal, so the `.drawio`
is shown to come from the branch's code.

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
| 6 | Every drawn row is measured | `go test ./internal/lint/ -run 'TestLint_CardText_EveryLayoutRowIsMeasured' -v` | for each card of each frozen example spec: shrinking only `W` to 1 px less than the widest `house.CardLayout` row gives exactly as many width findings as rows wider than the new space; shrinking only `H` to 1 px less than needed gives exactly one height finding |
| 7 | Renderer and lint share the layout | `grep -n 'CardLayout' cmd/noodle/render.go internal/lint/card_text.go` | both files call `house.CardLayout` |
| 8 | Moving the layout changed no drawing | `go test ./cmd/noodle/ -run 'TestRender_CardLayoutGolden' -v`; `git log --diff-filter=A --format=%h -- cmd/noodle/testdata/cardlayout/` | every frozen spec's cards and legend render byte-equal to their golden fragment; the testdata was added in a commit that does not touch `render.go` |

## LNT-02 · Header overlap

`house.HeaderBox(spec)` is `Rect{X: 40, Y: 30, W: Width − 376, H: bottom − 30}`: the dot at
x 40 to the right edge of the title and sub boxes, `Width − 336`. The bottom is y 74 plus
one 13 px subtitle line and, per meta entry, `ceil(TextWidth(line, 11) / (Width − 400))`
lines of 11 px, each times `LineHeightEm` (1.25). With `Width` 1600, a subtitle and three
one-line meta entries, `HeaderBox` is `Rect{40, 30, 1224, 101.5}`, bottom 131.5.

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | A zone under the header fails | `go test ./internal/lint/ -run 'TestLint_HeaderOverlap' -v` | subtest `zone_at_40`: exactly one finding naming the zone and `131.5` |
| 2 | The boundary is exact | same | subtests `zone_at_bottom` (top at 131.5): 0 findings; `zone_1px_above` (top at 130.5): exactly 1 |
| 3 | A wrapped meta line counts as two | same | subtest `meta_wraps`: a subtitle and one 190-character meta line (1254 px over 1200) give bottom 117.75; a zone at y 104, the one-line bottom, gets exactly one finding |
| 4 | Beside the header passes | same | subtest `zone_right_of_header`: a zone at y 40 and x 1265 gets 0 findings |
| 5 | Nodes, cards and notes are checked | same | subtests `node`, `card`, `note`: exactly one finding each, naming the element |
| 6 | No header, no rule | same | subtest `no_header`: a spec with no title, subtitle or meta yields 0 findings for an element at y 0 |
| 7 | Header geometry is shared | `grep -n 'HeaderBox' cmd/noodle/render.go internal/lint/header_overlap.go` | both use `house.HeaderBox` |

## LNT-03 · Text margin

`house.FitsWithMargin(textW, availW, fontSize)`: margin = max(2 × fontSize × 0.602,
0.04 × availW); true iff `textW + margin <= availW`, with `textW` measured at 0.602 em.
A plain node (no icon) has `availW = W − 18`, the padding `node_text.go` subtracts.

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The margin formula, exactly | `go test ./internal/house/ -run 'TestFitsWithMargin' -v` | at 12 px: available 200 (margin 14.448), 185 fits, 186 does not; available 600 (margin 24), 576 fits, 577 does not. At 10 px: available 200 (margin 12.04), 187 fits, 188 does not |
| 2 | Node titles use it on both sides | `go test ./internal/lint/ -run 'TestLint_NodeText_Margin' -v` | node `a`, plain box, `W` 200: subtest `just_under`, a 23-character title (166.2 px): 0 findings; `just_over`, a 24-character title (173.4 px): findings equal `[{Where: a, Msg: title overflows: 173px text + 14px margin in 182px}]` |
| 3 | Node lines use it on both sides | same | same node: `line_just_under`, a 28-character tech line (168.6 px): 0 findings; `line_just_over`, a 29-character line (174.6 px): findings equal `[{Where: a, Msg: "<the line>" overflows: 175px text + 12px margin in 182px}]` with the line written out in the test |
| 4 | Sequence boxes keep the margin | `go test ./internal/resolve/ -run 'TestBoxWidth_KeepsTheMargin' -v` | for `Velocity proxy [3.4.0-SNAPSHOT · LB Services]` and `Website + API [Laravel 12 · nginx + php-fpm]`, `FitsWithMargin(width at 0.602 em, W − 18, 12)` is true for the resolved `W`; subtest `main_box_fails`: `W` 350, `main`'s box for the Velocity label, gives false |
| 5 | The fixture keeps one line | `make examples`; `ls examples/regression-72/ specs/008-lint-robustness/visual/`; read `participants.light.png` and `participants.dark.png` | exit 0; `participants.drawio`, both PNGs and `regression-72-participants-before.png` listed; in each after image both labels sit on one line inside their box |

## LNT-04 · Port badge overlap

"Apart" is the gap between the two badges' facing edges, before inflation.

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The #73 repro fails with the exact finding | `go test ./internal/lint/ -run 'TestLint_PortsOverlap' -v` | subtest `issue_73` on `testdata/ports-73.yaml`: findings equal `[{Where: e1, Msg: port badge "TCP 443" overlaps the badge of e2}]` |
| 2 | The 2 px inflation is exact | same | subtests `3px_apart`: exactly 1 finding; `4px_apart`: 0 findings |
| 3 | A shared port is one badge | same | subtest `shared_port`: two connections to the same port of one node yield 0 findings |
| 4 | Property: finding iff inflated rects intersect | `go test ./internal/lint/ -run 'TestLint_PortsOverlap_Property' -v` | a `testing/quick` test over random badge pairs on one side asserts a finding exactly when `a.Inflate(2).Intersects(b.Inflate(2))` |

## LNT-05 · Annotations

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | An empty label keeps the badges alone | `go test ./internal/resolve/ -run 'TestResolve_AnnotationLabel' -v` | subtest `empty_label`: topology label exactly `!!⚠ G1!!`; `label`: `HTTPS` with G1 gives exactly `HTTPS · !!⚠ G1!!`; `sequence_empty_label`: a sequence message with `""` and G1 gives exactly `!!⚠ G1!!` |
| 2 | The resolver reports an undrawn annotation | `go test ./internal/resolve/ -run 'TestResolve_UndrawnAnnotation' -v` | a model with G1 on `c9` and a view that leaves `c9` out: resolving then linting gives exactly `[{Where: G1, Msg: targets nothing drawn in this view: c9, Warn: true}]` |
| 3 | The lint warns, never blocks | `go test ./internal/lint/ -run 'TestLint_UndrawnAnnotation' -v` | subtest `only_undrawn`: the same single warning; subtest `one_drawn_one_not`: 0 findings |
| 4 | The rendered label has no leading separator | `make examples`; `ls examples/regression-78/ specs/008-lint-robustness/visual/`; read `view.light.png` | exit 0; `view.drawio`, both PNGs and `regression-78-view-before.png` listed; the edge shows `⚠ G1` with no character before it |
| 5 | The lint measures the badge-only text | `go test ./internal/lint/ -run 'TestLint_AnnotationOnlyLabel' -v` | the edge's label box is the width of `⚠ G1`; a node placed 1 px into that box yields exactly one label-overlap finding |

## LNT-06 · Zone sub

For a zone at x 100, y 100, width 600, label `Platform` and the fixture sub: the label is
52.8 px wide, so the sub wraps in a 240 px column into three lines, the widest 234 px.

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The obstacle equals the drawn text | `go test ./internal/house/ -run 'TestZoneTitleBox_WrapsTheSub' -v` | `house.ZoneTitleBox` equals `Rect{X: 112, Y: 104, W: 234, H: 61.5}`; on `main` it is `Rect{X: 112, Y: 104, W: 618, H: 24}` |
| 2 | A lane right of the title is free | `go test ./internal/lint/ -run 'TestLint_ZoneSub' -v` | subtest `lane_free`: edge `e9` from (366, 40) to (366, 200) yields 0 findings; `through_title`: `e9` from (150, 40) to (150, 200) gives exactly `[{Where: e9, Msg: segment 0 crosses title of zone platform}]` |
| 3 | The fixture renders as the obstacle says | `make examples`; `ls examples/regression-87/ specs/008-lint-robustness/visual/`; read `view.light.png` | exit 0; `view.drawio`, both PNGs and `regression-87-view-before.png` listed; the sub sits on three lines under the label, and nothing overlaps it |

## LNT-07 · Canvas bounds and landscape side column

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Each kind past the width fails | `go test ./internal/lint/ -run 'TestLint_CanvasBounds' -v` | subtests `zone`, `node`, `card`, `note`, `tile` (a node of a landscape spec), `side_column` (a zone of a landscape spec): exactly one finding each, naming the element |
| 2 | Past the height fails | same | subtest `past_height`: exactly one finding naming the node |
| 3 | Touching the edge passes | same | subtest `touching`: an element whose right edge equals the canvas width yields 0 findings |
| 4 | The #90 landscape fits | `go test ./internal/resolve/ -run 'TestLandscape_SideColumnFits' -v` | on `testdata/landscape-90/`, every side-column zone and tile lies inside the canvas, and `lint.Lint` returns 0 findings |
| 5 | The landscape example is unchanged, or shown | `git diff --exit-code origin/main -- examples/landscape`; else `ls specs/008-lint-robustness/visual/` and read `examples/landscape/stack.light.png` | exit 0; or `landscape-stack-before.png` is listed and the image shows the side column inside the canvas, its `.drawio` current per `TestExamples_DrawioCurrent` |

## LNT-08 · Examples stay green, rules are not relaxed

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Every example passes | `make examples` | exit 0 |
| 2 | No rule or test removed | `git diff origin/main -- internal/lint \| grep -cE '^-func (check\|Test)'`; `git diff origin/main -- internal/lint/lint.go` | `0`; the `rules` list loses no entry |
| 3 | No tolerance widened | `git diff origin/main -- internal/lint`, read | no existing comparison is changed to accept more (larger slack, smaller gap, skipped kind) |
| 4 | Every changed drawing is shown | `git diff --name-only origin/main -- examples docs/diagrams \| cut -d/ -f1-3`; `go test ./cmd/noodle/ -run 'TestExamples_DrawioCurrent' -v`; `ls specs/008-lint-robustness/visual/` | each listed directory either has its `.drawio` equal to `main`'s, or a changed PNG and an `<example-dir>-<view>-before.png`; the evaluator reads each pair |
| 5 | The branch is green | `make check` | exit 0 |
