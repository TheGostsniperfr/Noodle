# Evidence · LNT-08

- Feature: LNT-08
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: e29f1e4
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Every example passes | PASS | `make examples` rendered all 13 example targets plus `docs/diagrams/contract-pipeline`. It printed `exit 0`, and `git status` was clean afterwards. |
| 2 | No rule or test removed | PASS | `git diff origin/main -- internal/lint \| grep -cE '^-func (check\|Test)'` prints `0`. In `git diff origin/main -- internal/lint/lint.go` the `rules` list only gains `checkCardText`, `checkHeaderOverlap` and `checkCanvasBounds`, with no `-` line. |
| 3 | No tolerance widened | PASS | I read `git diff origin/main -- internal/lint`. **`node_text.go`:** `TextWidth` at 0.6 em with `w > avail` becomes `ExportTextWidth` at 0.602 em plus `FitsWithMargin`. That is stricter: the text is measured wider and a margin is added. **`ports.go`:** the shared-port `seen` skip and the icon check are unchanged. A pairwise badge check with `Inflate(2)` is added. **`lint_test.go`:** only the base fixture `clean()` changed (canvas height 480→560, legend card height 100→190) so it passes the new card-height and canvas rules. One expected message also changed, to `title overflows: 173px text + 14px margin in 132px`. No case was removed and no threshold relaxed. **Outside `internal/lint`:** `house.ZoneTitleBox` (the zone title obstacle the lint uses) narrows only when the sub wraps under the label, so it still covers exactly what is drawn (FR-008). |
| 4 | Every changed drawing is shown | PASS | `git diff --name-only … \| cut -d/ -f1-3` lists 16 diagram directories. Each one commits its `.drawio`. `go test ./cmd/noodle/ -run TestExamples_DrawioCurrent -v` passes all 17 subtests. The 10 directories whose PNG changed each have an `<example-dir>-<view>-before.png` in `visual/`. The remaining 5 (catalog, cnp-access, messaging, topology-diff, contract-pipeline) change only their `.drawio`. That file was added in 5807cee, and `git log origin/main..5807cee -- internal cmd` is empty, so no code had changed yet; the test shows it still re-renders byte-equal. I read every before/after pair; see Findings. |
| 5 | The branch is green | PASS | `make check` printed `exit 0`. All packages are `ok`, `features (base HEAD)` and `spec-check` are silent. |

## Verify step

    $ make examples; echo "exit $?"
    == examples
    go run ./cmd/noodle render examples/cnp-runtime --view runtime
    ...
    go run ./cmd/noodle render docs/diagrams/contract-pipeline -icons docs/diagrams/icons
    exit 0

    $ git diff origin/main -- internal/lint | grep -cE '^-func (check|Test)'
    0

    $ go test ./cmd/noodle/ -run 'TestExamples_DrawioCurrent' -v
    --- PASS: TestExamples_DrawioCurrent (0.39s)
        --- PASS: .../examples/catalog/catalog.drawio
        ... (17 subtests, all PASS)
        --- PASS: .../docs/diagrams/contract-pipeline/contract-pipeline.drawio
    PASS

    $ ls specs/008-lint-robustness/visual/
    app-runtime-view-app-runtime-view-before.png  cnp-runtime-oidc-login-before.png
    cnp-runtime-runtime-before.png  landscape-stack-before.png
    platform-overview-platform-staging-before.png  regression-72-participants-before.png
    regression-78-view-before.png  regression-87-view-before.png
    repo-map-gitlab-repos-before.png  sequence-basics-request-before.png

    $ make check; echo "exit $?"
    ok  github.com/TheGostsniperfr/Noodle/internal/lint   (cached)
    ...
    == features (base HEAD)
    == spec-check
    exit 0

What I saw in each before/after pair:

- **regression-72:** both participant boxes are wider. Each title and tech line stays on one line with visible right slack, in light and dark.
- **regression-78:** before, the label reads `· ⚠ G1`. After, it reads `⚠ G1` with nothing in front.
- **regression-87:** before, the 90-character sub runs past the zone's right border and the edge crosses it. After, it wraps into three lines under `Platform` and the edge passes clear to the right.
- **landscape:** the canvas is wider and the Identity and Edge side column sits inside it.
- **cnp-runtime runtime, app-runtime-view, platform-overview and repo-map:** the canvas is 40 px wider and the cards shift right. The legend cards are wider, so "object reference, not traffic" is no longer tight against the edge. A few boxes are 10 px wider (cf-edge, vso, cilium, prometheus, app-backend). The drawings read the same, with no new overlap.
- **sequence-basics and oidc-login:** participant boxes are slightly wider and the labels stay on one line.

## Findings

- **No dedicated test for FR-010.** LNT-08 is proven by `make examples`, `TestExamples_DrawioCurrent` and the reviewer reading the diff. No automated test flags a loosened comparison. Criterion 3 relies on a human read, which I did. A test pinning the existing rule thresholds would let a future relaxation fail on its own.
- **`TestExamples_DrawioCurrent` (`cmd/noodle/drawio_test.go:53`) checks only the dark-theme `.drawio`.** It does not check that the committed PNGs match it; criterion 4 accepts that. A light-only drawing change would not be caught. This is minor.
- **Exemption wording versus history.** The criterion 4 exemption speaks of "this branch's first commit", but the branch also carries the unmerged spec 007 commits (f8c49db onwards), so 5807cee is not literally the first commit. The intent holds: no commit before 5807cee touches `internal/` or `cmd/`. Rewording the contract to "before any change to `internal/` or `cmd/`" would remove the ambiguity.
- **Scope.** `git diff origin/main...HEAD` includes all of spec 007: `tools/speccheck`, the Makefile, `.claude/`, CI and the spec 007 specs. None of it belongs to LNT-08. Merging spec 007 first, or rebasing spec 008 onto it once merged, would keep this PR's diff to spec 008.
- **oidc-login view (`examples/cnp-runtime/views/oidc-login.yaml:11-13`).** G1 and G2 are acknowledged in `hidden_annotations` with reasons. ADR-0026 allows this, and it is not a relaxed rule. Brian should still read the reasons to confirm the annotations really belong only to the runtime view.
