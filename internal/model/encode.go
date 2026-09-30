package model

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// EncodeFragment writes one item per line, so a rerun on changed input diffs line by line.
func EncodeFragment(w io.Writer, f *Fragment) error {
	doc := &yaml.Node{Kind: yaml.MappingNode}
	add := func(key string, val *yaml.Node) {
		doc.Content = append(doc.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, val)
	}
	node := func(v any, style yaml.Style) (*yaml.Node, error) {
		var n yaml.Node
		if err := n.Encode(v); err != nil {
			return nil, fmt.Errorf("EncodeFragment: %w", err)
		}
		n.Style = style
		return &n, nil
	}
	for _, kv := range []struct {
		key string
		v   any
	}{{"apiVersion", f.APIVersion}, {"kind", f.Kind}, {"provenance", f.Provenance}} {
		n, err := node(kv.v, yaml.FlowStyle)
		if err != nil {
			return err
		}
		add(kv.key, n)
	}
	lists := []struct {
		key   string
		items []any
	}{{"elements", toAny(f.Elements)}, {"connections", toAny(f.Connections)}, {"references", toAny(f.References)}}
	for _, l := range lists {
		if len(l.items) == 0 {
			continue
		}
		seq := &yaml.Node{Kind: yaml.SequenceNode}
		for _, it := range l.items {
			n, err := node(it, yaml.FlowStyle)
			if err != nil {
				return err
			}
			seq.Content = append(seq.Content, n)
		}
		add(l.key, seq)
	}
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return fmt.Errorf("EncodeFragment: %w", err)
	}
	return enc.Close()
}

func toAny[T any](items []T) []any {
	out := make([]any, len(items))
	for i, it := range items {
		out[i] = it
	}
	return out
}
