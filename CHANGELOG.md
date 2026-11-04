# Changelog

All notable changes to noodle, generated from conventional commits by `make changelog`.

## Unreleased

### Bug fixes

- **render**: Disable font ligatures in labels
- **render**: Lighten planned hatch so titles stay readable at slide scale
- **bin**: Rebuild the plugin binary when internal packages change
- **examples**: Use a severity the model schema accepts in cnp-access
- **render**: Draw matrix group frames as a border, without fill
- **resolve**: Keep the first matrix frame clear of the row headers
- **house**: Darken empty matrix cells in the light theme
- **deps**: Bump golang.org/x/text to 0.41.0 (#63)
- **lint**: Measure each br-separated line of a node desc on its own (#68)
- **process**: Apply contract review rounds 2 and 3
- **process**: Apply contract review round 4
- **process**: Apply contract review round 5
- **process**: Contract criteria are rows under the feature's section
- **process**: Contract criteria ignore fenced examples
- **process**: Count only numbered items as considered options
- **process**: The evidence gate reads only a filled verdict and a real base
- **process**: Review-contract writes through a temp file and needs one spec
- **process**: An Evaluator look-alike or an unreadable spec fails the gate

### CI

- **security**: Add trivy scan and complete dependabot coverage (#64)
- **deps**: Auto-merge dependabot patch and minor updates (#67)

### Chores

- **examples**: Re-render landscape and catalog PNGs with the lighter hatch
- **ci**: Bump actions/checkout from 5 to 7 (#65)
- **ci**: Bump actions/setup-go from 5 to 7 (#66)
- **process**: Make is the only entry point, with an evaluator agent
- **process**: DEV-01 passes
- **process**: DEV-02 passes
- **process**: DEV-03 passes
- **process**: DEV-05 passes
- **process**: DEV-04 passes

### Documentation

- Add session harness (progress log, feature checklist, init script) and spec 001
- Record field feedback as backlog, ADR-0008 and skill rules
- Add backlog B-19 to B-23 from field review
- **adr**: Accept ADR-0007 contract files; plan phase 1
- Illustrate the phase 1 pipeline and require visuals in PRs
- Log T03 to T04b in the progress file
- **spec**: Add spec 002 landscape and catalog views, ADR-0011 to ADR-0013
- **progress**: DockAir tech stack and catalogue rendered, P1b-07 passes
- **adr**: Renumber spec 002 ADRs to 0012-0014
- **spec**: Add spec 003 static discovery, ADR-0015 and ADR-0016
- **skill**: Rewrite the reference and workflow for v1alpha1 systems
- Reference the GitHub project board as the backlog
- **adr**: Add ADR-0017, yaml.v3 decoding and unresolved entries
- **spec**: Add the checkpoint and blind spots to spec 003, ADR-0018
- **adr**: Add ADR-0019 and ADR-0020 and spec 004 for the access matrix
- **examples**: Add the CNP access matrix example
- **skill**: Document badge-only step labels
- **process**: Contract review round 5 of spec 007
- **process**: Cap contract reviews at three rounds
- **process**: Evidence for DEV-01
- **process**: Evidence for DEV-02, needs work
- **process**: Evidence for DEV-02
- **process**: Evidence for DEV-03
- **process**: Evidence for DEV-04, first run
- **process**: Evidence for DEV-05
- **process**: Evidence for DEV-04
- **adr**: ADR-0025, work the backlog in batches and waves
- **progress**: Wave 1 under way
- **spec**: Spec 010, project and distribution (batch E)
- **spec**: Apply contract review round 1 of spec 010
- **spec**: Apply contract review round 2 of spec 010
- **spec**: Contract of spec 010 agreed after three review rounds
- **adr**: ADR-0030, Apache-2.0 without a DCO; spec 010 clarified
- **proj**: Add contributor files and idea and bug issue templates
- **guide**: Add the user guide, noodle help and a generated CLI reference
- **proj**: Position noodle, relate it to living documentation, add the Discovery reference

### Features

- **plugin**: Distribute noodle as a Claude Code plugin with embedded icons
- **harness**: Automate session handoff with hooks, CI and wrap-up skill
- **model**: Add v1alpha1 contract types and system loader
- **model**: Check references across model, views and layouts
- **schemas**: Add v1alpha1 JSON Schemas for model, view and layout
- **examples**: Write the CNP runtime example in the v1alpha1 format
- **model**: Name what enforces a denied connection (ADR-0010)
- **model**: Check routes, edge direction and card positions
- **resolve**: Resolve topology views to absolute geometry
- **cli**: Render a system directory with noodle render
- **model**: Sequence steps are messages over model connections (ADR-0011)
- **sequence**: Render sequence views and the CNP OIDC login
- **model**: Add target to planned elements, inherit status from zones
- **render**: Draw planned and deprecated status with target pill and legend
- **model**: Add tool elements and landscape views with item checks
- **resolve**: Resolve landscape views to a computed grid
- **lint**: Warn about landscape items without icon
- **examples**: Add landscape example to init and CI
- **model**: Add offerings and catalog views with reference checks
- **resolve**: Resolve catalog views to offering cards
- **examples**: Add catalog example to init and CI
- **icons**: Add logos for landscape and catalog views
- **cli**: Add -slide to render the drawing only, cropped to its content
- **agents**: Collect diagram facts in a Sonnet fact-finder subagent
- **cli**: Add noodle migrate from a v0 file to a v1alpha1 system
- **render**: Draw multiplicity as a stack, footprint known to the lint
- **model**: Add the Fragment contract and matches on elements
- **adapter**: Add the adapter registry and noodle discover
- **k8s**: Decode manifests into workloads, services and namespaces
- **k8s**: Connect routes, ingresses and services to their backends
- **model**: Add memberships, grants and the matrix view fields
- **access**: Add levels, grant ids and phases per state
- **resolve**: Resolve matrix views into cells, changes and counts
- **render**: Draw access matrices with level fills and change frames
- **render**: Frame matrix column groups and put the legend in a card
- **render**: Draw a topology view as a proposal diff
- **render**: Highlight a subset of a topology view
- **model**: Draw queues as pipes and broker routing as routes
- **process**: Traceable specs, ADR shape and evidence gate in tools/speccheck
- **proj**: Add Apache-2.0 LICENSE and NOTICE, declare the license in plugin, flake and skill
- **assets**: Attribute every embedded icon and test the attribution page
- **speccheck**: Validate every skill against the Agent Skills rules
- **cli**: Add noodle version set by ldflags, passed by the flake
- **proj**: Embed the skill and agents in the binary, record the Codex agent format
- **cli**: Add noodle setup for OpenCode and Codex
- **cli**: Mark files written by noodle setup and keep foreign ones

### Refactoring

- Move resolved geometry types to internal/diagram
- **lint**: One rule per file in internal/lint, house style in internal/house
- **examples**: Route the CNP runtime layout on named lanes
- **examples**: Migrate platform-regression to one system per diagram
- **cli**: Render systems only, v0 files go through noodle migrate (breaking)
- **skill**: Move Claude Code-only names under an In Claude Code section

### Tests

- Add anonymized platform regression fixtures linted in CI
- **process**: Drop a commands case that tested nothing
