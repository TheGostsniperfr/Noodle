# Progress

Handoff between sessions. The newest entry comes first. Update it at the end of every
session: what changed, what is open, where the next session starts. Keep entries short;
decisions go in ADRs, acceptance in `features.json`.

## Start here

1. `./scripts/init.sh`: fix anything red before new work.
2. Read the current spec: [`specs/001-model-view-layout-split/spec.md`](../specs/001-model-view-layout-split/spec.md).
3. Next action: settle the spec's four open questions with Brian, then write ADR-0007.
   Use plan mode for this phase.

## Open threads outside this repo

| Where | What | Status |
|---|---|---|
| `~/Documents/aepita/ing2/3-Istor/cnp-docs` | CNP runtime specs and diagrams (`architecture/`), the old generator `tools/archgen` and skill `.claude/skills/cnp-diagram` | **uncommitted**. The generator and skill are superseded by this plugin. Proposal: keep `architecture/specs` + diagrams, delete `tools/archgen` and the local skill, render with the plugin. |
| CNP platform | Security gaps found while mapping: G1 no Cilium tenant isolation in code, G2 no groups-claim authorization in SecurityPolicy, G3 plain HTTP in cluster including Vault (`tls_disable`), G4 tunnels bound to Envoy's hashed Service name, G5 one tunnel per app on `cloudflared:latest`, G6 token call probably hairpins through Cloudflare | documented in the diagram only. Deserve issues in the CNP repos. G1 and G2 first. |
| `~/.dotfiles` | noodle plugin installed declaratively (commit `9bf7e57`) | pushed. `nixos-rebuild switch` still to run. |

## Log

### 2026-09-26 · Bootstrap

- Built the v0 generator and house style on the CNP platform: Tailwind palettes (dark and
  light), official logos, port badges, bridges, lint for readability and semantics.
- Iterated semantics with Brian's feedback into ADR-0003 (arrow = connection, requests
  only, ports on the server, egress/ingress sides) and ADR-0004 (topology vs sequence, one
  step per connection).
- Chose to build our own stack over LikeC4 (ADR-0001). IcePanel and Ilograph are the
  main feature inspirations (`docs/INSPIRATIONS.md`).
- Wrote the vision (levels L0 to L6), architecture with contract sketches, roadmap.
- Shipped as a Claude Code plugin + marketplace (repo root), with a wrapper that finds or
  builds the binary, icons compiled in, a Nix package. Install tested from GitHub.
- Adopted the session harness: this file, `docs/features.json`, `scripts/init.sh`, and
  spec-first work in `specs/`.
