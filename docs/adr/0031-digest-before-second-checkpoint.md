---
status: accepted
date: 2026-11-05
deciders: [Brian Perret]
supersedes: []
extends: ["0018", "0029"]
specs: ["009"]
constitution: [C08, C09]
---

# ADR-0031 · After the stop, the agent gets a compact digest and the secret endpoints, then the checkpoint runs again

## Context and problem

The checkpoint of spec 009 T09 said `stop`
([`discovery-checkpoint.md`](../benchmarks/discovery-checkpoint.md)). Arm C, given the
fragments, cost 110 % of arm A and found 25/25 edges to A's 24/25. Running discovery
costs no tokens. The cost is in what the agent reads:

- about 38k tokens of fragments, mostly configuration references a diagram does not
  need, read back from the cache on every turn;
- 13 more turns, spent re-reading the repositories for the external services. Those
  addresses come from Secret keys and application config, which no fragment names.

FR-015 says to revise the spec after a stop. The question: drop discovery for the
agent, or change what the agent receives and measure again?

## Decision drivers

- Compute without AI is free. Tokens are what the checkpoint guards.
- Keep the edge that only C found, and recall at least A's.
- A decision that two rows can check, as in ADR-0029.
- No Secret value anywhere (C09).

## Considered options

1. **A digest for the agent and `secret-endpoint` first, then a second checkpoint**
   with the same protocol and threshold.
2. **Stop static discovery for the agent now.** Keep it for `drift`, `check` and
   `query` only.
3. **Change only the prompt of arm C** so that it trusts the fragments, then run again.

## Decision outcome

Chosen: **option 1**. It attacks both causes of the cost and keeps the 70 % rule.

- `noodle digest DIR` reads the fragments of a `discovered/` directory and prints one
  plain-text line per fact:
  - components: workloads, gateways and external elements, with a Service folded into
    the workload it selects;
  - traffic connections, routes resolved to the workload behind them, and the
    ExternalSecret path to its provider;
  - the `unresolved` entries, `secret-endpoint` included.
  
  Configuration references to ConfigMaps, Secrets and volumes are left out. Every line
  keeps a short `src` (`file:line`). The digest is a view for agents, not a contract:
  it has no schema and no `apiVersion`, and its format can change without a version
  bump. The fragment stays the contract (C06).
- On the PAE fragments of T09, the digest is at most 25 % of the fragments' estimated
  tokens.
- FR-017 (`secret-endpoint`) leaves the FR-015 gate and lands before the second
  checkpoint, because it targets the re-reading. FR-016, FR-018, FR-019 and FR-020 stay
  gated.
- Second checkpoint: same snapshot, model, tools and ground truth. Arm A is run again,
  because one run per arm is noisy. Arm C gets the digest instead of the fragments, and
  its paragraph says the digest is complete for in-cluster connections and lists the
  Secret keys behind external addresses. It is recorded in
  `docs/benchmarks/discovery-checkpoint-2.md`, with the rule of ADR-0029.
- If it says `stop` again, static discovery stops serving the agent. FR-022 (skill step 1
  runs discovery) and FR-023 (the three-arm benchmark) are dropped, wave 3 stays closed,
  and discovery is kept for `drift`, `check` and `query`.

### Consequences

- Good: a second `stop` ends the question with the fixes already tried. A `go` opens
  wave 3 on a measured base.
- Good: the digest is the input the skill would give the agent anyway (FR-022).
- Bad: about $9 for two more arms. The digest may drop a fact that only the full
  fragment holds, and recall would show it.
- Bad: DSC-09 #4 anchored the gate on `Decision: go` in the first checkpoint file. It
  now anchors on the second file and excludes the `secret-endpoint` files, through a
  contract follow-up.

### Confirmation

Spec 009 revision 1, features DSC-13 (digest and its size bound) and DSC-14 (second
checkpoint rows and decision), graded by `make evaluate`.

## Pros and cons of the options

### Digest, secret-endpoint, second checkpoint

- Good: removes the volume and the reason to re-read, then measures.
- Bad: one more run and one more command.

### Stop for the agent now

- Good: no more spend.
- Bad: gives up on the one edge C found, from a single noisy run, with an obvious fix
  untried.

### Prompt only

- Good: cheapest.
- Bad: the agent would still read 38k tokens on every turn. The prompt cannot supply
  the external addresses the fragments lack.
