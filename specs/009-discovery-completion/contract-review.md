# Contract review · Spec 009

- Review: **CHANGES**
- Evaluator: evaluator (claude-opus-5-5)
- Date: 2026-11-04

Round 3 was the last. Since then, commit `32aa8bf` recorded Brian's answers and rewrote these rows: DSC-03 #3, #3b and #5, DSC-04 #1, #3, #5 and #6, DSC-08 #4, DSC-10 #3 and #5. The PROCESS.md cap still applies to them, so only two kinds of point are marked **Apply**:

- a point that lets the evaluator write;
- a point that makes a criterion pass or fail whatever the code does.

All other points are **Follow-up**. Record them in the contract's follow-up table and start implementation.

What I checked:
- ADR-0027, ADR-0028 and ADR-0029 exist, and each has `status: accepted`.
- `make spec-check` exits 0.
- `go run ./tools/speccheck locate P6a-03` prints `specs/009-discovery-completion`.
- `origin/main` has no `DSC-*` entry, so the rewording of DSC-03 and DSC-04 in `features.json` changes only entries added on this branch. That does not break C12.
- `flake.nix` has no `kustomize` today.
- No ADR mentions `discovery.yaml` today, so DSC-10 #5 cannot be met before T13.

## Changes

| Feature | Criterion | Problem | Proposed criterion |
|---|---|---|---|
| DSC-04 | #1, subtest `legacy-fragment` | **Apply.** The subtest fails whatever the code does. The fixture `internal/model/testdata/minimal/discovered/k8s-manifests.yaml` is written by hand and double-quotes its ids, for example `id: "k8s:prod/namespace/prod"`. `model.EncodeFragment` writes every item in flow style. In the vendored yaml.v3, `vendor/gopkg.in/yaml.v3/emitterc.go`, a `:` inside a flow scalar sets `flow_indicators`, so `flow_plain_allowed` is false. Lines 1022-1026 then turn the plain style into `SINGLE_QUOTED`. The encoder therefore writes `id: 'k8s:prod/namespace/prod'`, and the "marshals it back byte-identical" check fails on line 5 even with a perfect ADR-0028 change. Making it pass would mean changing the encoder's quoting, which is out of scope and would change every fragment. | `legacy-fragment` loads the fixture, which must validate against `schemas/v1alpha1/fragment.schema.json`, then encodes it with `EncodeFragment` and loads the result again. It asserts with `Equal` that the reloaded `Fragment` equals the first load, and that no item of the encoded output has a top-level `object` key. A second check proves the new field adds no bytes when it is absent: `bytes.Equal(EncodeFragment(load(fixture)), EncodeFragment(load(EncodeFragment(load(fixture)))))`. |
| DSC-03 | #3b | **Follow-up.** `grep -c 'kustomize' flake.nix` ≥ `1` is also met by a comment or by another output. The rule "inside the `devShells` package list" cannot be read from a count. | `grep -cE '^\s*packages = with pkgs; \[[^]]*\bkustomize\b' flake.nix` = `1`. Today this line is line 27, in the `devShells` block. |
| DSC-04 | #6 | **Follow-up.** `grep -c 'noodle/v1alpha2' schemas/v1alpha1/fragment.schema.json` = `0` still holds after a bump made the usual way, through a new `schemas/v1alpha2/` directory or a new `const` elsewhere. | `grep -c '"apiVersion": {"const": "noodle/v1alpha1"}' schemas/v1alpha1/fragment.schema.json` = `1`, and `ls schemas/` lists only `v1alpha1`. |
| DSC-10 | #5 | **Follow-up.** `grep -l 'discovery.yaml' docs/adr/0*.md` must give "one ADR". A later ADR that only mentions the file makes the count two, and the criterion then fails for a reason that has nothing to do with T13. | Pick the ADR by its frontmatter: `grep -l '^specs: .*"009"' docs/adr/0*.md \| xargs grep -l 'discovery.yaml'` gives exactly one file, and that file has `status: accepted`. |
| (preamble) | "Agreed" | **Follow-up.** Line 17 says `Agreed: pending make review-contract`. Line 24 says `Agreed 2026-11-04`. An evaluator cannot tell from the file whether the contract was agreed. | Keep one line: `Agreed: 2026-11-04, after round 3 and the post-answer review`. Add one line for this review's single applied point. |
