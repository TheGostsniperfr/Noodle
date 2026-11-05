package graph_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/graph"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

func lookupGraph() *graph.Graph {
	m := &model.Model{Elements: []model.Element{{ID: "api", Kind: "backend", Matches: []string{"k8s:prod/deployment/api"}}}}
	at := func(line int) model.Src { return model.Src{File: "m.yaml", Line: line} }
	ref := func(id, from, to, kind string, line int) model.DiscoveredReference {
		return model.DiscoveredReference{Reference: model.Reference{ID: id, From: from, To: to, Kind: kind}, Src: at(line)}
	}
	f := &model.Fragment{ID: "k8s-a",
		Elements: []model.DiscoveredElement{deployment("k8s:prod/deployment/api", 1), deployment("k8s:prod/deployment/worker", 2)},
		References: []model.DiscoveredReference{
			ref("k8s:r/3", "k8s:prod/deployment/worker", "k8s:prod/secret/db", "secret-key", 30),
			ref("k8s:r/2", "k8s:prod/deployment/api", "k8s:prod/secret/db", "secret-key", 20),
			ref("k8s:r/1", "k8s:prod/deployment/api", "k8s:prod/secret/db", "mounts", 21),
			ref("k8s:r/4", "k8s:prod/deployment/api", "k8s:prod/configmap/cfg", "env-from", 22),
		},
		Connections: []model.DiscoveredConnection{{Connection: model.Connection{ID: "k8s:c/1", From: "k8s:prod/deployment/api", To: "k8s:prod/deployment/worker", Kind: "flow"}, Src: at(40)}},
	}
	return graph.Merge(m, f)
}

func TestLookup(t *testing.T) {
	g := lookupGraph()

	tests := []struct {
		name   string
		id     string
		want   graph.Node
		wantOK bool
	}{
		{"a curated id gives the merged element and its sources", "api",
			graph.Node{ID: "api", Element: model.Element{ID: "api", Kind: "backend", Tech: "Deployment · api:1", Matches: []string{"k8s:prod/deployment/api"}},
				Sources: []graph.Source{{ID: "k8s:prod/deployment/api", Src: model.Src{File: "manifests/api.yaml", Line: 1}, Object: &model.Object{APIVersion: "apps/v1", Kind: "Deployment", Namespace: "prod", Name: "api"}}}}, true},
		{"a discovered element id", "k8s:prod/deployment/worker",
			graph.Node{ID: "k8s:prod/deployment/worker", Element: model.Element{ID: "k8s:prod/deployment/worker", Kind: "backend", Tech: "Deployment · api:1"},
				Sources: []graph.Source{{ID: "k8s:prod/deployment/worker", Src: model.Src{File: "manifests/api.yaml", Line: 2}, Object: &model.Object{APIVersion: "apps/v1", Kind: "Deployment", Namespace: "prod", Name: "api"}}}}, true},
		{"an id only an edge names", "k8s:prod/secret/db", graph.Node{ID: "k8s:prod/secret/db", Element: model.Element{ID: "k8s:prod/secret/db"}}, true},
		{"an unknown id", "k8s:prod/secret/none", graph.Node{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := g.Lookup(tt.id)

			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNeighbours(t *testing.T) {
	g := lookupGraph()
	line := func(edges []graph.Edge) []string {
		out := []string{}
		for _, e := range edges {
			out = append(out, e.Kind+" "+e.From+" -> "+e.To)
		}
		return out
	}

	tests := []struct {
		name, id, kind string
		want           []string
	}{
		{"of a Secret, filtered by kind", "k8s:prod/secret/db", "secret-key", []string{
			"secret-key k8s:prod/deployment/api -> k8s:prod/secret/db", "secret-key k8s:prod/deployment/worker -> k8s:prod/secret/db"}},
		{"of a Secret, every kind", "k8s:prod/secret/db", "", []string{
			"mounts k8s:prod/deployment/api -> k8s:prod/secret/db", "secret-key k8s:prod/deployment/api -> k8s:prod/secret/db", "secret-key k8s:prod/deployment/worker -> k8s:prod/secret/db"}},
		{"of a curated id, through what it matches", "api", "flow", []string{"flow k8s:prod/deployment/api -> k8s:prod/deployment/worker"}},
		{"of an id nothing touches", "k8s:prod/secret/none", "", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, line(g.Neighbours(tt.id, tt.kind)))
		})
	}

	t.Run("an edge keeps its source", func(t *testing.T) {
		edges := g.Neighbours("k8s:prod/secret/db", "mounts")

		assert.Equal(t, []graph.Edge{{ID: "k8s:r/1", From: "k8s:prod/deployment/api", To: "k8s:prod/secret/db", Kind: "mounts", Reference: true, Src: model.Src{File: "m.yaml", Line: 21}}}, edges)
	})
}
