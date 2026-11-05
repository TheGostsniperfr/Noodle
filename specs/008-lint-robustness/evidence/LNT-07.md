# Evidence · LNT-07

- Feature: LNT-07
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 4fb1b56
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Each kind past the width fails | PASS | I ran `go test ./internal/lint/ -run 'TestLint_CanvasBounds' -v`. These subtests report `--- PASS`: `zone`, `node`, `card`, `note`, `tile` (a node in a spec with `Type: landscape`) and `side_column` (zone `side-0` in a landscape spec). Each one checks with `assert.Equal` that there is exactly one finding and that it names the element. For example: `{Where: "side-0", Msg: "extends past the canvas: right edge at 1712, canvas is 1600 wide"}` |
| 2 | Past the height fails | PASS | Same run: `--- PASS: TestLint_CanvasBounds/past_height`. It expects exactly `{Where: "n", Msg: "extends past the canvas: bottom edge at 910, canvas is 900 high"}` |
| 3 | Touching the edge passes | PASS | Same run: `--- PASS: TestLint_CanvasBounds/touching`. The node sits at (1500, 840) and is 100×60, so its right edge is 1600 and its bottom edge is 900. The test expects `nil` |
| 4 | The #90 landscape fits | PASS | I ran `go test ./internal/resolve/ -run 'TestLandscape_SideColumnFits' -v`. Both `--- PASS: TestLandscape_SideColumnFits` (`testdata/landscape-90`, width 2400, the cap) and `--- PASS: TestLandscape_SideColumnFits_Issue90View` (`testdata/landscape-90-playbook`, view `layers`) pass. Each checks that the `side-*` zone and its single tile are inside the canvas and that `lint.Lint` returns nothing. `tasks.md` T14 records that #90 does not reproduce on c2cf233 |
| 5 | The landscape example is unchanged, or shown | PASS | `git diff --stat origin/main -- examples/landscape` shows that `stack.drawio` was added and both PNGs changed, so the fallback applies. `specs/008-lint-robustness/visual/` lists `landscape-stack-before.png`. `TestExamples_DrawioCurrent/examples/landscape/stack.drawio` reports `--- PASS`. I read `stack.light.png`. The Identity and Edge side zones and their Keycloak and Cloudflare tiles are fully inside the canvas, with about 40 px of margin on the right. In the before image the layout is the same and the tiles are a little narrower |

## Verify step

    $ go test ./internal/lint/... ./internal/resolve/...
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint
    ok  	github.com/TheGostsniperfr/Noodle/internal/resolve

    $ go test ./internal/lint/ -run 'TestLint_CanvasBounds' -v
    --- PASS: TestLint_CanvasBounds (0.00s)
        --- PASS: TestLint_CanvasBounds/zone (0.00s)
        --- PASS: TestLint_CanvasBounds/past_height (0.00s)
        --- PASS: TestLint_CanvasBounds/card (0.00s)
        --- PASS: TestLint_CanvasBounds/side_column (0.00s)
        --- PASS: TestLint_CanvasBounds/node (0.00s)
        --- PASS: TestLint_CanvasBounds/tile (0.00s)
        --- PASS: TestLint_CanvasBounds/past_the_left_edge (0.00s)
        --- PASS: TestLint_CanvasBounds/touching (0.00s)
        --- PASS: TestLint_CanvasBounds/note (0.00s)
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint	0.007s

    $ go test ./internal/resolve/ -run 'TestLandscape_SideColumnFits' -v
    --- PASS: TestLandscape_SideColumnFits (0.00s)
    --- PASS: TestLandscape_SideColumnFits_Issue90View (0.00s)
    ok  	github.com/TheGostsniperfr/Noodle/internal/resolve	0.009s

    $ make examples
    examples exit 0

    $ git diff origin/main...HEAD -- internal/lint | grep -cE '^-func (check|Test)'
    0

## Findings

- **The resolver half of FR-009 is guarded, not shown to be fixed.**
  - Files: `internal/resolve/landscape_bounds_test.go:16` and `:47`.
  - Neither test would fail if this branch were reverted. The resolver has always placed the side column inside `spec.Width`.
  - The previous evaluation raised this. T14 now records why it is so: the real agent-playbook view of #90 does not clip on c2cf233 in the drawio, PNG or SVG export.
  - `plan.md` and `PROGRESS.md` now state this too, so the earlier "deviation unrecorded" finding is resolved.
  - The criterion is met as written. The cause of the clipping reported in #90 is still unknown, though.
  - Suggested fix: comment on issue #90 with the T14 finding before closing it, so the reporter can say whether the clipping was in their export pipeline.
- **The implementer amended criterion 4 after the evaluation, without a review round.**
  - Commit 4fb1b56 rewrote the threshold of LNT-07 #4 in `contract.md` after the NEEDS_WORK evidence. It added the playbook fixture and the T14 record.
  - The change makes the criterion stricter, not weaker, and it answers the earlier finding. So this breaks no criterion.
  - Still, PROCESS.md has no step for an implementer editing an agreed contract. A future change of this kind should go through `make review-contract` or be noted in the contract's header, as the round notes are.
- **The description still claims more than the rule checks.**
  - The description in `docs/features.json` says "anything drawn past the canvas".
  - `internal/lint/canvas_bounds.go:10` checks only zones, node footprints, cards and notes. Edge label boxes, port badges, flow arrows and the header are not checked.
  - FR-009 lists only the kinds that are checked, so no criterion is broken.
  - Suggested fix: extend the rule to the other drawn boxes, or narrow the wording in a new spec. The entry itself must not be edited (C12).
- **Scope and constitution checks found nothing.**
  - `lint.go` only adds `checkCanvasBounds` (plus the sibling rules of LNT-01 and LNT-02) to `rules`.
  - No `check` or `Test` function was removed.
  - The LNT-07 entry was added in d22e14b and has not been edited since. `passes` is still `false` (C12).
  - If the rule were removed, every `TestLint_CanvasBounds` subtest except `touching` would fail. So the lint half is really tested.
