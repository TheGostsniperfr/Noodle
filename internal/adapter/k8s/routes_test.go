package k8s_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// edge is a connection or reference without its hashed id and src, for comparison.
type edge struct{ From, To, Port, Protocol string }

func conns(f *model.Fragment) []edge {
	var out []edge
	for _, c := range f.Connections {
		out = append(out, edge{c.From, c.To, c.Port, c.Protocol})
	}
	return out
}

type ref struct{ From, To, Kind string }

func refs(f *model.Fragment) []ref {
	var out []ref
	for _, r := range f.References {
		out = append(out, ref{r.From, r.To, r.Kind})
	}
	return out
}

func unresolved(f *model.Fragment, kind string) []model.Unresolved {
	var out []model.Unresolved
	for _, u := range f.Unresolved {
		if u.Kind == kind {
			u.Src = model.Src{}
			out = append(out, u)
		}
	}
	return out
}

func docs(parts ...string) string { return strings.Join(parts, "---\n") }

func workload(kind, ns, name, labels, ports string) string {
	return "apiVersion: apps/v1\nkind: " + kind + "\nmetadata: {name: " + name + ", namespace: " + ns + "}\n" +
		"spec:\n  template:\n    metadata: {labels: " + labels + "}\n    spec:\n      containers:\n        - {name: c, image: x, ports: " + ports + "}\n"
}

func service(ns, name, selector, ports string) string {
	s := "apiVersion: v1\nkind: Service\nmetadata: {name: " + name + ", namespace: " + ns + "}\nspec:\n"
	if selector != "" {
		s += "  selector: " + selector + "\n"
	}
	if ports != "" {
		s += "  ports: " + ports + "\n"
	}
	return s
}

func TestDiscover_ConnectsAServiceToTheWorkloadsItSelects(t *testing.T) {
	t.Parallel()
	api := workload("Deployment", "n", "api", "{app: api, tier: web}", "[{name: http, containerPort: 8080}, {name: metrics, containerPort: 9090}]")

	tests := []struct {
		name, input string
		want        []edge
	}{
		{"targetPort by number names the container port",
			docs(api, service("n", "api", "{app: api}", "[{name: web, port: 80, targetPort: 8080}]")),
			[]edge{{"k8s:n/service/api", "k8s:n/deployment/api", "http", ""}}},
		{"targetPort by name",
			docs(api, service("n", "api", "{app: api}", "[{name: m, port: 9090, targetPort: metrics}]")),
			[]edge{{"k8s:n/service/api", "k8s:n/deployment/api", "metrics", ""}}},
		{"no targetPort means the Service port",
			docs(api, service("n", "api", "{app: api}", "[{port: 8080}]")),
			[]edge{{"k8s:n/service/api", "k8s:n/deployment/api", "http", ""}}},
		{"a port the workload does not declare keeps its number",
			docs(api, service("n", "api", "{app: api}", "[{port: 80, targetPort: 3000}]")),
			[]edge{{"k8s:n/service/api", "k8s:n/deployment/api", "3000", ""}}},
		{"one connection per target port",
			docs(api, service("n", "api", "{app: api}", "[{name: web, port: 80, targetPort: 8080}, {name: m, port: 9090, targetPort: 9090}]")),
			[]edge{{"k8s:n/service/api", "k8s:n/deployment/api", "http", ""}, {"k8s:n/service/api", "k8s:n/deployment/api", "metrics", ""}}},
		{"a Service without ports still fronts its workload",
			docs(api, service("n", "api", "{app: api}", "")),
			[]edge{{"k8s:n/service/api", "k8s:n/deployment/api", "", ""}}},
		{"every matching workload, whatever its kind",
			docs(api, workload("StatefulSet", "n", "api-db", "{app: api}", "[]"), service("n", "api", "{app: api}", "[{port: 8080}]")),
			[]edge{{"k8s:n/service/api", "k8s:n/deployment/api", "http", ""}, {"k8s:n/service/api", "k8s:n/statefulset/api-db", "8080", ""}}},
		{"a CronJob's pods are selected through its job template",
			docs("apiVersion: batch/v1\nkind: CronJob\nmetadata: {name: job, namespace: n}\nspec:\n  jobTemplate:\n    spec:\n      template:\n        metadata: {labels: {app: job}}\n        spec: {containers: [{name: c, image: x}]}\n",
				service("n", "job", "{app: job}", "[{port: 80}]")),
			[]edge{{"k8s:n/service/job", "k8s:n/cronjob/job", "80", ""}}},
		{"a selector value that differs selects nothing",
			docs(api, service("n", "api", "{app: web}", "[{port: 80}]")), nil},
		{"a workload in another namespace is not selected",
			docs(api, service("other", "api", "{app: api}", "[{port: 80}]")), nil},
		{"a Service without selector connects nothing and is not reported",
			docs(api, "apiVersion: v1\nkind: Service\nmetadata: {name: db, namespace: n}\nspec: {type: ExternalName, externalName: db.acme.io}\n"), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, _ := discover(t, tt.input)

			assert.ElementsMatch(t, tt.want, conns(f))
		})
	}
}

