# Spec 003 · Static discovery from Kubernetes

- Status: **Planned** (`plan.md`, `tasks.md`)
- Phase: 6a, a slice of phase 6 pulled forward (ADR-0016). Starts after T13, T16 and
  T17 of spec 001.
- Decisions it builds on: ADR-0002, ADR-0003, ADR-0005, ADR-0007, ADR-0008, ADR-0015,
  ADR-0016, ADR-0017, ADR-0018

## Why

An agent drawing a platform reads its manifests itself: 50 tool calls and about 41 % of
the session's cost on the one session measured (PAE, 2026-09-29). The facts it looks for
(workloads, ports, routes, secrets, policies) are in the manifests, in a form a program
can resolve without a model. Once they are in a fragment, the agent writes only what
code cannot say, and a drift report tells when the diagram no longer matches the system.

## User scenarios

1. **Platform engineer, new diagram.** Runs `helm template … | noodle discover k8s -`,
   then asks Claude for the platform overview. The skill reads the fragment, not the
   charts, and writes a curated model whose elements carry `matches`.
2. **Same engineer, a month later.** Reruns discovery; `noodle drift` lists a new
   workload nobody drew, a route whose backend changed, and the planned Image Updater
   that is still absent.
3. **Operator with read access.** Pipes `kubectl get … -A -o yaml` instead of files and
   gets the same fragment shape, from what is actually deployed.
4. **Contributor adding Terraform later.** Writes `internal/adapter/terraform` against
   the adapter interface; nothing in model, graph, lint or render changes.

## Requirements

- **FR-001** `kind: Fragment` in `schemas/v1alpha1`: elements, connections, references,
  fragment `provenance`, item `src`; `matches` on model elements (ADR-0015).
- **FR-002** An adapter interface and registry in `internal/adapter`; `noodle discover
  <adapter> [paths…|-] [-o file]` runs any registered adapter.
- **FR-003** The k8s adapter reads files, directories and stdin, multi-document YAML and
  `kind: List`, and ignores status, managed fields and owner-generated objects
  (ReplicaSet, Pod, EndpointSlice) so manifests and live output give the same fragment.
- **FR-004** Elements: one per workload (Deployment, StatefulSet, DaemonSet, CronJob),
  Service, Gateway, Ingress controller class, Secret and external target, with ports and
  namespace zones.
- **FR-005** Connections (client to server, listening port, ADR-0003): Gateway and
  Ingress to backend Services through HTTPRoute, GRPCRoute and Ingress rules; Service to
  the workloads its selector matches, on the target port; a workload to another
  Service when an env value or ConfigMap holds its cluster DNS name or URL, marked
  inferred.
- **FR-006** References: parentRef, backendRef, envFrom, secretKeyRef, secret volumes,
  ExternalSecret to SecretStore to provider, Argo CD Application to its destination
  namespace.
- **FR-007** NetworkPolicy: ingress and egress rules resolved to allowed pairs; a
  connection found by FR-005 that a policy in force denies is marked `denied` with
  `enforced_by` (ADR-0010).
- **FR-008** Every item has `src`; secret values are never read into the fragment,
  only names and keys.
- **FR-009** `internal/graph` merges model and fragments per ADR-0015; `noodle drift
  <system>` reports unmatched, gone, planned-and-absent, and unmodelled connections.
- **FR-010** The skill runs discovery first when manifests exist and hands only what is
  left (unknowns, non-Kubernetes sources) to `fact-finder`.
- **FR-011** A token benchmark: three arms (agent reads code, `fact-finder`, fragment) on
  the same diagram, cost and accuracy, with a transcript reader in Go.
- **FR-012** A workload reading a Secret key whose name holds an address gets a
  `secret-endpoint` entry per Secret, keys named, values never read (ADR-0018).
- **FR-013** Operator rules: CloudNativePG `Cluster`, Prometheus operator, cert-manager
  issuers, external-dns provider, each with a fixture (ADR-0018).
- **FR-014** A checkpoint after T05 and T06: the same diagram built by the agent alone
  and with the fragment; discovery work continues only if the fragment run costs clearly
  less and finds as many edges (ADR-0018).

## Success criteria

- Features P6a-01 to P6a-10 in `docs/features.json` pass their verify step.
- The checkpoint (FR-014) is passed before T07 starts; if it fails, the spec is revised
  rather than continued.
- `examples/discovery-k8s` discovers, merges and reports drift in `init.sh` and CI, with
  a golden fragment.
- Rerunning discovery on unchanged input gives an identical file.
- The benchmark is published (Notion page "Benchmark · Token cost of a diagram") with
  arm C measured on the PAE platform, and a recall of connections at least equal to
  arm A's.

## Out of scope

Auto-layout (phase 6, next spec), client-go live discovery, Terraform, Hubble flows,
facets and enrichment modules (reserved by ADR-0015), CI failing on drift, a graph
database.

## Open questions

- None blocking. Committing a fragment as a drift baseline is left out: ADR-0015 keeps
  fragments regenerated and uncommitted; revisit if a team wants drift across releases.
