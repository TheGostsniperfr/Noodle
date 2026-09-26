# Working on noodle

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
