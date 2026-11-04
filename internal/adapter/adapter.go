// Package adapter runs discovery: an Adapter reads a source and returns a Fragment, never
// an edit of the curated model (ADR-0015). Adapters live in their own packages and are
// registered by the command, so adding one changes no other module.
package adapter

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

type Adapter interface {
	// Name is the prefix of every id the adapter emits, e.g. k8s.
	Name() string
	Discover(ctx context.Context, src Source) (*model.Fragment, Stats, error)
}

// Stats is what an adapter read and left out, so a run shows whether discovery pays
// for itself against reading the source directly.
type Stats struct {
	InputBytes int
	// NoiseBytes is the part of InputBytes no reader would need, such as CRD schemas.
	NoiseBytes int
	Objects    int
	// Skipped counts objects left out, by reason.
	Skipped map[string]int
	// Namespaced and NoNamespace count the kept objects of a namespaced kind that name a
	// namespace and that name none, so a command can warn about a mixed input.
	Namespaced, NoNamespace int
}

// Source is either Paths (files or directories) or Reader, e.g. stdin.
type Source struct {
	Paths  []string
	Reader io.Reader
	// Namespace is where objects that name none are applied, as kubectl apply -n and
	// an Argo CD destination do. Rendered charts often leave it to the deployer.
	Namespace string
	// Kustomize is a directory rendered with `kustomize build` before discovery, with
	// --enable-helm when EnableHelm is set (ADR-0027).
	Kustomize  string
	EnableHelm bool
}

// Name is what provenance.source records.
func (s Source) Name() string {
	if s.Reader != nil {
		return "-"
	}
	var parts []string
	for _, p := range append(append([]string{}, s.Paths...), s.Kustomize) {
		if p != "" {
			parts = append(parts, RelativePath(p))
		}
	}
	return strings.Join(parts, ",")
}

// Slug names a fragment file after its source: the last element of the first path, or
// of the kustomize directory, lowercased to letters, digits and hyphens.
func (s Source) Slug() string {
	p := s.Kustomize
	if p == "" && len(s.Paths) > 0 {
		p = s.Paths[0]
	}
	if p == "" {
		return "stdin"
	}
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	var b strings.Builder
	for _, r := range strings.ToLower(filepath.Base(p)) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else if b.Len() > 0 && !strings.HasSuffix(b.String(), "-") {
			b.WriteByte('-')
		}
	}
	if slug := strings.Trim(b.String(), "-"); slug != "" {
		return slug
	}
	return "stdin"
}

// RelativePath is p relative to the working directory, with forward slashes, so a
// fragment never holds an absolute or temporary path (FR-007).
func RelativePath(p string) string {
	abs, err := filepath.Abs(p)
	if err == nil {
		if wd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(wd, abs); err == nil {
				return filepath.ToSlash(rel)
			}
		}
	}
	return filepath.ToSlash(p)
}

type Registry struct {
	adapters map[string]Adapter
}

func NewRegistry(adapters ...Adapter) (*Registry, error) {
	r := &Registry{adapters: map[string]Adapter{}}
	for _, a := range adapters {
		if !model.IsDiscoveredID(a.Name() + ":x") {
			return nil, fmt.Errorf("NewRegistry: adapter name %q must be lowercase letters and digits", a.Name())
		}
		if _, dup := r.adapters[a.Name()]; dup {
			return nil, fmt.Errorf("NewRegistry: adapter %q registered twice", a.Name())
		}
		r.adapters[a.Name()] = a
	}
	return r, nil
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.adapters))
	for n := range r.adapters {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) Get(name string) (Adapter, error) {
	a, ok := r.adapters[name]
	if !ok {
		available := "none"
		if names := r.Names(); len(names) > 0 {
			available = strings.Join(names, ", ")
		}
		return nil, fmt.Errorf("unknown adapter %q, available: %s", name, available)
	}
	return a, nil
}

// Run discovers, stamps the provenance and sorts every list by id, so the same input
// gives the same fragment byte for byte whatever order the adapter found things in.
func Run(ctx context.Context, a Adapter, src Source, ref string, observedAt time.Time) (*model.Fragment, Stats, error) {
	f, stats, err := a.Discover(ctx, src)
	if err != nil {
		return nil, stats, fmt.Errorf("%s: %w", a.Name(), err)
	}
	f.APIVersion, f.Kind = model.APIVersion, "Fragment"
	f.Provenance = model.Provenance{Adapter: a.Name(), Source: src.Name(), Ref: ref, ObservedAt: observedAt.UTC().Format(time.RFC3339)}
	sort.SliceStable(f.Elements, func(i, j int) bool { return f.Elements[i].ID < f.Elements[j].ID })
	sort.SliceStable(f.Connections, func(i, j int) bool { return f.Connections[i].ID < f.Connections[j].ID })
	sort.SliceStable(f.References, func(i, j int) bool { return f.References[i].ID < f.References[j].ID })
	sort.SliceStable(f.Unresolved, func(i, j int) bool {
		a, b := f.Unresolved[i], f.Unresolved[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Value < b.Value
	})
	if err := checkPrefix(f, a.Name()+":"); err != nil {
		return nil, stats, fmt.Errorf("%s: %w", a.Name(), err)
	}
	return f, stats, nil
}

// checkPrefix holds a third-party adapter to ADR-0015: its ids name it and are well formed.
func checkPrefix(f *model.Fragment, prefix string) error {
	var ids []string
	for _, e := range f.Elements {
		ids = append(ids, e.ID)
	}
	for _, c := range f.Connections {
		ids = append(ids, c.ID)
	}
	for _, r := range f.References {
		ids = append(ids, r.ID)
	}
	for _, id := range ids {
		if !strings.HasPrefix(id, prefix) || !model.IsDiscoveredID(id) {
			return fmt.Errorf("id %q is not a discovered id %s<path>", id, prefix)
		}
	}
	return nil
}
