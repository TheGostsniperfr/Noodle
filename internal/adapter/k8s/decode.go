package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/TheGostsniperfr/Noodle/internal/adapter"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// object is one Kubernetes object: its identity, its body as decoded maps, and where it
// was read.
type object struct {
	APIVersion, Kind, Namespace, Name string
	Labels, Annotations               map[string]string
	Body                              map[string]any
	Src                               model.Src
}

type input struct {
	objects    []*object
	unrendered []model.Unresolved
	stats      adapter.Stats
	seen       map[string]bool
	namespace  string
}

// Skip reasons, as counted in Stats.
const (
	skipOwned     = "owned"
	skipHelmHook  = "helm-hook"
	skipGenerated = "generated"
	skipNoKind    = "no-kind"
	skipDuplicate = "duplicate"
	skipCRD       = "crd"
)

// generatedKinds are written by controllers, never by a person, and only appear in
// kubectl output: reading them would make live output differ from the manifests. Pods
// and ReplicaSets are left out by their owner reference instead, so a standalone Pod
// still counts as a workload.
var generatedKinds = set("ControllerRevision", "Endpoints", "EndpointSlice", "Event", "Lease")

func read(ctx context.Context, src adapter.Source) (*input, error) {
	in := &input{stats: adapter.Stats{Skipped: map[string]int{}}, seen: map[string]bool{}, namespace: or(src.Namespace, "default")}
	if src.Reader != nil {
		raw, err := io.ReadAll(src.Reader)
		if err != nil {
			return in, fmt.Errorf("read stdin: %w", err)
		}
		return in, in.decode("-", raw)
	}
	for _, p := range src.Paths {
		if err := in.walk(ctx, p); err != nil {
			return in, err
		}
	}
	return in, nil
}

func (in *input) walk(ctx context.Context, root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			if tool, marker := unrenderedDir(path); tool != "" {
				in.unrendered = append(in.unrendered, model.Unresolved{
					Kind: "unrendered", Value: filepath.ToSlash(path), Hint: tool,
					Src: model.Src{File: filepath.ToSlash(marker), Line: 1}})
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if path != root && ext != ".yaml" && ext != ".yml" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		return in.decode(filepath.ToSlash(path), raw)
	})
}

// unrenderedDir names the tool that renders a Helm chart or a Kustomize directory, and
// the file that marks it. Their raw files are templates and patches, not objects.
func unrenderedDir(dir string) (tool, marker string) {
	for _, m := range []struct{ file, tool string }{
		{"Chart.yaml", "helm"}, {"kustomization.yaml", "kustomize"}, {"kustomization.yml", "kustomize"}, {"Kustomization", "kustomize"},
	} {
		p := filepath.Join(dir, m.file)
		if _, err := os.Stat(p); err == nil {
			return m.tool, p
		}
	}
	return "", ""
}

func (in *input) decode(file string, raw []byte) error {
	in.stats.InputBytes += len(raw)
	lines := lineOffsets(raw)
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	var docs []*yaml.Node
	for {
		var doc yaml.Node
		err := dec.Decode(&doc)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			if bytes.Contains(raw, []byte("{{")) {
				return fmt.Errorf("%s: %w; it looks like a template, render it first (helm template, kustomize build)", file, err)
			}
			return fmt.Errorf("%s: %w", file, err)
		}
		if len(doc.Content) > 0 {
			docs = append(docs, doc.Content[0])
		}
	}
	for i, n := range docs {
		end := len(raw)
		if i+1 < len(docs) {
			end = lines.offset(docs[i+1].Line)
		}
		if err := in.add(file, n, lines, end); err != nil {
			return err
		}
	}
	return nil
}

