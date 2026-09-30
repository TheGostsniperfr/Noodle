# ADR-0016 · Static Kubernetes discovery comes before the viewer

- Status: Accepted
- Date: 2026-09-30
- Amends: the phase order in `docs/ROADMAP.md`

## Context

The roadmap puts adapters in phase 6, after the SVG renderer, the MCP server and the
viewer, because discovered models need auto-layout and a way to look at them. Field use
changed the priority: a diagram of one platform costs about 100k tokens, and the largest
share is an agent reading manifests turn after turn. Discovery that only feeds the
existing authoring workflow needs neither the viewer nor auto-layout.

## Decision

- Pull a slice of phase 6 forward as **phase 6a**: the adapter interface, the `Fragment`
  contract (ADR-0015), a Kubernetes adapter reading manifests or `kubectl get -o yaml`,
  merge and drift report, and the skill using them.
- Build it on Kubernetes libraries (`k8s.io/apimachinery`, `sigs.k8s.io/yaml`,
  Gateway API types), not on an external tool. Existing tools (KubeDiagrams,
  steampipe, kube-lineage, KubeAtlas) produce graphs of Kubernetes objects; noodle needs
  connections from client to server with a listening port (ADR-0003), which none derive.
  KubeDiagrams serves as a test oracle for object relations, not as a dependency, so the
  plugin stays one Go binary.
- Terraform (InfraMap's provider rules are the reference), client-go live discovery,
  Hubble flows and auto-layout stay in phase 6.
- Measure: the same diagram built three ways (agent reads code, `fact-finder` subagent,
  static fragment), cost and accuracy compared.

## Consequences

- Phase 1 closes first: T13, T16 and T17 of spec 001 come before spec 003.
- Discovered systems still need a hand or agent written layout until auto-layout lands.
