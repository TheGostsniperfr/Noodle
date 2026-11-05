---
status: accepted
date: 2026-11-05
deciders: [Brian Perret]
supersedes: []
extends: ["0001"]
specs: ["011"]
constitution: [C06]
---

# ADR-0035 · noodle rasterizes PNGs itself with golang.org/x/image, with no external tool

## Context and problem

Question 3 of spec 011. Every PNG today comes from draw.io desktop under xvfb: 20 to 40 s
per export (#82), nothing in a Claude cloud session (#91), and cropped images past about
4000 px (#71). ADR-0027 accepted one external tool, `kustomize`, for discovery only.

## Decision drivers

- `noodle` alone renders a PNG, on any machine the release binary runs on (#91).
- Exports in well under a second, whatever the width.
- The PNG and the SVG show the same scene (ADR-0033).
- Few dependencies, all vendored.

## Considered options

1. **In the binary**: the scene drawn to an `image.RGBA` with `golang.org/x/image/vector`
   for fills, a stroker written in noodle for lines and dashes, and
   `golang.org/x/image/font/opentype` for text. One new vendored module.
2. **`resvg` as an external tool**, called on the SVG the way `kustomize` is.
3. **Both**: the Go rasterizer by default, `resvg` when it is on the PATH.

## Decision outcome

Chosen: **option 1**, because it is the only option where the release binary needs
nothing else.

- `golang.org/x/image` is maintained by the Go team, like the vendored `golang.org/x/text`.
- `x/image/vector` fills paths with anti-aliasing but does not stroke them, so noodle
  turns polylines, rounded corners and dashes into polygons itself, in
  `internal/render/png`.
- Icons are SVG: the rasterizer draws the subset of SVG the curated icons use (paths,
  basic shapes, fills, transforms) and `go test` fails an icon outside that subset.

### Consequences

- Good: one binary; `make examples` no longer needs draw.io; CI renders PNGs.
- Good: no width limit beyond memory.
- Bad: anti-aliasing is less refined than resvg's, and a stroker and a small SVG icon
  reader are code noodle now owns.
- Bad: a new icon using an SVG feature outside the subset needs code before it renders.

### Confirmation

`go list -deps ./cmd/noodle` lists no cgo package. A test renders every curated icon in
both variants to PNG and fails on an unsupported element. `make examples` runs with
draw.io absent from the PATH.

## Pros and cons of the options

### Option 1

- Good: self-contained, fast.
- Bad: rasterizer code to own.

### Option 2

- Good: best quality, full SVG support.
- Bad: one more install, the same problem #91 reported.

### Option 3

- Good: quality when available.
- Bad: two code paths, and PNGs that differ by machine, which breaks golden images.
