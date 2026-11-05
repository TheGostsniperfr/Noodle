# Contract review · Spec 008

- Review: **CHANGES**
- Evaluator: evaluator (claude-opus-5-5)
- Date: 2026-11-04

## Changes

| Feature | Criterion | Problem | Proposed criterion |
|---|---|---|---|
| LNT-05 | #5 | FR-012 says `hide_annotations: true` draws no badge "on its elements and edges". Nodes carry badges too: `internal/resolve/resolve.go:84` sets `Badge: strings.Join(badges[e.ID], " ")`. Row #5 only checks edge `c1`. If node badges are still drawn under `hide_annotations`, the row still passes, so the element half of FR-012 is passed whatever the code does. | Add subtest `node_badge`: with `hide_annotations: true` and an annotation G2 targeting element `a`, the resolved node `a` has `Badge == ""`. Without the flag, it has `Badge == "G2"`. Add subtest `sequence`: a sequence message with label `HTTPS` and G1 resolves to exactly `HTTPS` under the flag. `internal/resolve/sequence.go:47-52` builds badges on its own path. |
| LNT-05 | #2, #3 | ADR-0026 says "Annotations with no drawn target still need `hidden_annotations`" even when `hide_annotations: true`. No row sets the flag together with an undrawn annotation. An implementation that skips the undrawn check whenever `hide_annotations` is set passes every row. | Add subtest `hide_annotations_still_checks` to `TestCheck_UndrawnAnnotation`: view `v` with `hide_annotations: true` and G1 targeting only `c9` gives findings equal to the `undrawn` list of #2. Add subtest `hide_annotations_drawn`: the same view with G1 also targeting `a` gives 0 findings. |
| LNT-05 | #2, #3, #4 | "Drawn in the view" is only tested on a topology view that includes `a` and `b`. `model.Check` has to work out "drawn" without the resolver. Today `checkView` (`internal/model/check.go:216`) only validates `include` ids. Two cases are never exercised: a `/**` include and a sequence view, where a connection is drawn only if a message uses it. So the check can disagree with what `resolve` draws and still pass every row. | Add subtests `include_glob`: `include: [grp/**]` with G1 targeting a child of `grp` gives 0 findings. Add `sequence_view`: a sequence view whose messages do not use `c9`, with G1 on `c9`, gives the `undrawn` finding with `File: views/<seq>.yaml`. |
| LNT-05 | #8 | `grep -n 'hidden_annotations\|hide_annotations'` matches a description string, or a field of any type. "An object of non-empty strings" is then judged by reading, which is not an exact threshold. Today `labels` is `{type: object, additionalProperties: {type: string}}` with no `minLength`, so the same shape without `minLength` would look right and still accept `""`. | Verified by `jq -c '.properties.hidden_annotations \| {type, additionalProperties}, .properties.hide_annotations.type' schemas/v1alpha1/view.schema.json`. Threshold: exactly the two lines `{"type":"object","additionalProperties":{"type":"string","minLength":1}}` and `"boolean"`. |

I checked these LNT-05 numbers against the code, and they hold:

- **#7 label box:** `PlainText` strips the `!!` markers (`internal/house/geometry.go:123,136`). So `⚠ G1` is 4 × 6 + 8 = 32 px wide, and ` · ⚠ G1` is 50 px. `one_px_outside` falls inside `main`'s box, so the row fails on `main`.
- **#2 render path:** `cmd/noodle/split.go:50` already stops on any `model.Check` finding before drawing. So "fails `make render`" follows from the `model.Check` rows.
- **#2 file name:** `views/v.yaml` matches `checkView`'s `filepath.Join("views", v.ID+".yaml")`.

Round 3 recorded follow-ups for LNT-01 #1 to #4, LNT-02 #5 and LNT-07 #1 and #2. They stay follow-ups under PROCESS.md's cap, and I do not raise them again. The four rows above concern LNT-05, which Brian's answers rewrote. Three of them (#2 to #5) can pass whatever the code does. The fourth (#8) cannot be graded exactly as written.
