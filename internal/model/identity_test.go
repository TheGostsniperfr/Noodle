package model_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

func fragmentWith(items string) []byte {
	return []byte(fragmentHead + fragmentProvenance + items)
}

func TestIdentity(t *testing.T) {
	t.Parallel()
	schemas := compileSchemas(t)

	t.Run("schema-accepts", func(t *testing.T) {
		t.Parallel()
		tests := []struct{ name, doc string }{
			{"an element with its object", `elements: [{id: "k8s:prod/deployment/api", kind: backend, src: {file: f, line: 1}, object: {apiVersion: apps/v1, kind: Deployment, namespace: prod, name: api}}]`},
			{"a cluster-scoped object without namespace", `elements: [{id: "k8s:_cluster/ingressclass/nginx", kind: backend, src: {file: f}, object: {apiVersion: networking.k8s.io/v1, kind: IngressClass, name: nginx}}]`},
			{"an element without object", `elements: [{id: "k8s:prod/namespace/prod", kind: group, src: {file: f}}]`},
			{"a connection with its object", `connections: [{id: "k8s:c/1", from: "k8s:a/b/c", to: "k8s:a/b/d", kind: flow, src: {file: f}, object: {apiVersion: v1, kind: ConfigMap, namespace: prod, name: cfg}}]`},
			{"a reference with its object", `references: [{id: "k8s:r/1", from: "k8s:a/b/c", to: "k8s:a/b/d", src: {file: f}, object: {apiVersion: apps/v1, kind: Deployment, namespace: prod, name: api}}]`},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := schemas["Fragment"].Validate(yamlToJSONValue(t, fragmentWith(tt.doc+"\n")))

				assert.NoError(t, err)
			})
		}
	})

	t.Run("schema-rejects", func(t *testing.T) {
		t.Parallel()
		tests := []struct{ name, object string }{
			{"no apiVersion", `{kind: Deployment, namespace: prod, name: api}`},
			{"no kind", `{apiVersion: apps/v1, namespace: prod, name: api}`},
			{"no name", `{apiVersion: apps/v1, kind: Deployment, namespace: prod}`},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				doc := `elements: [{id: "k8s:prod/deployment/api", kind: backend, src: {file: f}, object: ` + tt.object + `}]` + "\n"

				err := schemas["Fragment"].Validate(yamlToJSONValue(t, fragmentWith(doc)))

				assert.Error(t, err)
			})
		}
	})

	t.Run("src-object-unchanged", func(t *testing.T) {
		t.Parallel()
		doc := `elements: [{id: "k8s:prod/deployment/worker", kind: backend, src: {object: prod/Deployment/worker}}]` + "\n"

		err := schemas["Fragment"].Validate(yamlToJSONValue(t, fragmentWith(doc)))

		assert.NoError(t, err)
	})

	t.Run("legacy-fragment", func(t *testing.T) {
		t.Parallel()
		const fixture = "testdata/minimal/discovered/k8s-manifests.yaml"
		raw, err := os.ReadFile(fixture)
		require.NoError(t, err)
		require.NoError(t, schemas["Fragment"].Validate(yamlToJSONValue(t, raw)))
		first := loadFragmentFile(t, fixture)

		encoded := encodeFragment(t, first)
		second := loadFragmentFile(t, writeFragment(t, encoded))
		reencoded := encodeFragment(t, second)

		assert.Equal(t, first, second)
		assert.True(t, bytes.Equal(encoded, reencoded), "re-encoding changed the bytes")
		var doc map[string]any
		require.NoError(t, yaml.Unmarshal(encoded, &doc))
		for _, list := range []string{"elements", "connections", "references"} {
			items, _ := doc[list].([]any)
			if list != "references" {
				require.NotEmpty(t, items, list)
			}
			for _, it := range items {
				assert.NotContains(t, it, "object", "%s item %v", list, it)
			}
		}
	})
}

func loadFragmentFile(t *testing.T, path string) *model.Fragment {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.CopyFS(dir, os.DirFS("testdata/minimal")))
	require.NoError(t, os.RemoveAll(filepath.Join(dir, "discovered")))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "discovered"), 0o755))
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "discovered", "k8s-manifests.yaml"), raw, 0o644))
	s, err := model.LoadSystem(dir)
	require.NoError(t, err)
	return s.Fragments["k8s-manifests"]
}

func writeFragment(t *testing.T, raw []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "k8s-manifests.yaml")
	require.NoError(t, os.WriteFile(path, raw, 0o644))
	return path
}

func encodeFragment(t *testing.T, f *model.Fragment) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, model.EncodeFragment(&buf, f))
	return buf.Bytes()
}
