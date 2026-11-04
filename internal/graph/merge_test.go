package graph_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/graph"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

func port(name string, n int) model.Port { return model.Port{Name: name, Protocol: "TCP", Number: n} }

func deployment(id string, line int, ports ...model.Port) model.DiscoveredElement {
	return model.DiscoveredElement{
		Element: model.Element{ID: id, Kind: "backend", Tech: "Deployment · api:1", Ports: ports},
		Src:     model.Src{File: "manifests/api.yaml", Line: line},
		Object:  &model.Object{APIVersion: "apps/v1", Kind: "Deployment", Namespace: "prod", Name: "api"},
	}
}

func fragment(id string, elements []model.DiscoveredElement, connections ...model.DiscoveredConnection) *model.Fragment {
	return &model.Fragment{ID: id, Elements: elements, Connections: connections}
}

func TestMerge(t *testing.T) {
	t.Run("fills-empty-field", func(t *testing.T) {
		m := &model.Model{Elements: []model.Element{{ID: "api", Kind: "backend", Title: "API", Matches: []string{"k8s:prod/service/api"}}}}
		f := fragment("k8s-a", []model.DiscoveredElement{deployment("k8s:prod/service/api", 24, port("https", 8080))})

		g := graph.Merge(m, f)

		assert.Equal(t, []model.Element{{ID: "api", Kind: "backend", Title: "API", Tech: "Deployment · api:1",
			Ports: []model.Port{port("https", 8080)}, Matches: []string{"k8s:prod/service/api"}}}, g.Model.Elements)
		assert.Empty(t, m.Elements[0].Ports, "the curated model is not modified")
	})

	t.Run("keeps-curated-field", func(t *testing.T) {
		m := &model.Model{Elements: []model.Element{{ID: "api", Kind: "backend", Tech: "Go", Ports: []model.Port{port("grpc", 9000)},
			Matches: []string{"k8s:prod/deployment/api"}}}}
		f := fragment("k8s-a", []model.DiscoveredElement{deployment("k8s:prod/deployment/api", 1, port("http", 8080))})

		g := graph.Merge(m, f)

		assert.Equal(t, []model.Element{{ID: "api", Kind: "backend", Tech: "Go", Ports: []model.Port{port("grpc", 9000)},
			Matches: []string{"k8s:prod/deployment/api"}}}, g.Model.Elements)
	})

	t.Run("takes-list-whole", func(t *testing.T) {
		m := &model.Model{Elements: []model.Element{{ID: "api", Kind: "backend", Matches: []string{"k8s:prod/deployment/none", "k8s:prod/deployment/api", "k8s:prod/service/api"}}}}
		none := deployment("k8s:prod/deployment/none", 1)
		first := deployment("k8s:prod/deployment/api", 5, port("http", 8080), port("metrics", 9090))
		second := deployment("k8s:prod/service/api", 9, port("https", 443))
		second.Tags = []string{"edge"}

		g := graph.Merge(m, fragment("k8s-a", []model.DiscoveredElement{none, first, second}))

		assert.Equal(t, []model.Port{port("http", 8080), port("metrics", 9090)}, g.Model.Elements[0].Ports,
			"the first matched element that has ports, whole, not merged with the next")
		assert.Equal(t, []string{"edge"}, g.Model.Elements[0].Tags)
	})

	t.Run("keeps-sources", func(t *testing.T) {
		m := &model.Model{Elements: []model.Element{{ID: "api", Kind: "backend", Matches: []string{"k8s:prod/deployment/api", "k8s:prod/service/api", "k8s:prod/gone"}}}}
		dep := deployment("k8s:prod/deployment/api", 5)
		svc := deployment("k8s:prod/service/api", 24)
		svc.Object = &model.Object{APIVersion: "v1", Kind: "Service", Namespace: "prod", Name: "api"}

		g := graph.Merge(m, fragment("k8s-a", []model.DiscoveredElement{dep, svc}))

		assert.Equal(t, []graph.Source{
			{ID: "k8s:prod/deployment/api", Src: model.Src{File: "manifests/api.yaml", Line: 5}, Object: dep.Object},
			{ID: "k8s:prod/service/api", Src: model.Src{File: "manifests/api.yaml", Line: 24}, Object: svc.Object},
		}, g.Sources["api"])
	})

	t.Run("maps-connection-to-curated-ids", func(t *testing.T) {
		m := &model.Model{
			Elements: []model.Element{
				{ID: "web", Kind: "frontend", Matches: []string{"k8s:prod/deployment/web"}},
				{ID: "api", Kind: "backend", Ports: []model.Port{port("http", 8080)}, Matches: []string{"k8s:prod/deployment/api", "k8s:prod/service/api"}},
				{ID: "db", Kind: "database", Matches: []string{"k8s:data/service/db"}},
			},
			Connections: []model.Connection{{ID: "c-web-api", From: "web", To: "api", Kind: "flow"}},
		}
		elements := []model.DiscoveredElement{
			deployment("k8s:prod/deployment/web", 1), deployment("k8s:prod/deployment/api", 2), deployment("k8s:prod/service/api", 3),
			deployment("k8s:data/service/db", 4), deployment("k8s:prod/deployment/worker", 5),
		}
		at := func(line int) model.Src { return model.Src{File: "manifests/api.yaml", Line: line} }
		conn := func(id, from, to, port string, line int) model.DiscoveredConnection {
			return model.DiscoveredConnection{Connection: model.Connection{ID: id, From: from, To: to, Port: port, Kind: "flow"}, Src: at(line)}
		}
		f := fragment("k8s-a", elements,
			conn("k8s:c/a", "k8s:prod/deployment/web", "k8s:prod/service/api", "https", 10),
			conn("k8s:c/b", "k8s:prod/deployment/api", "k8s:data/service/db", "5432", 11),
			conn("k8s:c/c", "k8s:prod/service/api", "k8s:prod/deployment/api", "http", 12),
			conn("k8s:c/d", "k8s:prod/deployment/worker", "k8s:prod/service/api", "http", 13),
			conn("k8s:c/e", "k8s:prod/deployment/web", "k8s:prod/deployment/api", "http", 14),
		)

		g := graph.Merge(m, f)

		assert.Equal(t, []model.Connection{
			{ID: "c-web-api", From: "web", To: "api", Kind: "flow"},
			{ID: "k8s:c/b", From: "api", To: "db", Kind: "flow"},
		}, g.Model.Connections, "a connection between curated ends is added once, one the model has is not repeated, an inner or unmatched one is left out")
		assert.Equal(t, []graph.Source{{ID: "k8s:c/a", Src: at(10)}, {ID: "k8s:c/e", Src: at(14)}}, g.Sources["c-web-api"])
		assert.Equal(t, []graph.Source{{ID: "k8s:c/b", Src: at(11)}}, g.Sources["k8s:c/b"])
	})

	t.Run("keeps-the-port-the-curated-end-has", func(t *testing.T) {
		m := &model.Model{Elements: []model.Element{
			{ID: "web", Kind: "frontend", Matches: []string{"k8s:prod/deployment/web"}},
			{ID: "api", Kind: "backend", Ports: []model.Port{port("http", 8080)}, Matches: []string{"k8s:prod/service/api"}},
		}}
		conn := model.DiscoveredConnection{Connection: model.Connection{ID: "k8s:c/a", From: "k8s:prod/deployment/web", To: "k8s:prod/service/api", Port: "http", Kind: "flow"}}
		f := fragment("k8s-a", []model.DiscoveredElement{deployment("k8s:prod/deployment/web", 1), deployment("k8s:prod/service/api", 2)}, conn)

		g := graph.Merge(m, f)

		assert.Equal(t, []model.Connection{{ID: "k8s:c/a", From: "web", To: "api", Port: "http", Kind: "flow"}}, g.Model.Connections)
	})

	t.Run("does-not-depend-on-the-order-of-fragments", func(t *testing.T) {
		m := &model.Model{Elements: []model.Element{{ID: "api", Kind: "backend", Matches: []string{"k8s:prod/deployment/api"}}}}
		a := fragment("k8s-a", []model.DiscoveredElement{deployment("k8s:prod/deployment/api", 1, port("a", 1))})
		b := fragment("k8s-b", []model.DiscoveredElement{deployment("k8s:prod/deployment/api", 2, port("b", 2))})

		assert.Equal(t, graph.Merge(m, a, b).Model, graph.Merge(m, b, a).Model)
		assert.Equal(t, []model.Port{port("a", 1)}, graph.Merge(m, b, a).Model.Elements[0].Ports)
	})
}

