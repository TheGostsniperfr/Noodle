# ADR-0005 · Git is the source of truth; artifacts are generated

- Status: Accepted
- Date: 2026-09-26

## Context

Diagrams must be reviewable, diffable and reproducible, and an architecture map is
sensitive: it exposes hostnames, ports and known security gaps.

## Decision

- Models, views and layouts live in Git and change through pull requests.
- SVG, PNG, draw.io and viewer bundles are generated in CI. They are not committed,
  except as example screenshots in this repository.
- Previews are static builds behind access control. Public exports go through a
  redaction profile.
- Live data (L5 onwards) is not stored in Git. Its store is a future ADR.

## Consequences

- Hand edits in draw.io are for throwaway variants only.
- Stable, semantic IDs are mandatory: diffs match elements by ID.
