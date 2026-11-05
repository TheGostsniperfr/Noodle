# Noodle · Constitution

The invariants every spec, plan, ADR and change is checked against. Each one already
exists somewhere in the vision, the ADRs or AGENTS.md; this page numbers them so a spec
can cite them (`C04`) and `make spec-check` can verify the citation. Changing an
invariant needs an ADR, like any other decision (C10).

| Id | Invariant | Source |
|---|---|---|
| **C01** | **Model first.** What exists is described once, in the model. Views and layouts reference it and never redefine it. | Vision 1, ADR-0002 |
| **C02** | **An arrow is a connection**, from the side that opens it to the side that listens, the listening port on the server. References are configuration, never traffic. Responses are notes. | Vision 5, ADR-0003 |
| **C03** | **Topology and sequence are separate views**; a topology has one step per connection. | ADR-0004, ADR-0011 |
| **C04** | **Readable by construction.** A diagram that fails the lint is not emitted as a deliverable: `-draft` draws it with its findings and exits 1 (ADR-0036). A lint rule is never relaxed to make a diagram pass: the diagram is fixed. | Vision 4, ADR-0006, ADR-0036 |
| **C05** | **Git is the source of truth.** Model, views and layouts are reviewed in PRs; rendered artifacts are generated, never edited by hand. | Vision 6, ADR-0005 |
| **C06** | **Contracts are the boundary.** A module depends on schemas, never on another module's internals. A contract change needs an ADR and a schema version bump. | Vision 7, ADR-0007 |
| **C07** | **Truth comes from code and infrastructure; humans curate.** Adapters write fragments and never edit the curated model. The gap between the two is reported, not hidden. | Vision 3, ADR-0015 |
| **C08** | **Static first.** What deterministic code can know is computed, not asked of a model. An agent handles only what static analysis reports it cannot resolve. | ADR-0016, ADR-0017 |
| **C09** | **Read-only by default.** Nothing touches live infrastructure in write mode; secret values are never read into any output. | Vision 8, spec 003 FR-008 |
| **C10** | **Decisions live in ADRs.** A choice a future reader would ask "why?" about gets one. An accepted ADR is never rewritten, only superseded. | AGENTS.md |
| **C11** | **Stay in the current phase.** The current phase is the first unchecked one in the roadmap. Pulling work forward needs Brian's agreement and an ADR. | AGENTS.md, ROADMAP |
| **C12** | **A feature passes on evidence, not on its author's word.** Its verify step is rerun and recorded; in full mode an evaluator other than the implementer grades it. A `features.json` entry is never reworded or deleted. | ADR-0024 |
| **C13** | **Review is visual.** A change to what noodle draws shows the rendered before and after, and someone looks at the PNG. | AGENTS.md |
| **C14** | **One way in.** Every developer command is a `make` target listed in `docs/COMMANDS.md`; hooks and CI call the same targets. | ADR-0024 |
| **C15** | **House conventions.** Everything in English. Go is gofmt'd and vetted, errors are wrapped with context, tests are table-driven with testify. | AGENTS.md |
