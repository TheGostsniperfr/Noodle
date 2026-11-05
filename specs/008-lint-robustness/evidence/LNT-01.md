# Evidence · LNT-01

- Feature: LNT-01
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: b1f5466
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | A legend row wider than its column fails | PASS | `go test -count=1 ./internal/lint/ -run 'TestLint_CardText' -v` → `--- PASS: TestLint_CardText/legend_420px_ref_row`. The test asserts exactly `[{Where: legend, Msg: "object reference, not traffic" needs 183px, its cell is 138px wide}]` (card_text_test.go:55). 138 = 420/2 − 72, the edge cell `house.CardLayout` gives the renderer |
| 2 | Legend rows taller than the card fail | PASS | `--- PASS: TestLint_CardText/legend_120px_six_kinds`. The test asserts exactly `[{Where: legend, Msg: text needs 318px height, card is 120px}]`. 318 = 52 + 6×26 + 8 + 2×26 + 50, so six node rows are counted |
| 3 | Listed lines taller than the card fail | PASS | `--- PASS: TestLint_CardText/listed_card_too_tall`. The test asserts exactly `[{Where: gaps, Msg: text needs 111px height, card is 80px}]` |
| 4 | Listed lines and notes wider than their space fail | PASS | `--- PASS: …/listed_line_too_wide` asserts one finding naming `gaps`, the 44-char line, 290 px and 164 px. `--- PASS: …/note_too_wide` asserts one finding naming note `n`, the 40-char line, 240 px and 100 px |
| 5 | Cards that fit pass | PASS | `--- PASS: TestLint_CardText/fitting_cards`. A legend card, a listed card and a note give `nil` findings |
| 6 | Every drawn row is measured | PASS | `go test -count=1 ./internal/lint/ -run 'TestLint_CardText_EveryLayoutRowIsMeasured' -v` → `--- PASS: TestLint_CardText_EveryLayoutRowIsMeasured`. It has width and height subtests for every card of the 17 frozen specs. The 85 `--- PASS` lines of the full `TestLint_CardText` run also include `TestLint_CardText_EveryNoteIsMeasured`, which covers the round-3 follow-up on notes. `go test ./cmd/noodle/ -run 'TestRender_EveryDrawnLegendCellIsALayoutRow\|TestRender_EveryCardRowIsDrawn' -v` → both `--- PASS`, which covers the follow-up "each drawn card row is a layout row" |
| 7 | Renderer and lint share the layout | PASS | `grep -n 'CardLayout' cmd/noodle/render.go internal/lint/card_text.go` → `render.go:591: layout := house.CardLayout(r.spec, c)` and `card_text.go:10: layout := house.CardLayout(l.s, c)` |
| 8 | Moving the layout changed no drawing | PASS | `go test -count=1 ./cmd/noodle/ -run 'TestRender_CardLayoutGolden' -v` → 34 subtests (17 specs × dark/light) `--- PASS`, `ok`. `git log --diff-filter=A --format=%h -- cmd/noodle/testdata/cardlayout/` → `ef7a5c6`. `git show --stat ef7a5c6` does not touch `render.go`. ef7a5c6 is the only commit that touches the testdata, so the goldens were not rewritten after the refactor (2d5986f). The 17 frozen specs include all 14 views that `origin/main` has under `examples/` and `docs/diagrams/` |

## Verify step

    $ go test -count=1 ./internal/lint/...
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint	0.254s

    $ go test -count=1 ./internal/lint/ -run 'TestLint_CardText' -v
    --- PASS: TestLint_CardText (0.00s)
        --- PASS: TestLint_CardText/legend_420px_ref_row (0.00s)
        --- PASS: TestLint_CardText/listed_line_too_wide (0.00s)
        --- PASS: TestLint_CardText/legend_120px_six_kinds (0.00s)
        --- PASS: TestLint_CardText/note_too_wide (0.00s)
        --- PASS: TestLint_CardText/listed_card_too_tall (0.00s)
        --- PASS: TestLint_CardText/fitting_cards (0.00s)
    --- PASS: TestLint_CardText_EveryNoteIsMeasured (0.01s)
    --- PASS: TestLint_CardText_EveryLayoutRowIsMeasured (0.02s)
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint	0.094s

    $ make examples
    exit 0

## Findings

- Tests vs. a revert. `checkCardText` is a new rule, registered at `internal/lint/lint.go:40`. Every failing case in `TestLint_CardText` asserts a non-nil, exact finding list with `assert.Equal`. If the rule were removed, the first five subtests would fail, so these tests catch the removal of FR-001 and FR-002. `EveryLayoutRowIsMeasured` compares the lint against `house.CardLayout` row by row, so it catches a row kind the lint skips. `TestRender_EveryDrawnLegendCellIsALayoutRow` catches a cell the renderer draws outside the layout.
- `internal/lint/card_text_test.go:53`: the contract says "using the `ref` edge kind", but no `ref` kind exists. The test uses `link`, whose legend row is "object reference, not traffic" (`internal/house/theme.go:94`). The behaviour tested is the one the spec scenario describes. Only the contract wording is wrong. It breaks no criterion.
- `internal/lint/card_text_test.go:147`: the width probe of #6 also sets `H` to 1e6, while the contract says "shrinking only `W`". It only counts findings that contain "cell is", and no row width depends on `H` in `house.CardLayout`. The deviation is harmless, but a comment or the contract wording should say so.
- `internal/lint/card_text_test.go:112`: `minWidthOf` assumes every cell width is affine in the card width and extrapolates from two layouts. That holds for the current `CardLayout`. It would silently mis-measure if a cell width ever got clamped. It breaks no criterion.
