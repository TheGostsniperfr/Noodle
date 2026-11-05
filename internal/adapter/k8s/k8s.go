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
	g, u := &graph{}, &unresolvedSet{}
	f := &model.Fragment{Unresolved: in.unrendered}
	f.Elements = append(f.Elements, workloads(idx)...)
	f.Elements = append(f.Elements, services(idx)...)
	f.Elements = append(f.Elements, gateways(idx)...)
	f.Elements = append(f.Elements, routes(idx, g, u)...)
	f.Elements = append(f.Elements, ingresses(idx, g, u)...)
	selectsPods(idx, g, u)
	uses := configRefs(idx, g)
	secretEndpoints(idx, uses, u)
	f.Elements = append(f.Elements, externalSecrets(idx, g, u)...)
	applications(idx, g)
	f.Elements = append(f.Elements, addresses(idx, g, uses)...)
	f.Elements = append(f.Elements, namespaces(idx, f.Elements)...)
	f.Connections, f.References = g.conns, g.refs
	f.Unresolved = append(f.Unresolved, u.list()...)
	f.Unresolved = append(f.Unresolved, unknownKinds(idx)...)
	return f, in.stats, nil
}

// id is k8s:<namespace>/<kind>/<name> (ADR-0015). A namespace is its own namespace.
// Elements carry no title: the id holds the name, and the curated model names things.
func id(ns, kind, name string) string {
	if ns == "" {
		ns = "_cluster"
	}
	return "k8s:" + ns + "/" + strings.ToLower(kind) + "/" + name
}

func (o *object) id() string { return id(o.Namespace, o.Kind, o.Name) }

// identity is the object block of every item built from o (ADR-0028). A nil object
// gives nil, for the elements synthesized without one.
func (o *object) identity() *model.Object {
	if o == nil {
		return nil
	}
	return &model.Object{APIVersion: o.APIVersion, Kind: o.Kind, Namespace: o.Namespace, Name: o.Name}
}

// srcAt is where a field of o sits. A rendered input has no meaningful line, so its
// objects carry none and neither does anything derived from them.
func (o *object) srcAt(line int) model.Src {
	if o.Src.Line == 0 || line == 0 {
		return o.Src
	}
	return model.Src{File: o.Src.File, Line: line}
}
