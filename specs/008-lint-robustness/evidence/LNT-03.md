# Evidence · LNT-03

- Feature: LNT-03
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: e29f1e4
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The margin formula, exactly | PASS | `go test ./internal/house/ -run 'TestFitsWithMargin' -v` shows `--- PASS` for all six subtests: `12px_in_200,_margin_14.448:_185_fits` and `186_does_not`, `12px_in_600,_margin_24:_576_fits` and `577_does_not`, `10px_in_200,_margin_12.04:_187_fits` and `188_does_not`. `internal/house/text.go` implements `max(2·fs·0.602, 0.04·availW)` and `textW+margin <= availW` |
| 2 | Node titles use it on both sides | PASS | `go test ./internal/lint/ -run 'TestLint_NodeText_Margin' -v` shows `--- PASS: .../just_under` and `--- PASS: .../just_over`. The test uses `assert.Equal` against `[{Where: a, Msg: "title overflows: 173px text + 14px margin in 182px"}]` on `clean()` node `a`, a plain box with W 200 |
| 3 | Node lines use it on both sides | PASS | Same run shows `--- PASS: .../line_just_under` (28 chars) and `--- PASS: .../line_just_over` (29 chars). The expected message is `"ttt…t" overflows: 175px text + 12px margin in 182px`, with the line written out in the test |
| 4 | Sequence boxes keep the margin | PASS | `go test ./internal/resolve/ -run 'TestBoxWidth_KeepsTheMargin' -v` shows `--- PASS` for `velocity`, `website`, `main_box_fails` and `the_resolved_frame_passes_the_lint`. The test loads `examples/regression-72` and checks title at 12 px and tech at 10 px against `W − 18`. `main_box_fails` gives false at 190 − 18 for 168.56 px of tech |
| 5 | The fixture keeps one line | PASS | `make examples` gives `exit 0`. `ls` lists `participants.drawio`, `participants.light.png`, `participants.dark.png` and `visual/regression-72-participants-before.png`. `TestExamples_DrawioCurrent/examples/regression-72/participants.drawio` gives `--- PASS`. I read both after PNGs: "Velocity proxy" / "3.4.0-SNAPSHOT · LB Services" and "Website + API" / "Laravel 12 · nginx + php-fpm" each sit on one line inside their box, with visible right-hand slack, in light and dark |

The fixture matches the contract. `examples/regression-72/model.yaml` has `velocity` with `title: "Velocity proxy"` and `tech: "3.4.0-SNAPSHOT · LB Services"`, and `website` with `title: "Website + API"` and `tech: "Laravel 12 · nginx + php-fpm"`.

**Would the tests catch a revert?**
- **Lint test:** with the old 0.6 em measure and no margin, the 24-character title is 172.8 px in 182 px, so there is no finding and `just_over` fails.
- **Resolver test:** with the old `boxWidth`, both boxes come out at W 190, so the `velocity` and `website` tech assertions fail.

Both tests would catch the removal.

## Verify step

    $ go test ./internal/lint/... ./internal/resolve/...
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint
    ok  	github.com/TheGostsniperfr/Noodle/internal/resolve
    $ make examples
    ... go run ./cmd/noodle render examples/regression-72 ...
    exit 0
    $ go test ./cmd/noodle/ -run 'TestExamples_DrawioCurrent' -v
    --- PASS: TestExamples_DrawioCurrent/examples/regression-72/participants.drawio
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle

## Findings

- `internal/resolve/boxwidth_test.go:31`: `main_box_fails` checks only `seq.Frame.Nodes[0].Tech`, but the contract says "for both". It is equivalent today because both tech lines are 28 characters (168.56 px). If one fixture label changes, the test would stop covering the other participant. Fix: loop over both nodes.
- `TestBoxWidth_KeepsTheMargin` never pins the resolved `W`, which is 200 for both participants. A later change that widens boxes far beyond the margin would still pass. This is not required by the contract and is only noted.
- `specs/008-lint-robustness/visual/regression-72-participants-before.png` shows both labels on one line already. On the export machine the 190 px boxes did not wrap, only the right-hand slack was near zero. So the fixture shows the margin was gained. It does not reproduce the #72 wrap itself.
- Scope: commit bb68704 also sizes landscape tiles with the margin (`internal/resolve/landscape.go`) and widens boxes in three example layouts. Both follow from the tighter `node_text` rule, which tiles also pass through. No rule was relaxed (C04), and the LNT-03 entry in `docs/features.json` is still `passes: false` (C12).
