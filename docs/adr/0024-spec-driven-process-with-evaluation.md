---
status: accepted
date: 2026-11-04
deciders: [Brian Perret]
supersedes: []
extends: []
specs: ["007"]
constitution: [C10, C12, C14]
---

# ADR-0024 · Specs are traceable and features pass on an independent evaluator's evidence

## Context and problem

Seventeen new issues (#16, #18, #93 to #107) need specs and ADRs before code. Most
work is done by agents. Two failure modes are known from this repository and from the
NoodleOpenCode fork: an agent declares done when its code compiles, and agents grade
their own work generously ("agent completion is not proof", the fork's
`docs/agent-lessons.md`). Specs 001 to 006 have prose requirements that no tool can
trace to a test, ADRs that record one option without the alternatives, and commands
spread over six scripts called from four places.

How do we make an issue become verified code with as little trust in any single agent
as possible, without a cost that kills small changes?

## Decision drivers

- A feature is verified by someone other than its author (Anthropic, "Harness design
  for long-running application development", 2026: a separate, tuned evaluator is far
  more tractable than self-critique).
- Checks that need no model run on every commit, at no token cost (C08).
- Agents find every command in one place (C14).
- A bug fix stays cheap.
- Old specs and ADRs are not rewritten (C10).

## Considered options

1. **Adopt GitHub Spec Kit** as is: its `/speckit.*` commands and directory layout.
2. **Adopt OpenSpec**: change folders with spec deltas merged into living specs.
3. **Keep the repository's spec, plan, tasks shape and add what the tools do best**:
   EARS requirements (Kiro), a constitution, clarify, analyze and converge steps
   (Spec Kit), MADR 4 ADRs with frontmatter, sprint contracts and a fresh-context
   evaluator with a default-fail gate (Anthropic), all enforced by `make` targets and a
   Go checker.
4. **No new process**: keep prose specs and self-verification.

## Decision outcome

Chosen: **option 3**, because it keeps six specs and twenty-three ADRs valid as they
are, adds the two gates that address the known failure modes (independent evaluator,
evidence before a flip), and enforces traceability with code instead of convention,
which neither Spec Kit nor OpenSpec does. Process: `docs/PROCESS.md`.

- Specs declare `Mode: full` or `Mode: quick`. Full mode requires EARS requirements
  with scenarios, `covers` on features, a contract and evaluator evidence. Quick mode
  accepts evidence written by the implementer. Specs without a mode are legacy.
- ADRs from 0024 on use `docs/adr/_template.md`; the index in `docs/adr/README.md` is
  generated.
- Invariants are numbered in `docs/CONSTITUTION.md` and cited as `Cnn`.
- `tools/speccheck` implements every check; `Makefile` is the only entry point;
  `docs/COMMANDS.md` documents it.

### Consequences

- Good: a flipped feature always has a file showing the verify output and who graded it.
- Good: spec, tasks, features and contract cannot silently disagree.
- Bad: a full-mode feature costs one or two extra agent runs (contract review, evaluation).
- Bad: the evaluator is still a model and can be lenient; its prompt needs tuning from
  its logs, as the Anthropic post reports.
- Bad: "read-only" means no Write or Edit tool and an allowlist of pinned commands. Two
  patterns can still write: `go test` (`-o`, `-coverprofile`) and `git diff/log/show`
  (`--output`), kept because verify steps need them. A sandboxed evaluator is left for later.
  Spec 010 adds one exact entry, `make release-snapshot`, for the release dry run of its
  contract: it runs goreleaser, builds from the working tree and writes only `dist/`, which
  is git-ignored. Through `nix develop` it may fetch tools from the Nix cache.
- Bad: the gate trusts the `- Evaluator:` line as text; anyone can write it by hand. The
  mitigation is review: an evaluation commit holds only its evidence file.

### Confirmation

- `make spec-check` (traceability, ADR shape, index, command reference) and
  `make features` (history and evidence) run in `make check` and in CI.
- A `PostToolUse` hook runs `make features` when `docs/features.json` is edited.
- Spec 007 itself is in full mode and its features were graded by the evaluator.

## Pros and cons of the options

### Spec Kit

- Good: widely used, agent-agnostic, rich command set.
- Bad: enforcement is convention only; its layout would replace ours and orphan specs 001 to 006.

### OpenSpec

- Good: living specs merged from deltas, good for a long-lived product spec.
- Bad: drift between deltas and living specs is managed by hand; a second layout to learn.

### Own shape, best pieces

- Good: incremental, enforced by code, fits the existing hooks and `features.json`.
- Bad: we maintain the checker and the templates ourselves.

### No new process

- Good: free.
- Bad: keeps both known failure modes.
