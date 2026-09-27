# ADR-0009 · A view's perspective is `type`, not `kind`

- Status: Accepted
- Date: 2026-09-27
- Supersedes: one line of ADR-0007 (View: `kind: topology | sequence`)

## Context

ADR-0007 gives every contract file a `kind` (`Model`, `View`, `Layout`) and also says a
view carries `kind: topology | sequence`. Both keys sit at the top level of a view file,
so YAML cannot hold them.

## Decision

`kind` names the file type everywhere. A view's perspective is `type: topology |
sequence`.

```yaml
apiVersion: noodle/v1alpha1
kind: View
type: sequence
```

## Consequences

The rest of ADR-0007 stands. The view schema and the loader use `type`.