func TestDiscover_ReportsSelectorsThatMatchNothing_OnceForEveryService(t *testing.T) {
	t.Parallel()
	input := docs(
		service("a", "pg", "{cnpg.io/cluster: db, role: primary}", "[{port: 5432}]"),
		service("b", "pg", "{role: primary, cnpg.io/cluster: db}", "[{port: 5432}]"),
		service("a", "other", "{app: x}", "[{port: 80}]"),
	)

	f, _ := discover(t, input)

	assert.Equal(t, []model.Unresolved{
		{Kind: "unmatched-selector", Value: "cnpg.io/cluster=db,role=primary", About: []string{"k8s:a/service/pg", "k8s:b/service/pg"}},
		{Kind: "unmatched-selector", Value: "app=x", About: []string{"k8s:a/service/other"}},
	}, unresolved(f, "unmatched-selector"))
}

const gateway = `apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata: {name: gw, namespace: infra}
spec:
  gatewayClassName: envoy
  listeners:
    - {name: https, port: 443, protocol: HTTPS}
    - {name: dns, port: 53, protocol: UDP}
`

func route(kind, parentRefs, backendRefs string) string {
	return "apiVersion: gateway.networking.k8s.io/v1\nkind: " + kind + "\nmetadata: {name: r, namespace: shop}\n" +
		"spec:\n  hostnames: [shop.acme.io, api.acme.io]\n  parentRefs: " + parentRefs + "\n  rules:\n    - backendRefs: " + backendRefs + "\n"
}

func TestDiscover_Gateway_IsAnElementListeningOnItsListeners(t *testing.T) {
	t.Parallel()

	f, _ := discover(t, gateway)

	e := element(t, f, "k8s:infra/gateway/gw")
	assert.Equal(t, "Gateway · envoy", e.Tech)
	assert.Equal(t, []model.Port{{Name: "dns", Protocol: "UDP", Number: 53}, {Name: "https", Protocol: "TCP", Number: 443}}, e.Ports)
}

