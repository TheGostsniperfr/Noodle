package adapter_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/adapter"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

type stubAdapter struct {
	name string
	ids  []string
}

func (s stubAdapter) Name() string { return s.name }

func (s stubAdapter) Discover(context.Context, adapter.Source) (*model.Fragment, error) {
	f := &model.Fragment{}
	for _, id := range s.ids {
		f.Elements = append(f.Elements, model.DiscoveredElement{Element: model.Element{ID: id, Kind: "backend"}, Src: model.Src{Object: id}})
	}
	return f, nil
}

func TestNewRegistry_Fails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		adapters []adapter.Adapter
		wantErr  string
	}{
		{"on a name that cannot prefix an id", []adapter.Adapter{stubAdapter{name: "K8s"}}, `adapter name "K8s"`},
		{"on the same name twice", []adapter.Adapter{stubAdapter{name: "k8s"}, stubAdapter{name: "k8s"}}, `"k8s" registered twice`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := adapter.NewRegistry(tt.adapters...)

			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestRegistry_Get_ListsTheAvailableAdapters_OnAnUnknownName(t *testing.T) {
	t.Parallel()
	reg, err := adapter.NewRegistry(stubAdapter{name: "tf"}, stubAdapter{name: "k8s"})
	require.NoError(t, err)

	_, err = reg.Get("helm")

	assert.EqualError(t, err, `unknown adapter "helm", available: k8s, tf`)
}

func TestRun_StampsProvenanceAndSortsByID(t *testing.T) {
	t.Parallel()
	a := stubAdapter{name: "k8s", ids: []string{"k8s:p/service/b", "k8s:p/deployment/a"}}
	at := time.Date(2026, 9, 30, 14, 0, 0, 0, time.FixedZone("CEST", 2*3600))

	f, err := adapter.Run(context.Background(), a, adapter.Source{Paths: []string{"a.yaml", "dir"}}, "abc123", at)

	require.NoError(t, err)
	assert.Equal(t, model.Provenance{Adapter: "k8s", Source: "a.yaml,dir", Ref: "abc123", ObservedAt: "2026-09-30T12:00:00Z"}, f.Provenance)
	assert.Equal(t, "k8s:p/deployment/a", f.Elements[0].ID)
}

func TestRun_Fails_OnAnIDWithoutTheAdapterPrefix(t *testing.T) {
	t.Parallel()
	a := stubAdapter{name: "k8s", ids: []string{"tf:aws_instance.a"}}

	_, err := adapter.Run(context.Background(), a, adapter.Source{Reader: strings.NewReader("")}, "", time.Now())

	assert.ErrorContains(t, err, `id "tf:aws_instance.a" is not a discovered id k8s:<path>`)
}

func TestRegistry_RunsAnAdapterEndToEnd_IntoAFragmentTheSystemLoads(t *testing.T) {
	t.Parallel()
	reg, err := adapter.NewRegistry(stubAdapter{name: "k8s", ids: []string{"k8s:p/deployment/a"}})
	require.NoError(t, err)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "model.yaml"), []byte("apiVersion: noodle/v1alpha1\nkind: Model\n"), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "discovered"), 0o755))

	a, err := reg.Get("k8s")
	require.NoError(t, err)
	f, err := adapter.Run(context.Background(), a, adapter.Source{Paths: []string{"manifests"}}, "", time.Unix(0, 0))
	require.NoError(t, err)
	var buf bytes.Buffer
	require.NoError(t, model.EncodeFragment(&buf, f))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "discovered", "k8s-manifests.yaml"), buf.Bytes(), 0o644))
	s, err := model.LoadSystem(dir)

	require.NoError(t, err)
	loaded := s.Fragments["k8s-manifests"]
	loaded.ID = ""
	assert.Equal(t, f, loaded)
}
