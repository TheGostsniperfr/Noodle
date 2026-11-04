---
status: accepted
date: 2026-11-04
deciders: [Brian Perret]
supersedes: []
extends: ["0015"]
specs: ["009"]
constitution: [C06]
---

# ADR-0028 · Fragments carry an optional object identity

## Context and problem

Question 3 of spec 009. To jump from a node to the YAML document that defines it, and to
answer queries such as "who reads this Secret?", a discovered item needs its Kubernetes
identity, not only a file and line. Today `src.object` is a free string. #16 also asks
for the field path behind each reference and git origins. A contract change needs an
ADR and, unless additive, a version bump (C06).

## Decision drivers

- Useful identity now, without breaking existing fragments.
- Keep `v1alpha1`; a bump belongs to the versioning work of #104.
- Small change first; field paths and origins when a consumer needs them.

## Considered options

1. **Additive `object`, field paths and origins in `v1alpha1`.**
2. **Wait for the `v1alpha2` bump of #104.**
3. **Additive `object` only, now**; field paths and git origins later.

## Decision outcome

Chosen: **option 3**.

- A discovered item built from one API object may carry a top-level
  `object: {apiVersion, kind, namespace, name}`, with `apiVersion`, `kind` and `name`
  required when present. Elements synthesized without an object behind them carry none;
  a connection carries the `object` whose value created it.
- The string `src.object` stays valid and unchanged.
- The field is optional, so `v1alpha1` fragments without it stay valid: no version bump.
- Field paths on references and git or API origins are out of scope, kept with #16.

### Consequences

- Good: `noodle query` and drift can name objects exactly; old fragments still load.
- Bad: two ways to name an object (`src.object` and `object`) until #104 retires one.

### Confirmation

Spec 009 DSC-04: schema tests accept and reject `object`, and a fragment committed
before the change loads and marshals back byte-identical.

## Pros and cons of the options

### Object, field paths and origins now

- Good: everything #16 asks for at once.
- Bad: a larger contract change before any consumer uses field paths.

### Wait for v1alpha2

- Good: one clean bump.
- Bad: queries and drift wait for #104.

### Object only, now

- Good: smallest additive change that unblocks queries.
- Bad: field paths come in a second change.
