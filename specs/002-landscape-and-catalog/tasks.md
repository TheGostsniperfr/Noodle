# Tasks 002 · Landscape and catalog views

Plan: [`plan.md`](plan.md). One task per commit. Each ends with `go vet ./... && go test ./...`
and `./scripts/init.sh` green. Tick the box in the same commit. Rendering tasks render the
example PNG and look at it.

## A · Planned status in the core (P1b-01, P1b-02)

- [x] **T01** `status` and `target` flow through resolve into `diagram.Node` and
  `diagram.Zone`, zone values inherited by children that set none. Check: `target`
  without `planned` fails. Schemas updated.
- [x] **T02** Render `planned` (hatch, dashed, 60 % text, 40 % icon), `deprecated`
  (dotted, struck title), the target pill, and legend lines for the statuses used.
  Flip **P1b-01**, **P1b-02**.
- [ ] **T03** Parity with DockAir's `postprocess-drawio.py` on its three diagrams;
  delete the script and its side files in `dockair-docs` (separate MR there).

## B · Landscape (P1b-03, P1b-04, P1b-05)

- [x] **T04** Model and schema: kind `tool`; view `type: landscape`, `bands`, `side`,
  `labels`. Checks: unknown item, duplicate item. Table-driven tests.
- [x] **T05** `resolve.Landscape`: bands to `band` zones, sections to `group` zones,
  items to `tile` nodes, computed grid. Test: a two-band system resolves to the expected
  geometry.
- [x] **T06** Render `band` zones and `tile` nodes; section flow arrows when `flow: true`.
  Lint warning for items without icon.
- [ ] **T07** `examples/landscape`: small system covering bands, side column, flow,
  a planned item with target, a `tool`. Dark and light in `init.sh` and CI. Flip
  **P1b-03**, **P1b-04**, **P1b-05**.

## C · Catalog (P1b-06)

- [ ] **T08** Model and schema: `offerings`, `backed_by` check; view `type: catalog`,
  `columns`.
- [ ] **T09** `resolve.Catalog` and the `offering` card; row heights from content; card
  text overflow in the lint.
- [ ] **T10** `examples/catalog` with a planned offering. Flip **P1b-06**.

## D · DockAir (P1b-07)

- [ ] **T11** Icons: `icon-curator` fetches the missing logos, verified with
  `noodle -list-icons`.
- [ ] **T12** DockAir model: tools and planned items with targets, offerings from its
  service catalog; views `tech-stack` and `service-catalog` in `dockair-docs`.
  `noodle:diagram-reviewer` review, findings fixed. Flip **P1b-07**.

## E · Skill and docs

- [ ] **T13** Skill reference: the two view types, `tool`, `offerings`, `target`, with
  one recipe each. README example list.
