# Contract 009 · Discovery, phase 6a completion

What "done" means for each feature, agreed before implementation (ADR-0024). One failed
criterion fails the feature. Every "Verified by" runs inside the evaluator's allowlist
(`EVALUATOR` in the Makefile). In a table cell, `\|` is the Markdown escape of a pipe:
run it as a plain `|`, which inside a `grep -E` pattern is alternation. A literal pipe
character to match is written `[\|]`, that is `[|]` once unescaped.

- **Answers of 2026-11-04** (spec Clarifications): Q1 a, ADR-0029 (fresh `claude -p`
  arms, transcripts without tool results, go when C costs at most 70 % of A with no
  recall loss); Q2 a made explicit, ADR-0027 (with `-kustomize` and no `kustomize`,
  discovery fails naming the tool; without the flag, a kustomization is `unrendered`
  with the rerun command); Q3 c, ADR-0028 (`object` only; field paths and origins out of
  scope); Q4 a (`<system>/discovery.yaml`, its format ADR written in T13); Q5 a. The
  rows they reopened, DSC-03 #3 and #5, DSC-04, DSC-08 #4, DSC-10 #3 and #5, were
  rewritten for them.
- Round 1 (2026-11-04): CHANGES, 33 points, all applied.
- Round 2: CHANGES, 21 points, all applied; phase `6a` now points at this spec, so the
  carried P6a features are evaluated in full mode here.
