# ADR-0017 · The k8s adapter decodes with yaml.v3 and reports what it cannot resolve

- Status: Accepted
- Date: 2026-09-30
- Amends: ADR-0016 (libraries); extends ADR-0015 (fragment contract)

## Context

ADR-0016 names `k8s.io/apimachinery` and `sigs.k8s.io/yaml` for the k8s adapter.
`sigs.k8s.io/yaml` goes through JSON and drops line numbers, which every `src` needs
(ADR-0015), so the adapter decodes with `yaml.v3` anyway. apimachinery would then only
bring label selectors, a few lines of code, for a large vendored tree.

Static discovery cannot settle everything: a URL in an env value, an unknown CRD, a
selector that matches nothing. The plan drops them, so the agent that completes the
model would have to reread the manifests to find them, the cost phase 6a exists to cut.

## Decision

- **Decoding.** `yaml.v3` only, with our own helpers for nested fields, `intstr` ports
  and label selectors (`matchLabels`, `matchExpressions`), table-tested. No
  apimachinery, no `sigs.k8s.io/yaml`. ADR-0016's intent stands: our own Go adapter,
  no external tool.
- **Deterministic resolvers first.** Before giving up on a host, the adapter matches it
  against Services, Ingress and HTTPRoute hostnames in the same input, then against a
  short table of well-known SaaS domains.
- **`unresolved` in the fragment.** What is left is listed, not dropped:
  `{about: [ids], kind, value, hint, count, src}`, one flow line per entry.
  Kinds form a closed list (`env-url`, `unknown-kind`, `unmatched-selector`,
  `missing-backend`, `unknown-provider`, `unrendered`), extended by ADR. `unrendered` is
  a Helm chart or a Kustomize directory met in a walk: read raw, its templates and
  patches would give wrong objects, so it is skipped and named with the command to run.
- **Token guards.** An entry exists only if its answer can add an element or a
  connection. The same value across workloads is one entry with every `about`. Unknown
  kinds are one entry per kind with a count. `hint` is a short code, not prose.
  `noodle discover` prints the fragment and `unresolved` sizes in estimated tokens.
- **Resolutions are remembered.** An answer from an agent or a person becomes a rule
  in the system, applied by the adapter on the next run, so each question costs tokens
  once. Its file format is set with the skill change (spec 003 T11).

## Consequences

- `src` keeps file and line; no new dependency, `vendorHash` unchanged.
- `fact-finder` receives the `unresolved` list, not the manifests, in its own context.
- The benchmark (T13) reports `unresolved` size and fails arm C if it costs more than
  arm A.
- Traffic observation (Hubble, OpenTelemetry) stays in phase 6: it resolves connections
  from what actually flows, as a complement to what is declared.
