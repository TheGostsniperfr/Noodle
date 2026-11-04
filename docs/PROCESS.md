# Noodle · Process

How an issue becomes verified code (ADR-0024). Every step writes a file, and every gate
is a script or an independent evaluator, never the word of the agent that did the work.
Commands are in [`COMMANDS.md`](COMMANDS.md); invariants in [`CONSTITUTION.md`](CONSTITUTION.md).

## Two modes

| Mode | When | Steps |
|---|---|---|
| **full** | touches a contract, the semantics (C02, C03), a new subsystem or a new command | all of the steps below |
| **quick** | a bug, a lint fix, a size S task inside an existing spec | spec in a few lines, tasks, implement, evidence written by the implementer |

The spec says which: `- Mode: full` or `- Mode: quick`. Specs written before ADR-0024
have no mode and keep their old shape.

## Full mode

| # | Step | Writes | Who | Gate |
|---|---|---|---|---|
| 1 | **Specify** | `specs/NNN-name/spec.md` from [`_template`](../specs/_template/spec.md) | Opus, high effort | requirements in EARS, each with a scenario |
| 2 | **Clarify** | answers in `spec.md` → "Clarifications" | Brian | at most five questions per pass; a direction-setting answer becomes an ADR |
| 3 | **Decide** | `docs/adr/NNNN-*.md` from [`_template`](adr/_template.md) | Opus, Brian accepts | two options or more, a Confirmation section |
| 4 | **Plan and tasks** | `plan.md`, `tasks.md`, new entries in `features.json` with `covers` | Opus | tasks in waves; `make spec-check` green |
| 5 | **Contract** | `contract.md`: per feature, criteria, how each is verified, its threshold | implementer drafts, `make review-contract` | APPROVED, or three rounds done (see below) |
| 6 | **Implement** | code, tests, one commit per task | Sonnet when the tasks are clear | `make check` green |
| 7 | **Evaluate** | `evidence/<feature>.md` | `make evaluate FEATURE=…`, a fresh read-only agent | `- Verdict: PASS`; anything else goes to step 8 |
| 8 | **Converge** | gap tasks under "Convergence" in `tasks.md` | implementer | back to 6 until the evaluator passes |
| 9 | **Flip** | `passes: true`, PROGRESS line, commit | implementer | `make features` refuses a flip without its evidence |

### Requirements in EARS

One requirement, one behaviour, one way to test it:

```markdown
- **FR-002** WHEN a curated element `matches` an id that no fragment contains, THE SYSTEM
  SHALL report it as `gone`, with the file and line of the element.
  *Scenario:* WHEN the Deployment `api` is removed from the manifests THEN `noodle drift`
  lists `api` under gone, citing `model.yaml:12`.
```

Patterns: `THE SYSTEM SHALL …` (always), `WHEN <event> …`, `WHILE <state> …`,
`IF <unwanted condition> THEN …`, `WHERE <option is set> …`.
A requirement that states a property for all inputs ("for every valid model, …") gets a
property-based test (`testing/quick`, or `rapid` once vendored).

### Traceability

`make spec-check` follows each thread with no model in the loop:

```
FR-NNN in spec.md ── named by a task ── covered by a feature (covers) ── criteria in contract.md
                                              └── evidence/<id>.md with Verdict: PASS (make features)
```

It also checks that every cited ADR and `Cnn` exists, that the ADR index is current,
and that `COMMANDS.md` lists exactly the Makefile targets.

### Contract rounds

At most **three** review rounds per contract. Each round's CHANGES are applied before the
next. After the third, only two kinds of point are still applied: one that lets the
evaluator write, and one that makes a criterion pass or fail whatever the code does.
Others are recorded in the contract as follow-ups, and implementation starts. The
evaluation of each feature reruns every command anyway, so a weak criterion still meets
a skeptical reader.

### The evaluator

`.claude/agents/evaluator.md`: a fresh process, no Write or Edit, Opus. It reruns the
verify step itself, grades each contract criterion against its threshold, reads the
diff for scope creep and tests that assert nothing, looks at PNGs for visual features,
and answers PASS or NEEDS_WORK. One failed criterion fails the feature. It is told to
be skeptical: the agent that wrote the code is the one most likely to be generous.

## Delegating a batch to an implementer

A batch (ADR-0025) is implemented by one fresh Sonnet agent in its own worktree, never
by a fork of the main session: a fork copies the whole conversation and costs about
600k tokens per run. The agent gets a short, self-contained prompt with:

1. **A closed scope**: the worktree path, the branch, the tasks to do and the ones to
   leave (anything that needs a human answer or publishes).
2. **A read order**: AGENTS.md, PROCESS.md, COMMANDS.md, CONSTITUTION.md, then the
   spec, plan, tasks and contract, follow-ups included, and the batch's ADRs.
3. **A done that can be measured**: one task per commit, `make check` green before each.
4. **Explicit don'ts**: no relaxed lint rule (C04), no edit of an existing
   `features.json` entry, no `ALLOW_FEATURE_EDIT`, no flip, no evaluation, no push.
5. **Stop, don't decide**: an ambiguity is reported, and independent tasks go on.
6. **No self-grading**: the main session runs `make evaluate` for every feature in
   parallel, and sends NEEDS_WORK findings back to the same agent for convergence.

The agent reports commits per task, tasks left and why, and ambiguities found.

## Quick mode

Spec of a few lines with its mode, `tasks.md`, implement, then write
`evidence/<feature>.md` yourself with the verify output and `- Evaluator: self`.
`make features` accepts it because the spec is in quick mode.

## Keeping the harness honest

Each piece of this process exists because a model, today, does worse without it
(Anthropic, "Harness design for long-running application development", 2026). When a
new model arrives, run one spec without the contract step and one without the
evaluator; drop a piece that no longer changes the outcome, and record it in an ADR.
