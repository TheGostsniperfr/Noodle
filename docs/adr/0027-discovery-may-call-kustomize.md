---
status: accepted
date: 2026-11-04
deciders: [Brian Perret]
supersedes: []
extends: ["0016", "0017"]
specs: ["009"]
constitution: [C08, C09]
---

# ADR-0027 · Discovery may call kustomize, and says so when it is missing

## Context and problem

Question 2 of spec 009. Every Arffornia app is a Kustomize directory with `helmCharts`;
`noodle discover k8s` reports each one `unrendered`, so users wrap discovery in a shell
loop (#81). ADR-0016 and ADR-0017 kept discovery to our own Go decoding, with no
external tool. Rendering Kustomize correctly means its patch, overlay and Helm logic,
which we will not reimplement.

## Decision drivers

- Discover what Argo CD actually deploys, not raw bases and patches.
- No silent skip: a user must know when a tool is needed and how to get it.
- Read-only (C09) and deterministic for the same input (C08).

## Considered options

1. **Call `kustomize build` behind an explicit `-kustomize` flag**, fail loudly when the
   tool is missing.
2. **No external tool**: keep `unrendered` and document the loop in the skill.
3. **Vendor the kustomize Go API** into the binary.

## Decision outcome

Chosen: **option 1**, because it renders exactly as Argo CD does at the cost of one
optional tool, and keeps the default path free of it.

- `noodle discover k8s -kustomize DIR [-enable-helm]` runs `kustomize build DIR`
  (`--enable-helm` only with `-enable-helm`), never writes to `DIR`, and discovers the
  output; the fragment is named after the directory.
- With `-kustomize` and no `kustomize` on PATH, discovery **fails** with a message that
  names the tool and how to get it: `nix shell nixpkgs#kustomize`, or `nix develop` in
  this repository.
- Without the flag, a directory holding a `kustomization.yaml` stays `unrendered`, and
  its hint is the exact command to rerun with `-kustomize`.
- `kustomize` joins the flake's dev shell.
- This extends ADR-0016 and ADR-0017: their "no external tool" holds for decoding; an
  explicitly requested renderer is the one exception.

### Consequences

- Good: Kustomize repositories discover in one command.
- Bad: output depends on the `kustomize` version; the error and the hint name the
  version found.

### Confirmation

Spec 009 DSC-03 #3: tests with a fake `kustomize` on PATH for render and argv, and with
none for the failure message and the `unrendered` hint.

## Pros and cons of the options

### Call kustomize behind a flag

- Good: exact rendering, opt-in.
- Bad: one external dependency for that path.

### No external tool

- Good: nothing to install.
- Bad: every Kustomize user writes a loop; discovery misses what is deployed.

### Vendor the kustomize API

- Good: no PATH lookup.
- Bad: a large dependency tree vendored for one feature, and its version still lags Argo CD's.
