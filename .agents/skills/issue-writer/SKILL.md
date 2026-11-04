---
name: issue-writer
description: Turn a raw idea or a bug report about noodle into a GitHub issue that follows the idea.yml or bug.yml template, with the next free backlog id B-NN. Use when asked to write, draft or file an issue, a ticket or a backlog item for the noodle repository. Drafts first and never creates the issue without approval.
license: Apache-2.0
---

# Writing a noodle issue

The issue form is `.github/ISSUE_TEMPLATE/idea.yml` for work to do and
`.github/ISSUE_TEMPLATE/bug.yml` for something wrong. Read the one you need first: its
`label:` lines are the headings of the issue, in that order.

## Steps

1. **Decide the kind.** A bug has a wrong output to reproduce. Everything else is an idea.
2. **Find the next free `B-NN`** for an idea. Take the highest of both sources and add one:

   ```bash
   jq -r '.backlog[].id' docs/features.json | sort -V | tail -1
   gh issue list --state all --limit 300 --json body --jq '.[].body' | grep -o 'B-[0-9]*' | sort -V | tail -1
   ```

   The id is in the issue body, under "Backlog id". A bug has none.
3. **Write the headings of `idea.yml`**, each short and concrete:
   - **Context**: what is hard or missing today, and who feels it.
   - **Proposal**: what to do, in a few sentences.
   - **Scope**: what is in, as a short list.
   - **Out of scope**: what this does not cover, so the issue stays small.
   - **Acceptance**: one checkable statement per line, each with the command or file that
     proves it.
   - **Related**: issues, ADRs, specs, pull requests.
   - **Backlog id**: the `B-NN` from step 2.

   For a bug, use the headings of `bug.yml`: What happens, Reproduce, Expected, Version.
   Reproduce is the smallest system directory and the exact `noodle` command.
4. **Check for a duplicate** with `gh issue list --search "<keywords>" --state all`.
5. **Show the draft and stop.** Create the issue with `gh issue create` only once the
   person has approved the text. Put the issue on the project board only if asked.

## Style

Plain sentences, in English, one idea per issue. If the acceptance cannot be checked by a
command or by looking at a file, the issue is not ready: say what is missing.
