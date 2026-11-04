# Evidence · LNT-07

- Feature: LNT-07
- Verdict: **NEEDS_WORK**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: e29f1e4
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Each kind past the width fails | PASS | `go test ./internal/lint/ -run 'TestLint_CanvasBounds' -v`: `--- PASS` for `zone`, `node`, `card`, `note`, `tile`, `side_column`. Each subtest asserts one exact finding with `assert.Equal`, for example `{Where: "side-0", Msg: "extends past the canvas: right edge at 1712, canvas is 1600 wide"}` |
| 2 | Past the height fails | PASS | same run: `--- PASS: TestLint_CanvasBounds/past_height`, exact finding `{Where: "n", Msg: "extends past the canvas: bottom edge at 910, canvas is 900 high"}` |
| 3 | Touching the edge passes | PASS | same run: `--- PASS: TestLint_CanvasBounds/touching`. The node at (1500, 840), 100×60, has its right edge at 1600 and its bottom at 900. The expected findings are `nil` |
| 4 | The #90 landscape fits | PASS (threshold) | `go test ./internal/resolve/ -run 'TestLandscape_SideColumnFits' -v`: `--- PASS`. The side zone and tile `backup` are inside a canvas 2400 wide, and `lint.Lint` returns nothing. This test also passes without the change: see Findings |
| 5 | The landscape example is unchanged, or shown | PASS | `git diff --stat origin/main -- examples/landscape`: `stack.drawio` was added and both PNGs changed, so the fallback applies. `specs/008-lint-robustness/visual/landscape-stack-before.png` is listed. `TestExamples_DrawioCurrent/examples/landscape/stack.drawio` reports `--- PASS`. In `stack.drawio`, `side-0` and `side-1` are at x 997 with width 243, so their right edge is 1240 on a 1280 canvas. I read `stack.light.png`: the Identity and Edge side zones and their Keycloak and Cloudflare tiles are fully inside the canvas, with margin on the right. The before image shows the same layout, only narrower tiles |

## Verify step

    $ go test ./internal/lint/... ./internal/resolve/...
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint
    ok  	github.com/TheGostsniperfr/Noodle/internal/resolve

    $ go test ./internal/lint/ -run 'TestLint_CanvasBounds' -v
    --- PASS: TestLint_CanvasBounds (0.00s)
        --- PASS: TestLint_CanvasBounds/zone (0.00s)
        --- PASS: TestLint_CanvasBounds/tile (0.00s)
        --- PASS: TestLint_CanvasBounds/past_height (0.00s)
        --- PASS: TestLint_CanvasBounds/side_column (0.00s)
        --- PASS: TestLint_CanvasBounds/note (0.00s)
        --- PASS: TestLint_CanvasBounds/card (0.00s)
        --- PASS: TestLint_CanvasBounds/node (0.00s)
        --- PASS: TestLint_CanvasBounds/touching (0.00s)
        --- PASS: TestLint_CanvasBounds/past_the_left_edge (0.00s)

    $ go test ./internal/resolve/ -run 'TestLandscape_SideColumnFits' -v
    --- PASS: TestLandscape_SideColumnFits (0.01s)

## Findings

- **The #90 test passes on `main` too, so it proves no fix.**
  - `internal/resolve/landscape_bounds_test.go:16`, `internal/resolve/testdata/landscape-90/`.
  - T06 (1e510bd) does not touch `internal/resolve/landscape.go`. The only change to that file on the branch is the export-width `tileWidth` from T07 (bb68704).
  - The side column placement (`sideX := width - landscapeMargin - sideW`, `landscape.go:53-62`) and `naturalWidth` are the same on `origin/main` and on c2cf233, the commit #90 was reported against. `git log c2cf233..origin/main -- internal/resolve/landscape.go` is empty.
  - The side column has always been placed inside `spec.Width` by construction. So this fixture never reproduced the clipping reported in #90, and `TestLandscape_SideColumnFits` would pass if T06 were reverted.
  - The second half of FR-009 is therefore graded by a test that does not exercise the failure, and #90's real cause is still unidentified. Candidates include tile text wider at export than `tileW` (which T07 may have fixed), an explicit `width:`, or the export crop.
  - Fix: reproduce #90 from the agent-playbook view (agent-playbook PR #5) on c2cf233 and find the cause. Then either make `landscape-90` fail on `origin/main` and pass on the branch, or record in an ADR or on issue #90 that T07 fixed it, with a fixture that fails before T07.
- **The deviation from the plan is unrecorded.**
  - `docs/PROGRESS.md:36` says the resolver is unchanged and refers to "see the report". No report exists: there is no comment on issue #90, and nothing in `specs/008-lint-robustness/` or `docs/adr/` explains it.
  - `plan.md` ("Canvas bounds") promised that `resolve/landscape.go` would compute the band width from the canvas minus the side column.
  - Fix: write the finding down (issue #90 comment or ADR) and add it to `tasks.md` as a Convergence task.
- **The description claims more than the rule checks.**
  - `internal/lint/canvas_bounds.go:10` checks zones, node footprints, cards and notes only.
  - Edge label boxes, port badges, flow arrows and the header are not checked. The feature description says "anything drawn past the canvas".
  - FR-009 lists only zones, nodes, cards, notes, tiles and side columns, so this breaks no criterion. Either narrow the wording or extend the rule.
- **Scope and constitution:** nothing found.
  - `lint.go` only adds `checkCanvasBounds` to `rules`.
  - `git diff origin/main...HEAD -- internal/lint | grep -cE '^-func (check|Test)'` gives `0`.
  - The LNT-07 entry in `docs/features.json` is unedited and `passes` is `false` (C12).
  - `TestLint_CanvasBounds` would fail on every subtest except `touching` if the rule were removed, so the lint half is properly tested.
