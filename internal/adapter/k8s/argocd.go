package k8s

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

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

// ArgoNamespace is the namespace an Application file deploys into, which objects that
// name none are applied to (FR-005).
func ArgoNamespace(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("ArgoNamespace: %w", err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	for {
		var doc yaml.Node
		if err := dec.Decode(&doc); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return "", fmt.Errorf("ArgoNamespace: %s: %w", path, err)
		}
		if len(doc.Content) == 0 {
			continue
		}
		app := doc.Content[0]
		if text(mapValue(app, "kind")) == "Application" {
			if ns := text(at(app, "spec", "destination", "namespace")); ns != "" {
				return ns, nil
			}
		}
	}
	return "", fmt.Errorf("ArgoNamespace: %s holds no Application with spec.destination.namespace", path)
}
