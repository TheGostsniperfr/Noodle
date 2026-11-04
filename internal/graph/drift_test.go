package graph_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/graph"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

func element(id string, line int) model.DiscoveredElement {
	return model.DiscoveredElement{Element: model.Element{ID: id, Kind: "backend"}, Src: model.Src{File: "m.yaml", Line: line}}
}

func flow(id, from, to string, line int) model.DiscoveredConnection {
	return model.DiscoveredConnection{Connection: model.Connection{ID: id, From: from, To: to, Kind: "flow"}, Src: model.Src{File: "m.yaml", Line: line}}
}

func TestDrift(t *testing.T) {
	curatedSrc := model.Src{File: "model.yaml"}

	t.Run("unmatched", func(t *testing.T) {
		m := &model.Model{Elements: []model.Element{{ID: "api", Kind: "backend", Matches: []string{"k8s:prod/deployment/api"}}}}
		zone := model.DiscoveredElement{Element: model.Element{ID: "k8s:prod/namespace/prod", Kind: "group"}, Src: model.Src{File: "m.yaml", Line: 1}}
		f := fragment("k8s-a", []model.DiscoveredElement{zone, element("k8s:prod/deployment/api", 2), element("k8s:prod/deployment/worker", 3)})

		r := graph.Merge(m, f).Drift()

		assert.Equal(t, []graph.Entry{{ID: "k8s:prod/deployment/worker", Src: model.Src{File: "m.yaml", Line: 3}}}, r.Unmatched,
			"a discovered component no element matches; a namespace zone is not one")
	})

	t.Run("gone", func(t *testing.T) {
		m := &model.Model{Elements: []model.Element{
			{ID: "api", Kind: "backend", Matches: []string{"k8s:prod/deployment/api", "k8s:prod/service/api"}},
			{ID: "secret-holder", Kind: "backend", Matches: []string{"k8s:prod/secret/db"}},
		}}
		f := fragment("k8s-a", []model.DiscoveredElement{element("k8s:prod/service/api", 2)})
		f.References = []model.DiscoveredReference{{Reference: model.Reference{ID: "k8s:r/1", From: "k8s:prod/service/api", To: "k8s:prod/secret/db"}, Src: model.Src{File: "m.yaml", Line: 9}}}

		r := graph.Merge(m, f).Drift()

		assert.Equal(t, []graph.Entry{{ID: "api", Detail: "k8s:prod/deployment/api", Src: curatedSrc}}, r.Gone,
			"an id some fragment still holds, even only as the end of an edge, is not gone")
	})

	t.Run("planned-absent", func(t *testing.T) {
		m := &model.Model{Elements: []model.Element{
			{ID: "search", Kind: "backend", Status: "planned"},
			{ID: "cache", Kind: "backend", Status: "planned", Matches: []string{"k8s:prod/deployment/cache"}},
			{ID: "queue", Kind: "backend", Status: "planned", Matches: []string{"k8s:prod/deployment/queue"}},
		}}
		f := fragment("k8s-a", []model.DiscoveredElement{element("k8s:prod/deployment/queue", 2)})

		r := graph.Merge(m, f).Drift()

		assert.Equal(t, []graph.Entry{{ID: "cache", Src: curatedSrc}, {ID: "search", Src: curatedSrc}}, r.Planned,
			"a planned element is absent when none of its matches is in a fragment")
		assert.Empty(t, r.Gone, "a planned element that was never built is not gone")
	})

	t.Run("unmodelled-connection", func(t *testing.T) {
		m := &model.Model{
			Elements: []model.Element{
				{ID: "web", Kind: "frontend", Matches: []string{"k8s:prod/deployment/web"}},
				{ID: "api", Kind: "backend", Matches: []string{"k8s:prod/deployment/api", "k8s:prod/service/api"}},
				{ID: "db", Kind: "database", Matches: []string{"k8s:data/service/db"}},
			},
			Connections: []model.Connection{{ID: "c-web-api", From: "web", To: "api", Kind: "flow"}},
		}
		f := fragment("k8s-a", []model.DiscoveredElement{element("k8s:prod/deployment/web", 1), element("k8s:prod/deployment/api", 2),
			element("k8s:prod/service/api", 3), element("k8s:data/service/db", 4), element("k8s:prod/deployment/worker", 5)},
			flow("k8s:c/1", "k8s:prod/deployment/web", "k8s:prod/service/api", 10),
			flow("k8s:c/2", "k8s:prod/deployment/api", "k8s:data/service/db", 11),
			flow("k8s:c/3", "k8s:prod/deployment/api", "k8s:data/service/db", 12),
			flow("k8s:c/4", "k8s:prod/deployment/worker", "k8s:data/service/db", 13),
			flow("k8s:c/5", "k8s:prod/service/api", "k8s:prod/deployment/api", 14),
		)

		r := graph.Merge(m, f).Drift()

		assert.Equal(t, []graph.Entry{{ID: "api -> db", Detail: "k8s:prod/deployment/api -> k8s:data/service/db", Src: model.Src{File: "m.yaml", Line: 11}}}, r.Unmodelled,
			"one entry per pair of curated elements, the model's own connection, an inner one and one with an unmatched end left out")
	})

	t.Run("sorted", func(t *testing.T) {
		m := &model.Model{Elements: []model.Element{
			{ID: "b-gone", Kind: "backend", Matches: []string{"k8s:p/z", "k8s:p/a"}},
			{ID: "a-gone", Kind: "backend", Matches: []string{"k8s:p/y"}},
			{ID: "z-plan", Kind: "backend", Status: "planned"},
			{ID: "a-plan", Kind: "backend", Status: "planned"},
			{ID: "x", Kind: "backend", Matches: []string{"k8s:p/x"}},
			{ID: "y", Kind: "backend", Matches: []string{"k8s:p/y2"}},
			{ID: "w", Kind: "backend", Matches: []string{"k8s:p/w"}},
		}}
		f := fragment("k8s-a", []model.DiscoveredElement{element("k8s:p/zz", 1), element("k8s:p/aa", 2), element("k8s:p/x", 3), element("k8s:p/y2", 4), element("k8s:p/w", 5)},
			flow("k8s:c/2", "k8s:p/y2", "k8s:p/w", 10), flow("k8s:c/1", "k8s:p/x", "k8s:p/w", 11))

		r := graph.Merge(m, f).Drift()

		assert.Equal(t, []graph.Entry{{ID: "k8s:p/aa", Src: model.Src{File: "m.yaml", Line: 2}}, {ID: "k8s:p/zz", Src: model.Src{File: "m.yaml", Line: 1}}}, r.Unmatched)
		assert.Equal(t, []graph.Entry{{ID: "a-gone", Detail: "k8s:p/y", Src: curatedSrc}, {ID: "b-gone", Detail: "k8s:p/a", Src: curatedSrc}, {ID: "b-gone", Detail: "k8s:p/z", Src: curatedSrc}}, r.Gone)
		assert.Equal(t, []graph.Entry{{ID: "a-plan", Src: curatedSrc}, {ID: "z-plan", Src: curatedSrc}}, r.Planned)
		assert.Equal(t, []graph.Entry{
			{ID: "x -> w", Detail: "k8s:p/x -> k8s:p/w", Src: model.Src{File: "m.yaml", Line: 11}},
			{ID: "y -> w", Detail: "k8s:p/y2 -> k8s:p/w", Src: model.Src{File: "m.yaml", Line: 10}},
		}, r.Unmodelled)
	})
}

