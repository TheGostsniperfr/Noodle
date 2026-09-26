# Inspirations

What others do well, and what noodle takes from it. Our need differs from most of them:
they are design tools for documentation, while noodle wants a model derived from code
and infra, reviewed in Git, and later connected to live systems.

| Tool | What it does well | What we take |
|---|---|---|
| [IcePanel](https://icepanel.io/) | C4 model, zoom between levels, **Flows** with 8 step types and playback, **Tags** as perspective overlays, **Drafts** of a future state, links from objects to real code, object-level history | the step types and playback model; tags as overlays (risk, owner, tech); drafts as Git branches with a visual diff; links to reality |
| [Ilograph](https://www.ilograph.com/features.html) | model-based YAML, **perspectives** (same resources, different relations), instant **focus**, sequences, walkthroughs, resource finder, Markdown docs per element, keyboard and screen-reader access, **self-contained HTML export** | focus-and-dim, perspectives, walkthroughs, accessibility, single-file HTML for portfolios and PR links |
| [LikeC4](https://likec4.dev/) | model/view split in a DSL, dynamic views with parallel/alt/loop blocks and a sequence mode, interactive viewer, MCP server, draw.io round-trip, a live-status showcase | the flow-control blocks of dynamic views; the idea of an MCP over the model. Not adopted as an engine (ADR-0001). |
| [Structurizr](https://structurizr.com/) / [C4 model](https://c4model.com/) | the reference for levels, notation and the name/technology/description triptych | node text and levels |
| [Cocoon-AI architecture-diagram-generator](https://github.com/Cocoon-AI/architecture-diagram-generator) | a dark, legible visual style | the Tailwind palette and look, adapted with a light theme |
| [Cilium Hubble UI](https://cilium.io/use-cases/service-map/) | a service map from real network flows | Hubble as the source of truth for connections and ports (L4), and for dropped flows (L5) |
| [Kiali](https://kiali.io/) | a traffic graph with health from Prometheus | health overlays on a graph |
| [KubeView](https://kubeview.benco.io/), KubeDiagrams | a raw resource graph of a cluster | starting point for the manifest and kubectl adapters; our value is the curated architecture level and the drift between the two |
| [Cartography (CNCF)](https://github.com/cartography-cncf/cartography) | an infra asset graph with provenance, in Neo4j | provenance on discovered facts |
| [HolmesGPT](https://github.com/HolmesGPT/holmesgpt) | an AI SRE agent working through MCP toolsets | not rebuilt: noodle provides the topology context it lacks, through MCP (L6) |
| Control-room mimic boards | power grids and railways shown as a live synoptic board | the target experience of L5 |
