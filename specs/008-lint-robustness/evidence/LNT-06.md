# Evidence · LNT-06

- Feature: LNT-06
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: b1f5466
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The obstacle equals the drawn text | PASS | I ran `go test ./internal/house/ -run 'TestZoneTitleBox_WrapsTheSub' -v -count=1` and it printed `--- PASS: TestZoneTitleBox_WrapsTheSub`. The test uses `assert.Equal` to check that the box is `Rect{X: 112, Y: 104, W: 234, H: 61.5}` and that the lines are 39, 39 and 10 characters long. On `main` the box is 618 wide, so this test would fail if the change were reverted. The new test `TestZoneTitleBox_WrappedSubStartsAfterTheIcon` also passes: with an icon the box is `W: 260`, which is 26 + 234. Without the T13 fix the width would be `max(90.8, 234) = 234`, so this test catches the earlier finding. |
| 2 | A lane right of the title is free | PASS | I ran `go test ./internal/lint/ -run 'TestLint_ZoneSub' -v -count=1`. It printed `--- PASS` for `lane_free` (x 366), `through_title` (x 150, exactly `[{e9, segment 0 crosses title of zone platform}]`), `icon_zone_lane_through_the_sub` (x 366 with an icon gives exactly that same one finding) and `icon_zone_lane_beside_the_sub` (x 386 gives 0). Every subtest checks the whole finding list with `assert.Equal`. On `main` the obstacle reaches x 730, so `lane_free` would fail if the change were reverted. |
| 3 | The fixture renders as the obstacle says | PASS | `make examples` exited 0, and `git status` was clean afterwards. `ls` listed `view.drawio`, `view.light.png`, `view.dark.png` and `visual/regression-87-view-before.png`. `TestExamples_DrawioCurrent/examples/regression-87/view.drawio` passed, so the re-render is byte-equal. I read `view.light.png`: "Platform" is on the first line, and the sub is on three lines under it (4 words, 4 words, then `abcdefghi.`). The HTTP edge comes in to the right of the text and nothing overlaps it. In the before image the sub is one line that runs past the right border of the zone, and the edge crosses it. |

## Verify step

    $ go test ./internal/lint/... ./internal/house/... -count=1
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint	0.270s
    ok  	github.com/TheGostsniperfr/Noodle/internal/house	0.010s

    $ make examples; echo exit=$?
    exit=0

    $ go test ./cmd/noodle/ -run 'TestExamples_DrawioCurrent' -v -count=1
    --- PASS: TestExamples_DrawioCurrent/examples/regression-87/view.drawio (0.00s)
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle	0.339s

    $ go test ./... -count=1    # only non-ok line:
    ?   	github.com/TheGostsniperfr/Noodle/assets	[no test files]

## Findings

None of these breaks a contract criterion. They are carried over from the previous evaluation, and T13 did not address them.

- **The previous blocking finding is fixed.** `internal/house/geometry.go:204-210` now adds `ZoneIconSize + 8` to each wrapped sub line. That matches `cmd/noodle/render.go:380`, where `spacingLeft = pad - 12` inside a vertex of width `title.W`. Before the fix, an icon zone's title vertex was only 208 px wide inside, so draw.io would also have re-wrapped the 234 px lines. The fix closes that too.
- **Minor: narrow zones are not handled.** At `internal/house/geometry.go:179` the sub always wraps at 240 px, whatever the zone's width. In a zone narrower than about 264 px (290 px with an icon), the sub lines run past the right border and nothing reports it. Two possible fixes: cap the column at `z.W - 24 - indent`, or add a lint finding for it.
- **Minor: the sub is measured at a slightly different width than it is drawn.** It is wrapped and measured with `TextWidth` at 0.6 em, but the export font is 0.602 em (LNT-03). A 39-character line is therefore drawn about 234.8 px wide, against an obstacle of 234 px. The gap is under 1 px.
- **Weak test:** `internal/house/zone_title_test.go:52`, `TestZoneSubLines_WrapAtTheColumnWidth`, only checks the upper bound of each line. It would still pass if the wrap were wrong in other ways. `TestZoneTitleBox_WrapsTheSub` pins the exact lines, so behaviour is still covered.
