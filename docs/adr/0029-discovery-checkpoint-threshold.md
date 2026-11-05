---
status: accepted
date: 2026-11-04
deciders: [Brian Perret]
supersedes: []
extends: ["0018"]
specs: ["009"]
constitution: [C08, C09]
---

# ADR-0029 · The discovery checkpoint passes when arm C costs at most 70 % of arm A, with no recall loss

## Context and problem

Question 1 of spec 009. ADR-0018 continues discovery work only if the agent given the
fragment "costs clearly less and finds as many edges" as the agent alone. "Clearly
less" has no number, so the checkpoint result could be argued either way. It also does
not say who runs the sessions.

## Decision drivers

- A decision a script and a reviewer can check from two rows.
- Runs reproducible and free of manual steps.
- No Secret value from a manifest committed with a transcript (C09).

## Considered options

1. **Claude runs each arm as a fresh `claude -p` process; C passes at 70 % of A's cost
   or less, with recall at least A's.**
2. **Brian runs the arms by hand**, same threshold.
3. **Run on `examples/discovery-k8s` instead of PAE**, same threshold.

## Decision outcome

Chosen: **option 1**.

- Each arm is one fresh `claude -p` process on a local PAE checkout, same model.
- `Decision: go` when arm C's cost is at most 70 % of arm A's **and** C finds at least
  as many of the 25 in-cluster edges as A; `Decision: stop` otherwise.
- Transcripts are committed in `docs/benchmarks/transcripts/` with every tool result
  removed, keeping usage records and tool inputs.

### Consequences

- Good: the go or stop follows from the numbers.
- Bad: one run per arm is noisy; the later benchmark (spec 009 FR-023) runs each arm twice.

### Confirmation

Spec 009 DSC-08: the checkpoint file's two rows, its Decision line checked against this
rule, and transcripts with no tool result.

## Pros and cons of the options

### Claude runs the arms

- Good: reproducible, no manual step.
- Bad: headless runs may differ slightly from interactive ones.

### Brian runs them

- Good: closest to real use.
- Bad: manual, slower, harder to repeat.

### The example instead of PAE

- Good: small and committed.
- Bad: too small to measure cost; ADR-0018 asked for a real platform.
