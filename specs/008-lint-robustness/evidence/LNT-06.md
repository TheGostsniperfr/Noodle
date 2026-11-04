# Evidence · LNT-06

- Feature: LNT-06
- Verdict: **NEEDS_WORK**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: e29f1e4
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The obstacle equals the drawn text | PASS | `go test ./internal/house/ -run 'TestZoneTitleBox_WrapsTheSub' -v -count=1` printed `--- PASS: TestZoneTitleBox_WrapsTheSub`. The test asserts `Rect{X: 112, Y: 104, W: 234, H: 61.5}` and the three lines 39/39/10. On `main` the box is 52.8+12+13.2+540 = 618 wide, so this test fails if the change is reverted. |
| 2 | A lane right of the title is free | PASS | `go test ./internal/lint/ -run 'TestLint_ZoneSub' -v -count=1` printed `--- PASS: TestLint_ZoneSub/lane_free` and `--- PASS: TestLint_ZoneSub/through_title`. Both use `assert.Equal` on the full finding list. `through_title` expects exactly `[{e9, segment 0 crosses title of zone platform}]`. On `main` the obstacle reaches x 730, so `lane_free` at x 366 fails if the change is reverted. |
| 3 | The fixture renders as the obstacle says | PASS | `make examples` gave exit 0. `ls` listed `view.drawio`, `view.light.png`, `view.dark.png` and `regression-87-view-before.png`. `TestExamples_DrawioCurrent/examples/regression-87/view.drawio` passed (byte-equal re-render). I read `view.light.png`: "Platform" on the first line, then the sub on three lines (4 words, 4 words, "abcdefghi."). The text ends about 234 px from its left edge. The HTTP edge comes in to the right of the text and crosses nothing. In the before image, the sub is on one line that runs past the zone border and the edge crosses it. |

## Verify step

    $ go test ./internal/lint/... ./internal/house/... -count=1
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint	0.127s
    ok  	github.com/TheGostsniperfr/Noodle/internal/house	0.005s

    $ make examples; echo $?
    0

## Findings

- **FR-008 fails for zones with an icon: the obstacle is narrower than the drawn text.** In `internal/house/geometry.go:201-209`, `ZoneTitleBox` takes `max(row.W, TextWidth(line))` with the box starting at `z.X + 12`. But `cmd/noodle/render.go:329-331,380` draws the whole title block, sub lines included, in one vertex with `spacingLeft = pad - 12`. With an icon that is `ZoneIconSize + 8` = 26 px. So for an icon zone whose sub wraps, the 234 px sub lines are drawn from X+26 to X+260, while the obstacle stops at X+234. The lint then misses an edge that crosses the last 26 px of the sub. That breaks "the zone title obstacle SHALL cover exactly the text drawn" and the feature's own description. The bug is the opposite of the old one: false negatives instead of false positives.
  - Fix: in the wrapped case, add the icon offset (`ZoneIconSize + 8` when `z.Icon != ""`) to each sub line's width before taking the max.
  - Add a test: `TestZoneTitleBox_WrapsTheSub` with `Icon` set, expecting `W: 260`.
  - Add a `TestLint_ZoneSub` subtest: an icon zone with an edge at X+250 must give exactly one finding.
- **No test covers the icon case or a zone narrower than the column.** All three contract criteria use a zone with no icon and width 600, so the bug above passes the contract. The contract should get an icon-zone row.
- Minor: `internal/house/geometry.go:178` wraps the sub at 240 px no matter how wide the zone is. A zone narrower than 264 px gets sub lines past its right border, and nothing reports it. Either cap the column at `z.W - 24`, or have the lint report it.
- Minor: the sub is wrapped and measured with `TextWidth` at `CharWidthEm` 0.6. LNT-03 settled the export font at 0.602 em, so a 39-character line is drawn about 234.8 px wide against a 234 px obstacle. This is under 1 px and the contract pins 234, so it breaks no criterion. It is still the same measuring gap LNT-03 closed for node text.
- Weak test: `TestZoneSubLines_WrapAtTheColumnWidth` (`internal/house/zone_title_test.go:52`) only checks that each line fits within 240 px. It would still pass if the sub were never wrapped under the label and stayed inline, because `ZoneSubLines` returns nil and the loop has nothing to check. It also passes if wrapping happens when it should not. It adds no protection on its own.
