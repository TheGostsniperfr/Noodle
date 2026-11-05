# Evidence · LNT-05

- Feature: LNT-05
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 4fb1b56
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | An empty label keeps the badges alone | PASS | I ran `go test ./internal/resolve/ -run 'TestResolve_AnnotationLabel' -v`. These show `--- PASS`: `TestResolve_AnnotationLabel/empty_label`, which asserts `!!⚠ G1!!`; `TestResolve_AnnotationLabel/label`, which asserts `HTTPS · !!⚠ G1!!`; and `TestResolve_AnnotationLabel_SequenceMessageWithAnEmptyText`. That last test cuts the exact prefix `[1] `, requires that it is present (`annotations_test.go:102-103`), and asserts `!!⚠ G1!!` after it. This matches the threshold as amended in 4fb1b56 (T15). |
| 2 | An undrawn annotation is an error | PASS | I ran `go test ./internal/model/ -run 'TestCheck_UndrawnAnnotation' -v`. It shows `--- PASS: .../undrawn`. The test runs `assert.Equal` against `[{views/v.yaml, G1, "annotation targets nothing drawn in this view (c9): draw a target or list it in hidden_annotations with a reason"}]` (`internal/model/annotations_test.go:44-52`). |
| 3 | Acknowledged, or one target drawn, passes | PASS | Same run: `--- PASS` for `acknowledged` and for `one_drawn_one_not`. Both expect `nil`. |
| 3b | "Drawn" agrees with the resolver | PASS | Same run: `--- PASS` for `include_glob` (`include: [grp/**]`, G1 on `kid`, expects `nil`). `--- PASS` for `sequence_view`, which expects exactly the `undrawn` finding with `File: views/v.yaml`. |
| 3c | `hide_annotations` does not skip the check | PASS | Same run: `--- PASS` for `hide_annotations_still_checks`, which expects exactly `[undrawn]`. `--- PASS` for `hide_annotations_drawn`, which expects `nil`. |
| 4 | A bad acknowledgement is an error | PASS | Same run: `--- PASS` for `empty_reason`, `unknown_id` and `drawn_id`. Each asserts one finding, and each message is exactly the one in the contract. `unknown_id` lists `{G1: why, G7: x}`, so the undrawn G1 adds no second finding. |
| 5 | `hide_annotations` drops the badges | PASS | I ran `go test ./internal/resolve/ -run 'TestResolve_HideAnnotations' -v`. It shows `--- PASS` for `edge_empty`, `edge_label`, `node_badge`, `sequence` and `sequence_participant_badge`. `edge_empty` gives an empty label and `LabelBox` returns `false`. `edge_label` gives `HTTPS`. `node_badge` gives `""` with the flag and `"G2"` without it. `sequence` gives `HTTPS` after the required `[1] ` prefix, which matches the amended threshold. |
| 6 | The rendered label has no leading separator | PASS | `make examples` printed `exit=0`, and `git status` was clean afterwards. `ls` lists `view.drawio`, `view.light.png`, `view.dark.png` and `regression-78-view-before.png`. `TestExamples_DrawioCurrent/examples/regression-78/view.drawio` shows `--- PASS`. I read `view.light.png`: the Client→API edge shows `⚠ G1` with nothing in front of it. The before image shows a stray `·` in front of `⚠ G1`. |
| 7 | The lint measures the badge-only text | PASS | I ran `go test ./internal/lint/ -run 'TestLint_AnnotationOnlyLabel' -v`. All three subtests show `--- PASS`. The first asserts `box.W == 32`, `oldBox.W == 50`, and the same centre. `one_px_in` gives exactly `[{e, "label overlaps node m"}]`. `one_px_outside` gives `nil`. |
| 8 | The view contract carries both fields | PASS | `jq -c '(.properties.hidden_annotations \| {type, additionalProperties}), .properties.hide_annotations.type' schemas/v1alpha1/view.schema.json` prints `{"type":"object","additionalProperties":{"type":"string","minLength":1}}` and then `"boolean"`. |

## Verify step

    $ go test -count=1 ./internal/resolve/... ./internal/model/... ./internal/lint/...
    ok  	github.com/TheGostsniperfr/Noodle/internal/resolve	0.048s
    ok  	github.com/TheGostsniperfr/Noodle/internal/model	0.160s
    ok  	github.com/TheGostsniperfr/Noodle/internal/lint	0.168s
    $ go test ./internal/resolve/ -run 'TestResolve_AnnotationLabel|TestResolve_HideAnnotations' -v
    --- PASS: TestResolve_AnnotationLabel_SequenceMessageWithAnEmptyText
    --- PASS: TestResolve_AnnotationLabel        (sequence_text, empty_label, sequence_label, label)
    --- PASS: TestResolve_HideAnnotations        (sequence, sequence_participant_badge, edge_empty, edge_label, node_badge)
    $ go test ./internal/model/ -run 'TestCheck_UndrawnAnnotation' -v
    --- PASS: TestCheck_UndrawnAnnotation        (13/13 subtests, incl. undrawn, acknowledged, empty_reason, unknown_id, drawn_id)
    $ make examples
    exit=0

The `features.json` verify step holds:
- An empty label with G1 resolves to the badge alone.
- An undrawn G1 fails the check, and so does each bad `hidden_annotations` entry.
- An acknowledged G1 passes.
- No badge is drawn under `hide_annotations`, on edges, nodes, sequence messages or participants.

## Findings

- `specs/008-lint-robustness/contract.md` LNT-05 #1 and #5 were amended after the contract was agreed (4fb1b56, T15). The change names the test that exists and states the sequence label after its `[1] ` step number.
  - The amendment is sound. ADR-0011 numbers every message, so the old "exactly `HTTPS`" could never hold. The test still requires the exact `[1] ` prefix, and FR-006 and FR-012 are not weakened.
  - It is recorded only in tasks.md T15. The contract header does not record it, and it was not re-reviewed.
  - This breaks no criterion. Fix: add one line to the contract header listing the post-agreement amendments: #8 jq precedence in 87ba0ac, and #1 and #5 in 4fb1b56.
- Tests would catch a revert:
  - `empty_label` and `..._SequenceMessageWithAnEmptyText` fail if `joinBadges` puts back the leading ` · ` (`internal/resolve/resolve.go:140-149`).
  - `undrawn`, `sequence_view` and `hide_annotations_still_checks` fail if `checkAnnotations` is removed (`internal/model/check.go:744-783`).
  - `empty_reason`, `unknown_id` and `drawn_id` fail without the acknowledgement checks.
  - `edge_label`, `node_badge`, `sequence` and `sequence_participant_badge` fail if `badges()` ignores `HideAnnotations`.
  - No test asserts nothing.
  - `TestLint_AnnotationOnlyLabel` builds its label by hand, so it would not catch a resolver revert. Row #1 covers that path.
- Scope and constitution:
  - Since the last LNT-05 evaluation, the only LNT-05 change is the contract amendment. The rest of 4fb1b56 is LNT-07.
  - The `docs/features.json` diff against `origin/main` only adds lines, and LNT-05 is still `passes: false` until this evidence lands (C12).
  - No lint rule is relaxed (C04).
- Minor, breaks no criterion: `model.System.Drawn` (`internal/model/check.go:705`) is a second definition of "drawn", separate from the resolver's edge selection. `include_glob`, `sequence_view` and `sequence_view_drawn_message` pin it only for those cases. A view that the resolver draws differently, such as references or routes filtered by kind, could diverge without a test noticing.
