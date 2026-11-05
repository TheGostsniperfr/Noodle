# Contributing to noodle

Thank you for helping. This page is the checklist a pull request is read against.
Code of conduct: [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md). Vulnerabilities:
[`SECURITY.md`](SECURITY.md).

## License of contributions

Noodle is Apache-2.0 ([`LICENSE`](LICENSE), [ADR-0030](docs/adr/0030-apache-2-license-without-dco.md)).
Opening a pull request puts your contribution under the same license, as section 5 of the
license says. There is no agreement to sign. Do not submit code you may not license this way.

## Start with an issue

Open an issue with the **Idea** or **Bug** template before a large change. The backlog is
the [project board](https://github.com/users/TheGostsniperfr/projects/5); an item has an id
`B-NN` that is also the key of its entry in [`docs/features.json`](docs/features.json).
A small fix needs no issue.

## The workflow

Work goes from a spec to verified code, as [`docs/PROCESS.md`](docs/PROCESS.md) describes:

1. **Specify.** A change that touches a contract, the connection semantics, a new
   subsystem or a new command starts from a spec made from `specs/_template/`, in full
   mode. A bug, a lint fix or a small task inside an existing spec uses quick mode.
2. **Decide.** A choice a future reader would ask "why?" about is an ADR from
   `docs/adr/_template.md`. Never rewrite an accepted ADR; supersede it.
3. **Implement** one task per commit, with tests: table-driven, with `testify`.
4. **Check.** `make check` must be green before every commit. It runs vet, tests, every
   example, the features gate and `make spec-check`. All commands are in
   [`docs/COMMANDS.md`](docs/COMMANDS.md); enter the dev shell with `nix develop`.
5. **Evaluate.** In full mode, `make evaluate FEATURE=<id>` has a fresh read-only agent
   grade the feature against the spec's `contract.md`. A feature's `passes` flips to
   `true` only once its evidence says PASS. Never edit or delete an entry of
   `docs/features.json`.

The invariants every change is checked against are in
[`docs/CONSTITUTION.md`](docs/CONSTITUTION.md). Two of them bite often: a lint rule is
never relaxed to make a diagram pass, and rendered files are generated, never edited by hand.

## Code

- Go: `gofmt`, `go vet`, errors wrapped with context, table-driven tests.
- Everything is in English: code, docs, diagrams.
- Dependencies are vendored. Run `go mod vendor` after touching `go.mod`.
- After editing `.claude-plugin/`, `skills/` or `agents/`, run `make plugin`.

## Commits and pull requests

- Commit messages follow Conventional Commits: `<type>(<scope>): <description>`, in the
  present tense, 72 characters at most. Types: `feat`, `fix`, `refactor`, `perf`, `test`,
  `docs`, `chore`, `ci`, `style`.
- Branch from `main`: `feat/<name>` or `fix/<name>`.
- Fill in the pull request template. Close the issue it implements with `Closes #NN`.

### Show the change

Noodle draws diagrams, so review is visual. A pull request that changes what noodle
draws puts a before/after in its description: the PNG from `main` next to the one from the
branch, cropped on the change, linked by commit SHA. Render the example that exercises
the change, or add a small one. A pull request with no visual change says so, and why.
