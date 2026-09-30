package k8s

import (
	"fmt"
	"sort"
)

func set(vs ...string) map[string]bool {
	m := map[string]bool{}
	for _, v := range vs {
		m[v] = true
	}
	return m
}

func field(m map[string]any, path ...string) any {
	var cur any = m
	for _, p := range path {
		mm, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = mm[p]
	}
	return cur
}

// str reads a scalar as text: YAML gives numbers and booleans their own types, and a
// name or a label value can be written unquoted as either.
func str(m map[string]any, path ...string) string {
	switch v := field(m, path...).(type) {
	case nil:
		return ""
	case string:
		return v
	case map[string]any, []any:
		return ""
	default:
		return fmt.Sprint(v)
	}
}

func num(m map[string]any, path ...string) (int, bool) {
	v, ok := field(m, path...).(int)
	return v, ok
}

func slice(m map[string]any, path ...string) []any {
	v, _ := field(m, path...).([]any)
	return v
}

func obj(m map[string]any, path ...string) map[string]any {
	v, _ := field(m, path...).(map[string]any)
	return v
}

func strMap(m map[string]any, path ...string) map[string]string {
	src := obj(m, path...)
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]string, len(src))
	for k := range src {
		out[k] = str(src, k)
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
