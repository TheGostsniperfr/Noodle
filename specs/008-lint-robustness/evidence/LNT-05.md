J'ai relancé toutes les commandes et regardé les PNG. J'imprime maintenant le fichier de preuve (en anglais, conformément à C15).

# Evidence · LNT-05

- Feature: LNT-05
- Verdict: **NEEDS_WORK**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: b1f5466
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | An empty label keeps the badges alone | FAIL | `go test ./internal/resolve/ -run 'TestResolve_AnnotationLabel' -v`. Subtests `empty_label` (`!!⚠ G1!!`) and `label` (`HTTPS · !!⚠ G1!!`) show `--- PASS` and assert the threshold values. The row also names the subtest `sequence_empty_label`, and that subtest is not in the output. The case now lives in a separate top-level test, `TestResolve_AnnotationLabel_SequenceMessageWithAnEmptyText`. The contract's preamble says an output that lacks a subtest the row names fails the row. The threshold also says the sequence message "gives exactly `!!⚠ G1!!`". The resolved label is `[1] !!⚠ G1!!` (`internal/resolve/sequence.go:95`). The test strips `[1] ` before it asserts (`annotations_test.go:102`). The contract was never amended to describe that. |
| 2 | An undrawn annotation is an error | PASS | `go test ./internal/model/ -run 'TestCheck_UndrawnAnnotation' -v` shows `--- PASS: .../undrawn`. The test uses `assert.Equal` on `[{views/v.yaml, G1, "annotation targets nothing drawn in this view (c9): draw a target or list it in hidden_annotations with a reason"}]`. |
| 3 | Acknowledged, or one target drawn, passes | PASS | Same run: `--- PASS` for `acknowledged` and `one_drawn_one_not`. Both expect `nil`. |
| 3b | "Drawn" agrees with the resolver | PASS | Same run: `--- PASS` for `include_glob` (`include: [grp/**]`, G1 on `kid`, `nil`). `--- PASS` for `sequence_view`, which expects exactly the `undrawn` list with `File: views/v.yaml`. |
| 3c | `hide_annotations` does not skip the check | PASS | Same run: `--- PASS` for `hide_annotations_still_checks`, which expects exactly the `undrawn` list. `--- PASS` for `hide_annotations_drawn`, which expects `nil`. |
| 4 | A bad acknowledgement is an error | PASS | Same run: `--- PASS` for `empty_reason`, `unknown_id` and `drawn_id`. Each asserts one finding with the contract's message. `unknown_id` uses `{G1: why, G7: x}` so that the undrawn G1 adds no second finding. That is a sound reading of the threshold. |
| 5 | `hide_annotations` drops the badges | FAIL | `go test ./internal/resolve/ -run 'TestResolve_HideAnnotations' -v`. All five subtests show `--- PASS`. `edge_empty`, `edge_label` and `node_badge` match the threshold. The threshold for `sequence` is "gives exactly `HTTPS`". The resolved label is `[1] HTTPS`. The test asserts `HTTPS` only after removing the prefix (`annotations_test.go:176-178`). The contract still states the unprefixed value. |
| 6 | The rendered label has no leading separator | PASS | `make examples` exits with code 0, and the worktree is still clean afterwards. `ls` lists `view.drawio`, `view.light.png`, `view.dark.png` and `regression-78-view-before.png`. `TestExamples_DrawioCurrent/examples/regression-78/view.drawio` shows `--- PASS`. I read `view.light.png`: the Client→API edge shows `⚠ G1` with nothing before it. The before image shows `· ⚠ G1`. |
| 7 | The lint measures the badge-only text | PASS | `go test ./internal/lint/ -run 'TestLint_AnnotationOnlyLabel' -v`. Three subtests show `--- PASS`: the 32 px box against 50 px, `one_px_in` and `one_px_outside`. |
| 8 | The view contract carries both fields | PASS | `jq -c '(.properties.hidden_annotations \| {type, additionalProperties}), .properties.hide_annotations.type' schemas/v1alpha1/view.schema.json` prints `{"type":"object","additionalProperties":{"type":"string","minLength":1}}` and then `"boolean"`. |

## Verify step

    $ go test ./internal/resolve/... ./internal/model/...
    ok  	github.com/TheGostsniperfr/Noodle/internal/resolve
    ok  	github.com/TheGostsniperfr/Noodle/internal/model
    $ go test ./internal/resolve/ -run 'TestResolve_AnnotationLabel|TestResolve_HideAnnotations' -v
    --- PASS: TestResolve_AnnotationLabel_SequenceMessageWithAnEmptyText
    --- PASS: TestResolve_AnnotationLabel   (empty_label, label, sequence_label, sequence_text)
    --- PASS: TestResolve_HideAnnotations   (edge_empty, edge_label, node_badge, sequence, sequence_participant_badge)
    $ go test ./internal/model/ -run 'TestCheck_UndrawnAnnotation' -v
    --- PASS: TestCheck_UndrawnAnnotation   (13/13 subtests)
    $ make examples
    exit=0

The `features.json` verify step holds in substance:
- An empty label resolves to the badge alone.
- An undrawn G1 fails the check, and so does each bad `hidden_annotations` entry.
- An acknowledged G1 passes.
- No badge is drawn under `hide_annotations`.

The failures are in contract rows #1 and #5. Their tests and thresholds still disagree, as the previous evaluation reported.

## Findings

- `specs/008-lint-robustness/contract.md` LNT-05 #1 and `internal/resolve/annotations_test.go:136`. Row #1 names the subtest `sequence_empty_label`, but the sequence empty-text case is the top-level test `TestResolve_AnnotationLabel_SequenceMessageWithAnEmptyText`. Fix it one of two ways:
  - make the case a subtest `sequence_empty_label` of `TestResolve_AnnotationLabel`;
  - or amend row #1 to name the test that exists.
- `contract.md` LNT-05 #1 and #5, against `internal/resolve/sequence.go:95`. Both thresholds say a sequence message gives "exactly" `!!⚠ G1!!` or `HTTPS`, but every sequence label carries `[n] ` (ADR-0011). T12 chose to assert on the label after the step number, which is defensible, and the test checks that the prefix is there (`annotations_test.go:103`). The contract was not amended to match, although row #8 was amended in 87ba0ac. Fix: amend #1 and #5 to "the label after its `[1] ` step number is exactly …", and record the amendment in the contract header as #8 was.
- Tests would catch a revert:
  - `empty_label` and `TestResolve_AnnotationLabel_SequenceMessageWithAnEmptyText` fail if the leading ` · ` comes back.
  - `undrawn`, `sequence_view` and `hide_annotations_still_checks` fail without the undrawn check.
  - `empty_reason`, `unknown_id` and `drawn_id` fail without the acknowledgement checks.
  - `edge_empty`, `node_badge`, `sequence` and `sequence_participant_badge` fail if `badges()` ignores `HideAnnotations`.
  - No test asserts nothing.
- Scope and constitution:
  - Since the last evaluation (e29f1e4), the LNT-05 changes are the tests in `annotations_test.go` and the jq fix in `contract.md`. The other `internal/` changes belong to LNT-01, LNT-04 and LNT-06.
  - `docs/features.json` only gains entries; no entry is reworded (C12).
  - No lint rule is relaxed (C04).
- Minor, breaks no criterion: `model.System.Drawn` is still a second definition of "drawn", separate from the resolver's edge selection. `include_glob` and `sequence_view_drawn_message` pin it for those cases only.
