---
name: release-notes
description: Write the release notes of a noodle release from its changelog section. Picks the highlights, reuses the before/after images of the pull requests that changed what noodle draws, and lists breaking changes with a link to their migration. Use when preparing or publishing a noodle release, or when asked for release notes or highlights.
license: Apache-2.0
---

# Release notes for a noodle release

The changelog is generated from conventional commits and is the source of facts. The notes
are what a user reads first: a few highlights, then the full list.

## Steps

1. **Get the section.** `make changelog` regenerates `CHANGELOG.md`. For one release, the
   top section is the one to use; `git-cliff --latest --strip header` prints it alone.
2. **Pick three to five highlights** among the Features and Bug fixes: what a user can do
   now that they could not, in plain sentences. Skip chores, CI and refactoring.
3. **Reuse the pictures.** For a highlight that changed what noodle draws, find its pull
   request (the `(#NN)` in the changelog line) and take its before/after table:

   ```bash
   gh pr view NN --json body --jq .body
   ```

   Copy the image links as they are. They are pinned by commit SHA, so they stay valid. Do
   not render new images for a release.
4. **List breaking changes** under their own heading. A breaking change is a commit marked
   `!`, a schema version bump, or an ADR that supersedes another. For each one say what
   breaks and link the migration: the ADR, the docs page, or the command to run, such as
   `noodle migrate`. A breaking change with no migration is a finding: report it before
   releasing.
5. **Assemble:** highlights with their images, breaking changes with migration links, then
   the full changelog section. Name the version from `noodle version` or the tag.
6. **Show the draft and stop.** Publishing, tagging and editing the GitHub release are done
   by a person.
