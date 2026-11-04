# Evidence · LNT-01

- Feature: LNT-01
- Verdict: **NEEDS_WORK**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: e29f1e4
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | A legend row wider than its column fails | FAIL | `go test ./internal/lint/ -run 'TestLint_CardText' -v` → `--- PASS: TestLint_CardText/legend_420px_ref_row`. The threshold asks for **exactly one finding**. The subtest asserts three, all on `legend`: `" 1  request step ·  A  background step" needs 239px, its cell is 190px wide`, `"listening port on the server" needs 176px, its cell is 140px wide` and `"object reference, not traffic" needs 183px, its cell is 138px wide`. The edge kind is `link`, which `house.LegendEdges` maps to "object reference, not traffic", so the kind is acceptable. The count is not. |
| 2 | Legend rows taller than the card fail | PASS | `--- PASS: TestLint_CardText/legend_120px_six_kinds`. It asserts exactly `[{legend, "text needs 318px height, card is 120px"}]`. |
| 3 | Listed lines taller than the card fail | PASS | `--- PASS: TestLint_CardText/listed_card_too_tall`. It asserts exactly `[{gaps, "text needs 111px height, card is 80px"}]`. |
| 4 | Listed lines and notes wider than their space fail | PASS | `--- PASS: .../listed_line_too_wide` asserts one finding on `gaps` naming the 44-char line, 290px against 164px. `--- PASS: .../note_too_wide` asserts one finding on `n` naming the 40-char line, 240px against 100px. |
| 5 | Cards that fit pass | PASS | `--- PASS: TestLint_CardText/fitting_cards`. It has a legend card, a listed card and a note, and expects `nil`. |
| 6 | Every drawn row is measured | PASS | `go test ./internal/lint/ -run 'TestLint_CardText_EveryLayoutRowIsMeasured' -v` gives 69 `--- PASS` and no `FAIL`. Each card of each frozen spec has a `/width` and a `/height` subtest. See the findings about the dead `H` assignment and the notes follow-up. |
| 7 | Renderer and lint share the layout | PASS | `grep -n CardLayout`: `cmd/noodle/render.go:591: layout := house.CardLayout(r.spec, c)` and `internal/lint/card_text.go:10: layout := house.CardLayout(l.s, c)`. |
| 8 | Moving the layout changed no drawing | PASS | `TestRender_CardLayoutGolden` passes all 34 subtests (17 specs × light/dark). `git log --diff-filter=A -- cmd/noodle/testdata/cardlayout/` gives `ef7a5c6`. `git show --stat ef7a5c6 \| grep -c render.go` gives `0`. |

## Verify step

    $ go test -count=1 ./internal/lint/...
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint	0.150s
    $ go test ./internal/lint/ -run 'TestLint_CardText$' -v
    --- PASS: TestLint_CardText (0.00s)
        --- PASS: TestLint_CardText/legend_420px_ref_row (0.00s)
        --- PASS: TestLint_CardText/listed_card_too_tall (0.00s)
        --- PASS: TestLint_CardText/legend_120px_six_kinds (0.00s)
        --- PASS: TestLint_CardText/note_too_wide (0.00s)
        --- PASS: TestLint_CardText/listed_line_too_wide (0.00s)
        --- PASS: TestLint_CardText/fitting_cards (0.00s)
    $ go test ./cmd/noodle/ -run 'TestRender_EveryDrawnLegendCellIsALayoutRow|TestExamples_DrawioCurrent'
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle	0.516s
    $ make examples
    exit 0

The verify text of the feature is met in substance. The 420 px card fails on "object reference, not traffic", the 120 px six-kind card fails, and fitting cards pass. Contract criterion 1 is stricter than that and is not met.

Reverting `checkCardText` would break both tests:
- `TestLint_CardText` would return `nil` where it expects findings.
- `TestLint_CardText_EveryLayoutRowIsMeasured` would count 0 findings where it expects at least 1.

I read `examples/cnp-runtime/runtime.light.png`. The widened legend card shows each node row, edge row, step row and port row on one line, and "object reference, not traffic" is fully visible. The notes block at the bottom of the card is intact and nothing overlaps.

## Findings

- `internal/lint/card_text_test.go:52`, subtest `legend_420px_ref_row`: contract LNT-01 #1 requires exactly one finding. The test asserts three, adding the steps row and the port row. A 420 px legend overflows the fixed steps row (needs 239px, cell 190px) and the port row whatever its edge kinds, so no fixture can meet "exactly one" with the current layout. There are two fixes:
  - Narrow the overflow in the layout so that only the edge row overflows at 420 px.
  - Reopen the contract through `make review-contract` to say "exactly these three findings", and record it in `contract.md`.

  The implementer must not silently keep a test that contradicts the agreed threshold.
- `internal/lint/card_text_test.go:152-153`: `narrow.H = 1e6` is set after `narrow` is copied into `probe.Cards`, so it has no effect. The width probe runs with the card's original height. The count still filters on `"cell is"`, so the result holds, but the line is dead and misleading. Set `H` before building `probe.Cards`.
- `internal/lint/card_text_test.go:133`: contract round 3 recorded a follow-up to cover "notes in LNT-01 #6", which the evaluator is to re-check. `TestLint_CardText_EveryLayoutRowIsMeasured` iterates `spec.Cards` only, and no note goes through the shrink probe. Add a note loop that shrinks `W` below `house.NoteRow(n).Widest()` and expects one finding.
- `cmd/noodle/golden_test.go`, `TestRender_EveryDrawnLegendCellIsALayoutRow`: the round-3 follow-up asked for "a test that each drawn card row is a layout row". The test checks only cells whose id starts with `legend-`. Card titles and listed lines are not checked against `CardLayout` rows. Extend it to every cell the card renderer emits.
- `internal/house/card.go:77`: for wrapping cells, `Widest` reports the widest word, not the line. FR-001 asks the finding to name "the line". On a multi-word overflowing line, the message names a word. This is minor and breaks no criterion, because the contract fixtures use single-word lines. Consider naming the source line and the offending word together.
