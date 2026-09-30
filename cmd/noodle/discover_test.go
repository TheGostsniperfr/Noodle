package main

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/adapter"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// lineAdapter turns each input line "name" into a workload, and "a->b" into a connection.
type lineAdapter struct{}

func (lineAdapter) Name() string { return "lines" }

func (lineAdapter) Discover(_ context.Context, src adapter.Source) (*model.Fragment, error) {
	f := &model.Fragment{}
	sc := bufio.NewScanner(src.Reader)
	for n := 1; sc.Scan(); n++ {
		at := model.Src{File: "-", Line: n}
		if from, to, ok := strings.Cut(sc.Text(), "->"); ok {
			f.Connections = append(f.Connections, model.DiscoveredConnection{
				Connection: model.Connection{ID: "lines:" + from + "-" + to, From: "lines:" + from, To: "lines:" + to, Kind: "flow"},
				Inferred:   true, Src: at})
			continue
		}
		f.Elements = append(f.Elements, model.DiscoveredElement{Element: model.Element{ID: "lines:" + sc.Text(), Kind: "backend", Title: sc.Text()}, Src: at})
	}
	return f, sc.Err()
}

func TestDiscoverCommand_WritesAFragmentTheSystemLoads_AndARerunIsIdentical(t *testing.T) {
	reg, err := adapter.NewRegistry(lineAdapter{})
	require.NoError(t, err)
	dir := t.TempDir()
	require.NoError(t, os.CopyFS(dir, os.DirFS("../../internal/model/testdata/minimal")))
	require.NoError(t, os.RemoveAll(filepath.Join(dir, "discovered")))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "discovered"), 0o755))
	out := filepath.Join(dir, "discovered", "lines-stdin.yaml")
	args := []string{"lines", "-", "-o", out, "-observed-at", "2026-09-30T12:00:00Z"}
	var stdout bytes.Buffer

	require.NoError(t, discoverCommand(context.Background(), reg, args, strings.NewReader("worker\napi\nworker->api\n"), &stdout))
	first, err := os.ReadFile(out)
	require.NoError(t, err)
	require.NoError(t, discoverCommand(context.Background(), reg, args, strings.NewReader("worker\napi\nworker->api\n"), &stdout))
	second, err := os.ReadFile(out)
	require.NoError(t, err)
	s, err := model.LoadSystem(dir)

	require.NoError(t, err)
	assert.Empty(t, model.Check(s))
	assert.Equal(t, string(first), string(second))
	assert.Equal(t, `apiVersion: noodle/v1alpha1
kind: Fragment
provenance: {adapter: lines, source: '-', observedAt: "2026-09-30T12:00:00Z"}
elements:
  - {id: 'lines:api', kind: backend, title: api, src: {file: '-', line: 2}}
  - {id: 'lines:worker', kind: backend, title: worker, src: {file: '-', line: 1}}
connections:
  - {id: 'lines:worker-api', from: 'lines:worker', to: 'lines:api', kind: flow, inferred: true, src: {file: '-', line: 3}}
`, string(first))
	assert.Empty(t, stdout.String())
}

func TestDiscoverCommand_Fails(t *testing.T) {
	reg, err := adapter.NewRegistry(lineAdapter{})
	require.NoError(t, err)

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"without an adapter, listing the adapters", nil, "adapters: [lines]"},
		{"on an unknown adapter", []string{"k8s"}, `unknown adapter "k8s", available: lines`},
		{"on stdin mixed with paths", []string{"lines", "a.yaml", "-"}, "cannot be mixed with paths"},
		{"on a malformed observation time", []string{"lines", "-observed-at", "yesterday"}, "-observed-at"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := discoverCommand(context.Background(), reg, tt.args, strings.NewReader(""), &bytes.Buffer{})

			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
