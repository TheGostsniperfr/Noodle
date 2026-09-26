# noodle

Take the spaghetti out of a large architecture and keep it clear.

Noodle turns an architecture described as code into readable, presentation-grade
diagrams, and aims further: interactive views, guided tours, visual PR diffs, discovery
from real clusters, live overlays, and a model AI agents can query.
See [the vision](docs/VISION.md).

![CNP runtime example](examples/cnp-runtime/runtime-request-path.dark.png)

## Status

**v0.** One generator (`cmd/noodle`) turns a single YAML spec into draw.io in a dark or
light theme. It refuses to emit a diagram whose lines cross boxes, whose labels collide,
or whose arrows break the connection semantics. The next phase splits the spec into
model, views and layout ([roadmap](docs/ROADMAP.md)).

## Quick start

```bash
nix develop
go run ./cmd/noodle -icons .claude/skills/noodle-diagram/icons \
  -theme dark -o /tmp/runtime.drawio examples/cnp-runtime/runtime-request-path.yaml
drawio -x -f png -s 2 --border 20 -o /tmp/runtime.png /tmp/runtime.drawio
```

Without `-o`, noodle only lints.

## Docs

- [Vision](docs/VISION.md): the problem, principles and capability levels L0 to L6
- [Architecture](docs/ARCHITECTURE.md): modules, contracts, interfaces, storage and security
- [Roadmap](docs/ROADMAP.md): phases and their exit criteria
- [Decisions](docs/adr/README.md): ADRs
- [Inspirations](docs/INSPIRATIONS.md): IcePanel, Ilograph, LikeC4 and others, and what we borrow
- [Diagram skill](.claude/skills/noodle-diagram/SKILL.md): the conventions an AI agent follows to write diagrams
