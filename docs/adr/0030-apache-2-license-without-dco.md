---
status: accepted
date: 2026-11-04
deciders: [Brian Perret]
supersedes: []
extends: []
specs: ["010"]
constitution: [C10]
---

# ADR-0030 · Noodle is licensed under Apache-2.0, without a DCO

## Context and problem

The repository has no license, so its default is "all rights reserved": nobody may
reuse, fork or redistribute it, and the first external contribution (NoodleOpenCode,
#94) arrived with no stated terms. Brian wants a classic license that only asks to
credit the product and its author (#93, spec 010 clarifications 1 and 5).

## Decision drivers

- Attribution to "Noodle" and "Brian Perret" stays visible in redistributions.
- Contributions are clearly under the same terms, without paperwork for contributors.
- A permissive license that companies already accept.
- Bundled logos keep their owners' terms whatever the code license.

## Considered options

1. **Apache-2.0 with a NOTICE file, no DCO.**
2. **MIT, no DCO.**
3. **Apache-2.0 with a DCO sign-off on every commit.**

## Decision outcome

Chosen: **option 1**, because the NOTICE file carries the attribution Brian asked for
into every redistribution, section 5 of the license already puts contributions under the
same terms, and the explicit patent grant reassures companies. A DCO would add a
`Signed-off-by` step that section 5 makes redundant for a project of this size.

- `LICENSE` holds the Apache-2.0 text; `NOTICE` names Noodle, Brian Perret and
  `https://github.com/TheGostsniperfr/Noodle`.
- `.claude-plugin/plugin.json`, `flake.nix` (`lib.licenses.asl20`) and the skills'
  frontmatter declare `Apache-2.0`.
- Logos in `assets/icons/` are listed with their source and license in
  `assets/icons/ATTRIBUTION.md`; they remain trademarks of their owners.
- No part of an external contribution is merged before `LICENSE` is on `main`.

### Consequences

- Good: noodle may be used, forked and redistributed, with attribution preserved.
- Good: contributors need no sign-off or agreement beyond opening a PR.
- Bad: the license text is long, and every redistribution must ship `NOTICE`.
- Bad: without a DCO there is no per-commit statement of origin; review is the safeguard.

### Confirmation

Spec 010 feature PRJ-01: `jq -r .license .claude-plugin/plugin.json` and the skills'
`license:` print `Apache-2.0`, `LICENSE` holds `Version 2.0, January 2004`, `NOTICE`
names product, author and URL, and `go test ./assets/` fails for an icon without
attribution. PRJ-05 #8: `CONTRIBUTING.md` asks for no `Signed-off-by`.

## Pros and cons of the options

### Apache-2.0, no DCO

- Good: NOTICE-based attribution, contribution clause, patent grant.
- Bad: long text, NOTICE to carry.

### MIT, no DCO

- Good: shortest, most familiar.
- Bad: attribution is only the copyright line; no contribution clause or patent grant.

### Apache-2.0 with a DCO

- Good: a recorded statement of origin per commit.
- Bad: friction for every contributor, redundant with section 5 at this size.
