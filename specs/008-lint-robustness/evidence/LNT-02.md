# Evidence · LNT-02

- Feature: LNT-02
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: e29f1e4
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | A zone under the header fails | PASS | `go test ./internal/lint/ -run 'TestLint_HeaderOverlap' -v -count=1` → `--- PASS: TestLint_HeaderOverlap/zone_at_40`. The test asserts `assert.Equal` against `[{Where: z, Msg: "overlaps the header, which ends at y 131.5"}]` |
| 2 | The boundary is exact | PASS | same run → `--- PASS: .../zone_at_bottom` (y 131.5, expected `nil`) and `--- PASS: .../zone_1px_above` (y 130.5, exactly one finding at 131.5). `Rect.Intersects` is strict (`diagram.go:16-18`), so touching the bottom edge does not count |
| 3 | A wrapped meta line counts as two | PASS | same run → `--- PASS: .../meta_wraps`. It uses a subtitle and `strings.Repeat("m", 190)`, with a zone at y 104. The expected finding is exactly one, naming `117.75` |
| 4 | Beside the header passes | PASS | same run → `--- PASS: .../zone_right_of_header` (x 1265, y 40, expected `nil`). The box right edge is 40 + 1224 = 1264 |
| 5 | Nodes, cards and notes are checked | PASS | same run → `--- PASS: .../node`, `.../card`, `.../note`. Each asserts one exact finding naming `n`, `c` or `n` with the message at `131.5`. This settles the round-3 follow-up asking for exact messages |
| 6 | No header, no rule | PASS | same run → `--- PASS: .../no_header` (no title, subtitle or meta, node at y 0, expected `nil`) |
| 7 | Header geometry is shared | PASS | `grep -n 'HeaderBox' cmd/noodle/render.go internal/lint/header_overlap.go` → `render.go:306: if house.HeaderBox(s) == (diagram.Rect{})` and `header_overlap.go:7: header := house.HeaderBox(l.s)`. `HeaderBox` is `Header(s).Box` (`house/header.go:32`), and the renderer draws the dot, title and sub cells from that same `house.Header` layout |

## Verify step

    $ go test ./internal/lint/ -run 'TestLint_HeaderOverlap' -v -count=1
    --- PASS: TestLint_HeaderOverlap (0.00s)
        --- PASS: TestLint_HeaderOverlap/zone_at_40 (0.00s)
        --- PASS: TestLint_HeaderOverlap/zone_right_of_header (0.00s)
        --- PASS: TestLint_HeaderOverlap/meta_wraps (0.00s)
        --- PASS: TestLint_HeaderOverlap/zone_1px_above (0.00s)
        --- PASS: TestLint_HeaderOverlap/zone_at_bottom (0.00s)
        --- PASS: TestLint_HeaderOverlap/card (0.00s)
        --- PASS: TestLint_HeaderOverlap/no_header (0.00s)
        --- PASS: TestLint_HeaderOverlap/node (0.00s)
        --- PASS: TestLint_HeaderOverlap/note (0.00s)
    PASS
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint	0.008s

    $ go test ./internal/lint/... -count=1
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint	0.139s

    $ set -o pipefail; make examples > /tmp/ex.log 2>&1; echo "exit=$?"
    exit=0          (no "header", "error" or "fail" line in the log)

    $ gofmt -l internal/lint internal/house cmd/noodle; go vet ./internal/lint/ ./internal/house/ ./cmd/noodle/
    (gofmt lists no files) vet-ok

## Findings

- **Tests would catch a revert.** If `checkHeaderOverlap` is removed from `rules` (`internal/lint/lint.go:41`), `zone_at_40`, `zone_1px_above`, `meta_wraps`, `node`, `card` and `note` all fail, because each expects one exact finding. If `Intersects` were made inclusive, `zone_at_bottom` would fail.
- **Scope is clean.** Commit 0b70d00 touches only `header_overlap.go` and its test, adds one line to the `rules` list in `lint.go`, and reorders the emptiness check in `render.go` so it uses `HeaderBox`. It removes or loosens no existing lint rule, so C04 holds, and it changes no `features.json` entry, so C12 holds.
- **Minor, breaks no criterion: the subtitle never wraps in the measurement.** `internal/house/header.go:48-50` always counts the subtitle as one 13 px line. The meta lines are wrapped by measured width, but the subtitle is not, so a subtitle longer than `Width − 400` still lets an element sit under its second line without a finding. The contract fixes the subtitle at one line, but FR-003 says "measured as drawn". Fix: measure the subtitle the way the meta loop does, with `ceil(TextWidth(PlainText(s.Subtitle), 13) / textW)` lines, and add a subtest for a wrapped subtitle.
- **Minor: comment style.** `header_overlap.go:5` has a comment that says what the function does, which goes against the house comment rule. It breaks no criterion.
