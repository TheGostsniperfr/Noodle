---
name: contrib-check
description: Run the checklist of CONTRIBUTING.md on a noodle branch or pull request and report what is missing. Checks make check, conventional commits, the spec, contract and ADR a change needs, the features.json history, and the before/after images for a visual change. Use before opening or approving a pull request on the noodle repository.
license: Apache-2.0
---

# Checking a branch against CONTRIBUTING.md

Read `CONTRIBUTING.md` first: it is the checklist, and this skill runs it. Report each item
as pass, fail or not applicable, with the file or command behind it. Change nothing.

## Steps

1. **Scope.** `git diff --stat origin/main...HEAD` shows what the branch touches. A change to
   a contract, the connection semantics, a new subsystem or a new command needs a full-mode
   spec under `specs/`; a small fix does not.
2. **`make check`** must pass. Run it and quote the first error if it does not. After a
   change to `.claude-plugin/`, `skills/` or `agents/`, run `make plugin` too.
3. **Commits.** `git log --format=%s origin/main..HEAD`: each subject is a conventional
   commit, `<type>(<scope>): <description>`, in the present tense, 72 characters at most.
4. **Spec and decisions.** A full-mode change has `spec.md`, `plan.md`, `tasks.md` and
   `contract.md` that `make spec-check` accepts. A choice a future reader would ask "why?"
   about has an ADR, and no accepted ADR was rewritten.
5. **Features.** `make features BASE=origin/main` passes: no entry of `docs/features.json`
   was edited or removed, and every `passes: true` has its evidence.
6. **Show the change.** If the diff changes what noodle draws, the pull request body has a
   before/after table with images linked by commit SHA. If not, it says so and why.
7. **Code.** Go is formatted and vetted, errors are wrapped with context, tests are
   table-driven. A lint rule was never relaxed to make a diagram pass. Dependencies are
   vendored after a `go.mod` change.
8. **Verdict.** List failures first, then what is not applicable. Do not fix anything unless
   asked.
