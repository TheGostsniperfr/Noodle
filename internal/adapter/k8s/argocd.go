package k8s

// applications links each Argo CD Application to the namespace it syncs into (FR-002).
func applications(idx *index, g *graph) {
	for _, app := range idx.byKind["Application"] {
		ns := at(app.Node, "spec", "destination", "namespace")
		if text(ns) == "" {
			continue
		}
		g.refer(app.id(), id(text(ns), "Namespace", text(ns)), "syncs", app, app.srcAt(ns.Line))
	}
}
