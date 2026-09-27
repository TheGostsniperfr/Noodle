package model_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const schemaDir = "../../schemas/v1alpha1"

func compileSchemas(t *testing.T) map[string]*jsonschema.Schema {
	t.Helper()
	c := jsonschema.NewCompiler()
	out := map[string]*jsonschema.Schema{}
	for _, kind := range []string{"Model", "View", "Layout"} {
		path, err := filepath.Abs(filepath.Join(schemaDir, strings.ToLower(kind)+".schema.json"))
		require.NoError(t, err)
		s, err := c.Compile(path)
		require.NoError(t, err, kind)
		out[kind] = s
	}
	return out
}

// yamlToJSONValue goes through encoding/json so numbers and maps have the types the
// validator expects.
func yamlToJSONValue(t *testing.T, raw []byte) any {
	t.Helper()
	var v any
	require.NoError(t, yaml.Unmarshal(raw, &v))
	b, err := json.Marshal(v)
	require.NoError(t, err)
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	require.NoError(t, err)
	return doc
}

// contractFiles returns every model.yaml, views/*.yaml and layouts/*.yaml under root.
func contractFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return err
		}
		switch {
		case d.Name() == "model.yaml":
			files[path] = "Model"
		case filepath.Base(filepath.Dir(path)) == "views":
			files[path] = "View"
		case filepath.Base(filepath.Dir(path)) == "layouts":
			files[path] = "Layout"
		}
		return nil
	})
	require.NoError(t, err)
	return files
}

func TestSchemas_AcceptEveryContractFileInTheRepository(t *testing.T) {
	t.Parallel()
	schemas := compileSchemas(t)
	files := contractFiles(t, "testdata")
	for path, kind := range contractFiles(t, "../../examples") {
		files[path] = kind
	}
	require.NotEmpty(t, files)

	for path, kind := range files {
		t.Run(path, func(t *testing.T) {
			raw, err := os.ReadFile(path)
			require.NoError(t, err)

			err = schemas[kind].Validate(yamlToJSONValue(t, raw))

			assert.NoError(t, err)
		})
	}
}

func TestSchemas_Reject(t *testing.T) {
	t.Parallel()
	schemas := compileSchemas(t)

	tests := []struct {
		name, kind, doc string
	}{
		{"a blocked connection kind, replaced by denied", "Model", modelHead + "connections: [{id: c, from: a, to: b, kind: blocked}]\n"},
		{"multiplicity as a boolean flag", "Model", modelHead + "elements: [{id: a, kind: backend, multiplicity: true}]\n"},
		{"a port number out of range", "Model", modelHead + "elements: [{id: a, kind: backend, ports: [{name: p, protocol: TCP, port: 70000}]}]\n"},
		{"a view without type", "View", viewHead + "title: T\n"},
		{"participants on a topology view", "View", viewHead + "type: topology\nparticipants: [a]\n"},
		{"an endpoint without side", "Layout", layoutHead + "edges: {c: {from: a, to: b.left}}\n"},
		{"a ratio above 100 %", "Layout", layoutHead + "edges: {c: {from: a.right@120%, to: b.left}}\n"},
		{"a lane with both x and y", "Layout", layoutHead + "lanes: {bus: {x: 1, y: 2}}\n"},
		{"absolute v0 path field", "Layout", layoutHead + "edges: {c: {from: a.right, to: b.left, path: [[0, 0]]}}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := schemas[tt.kind].Validate(yamlToJSONValue(t, []byte(tt.doc)))

			assert.Error(t, err)
		})
	}
}
