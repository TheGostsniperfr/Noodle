// Package k8s discovers elements, connections and references from Kubernetes manifests
// or `kubectl get -o yaml` output (spec 003). It reads YAML with yaml.v3 so every item
// keeps its file and line (ADR-0017), and only the few fields each rule needs.
package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/adapter"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

type Adapter struct{}

func (Adapter) Name() string { return "k8s" }

func (Adapter) Discover(ctx context.Context, src adapter.Source) (*model.Fragment, adapter.Stats, error) {
	in, err := read(ctx, src)
	if err != nil {
		return nil, in.stats, fmt.Errorf("Discover: %w", err)
	}
	idx := newIndex(in.objects)
	f := &model.Fragment{Unresolved: in.unrendered}
	f.Elements = append(f.Elements, workloads(idx)...)
	f.Elements = append(f.Elements, services(idx)...)
	f.Elements = append(f.Elements, namespaces(idx, f.Elements)...)
	f.Unresolved = append(f.Unresolved, unknownKinds(idx)...)
	return f, in.stats, nil
}

// id is k8s:<namespace>/<kind>/<name> (ADR-0015). A namespace is its own namespace.
// Elements carry no title: the id holds the name, and the curated model names things.
func id(ns, kind, name string) string {
	return "k8s:" + ns + "/" + strings.ToLower(kind) + "/" + name
}

func (o *object) id() string { return id(o.Namespace, o.Kind, o.Name) }