func TestDiscover_Routes(t *testing.T) {
	t.Parallel()
	api := service("shop", "api", "", "[{name: http, port: 8080}, {name: admin, port: 9000}]")
	web := service("shop", "web", "", "[{name: web, port: 80}]")
	gw := "[{name: gw, namespace: infra}]"

	tests := []struct {
		name, input string
		wantConns   []edge
		wantRefs    []ref
		wantMissing []model.Unresolved
	}{
		{"an HTTPRoute joins its Gateway to its Service on the named port",
			docs(gateway, api, route("HTTPRoute", gw, "[{name: api, port: 8080}]")),
			[]edge{{"k8s:infra/gateway/gw", "k8s:shop/service/api", "http", "HTTP"}},
			[]ref{{"k8s:shop/httproute/r", "k8s:infra/gateway/gw", "parentRef"}, {"k8s:shop/httproute/r", "k8s:shop/service/api", "backendRef"}},
			nil},
		{"a GRPCRoute carries gRPC",
			docs(gateway, api, route("GRPCRoute", gw, "[{name: api, port: 9000}]")),
			[]edge{{"k8s:infra/gateway/gw", "k8s:shop/service/api", "admin", "gRPC"}},
			[]ref{{"k8s:shop/grpcroute/r", "k8s:infra/gateway/gw", "parentRef"}, {"k8s:shop/grpcroute/r", "k8s:shop/service/api", "backendRef"}},
			nil},
		{"a backendRef without port to a single-port Service",
			docs(gateway, web, route("HTTPRoute", gw, "[{name: web}]")),
			[]edge{{"k8s:infra/gateway/gw", "k8s:shop/service/web", "web", "HTTP"}},
			[]ref{{"k8s:shop/httproute/r", "k8s:infra/gateway/gw", "parentRef"}, {"k8s:shop/httproute/r", "k8s:shop/service/web", "backendRef"}},
			nil},
		{"a backendRef without port to a multi-port Service leaves the port out",
			docs(gateway, api, route("HTTPRoute", gw, "[{name: api}]")),
			[]edge{{"k8s:infra/gateway/gw", "k8s:shop/service/api", "", "HTTP"}},
			[]ref{{"k8s:shop/httproute/r", "k8s:infra/gateway/gw", "parentRef"}, {"k8s:shop/httproute/r", "k8s:shop/service/api", "backendRef"}},
			nil},
		{"a Gateway missing from the input is reported, its Service still referenced",
			docs(api, route("HTTPRoute", gw, "[{name: api, port: 8080}]")),
			nil,
			[]ref{{"k8s:shop/httproute/r", "k8s:shop/service/api", "backendRef"}},
			[]model.Unresolved{{Kind: "missing-backend", Value: "k8s:infra/gateway/gw", Hint: "gateway", About: []string{"k8s:shop/httproute/r"}}}},
		{"a Service missing from the input is reported",
			docs(gateway, route("HTTPRoute", gw, "[{name: api, port: 8080}]")),
			nil,
			[]ref{{"k8s:shop/httproute/r", "k8s:infra/gateway/gw", "parentRef"}},
			[]model.Unresolved{{Kind: "missing-backend", Value: "k8s:shop/service/api", Hint: "service", About: []string{"k8s:shop/httproute/r"}}}},
		{"a backend that is not a Service is reported",
			docs(gateway, route("HTTPRoute", gw, "[{kind: Backend, group: gateway.envoyproxy.io, name: ext}]")),
			nil,
			[]ref{{"k8s:shop/httproute/r", "k8s:infra/gateway/gw", "parentRef"}},
			[]model.Unresolved{{Kind: "missing-backend", Value: "k8s:shop/backend/ext", Hint: "not-a-service", About: []string{"k8s:shop/httproute/r"}}}},
		{"a parentRef to a Service, as a mesh route has, is not a Gateway",
			docs(api, route("HTTPRoute", "[{kind: Service, name: api}]", "[{name: api, port: 8080}]")),
			nil,
			[]ref{{"k8s:shop/httproute/r", "k8s:shop/service/api", "backendRef"}},
			nil},
		{"a parentRef without namespace is the route's own",
			docs(strings.Replace(gateway, "namespace: infra", "namespace: shop", 1), web, route("HTTPRoute", "[{name: gw}]", "[{name: web}]")),
			[]edge{{"k8s:shop/gateway/gw", "k8s:shop/service/web", "web", "HTTP"}},
			[]ref{{"k8s:shop/httproute/r", "k8s:shop/gateway/gw", "parentRef"}, {"k8s:shop/httproute/r", "k8s:shop/service/web", "backendRef"}},
			nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, _ := discover(t, tt.input)

			assert.Equal(t, tt.wantConns, conns(f))
			assert.Equal(t, tt.wantRefs, refs(f))
			assert.Equal(t, tt.wantMissing, unresolved(f, "missing-backend"))
		})
	}
}

func TestDiscover_ARouteElementNamesItsHosts(t *testing.T) {
	t.Parallel()

	f, _ := discover(t, route("HTTPRoute", "[]", "[]"))

	e := element(t, f, "k8s:shop/httproute/r")
	assert.Equal(t, "HTTPRoute", e.Tech)
	assert.Equal(t, "api.acme.io, shop.acme.io", e.Desc)
}

func TestDiscover_TwoRoutesToTheSameService_GiveOneConnection(t *testing.T) {
	t.Parallel()
	second := strings.Replace(route("HTTPRoute", "[{name: gw, namespace: infra}]", "[{name: web}]"), "name: r,", "name: r2,", 1)
	input := docs(gateway, service("shop", "web", "", "[{port: 80}]"), route("HTTPRoute", "[{name: gw, namespace: infra}]", "[{name: web}]"), second)

	f, _ := discover(t, input)

	require.Len(t, f.Connections, 1)
	assert.Equal(t, element(t, f, "k8s:shop/httproute/r").Src, f.Connections[0].Src, "the first route read")
}