func TestCrossFragment(t *testing.T) {
	route := model.DiscoveredElement{Element: model.Element{ID: "k8s:apps/httproute/api", Kind: "backend"}, Src: model.Src{File: "apps.yaml", Line: 3}}
	gateway := element("k8s:infra/gateway/shared", 7)
	apps := fragment("k8s-apps", []model.DiscoveredElement{route})
	apps.Unresolved = []model.Unresolved{
		{Kind: "missing-backend", Value: "k8s:infra/gateway/shared", Hint: "gateway", About: []string{"k8s:apps/httproute/api"}, Src: model.Src{File: "apps.yaml", Line: 3}},
		{Kind: "missing-backend", Value: "k8s:apps/service/ghost", Hint: "service", About: []string{"k8s:apps/httproute/api"}, Src: model.Src{File: "apps.yaml", Line: 3}},
	}
	infra := fragment("k8s-gateway-infra", []model.DiscoveredElement{gateway})

	g := graph.Merge(&model.Model{}, apps, infra)

	assert.Equal(t, []model.Unresolved{apps.Unresolved[1]}, g.Unresolved, "only the target no fragment defines stays missing")
	assert.Equal(t, []graph.Edge{{ID: "xref:k8s:apps/httproute/api->k8s:infra/gateway/shared", From: "k8s:apps/httproute/api", To: "k8s:infra/gateway/shared",
		Kind: "parentRef", Reference: true, Src: model.Src{File: "apps.yaml", Line: 3}}}, g.Neighbours("k8s:infra/gateway/shared", ""),
		"the route and its Gateway are linked")
}
