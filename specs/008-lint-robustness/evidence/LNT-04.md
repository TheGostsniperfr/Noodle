# Evidence · LNT-04

- Feature: LNT-04
- Verdict: **NEEDS_WORK**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: e29f1e4
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The #73 repro fails with the exact finding | PASS | `go test ./internal/lint/ -run 'TestLint_PortsOverlap' -v` → `--- PASS: TestLint_PortsOverlap/issue_73`. The test decodes `testdata/ports-73.yaml` with `KnownFields(true)` and checks `assert.Equal(..., []lint.Finding{{Where: "e1", Msg: "port badge \"TCP 443\" overlaps the badge of e2"}}, findings)`. In the fixture the edges land on `b` (x 300) at x 400 and x 430, so at `b.top@100px` and `b.top@130px`, with ports `TCP 443` and `TCP 8443`. |
| 2 | The 2 px inflation is exact | PASS | Same run: `--- PASS: .../3px_apart` checks an exact one-finding list, and `--- PASS: .../4px_apart` checks `assert.Empty`. `twoBadges` puts the second badge at `400 + PortWidth + gap`, so the gap is measured between the facing edges, as the contract defines "apart". `Rect.Intersects` is strict, so at 4 px the inflated rects only touch. |
| 3 | A shared port is one badge | **FAIL** | The threshold is "two connections to the same port of one node yield **0 findings**". `--- PASS: .../shared_port` passes, but it asserts `[]lint.Finding{{Where: "e1", Msg: "runs on top of e2 (vertical, x≈400)"}}`, which is one finding. The test passes only because it expects a finding the threshold rules out. |
| 4 | Property: finding iff inflated rects intersect | PASS | `go test ./internal/lint/ -run 'TestLint_PortsOverlap_Property' -v` → `--- PASS: TestLint_PortsOverlap_Property (0.11s)`. This is `testing/quick` with 5000 cases and seed 73. Badges are placed on all four sides, so the cases include same-side and adjacent-side pairs, which covers the round-3 follow-up. The expected value is `badges["e1"].Rect.Inflate(2).Intersects(badges["e2"].Rect.Inflate(2))`. |

## Verify step

    $ go test ./internal/lint/ -run 'TestLint_PortsOverlap' -v
    --- PASS: TestLint_PortsOverlap (0.00s)
        --- PASS: TestLint_PortsOverlap/4px_apart (0.00s)
        --- PASS: TestLint_PortsOverlap/3px_apart (0.00s)
        --- PASS: TestLint_PortsOverlap/shared_port (0.00s)
        --- PASS: TestLint_PortsOverlap/issue_73 (0.00s)
    --- PASS: TestLint_PortsOverlap_Property (0.11s)
    PASS
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint	0.118s

    $ go test ./internal/lint/...
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint	(cached)

    $ gofmt -l internal/lint/ ; go vet ./internal/lint/
    (no files) vet-ok

The features.json verify step says the test "passes two connections sharing one port". No test shows a lint pass for a shared port. The only shared-port spec in the tests returns a finding.

## Findings

- **`internal/lint/ports_overlap_test.go:30-39, 74-81` (fails criterion 3):**
  - **The problem:** `twoBadges(t, 0, true)` gives `e2` the same path as `e1`, so the two connections lie on top of each other. That trips the existing collinear-edge rule, and the subtest then asserts that finding. The threshold of criterion 3 and the verify step both ask for 0 findings.
  - **The fix:** make the two connections come from different sources, or bend one of them, so that they share only the final point on `b.top`. Then assert `assert.Empty(t, findings)`.
  - **Would the test catch a reverted feature?** Yes. If the `seen[b.ID]` dedupe were removed, both badges would get the same rect, and a second finding would make the subtest fail. The problem is only the threshold.
- **`internal/house/geometry.go:53` with `internal/lint/ports.go:20` (gap in FR-005, breaks no criterion):**
  - **The problem:** a badge's ID is built from node, side and position only. It ignores the port text. Two connections can reach the same point with *different* ports, say `TCP 443` and `TCP 8443`. They then share one ID, the renderer (`cmd/noodle/render.go:515-519`) draws only the first badge, and the lint reports nothing. The second port silently vanishes from the drawing, which is the kind of unreported drop this spec exists to close.
  - **The property test hides it:** it skips that exact case (`if toA == toB { return true }`).
  - **The fix:** put `e.Port` in the badge ID, so the two ports become distinct badges and the overlap rule reports them. Or add a lint error for different ports on one point.
- **Tests that test nothing:** none found. `issue_73` and `3px_apart` would fail if the pairwise loop in `ports.go:35-41` were removed. `4px_apart` would fail with a wider inflation, and the property test would fail if the check were removed.
- **Scope:** commit `a64c2f2` touches only `internal/lint/ports.go`, the new test, the fixture, `tasks.md` and `PROGRESS.md`. No existing rule in `ports.go` was loosened (C04). The `features.json` entry for LNT-04 is new on this branch and still `passes: false` (C12).