func TestDiscover_Ingress(t *testing.T) {
	t.Parallel()
	api := service("shop", "api", "", "[{name: http, port: 8080}, {name: admin, port: 9000}]")
	ingress := func(spec string) string {
		return "apiVersion: networking.k8s.io/v1\nkind: Ingress\nmetadata:\n  name: ing\n  namespace: shop\n" + spec
	}
	rules := "  rules:\n    - host: shop.acme.io\n      http:\n        paths:\n          - {path: /, backend: {service: {name: api, port: {number: 8080}}}}\n"

	tests := []struct {
		name, input string
		wantConns   []edge
		wantRefs    []ref
	}{
		{"the controller behind the class sends to the Service",
			docs(api, ingress("spec:\n  ingressClassName: nginx\n"+rules)),
			[]edge{{"k8s:_cluster/ingressclass/nginx", "k8s:shop/service/api", "http", "HTTP"}},
			[]ref{{"k8s:shop/ingress/ing", "k8s:_cluster/ingressclass/nginx", "ingressClassName"}, {"k8s:shop/ingress/ing", "k8s:shop/service/api", "backend"}}},
		{"the class annotation older charts use",
			docs(api, ingress("  annotations: {kubernetes.io/ingress.class: traefik}\nspec:\n"+rules)),
			[]edge{{"k8s:_cluster/ingressclass/traefik", "k8s:shop/service/api", "http", "HTTP"}},
			[]ref{{"k8s:shop/ingress/ing", "k8s:_cluster/ingressclass/traefik", "ingressClassName"}, {"k8s:shop/ingress/ing", "k8s:shop/service/api", "backend"}}},
		{"without a class the Ingress itself sends",
			docs(api, ingress("spec:\n"+rules)),
			[]edge{{"k8s:shop/ingress/ing", "k8s:shop/service/api", "http", "HTTP"}},
			[]ref{{"k8s:shop/ingress/ing", "k8s:shop/service/api", "backend"}}},
		{"a default backend with a port name",
			docs(api, ingress("spec:\n  defaultBackend: {service: {name: api, port: {name: admin}}}\n")),
			[]edge{{"k8s:shop/ingress/ing", "k8s:shop/service/api", "admin", "HTTP"}},
			[]ref{{"k8s:shop/ingress/ing", "k8s:shop/service/api", "backend"}}},
		{"the serviceName and servicePort form",
			docs(api, ingress("spec:\n  rules:\n    - http: {paths: [{backend: {serviceName: api, servicePort: 9000}}]}\n")),
			[]edge{{"k8s:shop/ingress/ing", "k8s:shop/service/api", "admin", "HTTP"}},
			[]ref{{"k8s:shop/ingress/ing", "k8s:shop/service/api", "backend"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, _ := discover(t, tt.input)

			assert.Equal(t, tt.wantConns, conns(f))
			assert.Equal(t, tt.wantRefs, refs(f))
		})
	}
}

func TestDiscover_IngressClass_TakesItsControllerFromTheClassObject(t *testing.T) {
	t.Parallel()
	input := docs(
		"apiVersion: networking.k8s.io/v1\nkind: IngressClass\nmetadata: {name: nginx}\nspec: {controller: k8s.io/ingress-nginx}\n",
		"apiVersion: networking.k8s.io/v1\nkind: Ingress\nmetadata: {name: ing, namespace: shop}\nspec: {ingressClassName: nginx}\n",
	)

	f, _ := discover(t, input)

	e := element(t, f, "k8s:_cluster/ingressclass/nginx")
	assert.Equal(t, "IngressClass · k8s.io/ingress-nginx", e.Tech)
	assert.Equal(t, 1, e.Src.Line)
	assert.Empty(t, e.Parent)
}

func TestDiscover_EdgeIDs_AreShortStableAndDistinctPerPort(t *testing.T) {
	t.Parallel()
	input := docs(workload("Deployment", "n", "api", "{app: api}", "[{name: a, containerPort: 1}, {name: b, containerPort: 2}]"),
		service("n", "api", "{app: api}", "[{port: 1}, {port: 2}]"))

	first, _ := discover(t, input)
	again, _ := discover(t, input)

	require.Len(t, first.Connections, 2)
	assert.NotEqual(t, first.Connections[0].ID, first.Connections[1].ID)
	assert.Equal(t, first.Connections[0].ID, again.Connections[0].ID)
	assert.Regexp(t, `^k8s:c/[0-9a-f]{10}$`, first.Connections[0].ID)
}
