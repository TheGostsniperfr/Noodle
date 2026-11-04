package k8s

import "gopkg.in/yaml.v3"

// The rules that cite a line read the yaml.Node tree, which keeps one for every value,
// while the maps of object.Body drop it.

func resolve(n *yaml.Node) *yaml.Node {
	for n != nil && n.Kind == yaml.AliasNode {
		n = n.Alias
	}
	return n
}

// mapValue returns the value of key in a mapping node. As with Body, the last of a
// repeated key wins and a merge key supplies what the mapping does not set.
func mapValue(n *yaml.Node, key string) *yaml.Node {
	n = resolve(n)
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}
	var found, merged *yaml.Node
	for i := 0; i+1 < len(n.Content); i += 2 {
		switch n.Content[i].Value {
		case key:
			found = n.Content[i+1]
		case "<<":
			if v := mapValue(n.Content[i+1], key); v != nil && merged == nil {
				merged = v
			}
		}
	}
	if found != nil {
		return found
	}
	return merged
}

func at(n *yaml.Node, path ...string) *yaml.Node {
	for _, key := range path {
		if n = mapValue(n, key); n == nil {
			return nil
		}
	}
	return resolve(n)
}

func seq(n *yaml.Node) []*yaml.Node {
	if n = resolve(n); n == nil || n.Kind != yaml.SequenceNode {
		return nil
	}
	out := make([]*yaml.Node, len(n.Content))
	for i, c := range n.Content {
		out[i] = resolve(c)
	}
	return out
}

// text reads a scalar the way str reads a map value.
func text(n *yaml.Node) string {
	if n = resolve(n); n == nil || n.Kind != yaml.ScalarNode {
		return ""
	}
	return n.Value
}

func lineOf(n *yaml.Node) int {
	if n == nil {
		return 0
	}
	return n.Line
}

// firstLine is the line of the first node that exists.
func firstLine(nodes ...*yaml.Node) int {
	for _, n := range nodes {
		if n != nil {
			return n.Line
		}
	}
	return 0
}

// containers are the init and regular containers of a workload, in spec order.
func (o *object) containers() []*yaml.Node {
	spec := at(o.Node, podSpecPath[o.Kind]...)
	return append(seq(mapValue(spec, "initContainers")), seq(mapValue(spec, "containers"))...)
}

func (o *object) podSpecNode() *yaml.Node {
	return at(o.Node, podSpecPath[o.Kind]...)
}
