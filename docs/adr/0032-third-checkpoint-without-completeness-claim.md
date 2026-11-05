---
status: accepted
date: 2026-11-05
deciders: [Brian Perret]
supersedes: []
extends: ["0029", "0031"]
specs: ["009"]
constitution: [C08, C09]
---

# ADR-0032 · One last checkpoint, with arm C told what the digest never sees

## Context and problem

ADR-0031 ends the question after a second `stop`. The second checkpoint said stop
([`discovery-checkpoint-2.md`](../benchmarks/discovery-checkpoint-2.md)), but not for
the reason ADR-0031 tested:

- cost passed: arm C at 63 % of arm A;
- recall failed on three edges the digest cannot hold (Grafana to Prometheus, Prometheus
  to Alertmanager, image pulls), and C's prompt called the digest "complete for the
  connections between components inside the cluster". That claim was false, and it
  told C not to look.

Phase 6a's exit criterion is this checkpoint (ROADMAP). Is the stop a verdict on the
digest, or on a false sentence in the prompt?

## Decision drivers

- Measure the digest, not a wrong claim about it.
- No tuning toward a `go`: the threshold, the snapshot and the ground truth do not move.
- One final answer, so phase 6a closes either way.

## Considered options

1. **A third and last run.** Only C's paragraph changes: it drops the completeness
   claim and lists what discovery never sees.
2. **Accept the second stop** as ADR-0031 says.

## Decision outcome

Chosen: **option 1**, with Brian's agreement of 2026-11-05.

- C's paragraph ([`checkpoint-prompt-c3.md`](../benchmarks/checkpoint-prompt-c3.md))
  describes the digest as it is. It then names the edges discovery never sees:
  connections implied by operator custom resources or by chart values such as
  datasources, image pulls, and anything outside the cluster. That list is what the
  skill must state anyway (FR-022).
- Everything else is as in round 2: snapshot, model, tools, ground truth, both arms run
  again, and ADR-0029's rule. The result goes to `docs/benchmarks/discovery-checkpoint-3.md`.
- This round is final. On `stop`, ADR-0031's consequence applies: discovery stops serving
  the agent, FR-022 and FR-023 are dropped, and phase 6a closes without its exit
  criterion. On `go`, wave 3 opens, operator rules first (FR-018), since they carry the
  edges that rounds 2 and 3 cannot see.

### Consequences

- Good: the decision rests on an honest prompt.
- Bad: the protocol changed after a result was known. This record states the change and
  its single cause, so it can be challenged.
- Bad: about $8 more.

### Confirmation

Spec 009 revision 2, feature DSC-15: rows A and C, and a decision that follows ADR-0029,
in `discovery-checkpoint-3.md`.

## Pros and cons of the options

### Third run

- Good: separates the digest from the claim.
- Bad: one more run, and a protocol change after the fact.

### Accept the stop

- Good: keeps ADR-0031 as written.
- Bad: phase 6a would close on a result caused by a sentence we know is false.
