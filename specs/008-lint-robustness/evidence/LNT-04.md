# Evidence · LNT-04

- Feature: LNT-04
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: b1f5466
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The #73 repro fails with the exact finding | PASS | I ran `go test ./internal/lint/ -run 'TestLint_PortsOverlap' -v -count=1` and got `--- PASS: TestLint_PortsOverlap/issue_73`. The test decodes `testdata/ports-73.yaml` with `KnownFields(true)`. In that fixture, `e1` (`TCP 443`) lands at x 400 and `e2` (`TCP 8443`) at x 430 on `b` (x 300), which is `b.top@100px` and `@130px`. The test then checks `assert.Equal(t, []lint.Finding{{Where: "e1", Msg: `port badge "TCP 443" overlaps the badge of e2`}}, findings)`. |
| 2 | The 2 px inflation is exact | PASS | Same run: `--- PASS: .../3px_apart` asserts that exact one-finding list, and `--- PASS: .../4px_apart` asserts `assert.Empty`. `twoBadges` places the second badge at `400 + PortWidth(text) + gap`, so the gap is measured between the facing edges, as the contract defines "apart". |
| 3 | A shared port is one badge | PASS | Same run: `--- PASS: .../shared_port`. Edge `e1` runs from `a` and edge `e2` from `c`, and both carry `TCP 443`. They meet only on the final point `(400, 220)` of `b.top`, and the test asserts `assert.Empty(t, findings)`. This fixes the earlier NEEDS_WORK (evidence at e29f1e4), where this subtest expected a collinear finding. |
| 4 | Property: finding iff inflated rects intersect | PASS | I ran `go test ./internal/lint/ -run 'TestLint_PortsOverlap_Property' -v -count=1` and got `--- PASS: TestLint_PortsOverlap_Property (0.15s)`. It is a `testing/quick` test with 5000 cases and seed 73. Badges go on all four sides, so the cases include same-side and adjacent-side pairs. The expected value is `badges["e1"].Rect.Inflate(2).Intersects(badges["e2"].Rect.Inflate(2))`. |

## Verify step

    $ go test ./internal/lint/ -run 'TestLint_PortsOverlap' -v -count=1
    --- PASS: TestLint_PortsOverlap (0.00s)
        --- PASS: TestLint_PortsOverlap/shared_port (0.00s)
        --- PASS: TestLint_PortsOverlap/issue_73 (0.00s)
        --- PASS: TestLint_PortsOverlap/3px_apart (0.00s)
        --- PASS: TestLint_PortsOverlap/4px_apart (0.00s)
        --- PASS: TestLint_PortsOverlap/two_ports_on_one_point (0.00s)
    --- PASS: TestLint_PortsOverlap_Property (0.15s)
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint	0.159s

    $ go test ./internal/lint/... -count=1
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint	0.199s

    $ gofmt -l internal/lint ; go vet ./internal/lint/ && echo vet-ok
    vet-ok

    $ make examples ; echo exit=$?
    exit=0

The `features.json` verify step asks for two things, and both hold. The #73 repro fails, with a finding that names `e1` and `e2`. Two connections sharing one port pass with 0 findings.

## Findings

- **Do the tests catch a revert?** Yes:
  - `issue_73` and `3px_apart` would fail without the pairwise loop at `internal/lint/ports.go:43-49`.
  - `4px_apart` would fail if the inflation were wider than 2 px.
  - `shared_port` would fail without the `seen[b.ID]` dedupe, because both badges would get the same rect and the lint would report an overlap.
  - The property test would fail if the check were removed.
- **Behaviour beyond the contract (no criterion broken):**
  - `ports.go:24-28` adds an error when two different ports land on the same point: `port %q lands on the same point of %s as %q of %s`. The previous evaluation asked for it.
  - It stays within FR-005: the second port's badge was silently dropped before.
  - The `two_ports_on_one_point` subtest pins it with an exact `assert.Equal`.
  - The property test still skips `toA == toB` (`ports_overlap_test.go:131`). That case is now covered only by this one subtest. This is acceptable.
- **Scope and constitution:**
  - Since `origin/main`, the LNT-04 code changes touch only `internal/lint/ports.go`, `ports_overlap_test.go` and `testdata/ports-73.yaml`. Convergence commit `af2d0cc` also adds `tasks.md` (T10) and `PROGRESS.md`.
  - No existing rule is loosened (C04).
  - The LNT-04 entry in `features.json` is new on this branch and still `passes: false` (C12).
  - `gofmt` and `go vet` are clean (C15).
- **Visual check:** LNT-04 is not a visual feature, so there was no PNG to read. `make examples` exits 0, so no committed example trips the new same-point rule.
