# Evidence · LNT-05

- Feature: LNT-05
- Verdict: **NEEDS_WORK**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: e29f1e4
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | An empty label keeps the badges alone | FAIL | `go test ./internal/resolve/ -run 'TestResolve_AnnotationLabel' -v`: all 5 subtests `--- PASS`. `empty_label` asserts `!!⚠ G1!!` and `label` asserts `HTTPS · !!⚠ G1!!`, both as the contract says. `sequence_empty_label` does not test what the threshold names. The threshold says "a sequence message with `""` and G1 gives exactly `!!⚠ G1!!`". The test uses step `over: c2`, which has no protocol and no `text`, targeted by G3. It asserts `[1] !!⚠ G3!!` (annotations_test.go:91). The asserted value is not the threshold's value, and the `""` input is not exercised. |
| 2 | An undrawn annotation is an error | PASS | `go test ./internal/model/ -run 'TestCheck_UndrawnAnnotation' -v`: `--- PASS: .../undrawn`, with an exact `assert.Equal` on `{views/v.yaml, G1, "annotation targets nothing drawn in this view (c9): draw a target or list it in hidden_annotations with a reason"}` |
| 3 | Acknowledged, or one target drawn, passes | PASS | same run: `--- PASS: .../acknowledged` and `--- PASS: .../one_drawn_one_not`, both expecting `nil` |
| 3b | "Drawn" agrees with the resolver | PASS | same run: `--- PASS: .../include_glob` (`include: [grp/**]`, G1 on `kid`, `nil`) and `--- PASS: .../sequence_view` (exact `undrawn` list, `File: views/v.yaml`) |
| 3c | `hide_annotations` does not skip the check | PASS | same run: `--- PASS: .../hide_annotations_still_checks` (exact `undrawn` list) and `--- PASS: .../hide_annotations_drawn` (`nil`) |
| 4 | A bad acknowledgement is an error | PASS | same run: `--- PASS` for `empty_reason`, `unknown_id` and `drawn_id`. Each gives exactly one finding, with the contract's message. `unknown_id` lists `{G1: why, G7: x}`, not `{G7: "x"}`. With the contract's input, the undrawn G1 would add a second finding, so the test's input is a sound reading. |
| 5 | `hide_annotations` drops the badges | FAIL | `go test ./internal/resolve/ -run 'TestResolve_HideAnnotations' -v`: all subtests `--- PASS`. `edge_empty`, `edge_label` and `node_badge` match the threshold. The `sequence` subtest asserts `[1] HTTPS` (annotations_test.go:144), but the threshold is "exactly `HTTPS`". |
| 6 | The rendered label has no leading separator | PASS | `make examples` → `exit 0`, worktree still clean. `ls` lists `view.drawio`, `view.light.png`, `view.dark.png` and `regression-78-view-before.png`. The drawio cell value is `⚠ G1` alone. `TestExamples_DrawioCurrent/examples/regression-78/view.drawio` passes. Read `view.light.png`: the edge Client→API shows `⚠ G1` with nothing before it. The before image shows `· ⚠ G1`. |
| 7 | The lint measures the badge-only text | PASS | `go test ./internal/lint/ -run 'TestLint_AnnotationOnlyLabel' -v`: 3 subtests `--- PASS`. The box is 32 px wide against 50 px, with the same centre. `one_px_in` gives `[{e, label overlaps node m}]` and `one_px_outside` gives `nil`. |
| 8 | The view contract carries both fields | FAIL | The command as written in the contract, `jq -c '.properties.hidden_annotations \| {type, additionalProperties}, .properties.hide_annotations.type' ...`, prints `{"type":"object","additionalProperties":{"type":"string","minLength":1}}` and then **`null`**, not `"boolean"`. The jq pipe binds looser than the comma, so the second path is evaluated on `hidden_annotations`. The schema itself is right: with parentheses the command prints `"boolean"`. The threshold is still not met by the command the contract specifies. |

## Verify step

    $ go test ./internal/resolve/... ./internal/model/...
    ok  	github.com/TheGostsniperfr/Noodle/internal/resolve
    ok  	github.com/TheGostsniperfr/Noodle/internal/model
    $ go test ./internal/resolve/ -run 'TestResolve_AnnotationLabel|TestResolve_HideAnnotations' -v
    --- PASS: TestResolve_AnnotationLabel (5/5 subtests PASS)
    --- PASS: TestResolve_HideAnnotations (5/5 subtests PASS)
    $ go test ./internal/model/ -run 'TestCheck_UndrawnAnnotation' -v
    --- PASS: TestCheck_UndrawnAnnotation (13/13 subtests PASS)
    $ jq -c '.properties.hidden_annotations | {type, additionalProperties}, .properties.hide_annotations.type' schemas/v1alpha1/view.schema.json
    {"type":"object","additionalProperties":{"type":"string","minLength":1}}
    null

The verify step of `docs/features.json` holds in substance. An empty label resolves to the badge alone. An undrawn G1 fails the check, and so does each bad `hidden_annotations` entry. An acknowledged G1 passes. No badge is drawn under `hide_annotations`.

## Findings

- `specs/008-lint-robustness/contract.md` LNT-05 #1 and #5, against `internal/resolve/annotations_test.go:91` and `:144`. Sequence message labels always carry the step prefix `[n] ` (`internal/resolve/sequence.go:95`), so "exactly `!!⚠ G1!!`" and "exactly `HTTPS`" cannot hold for a sequence message. The tests assert `[1] !!⚠ G3!!` and `[1] HTTPS`. `sequence_empty_label` also uses G3 on a connection with no protocol, not "`""` and G1". A sequence step cannot carry `""` at all: an empty `text` falls back to verb·protocol. The fix is to amend the two thresholds to the prefixed values and to an input a sequence view can express, recorded with Brian per PROCESS.md. The other way is to have the test match the contract. Either way, the test and the contract must agree.
- `specs/008-lint-robustness/contract.md` LNT-05 #8: the "Verified by" command prints `null` for its second line because of jq precedence. Fix it to `jq -c '(.properties.hidden_annotations | {type, additionalProperties}), .properties.hide_annotations.type' ...`, which prints the expected two lines on this commit. The code needs no change.
- Tests would catch a revert. `empty_label` fails if the ` · ` join comes back. `undrawn` and `hide_annotations_still_checks` fail without `checkAnnotations`. `empty_reason`, `unknown_id` and `drawn_id` fail without the acknowledgement checks. `edge_empty`, `node_badge` and `sequence_participant_badge` fail if `badges()` ignores `HideAnnotations`. No empty test found.
- Scope and constitution: the diff of T05 (bdfcf86) stays in `internal/model`, `internal/resolve`, `internal/lint` tests, the view schema, the skill reference, the regression-78 fixture and its Makefile line. It also adds `hidden_annotations` to `examples/cnp-runtime/views/oidc-login.yaml`, which the new check requires, with a reason for each id. ADR-0026:53 records the schema fields as optional additive alpha fields (C06). No lint rule is relaxed and no `features.json` entry is edited (C04, C12).
- Minor, breaks no criterion: `System.Drawn` (`internal/model/check.go:705`) marks a topology edge as drawn when both ends are included. This is a second definition of "drawn" next to the resolver's own edge selection. It agrees for the cases tested (`include_glob`), but nothing pins the two together for background or route edges.