// scenario is a curated model and the fragment matched to it, both drawn from a seed.
func scenario(seed int64) (*model.Model, *model.Fragment) {
	r := rand.New(rand.NewSource(seed))
	pick := func(prefix string) string {
		if r.Intn(3) == 0 {
			return ""
		}
		return fmt.Sprintf("%s%d", prefix, r.Intn(3))
	}
	randPorts := func() []model.Port {
		var ports []model.Port
		for range r.Intn(3) {
			ports = append(ports, port(fmt.Sprintf("p%d", r.Intn(4)), 1+r.Intn(9000)))
		}
		return ports
	}
	m, f := &model.Model{}, &model.Fragment{ID: "k8s-a"}
	for i := range 1 + r.Intn(4) {
		id := fmt.Sprintf("k8s:prod/deployment/e%d", i)
		curated := model.Element{ID: fmt.Sprintf("e%d", i), Kind: "backend", Title: pick("title"), Tech: pick("tech"), Desc: pick("desc"),
			Icon: pick("icon"), Multiplicity: pick("mult"), Ports: randPorts(), Tags: []string{pick("tag")}, Matches: []string{id}}
		if curated.Tech == "" {
			curated.Tech = "curated"
		}
		if len(curated.Ports) == 0 {
			curated.Ports = []model.Port{port("curated", 1)}
		}
		m.Elements = append(m.Elements, curated)
		found := model.Element{ID: id, Kind: "database", Title: pick("title"), Tech: pick("tech"), Desc: pick("desc"), Icon: pick("icon"),
			Multiplicity: pick("mult"), Ports: randPorts(), Tags: []string{pick("tag")}}
		f.Elements = append(f.Elements, model.DiscoveredElement{Element: found, Src: model.Src{File: "f", Line: i + 1}})
	}
	return m, f
}

func TestMerge_Property_NeverOverridesACuratedField(t *testing.T) {
	setByBoth := 0
	neverOverrides := func(seed int64) bool {
		m, f := scenario(seed)
		before := append([]model.Element{}, m.Elements...)
		for i := range before {
			before[i].Ports = append([]model.Port{}, before[i].Ports...)
			before[i].Tags = append([]string{}, before[i].Tags...)
		}

		g := graph.Merge(m, f)

		for i, e := range g.Model.Elements {
			was, now, found := reflect.ValueOf(before[i]), reflect.ValueOf(e), reflect.ValueOf(f.Elements[i].Element)
			for n := range was.NumField() {
				if was.Field(n).IsZero() {
					continue
				}
				if !reflect.DeepEqual(was.Field(n).Interface(), now.Field(n).Interface()) {
					return false
				}
				if !found.Field(n).IsZero() && n != 0 {
					setByBoth++
				}
			}
		}
		return true
	}

	err := quick.Check(neverOverrides, &quick.Config{MaxCount: 500})

	assert.NoError(t, err)
	assert.Positive(t, setByBoth, "no case had a field set on both sides, the property tested nothing")
}
