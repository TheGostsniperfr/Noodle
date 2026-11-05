# Working on noodle

## Session protocol

Every session starts without memory of the previous one. The repository carries the state,
and hooks in `.claude/settings.json` keep it current.

**Start.** The `SessionStart` hook injects "Start here", the next features and the git state.
Run `make init` and fix anything red before new work. Every command is a `make` target,
listed with when to use it in [`docs/COMMANDS.md`](docs/COMMANDS.md).

**Work, one feature at a time, through [`docs/PROCESS.md`](docs/PROCESS.md)** (ADR-0024).
New work starts from a spec made from `specs/_template/`: `spec.md` in full or quick mode,
then `plan.md`, `tasks.md` and, in full mode, `contract.md`. Requirements are EARS with a
scenario; decisions are ADRs from `docs/adr/_template.md`; invariants are cited from
[`docs/CONSTITUTION.md`](docs/CONSTITUTION.md). Ask before choices the spec leaves open
(five questions at most per pass), and record the answer in an ADR. `make spec-check`
tells you what is missing.

**After each feature, not only at the end:**
1. commit the work, then `make evaluate FEATURE=<id>` (full mode) or write
   `evidence/<id>.md` yourself (quick mode); on NEEDS_WORK, add the findings as
   Convergence tasks and fix them;
2. flip its `passes` to `true` in `docs/features.json` only once the evidence says PASS;
   never edit or delete an entry: `make features`, a hook and CI reject it;
3. add a line under today's entry in `docs/PROGRESS.md`;
4. commit with a conventional message.

The `Stop` hook blocks ending a turn when files changed but `docs/PROGRESS.md` did not.
A one-line "in progress" note satisfies it.

**End.** When Brian stops ("je reprends demain", "on s'arrête là"), run the `wrap-up` skill.

## Backlog

Planning lives on the [GitHub project board](https://github.com/users/TheGostsniperfr/projects/5)
(Idea, To Do, In Progress, Done), one issue per item. Read it with
`gh project item-list 5 --owner TheGostsniperfr`. The board says what to work on next;
`docs/features.json` says what is verified. A backlog id `B-NN` names the same item in
both: put it in the issue body. Close the issue in the PR that flips its `passes`
(`Closes #NN`), and open an issue for any new backlog item or spec.

## Model guidance

The main session model is Brian's choice (`/model`). Subagents pin theirs.

| Work | Model |
|---|---|
| Specs, ADRs, architecture, open questions, anything that sets direction | Opus, high effort |
| Implementing tasks from a `tasks.md` that is already clear | Sonnet |
| Collecting a diagram's facts from code (skill step 1) | `noodle:fact-finder` agent (Sonnet) |
| Reviewing a rendered diagram as a tech lead | `noodle:diagram-reviewer` agent (Opus) |
| Grading a feature, reviewing a contract | `evaluator` agent (Opus, read-only), through `make evaluate` and `make review-contract` |
| Mechanical work: fetching and recolouring logos, variants | `noodle:icon-curator` agent (Haiku) |

## Read first

`docs/VISION.md`, `docs/ARCHITECTURE.md`, `docs/ROADMAP.md` and the ADRs in `docs/adr/`.
They are the source of truth for scope and design. The current phase is the first
unchecked one in the roadmap.

## Rules

- **Contracts are the boundary.** A module depends on schemas, never on another
  module's internals. Changing a contract needs a new ADR and a schema version bump.
- **Decisions go in ADRs.** A choice that a future reader would ask "why?" about gets an
  ADR. Never rewrite an accepted ADR; supersede it.
- **Stay in the current phase.** Do not build ahead of the roadmap without asking.
- **Semantics are fixed by ADR-0003 and ADR-0004.** Do not relax a lint rule to make a
  diagram pass; fix the diagram.
- Go: `gofmt`, `go vet`, errors wrapped with context, table-driven tests with testify.
- Everything in English: code, docs, diagrams.

## Layout

The repository root is both the Claude Code plugin (`.claude-plugin/plugin.json`,
`skills/`, `bin/`) and its marketplace (`.claude-plugin/marketplace.json`). `bin/noodle`
is the wrapper users run; `assets/icons/` is compiled into the binary.
`.claude/skills/noodle-diagram` is a symlink so sessions in this repo see the skill. `agents/` ships the
plugin's subagents. `.claude/skills/wrap-up` and the hooks are for working on this repo only.
Dependencies are vendored: run `go mod vendor` after touching `go.mod`.

## Check before handing back

```bash
make check     # vet, tests, every example, the features gate, spec-check
make plugin    # after touching .claude-plugin/, skills/ or agents/
```

For diagram work, render the PNG and look at it. The lint catches geometry, not taste.

## Show it in the PR

noodle is a drawing tool, so review is visual. A PR that changes what noodle draws puts
a before/after in its description (`.github/pull_request_template.md`): the PNG from
`main` next to the one from the branch, cropped on the change, linked by commit SHA
(`https://github.com/TheGostsniperfr/Noodle/blob/<sha>/<path>.png?raw=true`). Render
the example that exercises the change, or add one. A PR with no visual change says so.
A design decision with no rendering yet (an ADR, a plan) can be illustrated by a noodle
diagram in `docs/diagrams/`.
