# ADR-0002 · Separate model, views and layout

- Status: Accepted
- Date: 2026-09-26

## Context

The v0 spec mixes what exists (components, connections), what is shown (which path,
which numbers) and where it is drawn (coordinates). That is why the CNP runtime diagram
could show only one path. It also blocks the other goals: an AI cannot read the
architecture without layout noise, discovery cannot produce coordinates, and a PR diff
has to compare drawings instead of facts.

## Decision

Three contracts, three kinds of files:

- **Model**: elements, connections, references, ports, zones, annotations. No coordinates.
- **View**: a selection of the model, scenarios and flows, tags, tours, redaction.
- **Layout**: per view, positions and routes. Pinned values survive auto-layout.

Schemas are versioned (`noodle/v1alpha1`) and published as JSON Schema.

## Consequences

- Any number of views on one model; the AI and the MCP server read the model only.
- Discovery adapters produce model fragments; layout is computed afterwards.
- The v0 format stays readable by the CLI until the example is migrated, then is removed.