// add reads one document or List item. end is the byte offset where it stops, so its
// size can be counted without encoding it again.
func (in *input) add(file string, n *yaml.Node, lines lineIndex, end int) error {
	if n.Kind != yaml.MappingNode {
		return nil
	}
	body, _ := value(n).(map[string]any)
	kind := str(body, "kind")
	if items := itemNodes(n); items != nil && strings.HasSuffix(kind, "List") {
		for i, it := range items {
			itemEnd := end
			if i+1 < len(items) {
				itemEnd = lines.offset(items[i+1].Line)
			}
			if err := in.add(file, it, lines, itemEnd); err != nil {
				return err
			}
		}
		return nil
	}
	if kind == "" || str(body, "apiVersion") == "" {
		in.stats.Skipped[skipNoKind]++
		return nil
	}
	in.stats.Objects++
	o := &object{
		APIVersion: str(body, "apiVersion"), Kind: kind,
		Namespace: str(body, "metadata", "namespace"), Name: str(body, "metadata", "name"),
		Labels: strMap(body, "metadata", "labels"), Annotations: strMap(body, "metadata", "annotations"),
		Body: body, Src: model.Src{File: file, Line: n.Line},
	}
	switch {
	case kind == "CustomResourceDefinition":
		in.stats.Skipped[skipCRD]++
		in.stats.NoiseBytes += end - lines.offset(n.Line)
	case generatedKinds[kind]:
		in.stats.Skipped[skipGenerated]++
	case ownedByController(body):
		in.stats.Skipped[skipOwned]++
	case offPathHook(o.Annotations["helm.sh/hook"]):
		in.stats.Skipped[skipHelmHook]++
	default:
		if o.Namespace == "" && !clusterScoped[kind] {
			o.Namespace = in.namespace
		}
		// The first definition wins, in path order, so reruns agree.
		if key := o.APIVersion + " " + o.id(); in.seen[key] {
			in.stats.Skipped[skipDuplicate]++
		} else {
			in.seen[key] = true
			in.objects = append(in.objects, o)
		}
	}
	return nil
}

// value converts a node to maps, slices and scalars. Unlike yaml.v3's own decoding, a
// key repeated in a mapping keeps its last value, as the Kubernetes API and Helm do:
// rendered charts contain such keys and the cluster accepts them.
func value(n *yaml.Node) any {
	switch n.Kind {
	case yaml.MappingNode:
		m := make(map[string]any, len(n.Content)/2)
		for i := 0; i+1 < len(n.Content); i += 2 {
			if n.Content[i].Value == "<<" {
				if merged, ok := value(n.Content[i+1]).(map[string]any); ok {
					for k, v := range merged {
						if _, set := m[k]; !set {
							m[k] = v
						}
					}
				}
				continue
			}
			m[n.Content[i].Value] = value(n.Content[i+1])
		}
		return m
	case yaml.SequenceNode:
		out := make([]any, len(n.Content))
		for i, c := range n.Content {
			out[i] = value(c)
		}
		return out
	case yaml.AliasNode:
		return value(n.Alias)
	case yaml.DocumentNode:
		if len(n.Content) == 0 {
			return nil
		}
		return value(n.Content[0])
	default:
		var v any
		if err := n.Decode(&v); err != nil {
			return n.Value
		}
		return v
	}
}

// lineIndex maps a 1-based line number to the byte offset where it starts.
type lineIndex []int

func lineOffsets(raw []byte) lineIndex {
	idx := lineIndex{0, 0}
	for i, b := range raw {
		if b == '\n' {
			idx = append(idx, i+1)
		}
	}
	return idx
}

func (l lineIndex) offset(line int) int {
	if line < len(l) {
		return l[line]
	}
	return l[len(l)-1]
}

// offPathHook is a Helm hook that never runs while the release serves: tests, and
// delete or rollback steps. Install and upgrade hooks, such as migrations, are kept.
func offPathHook(hooks string) bool {
	if hooks == "" {
		return false
	}
	for _, h := range strings.Split(hooks, ",") {
		switch h = strings.TrimSpace(h); {
		case strings.HasPrefix(h, "test"), strings.HasSuffix(h, "-delete"), strings.HasSuffix(h, "-rollback"):
		default:
			return false
		}
	}
	return true
}

// itemNodes returns the items of a List, as nodes so each keeps its line.
func itemNodes(n *yaml.Node) []*yaml.Node {
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == "items" && n.Content[i+1].Kind == yaml.SequenceNode {
			return n.Content[i+1].Content
		}
	}
	return nil
}

func ownedByController(body map[string]any) bool {
	for _, ref := range slice(body, "metadata", "ownerReferences") {
		if m, ok := ref.(map[string]any); ok && m["controller"] == true {
			return true
		}
	}
	return false
}
