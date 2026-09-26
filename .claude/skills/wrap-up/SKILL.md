---
name: wrap-up
description: End-of-session handoff for the noodle repository. Use when Brian says he is stopping or will resume later ("je reprends demain", "on s'arrête là", "fin de session", "wrap up", "on continue plus tard"), or asks to save where things stand.
---

# Wrap up the session

The next session starts with no memory of this one. Leave the repository so that
`scripts/session-context.sh` and `docs/PROGRESS.md` are enough to resume.

1. Run `./scripts/init.sh`. If something is red, fix it or record it as the first item of
   "Start here".
2. `docs/features.json`: flip `passes` to `true` only for features whose `verify` step
   you ran in this session. Add new entries for work discovered; never edit or delete
   existing ones.
3. `docs/PROGRESS.md`:
   - **Start here**: the exact next action, specific enough to act on without rereading
     the log, e.g. "answer the 4 open questions of spec 001", not "continue phase 1".
   - **Open threads**: add, update or remove rows, including things outside this repo.
   - **Log**: a dated entry at the top with what changed, decisions (link the ADR), and
     what was tried and abandoned, with the reason.
4. A decision a future reader would question and that has no ADR yet: write it.
5. Commit with a conventional message. The tree must be clean. Ask Brian before pushing.
6. Tell Brian in three lines: what was done, what is next, anything he must do himself.