- Round 3, the last (PROCESS.md cap): CHANGES; the three points that made a criterion
  pass or fail whatever the code does are applied (DSC-04 #2 and #3, the FR-015 gate of
  DSC-09 #4); nine follow-ups are recorded below, for the implementer to apply in the
  first commit of the matching task.
- Post-answer review (2026-11-04): CHANGES, 4 points: DSC-04 #1 `legacy-fragment`
  applied (a byte-identical round trip failed whatever the code does, since yaml.v3
  single-quotes ids in flow style); DSC-03 #3b, DSC-04 #6 and DSC-10 #5 tightened as
  proposed.
- Agreed: 2026-11-04, after round 3 and the post-answer review.

Test names below are the ones the implementation must use. Each named test is
table-driven and asserts with `Equal` on the whole output it checks (a full reference or
connection list, a full stdout), so extra or duplicate items fail it. The example
`examples/discovery-k8s` has namespaces `prod` and `data`, manifests directly in
`manifests/` (no subdirectory), and Deployment `worker` deliberately matched by no
curated element. Committed transcripts hold no `tool_result` record (plan 009).

## DSC-01 · Configuration references

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | One fixture per reference kind of FR-001 and FR-002 | `go test ./internal/adapter/k8s/ -run TestReferences -v` | exit 0, with subtests `env-from`, `secret-key`, `config-key`, `mounts-secret`, `mounts-configmap`, `mounts-pvc`, `externalsecret-target`, `externalsecret-store`, `clusterstore-provider`, `store-provider-1password`, `store-provider-vault`, `store-provider-aws`, `argocd-syncs`; each asserts with `Equal` the fragment's full `references` list, `src` lines included |
| 2 | Providers are typed | same | the three `store-provider-*` subtests assert an element of kind `external` with tech `1Password`, `Vault`, `AWS` |
| 3 | No Secret value reaches a fragment, for a fixture and for every input | `go test ./internal/adapter/k8s/ -run 'TestReferences_NoSecretValue\|TestReferences_Property_NoSecretValue' -v` | exit 0: the fixture with `data` and `stringData` yields no value nor its base64 form in the fragment; the property test (`testing/quick`, `MaxCount` ≥ 200) does the same over generated Secrets with random values |

## DSC-02 · Addresses

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The test passes | `go test ./internal/adapter/k8s/ -run TestAddresses -v` | exit 0 |
| 2 | The full matrix of FR-003 forms | `go test ./internal/adapter/k8s/ -run TestAddresses -v \| grep -cE -- '--- PASS: TestAddresses/(svc-ns\|svc-ns-svc\|svc-ns-svc-cluster-local)/(port\|noport)/(scheme\|noscheme)/(configmap\|env\|cr)'` | `36`; each of these subtests asserts with `Equal` one connection with `inferred: true`, the port of plan 009's port rule, and the value's `src` line |
| 3 | The port rule without a port in the value | `TestAddresses/noport-multiport` in #1 | a Service with ports `metrics` 9090 and `http` 8080 yields the connection on 8080 |
| 4 | Unknown hosts are not guessed | `TestAddresses/unknown-service` and `TestAddresses/external-host` in #1 | no connection and no new element |
| 5 | Public hosts map to their entry point | `TestAddresses/route-hostname` and `TestAddresses/ingress-host` in #1 | exactly one inferred connection from the reading workload to the Gateway of the HTTPRoute, resp. to the Ingress controller element |

## DSC-03 · Inputs

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The namespace warning | `go test ./cmd/noodle/ -run TestNamespace -v` | exit 0: subtest `mixed` prints exactly `3 objects name no namespace; pass -namespace` on stderr; `all-named` prints no such line; `namespace-flag` with `-namespace web` prints no such line and gives the 3 objects ids under `k8s:web/` |
| 2 | `-argocd-app` sets the default namespace, silently | same, subtest `argocd-app` | the 3 objects get ids under `k8s:web/`, and stderr has no `name no namespace` line |
| 3 | `-kustomize` renders read-only, passes `-enable-helm` only when asked, names the fragment | `go test ./internal/adapter/k8s/ ./cmd/noodle/ -run TestKustomize -v` | exit 0, with a fake `kustomize` on PATH that records its argv: `render` discovers its output, asserts argv is exactly `build <DIR>`, and the input directory's file hashes are unchanged; `enable-helm` asserts argv exactly `build <DIR> --enable-helm`; `fragment-name` asserts the file written is `discovered/k8s-shop.yaml`; `missing-tool` runs `-kustomize` with no `kustomize` on PATH and asserts a non-nil error whose message contains `kustomize` and `nix shell nixpkgs#kustomize`, and that no fragment is written; `hint-without-flag` walks a directory holding `kustomization.yaml` without the flag and asserts with `Equal` one `unrendered` entry whose hint is `noodle discover k8s -kustomize <DIR>` |
| 3b | The tool is in the dev shell | `grep -cE '^\s*packages = with pkgs; \[[^]]*\bkustomize\b' flake.nix` | `1` (the `devShells` package list) |
| 4 | `src` paths are relative | `go test ./internal/adapter/k8s/ -run TestSrcPaths -v`; `grep -rcE '(^\|[ {,-])file: */' examples/discovery-k8s/discovered/` | exit 0, subtests `manifests` and `kustomize` assert that every `src.file` starts with the input directory relative to the working directory, none starts with `/`, none contains `os.TempDir()`; grep gives `0` for every file |
| 5 | The decision is recorded | `grep -c '^status: accepted' docs/adr/0027-discovery-may-call-kustomize.md`; `make spec-check` | `1`; exit 0 |

## DSC-04 · Object identity

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The contract: optional top-level `object`, `src.object` kept | `go test ./internal/model/ -run TestIdentity -v` | exit 0: `schema-accepts` (an item with `object {apiVersion, kind, namespace, name}`, and an item without `object`); `schema-rejects` with three cases, no `apiVersion`, no `kind`, no `name`, each refused; `src-object-unchanged` (a fragment using only the string `src.object` validates); `legacy-fragment` loads `internal/model/testdata/minimal/discovered/k8s-manifests.yaml` (which validates against `schemas/v1alpha1/fragment.schema.json`), encodes it with `EncodeFragment` and loads the result again: `Equal` on the two `Fragment` values, no encoded item has a top-level `object` key, and `bytes.Equal(EncodeFragment(load(fixture)), EncodeFragment(load(EncodeFragment(load(fixture)))))` |
| 2 | The legacy fixture predates this spec | `git log --format=%h origin/main -- internal/model/testdata/minimal/discovered/k8s-manifests.yaml` | non-empty |
| 3 | The adapter fills them where an object stands behind the item | `go test ./internal/adapter/k8s/ -run TestIdentity -v` | exit 0: `scenario` asserts FR-008's scenario exactly; `every-item`, on the fragment of `examples/discovery-k8s/manifests`, asserts 0 object-derived items without `object`, and with `Equal` the full list of ids that carry no `object` (the synthesized elements only); no item or reference has a `field` key (out of scope, ADR-0028); `connection-object` asserts an inferred connection carries the `object` of the ConfigMap whose value created it |
| 4 | Existing fixtures are not rewritten | `git diff --name-only --diff-filter=MD origin/main...HEAD -- internal/adapter/k8s/testdata cmd/noodle/testdata internal/model/testdata` | no output |
| 5 | The decision is recorded (C06) | `grep -c '^status: accepted' docs/adr/0028-fragments-carry-object-identity.md`; `make spec-check` | `1`; exit 0 |
| 6 | No version bump | `ls schemas/`; `grep -c 'noodle/v1alpha1' schemas/v1alpha1/fragment.schema.json` | only `v1alpha1`; at least `1`, and no other apiVersion string in that file |

## DSC-05 · The example

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Golden from manifests, twice, and from a List | `go test ./cmd/noodle/ -run TestDiscoverExample -v` | exit 0, subtests `manifests-first`, `manifests-second`, `kubectl-list`, each byte-equal to `examples/discovery-k8s/discovered/k8s-manifests.yaml` |
| 2 | Determinism for every input | `go test ./internal/adapter/k8s/ -run TestDiscover_Property_Deterministic -v` | exit 0: `testing/quick`, `MaxCount` ≥ 200, over generated manifest sets in shuffled order, two runs byte-identical |
| 3 | The example has what spec 003's plan lists | for each kind K, `cat examples/discovery-k8s/manifests/*.yaml \| grep -c '^kind: K$'`; `grep -c 'matches:' examples/discovery-k8s/model.yaml`; `grep -c 'status: planned' examples/discovery-k8s/model.yaml` | Gateway ≥ 1, HTTPRoute ≥ 2, Deployment ≥ 3, StatefulSet ≥ 1, Secret ≥ 1, ExternalSecret ≥ 1, NetworkPolicy ≥ 2; matches ≥ `2`; planned = `1` |
| 4 | `make examples` reruns and diffs | `make examples`; `grep -cE 'discover k8s examples/discovery-k8s/manifests' Makefile`; `grep -cE '(diff\|cmp).*examples/discovery-k8s/discovered/k8s-manifests.yaml' Makefile` | exit 0; `1`; `1` |

## DSC-06 · Merge

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | ADR-0015 merge rules | `go test ./internal/graph/ -run TestMerge -v` | exit 0, subtests `fills-empty-field`, `keeps-curated-field`, `maps-connection-to-curated-ids`, `takes-list-whole`, each asserting with `Equal` the merged element or connection |
| 2 | Sources survive the merge | same, subtest `keeps-sources` | the merged element lists the `src` and `object` of each matched item |
| 3 | Commands load the merged graph | `go test ./cmd/noodle/ -run TestLoad_MergesFragments -v` | exit 0: loading `examples/discovery-k8s` through the command loader yields the merged `api` element with port 8080 from the fragment |
| 4 | Curated fields never overridden, for every input | `go test ./internal/graph/ -run TestMerge_Property -v` | exit 0: `testing/quick`, `MaxCount` ≥ 500; the generator sets at least one scalar and one list field on each curated element; the test asserts every curated field is equal after the merge, and that at least one case had a field set by both sides |

## DSC-07 · Drift

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The four drift families, sorted | `go test ./internal/graph/ -run TestDrift -v` | exit 0, subtests `unmatched`, `gone`, `planned-absent`, `unmodelled-connection`, each asserting with `Equal` the entry and its `src`; `sorted` feeds two entries per family in reverse order and asserts sorted output with `Equal` |
| 2 | Cross-fragment resolution | `go test ./internal/graph/ -run TestCrossFragment -v` | exit 0: a route in one fragment and its Gateway in another are linked, and no `missing-backend` remains for that route |
| 3 | The example's report, byte for byte | `go test ./cmd/noodle/ -run TestDrift -v` | exit 0: `noodle drift examples/discovery-k8s` stdout equals `examples/discovery-k8s/drift.txt`, and the command exits 0 |
| 4 | The report is not empty and holds FR-011's scenario | `cat examples/discovery-k8s/drift.txt`; `grep -cE '^(unmatched\|gone\|planned\|unmodelled)' examples/discovery-k8s/drift.txt`; `grep -cE 'worker.*manifests/.*:[0-9]+' examples/discovery-k8s/drift.txt` | `4` headings, each followed by at least one entry line; `1`, and that line sits under `unmatched` |

## DSC-08 · Token cost and checkpoint

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The transcript reader prints all four measures | `go test ./tools/tokencost/ -v`; `grep -cE '^(category\|model\|turns\|wall)' tools/tokencost/testdata/expected.txt` | exit 0, the test asserting with `Equal` the full stdout on the fixture against `expected.txt`; each of the four prefixes at least once |
| 2 | The checkpoint rows and transcripts | `grep -cE '^[\|] [AC] [\|] \$[0-9]+\.[0-9]{2} [\|] [0-9]+/25 [\|] docs/benchmarks/transcripts/[^ ]+\.jsonl [\|]$' docs/benchmarks/discovery-checkpoint.md`; `ls` of each transcript path | `2`; each path exists |
| 3 | Transcripts hold no tool results (C09) | `grep -c '"type":"tool_result"' docs/benchmarks/transcripts/*.jsonl` | `0` for every file |
| 4 | The decision follows ADR-0029 | `cat docs/benchmarks/discovery-checkpoint.md`; `grep -c '^status: accepted' docs/adr/0029-discovery-checkpoint-threshold.md` | the file cites ADR-0029 and has exactly one line `Decision: go` or `Decision: stop`, `go` exactly when C's cost is at most 70 % of A's and C's edges are at least A's, given the two rows; `1` |
| 5 | PROGRESS records cost and recall of both arms on platform-overview (P6a-08) | `grep -cE 'checkpoint.*platform-overview.*A.*\$[0-9.]+.*[0-9]+/25.*C.*\$[0-9.]+.*[0-9]+/25.*(go\|stop)' docs/PROGRESS.md` | `1` |

## DSC-09 · Rules after the checkpoint

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Each rule has its fixtures | `go test ./internal/adapter/k8s/ -run 'TestNetworkPolicy\|TestSecretEndpoint\|TestOperatorRules' -v` | exit 0, subtests `netpol-deny`, `netpol-deny-route`, `netpol-egress`, `netpol-allow`, `secret-endpoint-env`, `secret-endpoint-envfrom`, `secret-endpoint-externalsecret`, `secret-endpoint-ignores-password`, `cnpg-cluster`, `prometheus-alertmanager`, `certmanager-issuer`, `external-dns`, `servicemonitor`, `podmonitor`; each asserts with `Equal` the fixture's full list of the items it produces (connections, entries or elements) |
| 2 | FR-016's scenario, both halves | `TestNetworkPolicy/netpol-deny` in #1 | the full connection list has `web → postgres` with `denied: true` and `enforced_by: k8s:data/networkpolicy/default-deny`, and `api → postgres` with neither field |
| 3 | `secret-endpoint` names keys and readers, never values | `TestSecretEndpoint/secret-endpoint-env` and `TestSecretEndpoint/secret-endpoint-ignores-password` in #1 | value `db/DB_URL`, `about` equal to `[k8s:prod/deployment/api, k8s:prod/deployment/worker]`, no Secret value in the fragment; no entry for key `PASSWORD` |
| 4 | Started only after `Decision: go` (FR-015) | anchor `git log --format=%h -S'Decision: go' --reverse -- docs/benchmarks/discovery-checkpoint.md \| head -1`; content check: `git log --format=%h -G'func Test(NetworkPolicy\|SecretEndpoint\|OperatorRules\|Rules\|ReferenceKinds)\|enforced_by\|EnforcedBy\|secret-endpoint' -- internal/adapter/k8s/` and the same limited to `<anchor>..HEAD`; file check: for `netpol.go`, `secretendpoint.go`, `operators.go`, `rules.go`, `refkinds.go` of `internal/adapter/k8s/` and their `_test.go`, `git log --format=%h <anchor>..HEAD -- <files>` and `git log --format=%h -- <files>` | the anchor exists, and in each check the two lists are equal; if the file says `Decision: stop` instead, `grep -c '^## Revision' specs/009-discovery-completion/spec.md` ≥ `1` and both full lists are empty |

## DSC-10 · Rules file and reference kinds

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | A rule in `<system>/discovery.yaml` maps a kind to references and to elements | `go test ./internal/adapter/k8s/ -run TestRules -v` | exit 0: `vaultsecret` asserts FR-019's scenario (14 references, no `unknown-kind` for `VaultSecret`); `element` asserts one element per object, with `Equal` on `{id, kind, object, src}`; `template-ns`, `template-name` and `template-fanout` (N references for N `targetRefs`) assert with `Equal` |
| 2 | A malformed rule is refused | same, subtest `bad-template` | discovery fails with an error matching `^.*discovery\.yaml:[0-9]+: ` followed by the template, asserted with `Regexp` |
| 3 | The six reference kinds of FR-020 | `go test ./internal/adapter/k8s/ -run TestReferenceKinds -v` | exit 0, subtests `owner`, `selects`, `runs-as`, `bound-by`, `scales`, `issues`, each asserting `{from, to, kind, src.line}` with `Equal` (no `field`, ADR-0028) |
| 4 | Started only after `Decision: go` | DSC-09 #4, which covers `rules.go` and `refkinds.go` | same threshold |
| 5 | The rules file format is decided | `grep -l '^specs: .*"009"' docs/adr/0*.md`, then `grep -l 'discovery.yaml'` on those files; for the result, `grep -c '^status: accepted'`; `make spec-check` | exactly one file, written in T13; `1`; exit 0 |

## DSC-11 · Query

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Neighbours, filtered and not | `go test ./cmd/noodle/ -run TestQuery -v` | exit 0: `neighbours-kind` runs `noodle query examples/discovery-k8s neighbours k8s:prod/secret/db -kind secret-key` and asserts stdout equal to a golden listing exactly `k8s:prod/deployment/api` and `k8s:prod/deployment/worker`, each line with kind `secret-key` and a `file:line`; `neighbours-all` runs the same without `-kind` against its own golden |
| 2 | Shortest path, stable, and no path | same, subtests `path`, `path-tie`, `no-path` | `path` runs `path k8s:prod/deployment/web k8s:data/statefulset/postgres` and asserts stdout equal to its golden, each edge with kind and `src`; `path-tie` runs 20 times on a graph with two equal paths and asserts the same output each time, following plan 009's lexicographic rule; `no-path` prints exactly `no path` and exits 1 |

## DSC-12 · Skill and benchmark

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Discovery is step 1 of the skill | `grep -nE '^[0-9]+\. ' skills/noodle-diagram/SKILL.md`; `grep -n 'noodle discover' skills/noodle-diagram/SKILL.md` | the first `noodle discover` line comes after the `1. ` item and before the `2. ` item |
| 2 | `fact-finder` gets only what is left; the rules file is documented | `grep -n 'fact-finder' skills/noodle-diagram/SKILL.md`; `grep -n 'unresolved' skills/noodle-diagram/SKILL.md`; `grep -c 'discovery.yaml' skills/noodle-diagram/reference.md`; `grep -c 'matches' skills/noodle-diagram/reference.md` | the first `unresolved` line lies between the item naming `fact-finder` and the next numbered item; ≥ `1`; ≥ `1` |
| 3 | The skill names the edges discovery never sees | `grep -niE 'never sees' skills/noodle-diagram/SKILL.md skills/noodle-diagram/reference.md` | at least one match listing outside-cluster, CI and secret-value edges |
| 4 | The skill's run reads no manifest in any way (FR-022, P6a-06) | `grep '"type":"tool_use"' docs/benchmarks/transcripts/skill-discovery-k8s.jsonl \| grep 'examples/discovery-k8s/manifests' \| grep -vc 'noodle discover'`; `grep -c 'noodle discover' docs/benchmarks/transcripts/skill-discovery-k8s.jsonl` | `0`; ≥ `1` |
| 5 | The benchmark rows and means | `grep -cE '^[\|] [ABC] [\|] [12] [\|] \$[0-9.]+ [\|] [0-9]+/[0-9]+ [\|] [0-9]+m[0-9]+s [\|] docs/benchmarks/transcripts/[^ ]+\.jsonl [\|]$' docs/benchmarks/discovery-tokens.md`; `ls` of each transcript; `grep -c '"type":"tool_result"' docs/benchmarks/transcripts/*.jsonl` | `6`; each exists; `0` for every file; the file states the mean recall of A and C, C ≥ A |
| 6 | The Notion publication | `grep -cE '(notion\.(so\|site)\|app\.notion\.com)/[^ ]+.*20[0-9]{2}-[0-9]{2}-[0-9]{2}' docs/benchmarks/discovery-tokens.md`; `grep -c 'Checked on Notion by Brian' docs/PROGRESS.md` | ≥ `1`; ≥ `1` |

## Immutable features carried by this spec

Phase `6a` points at this spec, so `make evaluate FEATURE=P6a-0N` reads this contract and
writes `specs/009-discovery-completion/evidence/P6a-0N.md`, in full mode (C12).

| Feature | Graded by |
|---|---|
| P6a-03 | DSC-05 #1 and #4 |
| P6a-04 | `go test ./internal/adapter/k8s/ -run TestExample_ManifestsEqualList -v`, exit 0: the fragment of `examples/discovery-k8s/manifests` equals byte for byte the fragment of `list.yaml`, so P6a-04's own verify (`./internal/adapter/k8s/...`) makes the comparison |
| P6a-05 | DSC-07 #3 and #4 |
| P6a-06 | DSC-12 #1 to #4 |
| P6a-07 | the committed `discovery-tokens.md` (DSC-12 #5) with the Notion URL and its last-edited date, and DSC-12 #6, since the evaluator cannot open Notion |
| P6a-08 | DSC-08 #2 to #5 |
| P6a-09 | DSC-09 #1 and #3 (env, envFrom and ExternalSecret targets) |
| P6a-10 | DSC-09 #1, operator subtests |
| all | `go run ./tools/speccheck locate P6a-03` prints exactly `specs/009-discovery-completion` |

## Follow-ups from round 3

Not graded until the implementer applies them in the first commit of the matching task;
from that commit on, they replace the criterion they name.

| Feature | Criterion | Follow-up |
|---|---|---|
| DSC-12 | #4 | also catch indirect reads: `grep '"type":"tool_use"' … \| grep -E 'manifests\|list\.yaml\|"name":"(Grep\|Glob)".*discovery-k8s' \| grep -vc 'noodle discover'` = `0` |
| DSC-08, DSC-12 | DSC-08 #3, DSC-12 #5 | `grep -cE '"type": *"tool_result"\|"toolUseResult"' docs/benchmarks/transcripts/*.jsonl` = `0` for every file |
| DSC-12 | #5 | for each X in A, B, C and N in 1, 2: `grep -cE '^[\|] X [\|] N [\|]' docs/benchmarks/discovery-tokens.md` = `1`; add `cat` of the file so the C ≥ A means can be read |
| DSC-08 | #4 | done: #4 now checks ADR-0029's `status: accepted` |
| DSC-09 | #3 | the three `secret-endpoint-env`, `-envfrom` and `-externalsecret` subtests each assert no Secret value and no base64 form of it in the fragment |
| DSC-08 | #1 | four counts, `^category`, `^model`, `^turns`, `^wall` on `expected.txt`, each ≥ `1` |
| DSC-11 | #2 | plan 009 decides whether `path` crosses `denied` connections (proposed: it skips them), and subtest `path-denied` asserts it |
| DSC-04 | scope | done: `schemas/v1alpha1/` is in plan 009's table |
| DSC-03 | #4 | `grep -rcE '(^\|[ {,-])file: *"?/' examples/discovery-k8s/discovered/` = `0` for every file |
