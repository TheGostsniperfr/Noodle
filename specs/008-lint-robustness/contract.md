# Contract 008 · Lint robustness

What "done" means for each feature, agreed before implementation (ADR-0024).
One failed criterion fails the feature. Every "Verified by" runs inside the evaluator's
allowlist (`EVALUATOR` in the Makefile); a command outside it is refused, and fails.
In a table cell, `\|` is the Markdown escape of a pipe: run every command with a plain `|`.

- Agreed: pending `make review-contract SPEC=008`.
- Every lint test below asserts the exact finding list (`assert.Equal`), so a rule that
  fires on the wrong element or with the wrong message fails.
- C04 holds throughout: no existing rule of `internal/lint` is removed, skipped or given a
  wider tolerance; LNT-08 #2 checks it.

## LNT-01 · Cards, notes and legend are measured

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | A legend row wider than its column fails | `go test ./internal/lint/ -run 'CardText' -v` | exit 0; the subtest with a 420 px legend card using the `ref` edge kind asserts exactly one finding on that card whose message contains `object reference, not traffic` and two widths in px |
| 2 | Card content taller than the card fails | same | the subtest with a 120 px legend card listing six kinds asserts exactly one finding naming the card, the needed and the available height |
| 3 | Listed lines and notes are measured too | same | one subtest each for a listed card line and a note text too wide, each asserting exactly one finding naming the card or note and the line |
| 4 | Cards that fit pass | same | a fitting legend card, listed card and note yield 0 findings |
| 5 | The lint measures what the renderer draws | `grep -n 'CardLayout' cmd/noodle/render.go internal/lint/card_text.go` | both files call `house.CardLayout`; neither computes legend row positions on its own |

## LNT-02 · Header overlap

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | An element under the header fails | `go test ./internal/lint/ -run 'HeaderOverlap' -v` | exit 0; a top zone at y 40 under a title, subtitle and three meta lines yields exactly one finding naming that zone and the header bottom |
| 2 | Nodes, cards and notes are checked as well | same | one subtest each, exactly one finding naming the element |
| 3 | Below the header passes, and no header means no rule | same | the same zone at the header bottom plus `MinZoneGap` yields 0 findings; a spec with no title, subtitle or meta yields 0 findings |
| 4 | Header geometry is shared | `grep -n 'HeaderBox' cmd/noodle/render.go internal/lint/header_overlap.go` | both use `house.HeaderBox` |

## LNT-03 · Text margin

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The margin is enforced on both sides of the threshold | `go test ./internal/lint/ -run 'NodeText' -v` | exit 0; a node title whose width plus margin exceeds the available width by 1 px yields exactly one finding; at 1 px under, 0 findings |
| 2 | The margin is two characters or 4 %, whichever is larger | `go test ./internal/house/ -run 'FitsWithMargin' -v` | exit 0; table cases on a narrow box (two characters wins) and a wide box (4 % wins) |
| 3 | Sequence boxes keep the margin | `go test ./internal/resolve/ -run 'BoxWidth' -v` | exit 0; for each of the four participant labels of #72, the resolved box width minus the text width is at least the margin |
| 4 | The exported participants keep one line | the committed PNG of the regression fixture holding the four labels, read as an image | each label is on one line inside its box |

## LNT-04 · Port badge overlap

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The #73 repro fails, naming both edges | `go test ./internal/lint/ -run 'Ports' -v` | exit 0; the fixture with edges at `b.top@100px` and `b.top@130px` yields exactly one finding whose message names both edge ids |
| 2 | Badges 2 px apart count as overlapping, 3 px apart do not | same | one subtest each: 1 finding and 0 findings |
| 3 | A shared port is one badge | same | two connections to the same port of the same node yield 0 findings |

## LNT-05 · Annotations

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | An empty label keeps the badges alone | `go test ./internal/resolve/ -run 'Annotation' -v` | exit 0; `labels: {c1: ""}` with G1 on `c1` resolves the label to exactly `!!⚠ G1!!`; a non-empty label still resolves to `<label> · !!⚠ G1!!` |
| 2 | An annotation with no drawn target is reported | same, or `go test ./internal/lint/ -run 'Annotation' -v` per clarification 2 | an annotation whose only target is outside the view is reported once, by id, with its targets, at the severity clarification 2 chose |
| 3 | An annotation with one drawn target is not reported | same | 0 reports |
| 4 | The rendered label has no leading separator | the committed before/after PNG of the fixture, read as an image | the edge shows `⚠ G1` with nothing before it |

## LNT-06 · Zone sub

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The obstacle equals the drawn text | `go test ./internal/house/ -run 'ZoneTitleBox' -v` | exit 0; for a zone with a 90-character sub, the box covers the label line and the sub as laid out by clarification 3, and no more |
| 2 | A lane right of the title is free | `go test ./internal/lint/ -run 'ZoneSub' -v` | exit 0; an edge entering the zone 20 px right of its title text yields 0 findings; an edge through the title text still yields exactly one |
| 3 | The renderer draws what the obstacle covers | the committed before/after PNG of the fixture, read as an image | the sub appears where the obstacle says, and nothing overlaps it |

## LNT-07 · Canvas bounds and landscape side column

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Anything past the canvas fails | `go test ./internal/lint/ -run 'CanvasBounds' -v` | exit 0; one subtest each for a zone, node, card and note extending past the width or height, each exactly one finding naming it |
| 2 | Inside the canvas passes | same | an element touching the canvas edge but not past it yields 0 findings |
| 3 | The #90 landscape fits | `go test ./internal/resolve/ -run 'Landscape' -v` | exit 0; for the #90 view, every side-column rect lies inside the canvas, and `Lint` on the result returns 0 findings |
| 4 | The landscape example is unchanged or shown | `make examples` and the PR's before/after PNG of `examples/landscape` | exit 0; any visual change is in the PR |

## LNT-08 · Examples stay green, rules are not relaxed

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Every example passes | `make examples` | exit 0 |
| 2 | No rule loosened | `git diff origin/main -- internal/lint` | no existing finding condition is removed or given a wider tolerance; only additions and tightenings |
| 3 | Changed diagrams are shown | `git diff --name-only origin/main -- examples docs/diagrams` and the PR description | every changed example has a before/after PNG pair in the PR |
| 4 | The branch is green | `make check` | exit 0 |
