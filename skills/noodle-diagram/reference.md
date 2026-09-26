# Spec reference (v0)

One YAML file per diagram. Coordinates are absolute pixels, origin top-left, y down.
Unknown fields are rejected. A full working example is at the end.

## Top level

| Field | Type | Notes |
|---|---|---|
| `id` | string | stable, used in draw.io page ids |
| `title` | string | one line, `Project · Topic · Question` |
| `subtitle` | string | the one question this diagram answers |
| `meta` | string[] | scope, out of scope, version and source path, one line each |
| `width`, `height` | number | canvas size; leave ~40 px margins |
| `zones` | Zone[] | drawn first, in order |
| `nodes` | Node[] | |
| `edges` | Edge[] | |
| `notes` | Note[] | free text next to a component, e.g. a response that changes the path |
| `cards` | Card[] | legend, steps and gaps under the diagram |

## Zone

| Field | Notes |
|---|---|
| `id` | |
| `kind` | `region` (infra or trust perimeter, dashed 8·4) or `group` (functional category, dashed 4·4) |
| `label` | uppercase, role first: `GATEWAY API`, `IDENTITY` |
| `sub` | small grey text after the label, e.g. the namespace |
| `color` | `cyan` `emerald` `violet` `amber` `rose` `orange` `slate` `indigo` `sky`; siblings differ |
| `icon` | environment on regions (`globe`, `rack`, a provider logo), product or `k8s-ns` on groups |
| `x` `y` `w` `h` | a zone fully contains its nodes and child zones |

## Node

| Field | Notes |
|---|---|
| `id` | semantic and stable, never renamed for cosmetics |
| `kind` | `frontend` `backend` `database` `cloud` `security` `bus` `external` |
| `shape` | `box` (default) · `cylinder` for a datastore · `actor` for a human (icon with label below) |
| `icon` | see `noodle -list-icons`; project icons via `-icons DIR` |
| `title` | bold name |
| `tech` | `[Technology · variant]`, shown in italics |
| `desc` | one line: the responsibility |
| `badge` | e.g. `G2`: links to an entry of the gaps card |
| `x` `y` `w` `h` | typical box 220–320 × 72–120; actor 56 × 56 |

Text width is estimated at 0.6 em per character: 7.2 px at 12 px (title), 6 px at 10 px
(tech, desc). Available width is `w − 58` with an icon. The lint reports any overflow.

## Edge

| Field | Notes |
|---|---|
| `id` | `e-<step>-<from>-<to>` for connections, `r-…` for references, `x-…` for blocked |
| `from`, `to` | node or zone ids; `from` opens the connection |
| `kind` | `flow` · `auth` · `tunnel` · `async` · `blocked` · `link` (reference, not traffic) |
| `label` | `[step] verb · protocol`; `<br>` for a second line; `[1]` numbers, `[A]` letters |
| `port` | listening port on `to`, e.g. `TCP 443`, `UDP 7844`; drawn as a badge on its border |
| `against_flow` | `true` only for outbound connections running against the reading direction |
| `path` | `[[x, y], …]` orthogonal; first point on the border of `from`, last on the border of `to` |
| `label_at` | `[x, y]` on the drawn path; default is the middle of the longest segment |
| `label_offset` | `[dx, dy]` to push a long label beside a short segment |

Side rule, except `link` and `against_flow`: the first point is on the right or bottom
border of `from`, the last on the left or top border of `to`. With a `port`, the drawn
path stops at the outer edge of the badge; `label_at` must lie on that drawn path.

## Note and Card

```yaml
notes:
  - {id: n-302, text: "Without a session, Envoy replies **302** …", x: 52, y: 590, w: 196, h: 90}
cards:
  - {id: card-legend, title: "Legend", color: slate, legend: true, x: 40, y: 1340, w: 540, h: 420}
  - id: card-steps
    title: "Request path · data plane"
    color: cyan
    x: 610
    y: 1340
    w: 620
    h: 420
    lines:
      - "[1] Browser sends GET … Cloudflare terminates TLS."
```

Inline markup everywhere text is shown: `[1]` filled step badge, `[A]` outlined step
badge, `**bold**`, `!!warning!!` in orange.

## Minimal example

```yaml
id: hello
title: "Demo · Web app"
subtitle: "How a request reaches the database."
meta: ["Scope: demo", "v0.1"]
width: 1100
height: 520
zones:
  - {id: z-internet, kind: region, color: slate, icon: globe, label: "INTERNET", x: 40, y: 40, w: 220, h: 300}
  - {id: z-cluster, kind: region, color: cyan, icon: kubernetes, label: "CLUSTER", x: 320, y: 40, w: 740, h: 300}
nodes:
  - {id: user, kind: frontend, shape: actor, icon: user, title: "User", desc: "browser", x: 122, y: 150, w: 56, h: 56}
  - {id: api, kind: backend, icon: k8s-deploy, title: "API", tech: "[Deployment]", desc: "Serves requests", x: 380, y: 140, w: 240, h: 80}
  - {id: db, kind: database, shape: cylinder, icon: cnpg, title: "Postgres", tech: "[CloudNativePG]", desc: "App data", x: 760, y: 130, w: 240, h: 100}
edges:
  - {id: e-1-user-api, kind: flow, from: user, to: api, port: "TCP 443", label: "[1] GET · HTTPS", path: [[178, 200], [380, 200]]}
  - {id: e-2-api-db, kind: flow, from: api, to: db, port: "TCP 5432", label: "[2] query · SQL", path: [[620, 200], [760, 200]]}
cards:
  - {id: card-legend, title: "Legend", color: slate, legend: true, x: 40, y: 380, w: 1020, h: 120}
```

The CNP runtime example in the noodle repository (`examples/cnp-runtime/`) shows every
feature at scale.
