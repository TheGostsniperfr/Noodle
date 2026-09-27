# Plan 001 · Split model, views and layout

Spec: [`spec.md`](spec.md). Decisions: ADR-0007 (files, references, geometry), ADR-0008
(status, multiplicity).

## Approach

Keep the renderer and the geometry lint of `cmd/noodle` working as they are, and put a
**resolver** in front of them: model + view + layout in, the v0 `Spec` geometry out.
Phase 1 then changes inputs, not drawing. The SVG renderer (phase 2) will consume the
resolved geometry too, so the v0 `Spec` type becomes an internal "resolved diagram"
rather than a file format.

```
system dir ──load──▶ Model, View, Layout ──check refs──▶ resolve ──▶ Diagram (v0 Spec) ──▶ lint geometry ──▶ render
                         ▲ JSON Schema (tests)                           ▲ sequence views: computed, no layout
```

## Packages

| Package | Content | Depends on |
|---|---|---|
| `internal/model` | `Model`, `View`, `Layout` types; `LoadSystem(dir)`; reference checks (FR-005) | `gopkg.in/yaml.v3` |
| `schemas/v1alpha1/` | `model.schema.json`, `view.schema.json`, `layout.schema.json`, hand-written | none |
| `internal/diagram` | the resolved geometry, today's `Spec`, `Zone`, `Node`, `Edge`, `Card`, `Note`, `Rect`, `Point` moved out of `cmd/noodle/spec.go` | none |
| `internal/resolve` | topology: relative positions → absolute, `element.side@ratio` → point, `lane:` → coordinate, steps → `[n]` label prefixes, annotations → badges, `denied` → blocked style; sequence: computed columns and rows | `model`, `diagram` |
| `internal/lint` | one rule per file, moved from `cmd/noodle/lint.go` with the same messages (FR-006) | `diagram` |
| `cmd/noodle` | CLI and draw.io renderer | all of the above |

Contracts first: `internal/lint` and the renderer depend on `internal/diagram` only,
never on `internal/model`. That keeps the rule "modules depend on contracts".

## Key choices left to the tasks

- **Schema validator** for tests: `github.com/santhosh-tekuri/jsonschema/v6` (pure Go,
  draft 2020-12). Vendored. Schemas are the published contract; Go types are checked
  against them by validating every example file in `go test`.
- **Endpoint syntax.** `id.side` or `id.side@NN%`, side in `left|right|top|bottom`,
  ratio default 50 %. Parsed in `internal/resolve`, errors carry file and edge id.
- **Label text.** Default `verb · protocol` from the model connection (ADR-0003); a
  view override replaces it; the step prefix `[n]` is added by the resolver.
- **Ports.** v0 `port: "TCP 443"` becomes a named port on the target element; the
  resolver renders the badge text from protocol and number.

## Migration

1. CNP example: `examples/cnp-runtime/model.yaml`, `views/runtime.yaml`,
   `layouts/runtime.yaml` (P1-03 verify step names `--view runtime`). Check: the
   resolved draw.io matches the v0 output, PNG diffed with `magick compare`.
2. OIDC sequence: `views/oidc-login.yaml`, no layout (P1-06).
3. Platform-regression fixtures migrate after CNP, one directory each.
4. Remove the v0 single-file reader and the old example file (FR-009).

## Risks

- **Pixel parity with v0.** Relative positions and symbolic endpoints must round-trip
  to the same absolute points. Mitigation: migrate CNP with absolute-equivalent values
  first, prove parity, then rewrite to symbolic endpoints and re-check.
- **Sequence renderer scope.** Keep it to lifelines, numbered arrows, dashed returns
  (FR-008). Alternates and parallel blocks are phase 4.
