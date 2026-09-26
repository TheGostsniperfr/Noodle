# Working on noodle

## Session protocol

Every session starts without memory of the previous one. The repository carries the state.

**Start**
1. Run `./scripts/init.sh` and fix anything red before new work.
2. Read `docs/PROGRESS.md` ("Start here" and the latest log entry).
3. Pick the first feature with `"passes": false` in `docs/features.json`, within the
   current spec in `specs/`.

**Work**
- One feature at a time. New work on a phase starts from a spec: `specs/NNN-name/spec.md`
  (what and why), then `plan.md`, then `tasks.md`.
- Ask before choices the spec leaves open; record the answer in an ADR.

**End**
1. Flip `passes` to `true` only for features whose `verify` step you actually ran.
   Never edit or delete a feature to make it pass.
2. Add a dated entry at the top of the log in `docs/PROGRESS.md`, and update "Start here"
   and "Open threads".
3. Commit with a conventional message. Leave the tree clean.

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
`.claude/skills/noodle-diagram` is a symlink so sessions in this repo see the skill.
Dependencies are vendored: run `go mod vendor` after touching `go.mod`.

## Check before handing back

```bash
go vet ./... && go test ./...
go run ./cmd/noodle examples/cnp-runtime/runtime-request-path.yaml
claude plugin validate .
```

For diagram work, render the PNG and look at it. The lint catches geometry, not taste.
