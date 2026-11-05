# Plan 007 · Engineering process

Spec: [`spec.md`](spec.md). Decisions: ADR-0024.

## Approach

One Go command, `tools/speccheck`, implements every check, so the rules are tested like
the rest of the code and need no `jq`. The `Makefile` wraps it and every other command;
`docs/COMMANDS.md` documents the targets and is itself checked against the Makefile.
The evaluator is a project subagent run with `claude -p --agent evaluator`, write tools
disallowed, its stdout saved as the evidence file.

```
make spec-check ─▶ speccheck all ─┬─ spec: Cnn, ADR refs, flips; full mode: EARS, tasks, covers, contract
                                   ├─ adr: frontmatter shape, generated index
                                   └─ commands: Makefile ⇄ COMMANDS.md
make features   ─▶ speccheck features -base REF: history, evidence for each new pass
make evaluate   ─▶ claude -p --agent evaluator ─▶ <spec>/evidence/<id>.md
```

## Packages and files

| Path | Change |
|---|---|
| `tools/speccheck` | new: `spec`, `adr`, `commands`, `features`, `locate`, `all` |
| `Makefile`, `docs/COMMANDS.md` | new; replace `scripts/check-features.sh`; `scripts/init.sh` becomes `exec make init`, since immutable verify steps cite it |
| `docs/CONSTITUTION.md`, `docs/PROCESS.md` | new |
| `specs/_template/`, `docs/adr/_template.md` | new |
| `.claude/agents/evaluator.md` | new, repository-only (not shipped in `agents/`) |
| `.claude/settings.json` | `PostToolUse` runs `make features` on `docs/features.json` |
| `.github/workflows/ci.yml` | steps call `make` |
| `AGENTS.md`, `README.md`, `docs/PROGRESS.md`, `scripts/session-context.sh` | point to `make` and the process |
| `flake.nix` | `gnumake` in the dev shell |

## Alternatives set aside

A Bash checker over `jq` (the current `check-features.sh`): untested, and a second
language for logic that grows. A Stop hook running the evaluator after every turn: too
costly for turns that change no feature.

## Risks

| Risk | Mitigation |
|---|---|
| The evaluator passes weak work | skeptical prompt with calibration cases; tune it from its reports |
| `claude -p` unavailable in CI | evaluation is a local step; CI checks only that the evidence file exists |
| The gate blocks legacy flips | only flips relative to the base ref are gated |
