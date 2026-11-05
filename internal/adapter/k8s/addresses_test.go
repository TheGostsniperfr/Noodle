package k8s_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// inferredLines prints every inferred connection as "from -> to :port line object",
// sorted, so a test states the whole list at once.
func inferredLines(f *model.Fragment) []string {
	out := []string{}
	for _, c := range f.Connections {
		if !c.Inferred {
			continue
		}
		obj := "-"
		if c.Object != nil {
			obj = c.Object.Kind + "/" + c.Object.Name
		}
		out = append(out, fmt.Sprintf("%s -> %s :%s %d %s", c.From, c.To, c.Port, c.Src.Line, obj))
	}
	sort.Strings(out)
	return out
}

const postgresService = `---
apiVersion: v1
kind: Service
metadata: {name: postgres, namespace: data}
spec:
  selector: {app: postgres}
  ports: [{port: 5432}]
`

// holders place VALUE on a known line: the key of a ConfigMap read by web, an env var of
// web, and the spec of a custom resource.
var holders = map[string]struct {
	manifests string
	want      string
}{
	"configmap": {`apiVersion: v1
kind: ConfigMap
metadata: {name: settings, namespace: prod}
data:
  DB_URL: "VALUE"
---
` + webWith("          envFrom: [{configMapRef: {name: settings}}]\n"),
		"k8s:prod/deployment/web -> k8s:data/service/postgres :5432 5 ConfigMap/settings"},
	"env": {webWith("          env:\n            - name: DB_HOST\n              value: \"VALUE\"\n"),
		"k8s:prod/deployment/web -> k8s:data/service/postgres :5432 12 Deployment/web"},
	"cr": {`apiVersion: example.com/v1
kind: Database
metadata: {name: main, namespace: prod}
spec:
  connection:
    url: "VALUE"
`, "k8s:prod/database/main -> k8s:data/service/postgres :5432 6 Database/main"},
}

func webWith(containerFields string) string {
	return `apiVersion: apps/v1
kind: Deployment
metadata: {name: web, namespace: prod}
spec:
  template:
    spec:
      containers:
        - name: web
          image: acme/web:1
` + containerFields
}

func TestAddresses(t *testing.T) {
	forms := []struct{ name, host string }{
		{"svc-ns", "postgres.data"},
		{"svc-ns-svc", "postgres.data.svc"},
		{"svc-ns-svc-cluster-local", "postgres.data.svc.cluster.local"},
	}
	for _, form := range forms {
		t.Run(form.name, func(t *testing.T) {
			for _, port := range []string{"port", "noport"} {
				t.Run(port, func(t *testing.T) {
					for _, scheme := range []string{"scheme", "noscheme"} {
						t.Run(scheme, func(t *testing.T) {
							value := form.host
							if port == "port" {
								value += ":5432"
							}
							if scheme == "scheme" {
								value = "postgresql://" + value + "/app"
							}
							for _, source := range []string{"configmap", "env", "cr"} {
								t.Run(source, func(t *testing.T) {
									h := holders[source]

									f, _ := discover(t, strings.ReplaceAll(h.manifests, "VALUE", value)+postgresService)

									assert.Equal(t, []string{h.want}, inferredLines(f))
								})
							}
						})
					}
				})
			}
		})
	}

	t.Run("noport-multiport", func(t *testing.T) {
		multi := strings.Replace(postgresService, "ports: [{port: 5432}]", "ports: [{name: metrics, port: 9090}, {name: http, port: 8080}]", 1)

		f, _ := discover(t, webWith("          env: [{name: DB, value: 'postgres.data'}]\n")+multi)

		assert.Equal(t, []string{"k8s:prod/deployment/web -> k8s:data/service/postgres :http 10 Deployment/web"}, inferredLines(f))
	})

	t.Run("noport-multiport-without-http-takes-the-first", func(t *testing.T) {
		multi := strings.Replace(postgresService, "ports: [{port: 5432}]", "ports: [{name: metrics, port: 9090}, {name: admin, port: 8080}]", 1)

		f, _ := discover(t, webWith("          env: [{name: DB, value: 'postgres.data'}]\n")+multi)

		assert.Equal(t, []string{"k8s:prod/deployment/web -> k8s:data/service/postgres :metrics 10 Deployment/web"}, inferredLines(f))
	})

	t.Run("port-picks-the-service-port-with-that-number", func(t *testing.T) {
		multi := strings.Replace(postgresService, "ports: [{port: 5432}]", "ports: [{name: metrics, port: 9090}, {name: http, port: 8080}]", 1)

		f, _ := discover(t, webWith("          env: [{name: DB, value: 'postgres.data:9090'}]\n")+multi)

		assert.Equal(t, []string{"k8s:prod/deployment/web -> k8s:data/service/postgres :metrics 10 Deployment/web"}, inferredLines(f))
	})

	t.Run("unknown-service", func(t *testing.T) {
		f, _ := discover(t, webWith("          env: [{name: DB, value: 'ghost.data.svc:5432'}, {name: B, value: 'postgres.other:5432'}, {name: C, value: 'postgres.data:9999'}]\n")+postgresService)

		assert.Empty(t, inferredLines(f))
		assert.ElementsMatch(t, []string{"k8s:data/namespace/data", "k8s:data/service/postgres", "k8s:prod/deployment/web", "k8s:prod/namespace/prod"}, ids(f))
	})

	t.Run("external-host", func(t *testing.T) {
		f, _ := discover(t, webWith("          env: [{name: DB, value: 'postgres://db.example.com:5432/app'}, {name: B, value: 'postgres.data.example.com'}]\n")+postgresService)

		assert.Empty(t, inferredLines(f))
		assert.ElementsMatch(t, []string{"k8s:data/namespace/data", "k8s:data/service/postgres", "k8s:prod/deployment/web", "k8s:prod/namespace/prod"}, ids(f))
	})

	t.Run("a-secret-is-never-scanned", func(t *testing.T) {
		secret := "apiVersion: v1\nkind: Secret\nmetadata: {name: db, namespace: prod}\nstringData: {URL: 'postgres.data:5432'}\n---\n"

		f, _ := discover(t, secret+webWith("          envFrom: [{secretRef: {name: db}}]\n")+postgresService)

		assert.Empty(t, inferredLines(f))
	})

	t.Run("a-configmap-nobody-reads-yields-nothing", func(t *testing.T) {
		f, _ := discover(t, "apiVersion: v1\nkind: ConfigMap\nmetadata: {name: settings, namespace: prod}\ndata: {DB: 'postgres.data:5432'}\n"+postgresService)

		assert.Empty(t, inferredLines(f))
	})

	t.Run("a-config-key-reads-that-key-only", func(t *testing.T) {
		cm := "apiVersion: v1\nkind: ConfigMap\nmetadata: {name: settings, namespace: prod}\ndata:\n  other: 'postgres.data:5432'\n  mine: 'unrelated'\n---\n"

		f, _ := discover(t, cm+webWith("          env: [{name: M, valueFrom: {configMapKeyRef: {name: settings, key: mine}}}]\n")+postgresService)

		assert.Empty(t, inferredLines(f))
	})

	t.Run("route-hostname", func(t *testing.T) {
		manifests := gatewayAndRoute + webWith("          env: [{name: PUBLIC_URL, value: 'https://shop.example.com/cart'}]\n")

		f, _ := discover(t, manifests)

		assert.Equal(t, []string{"k8s:prod/deployment/web -> k8s:infra/gateway/shared :https 27 Deployment/web"}, inferredLines(f))
	})

	t.Run("ingress-host", func(t *testing.T) {
		ingress := `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata: {name: shop, namespace: infra}
spec:
  ingressClassName: nginx
  rules:
    - host: shop.example.com
      http: {paths: [{path: /, pathType: Prefix, backend: {service: {name: shop, port: {number: 80}}}}]}
---
`

		f, _ := discover(t, ingress+webWith("          env: [{name: PUBLIC_URL, value: 'https://user:pw@shop.example.com/'}]\n"))

		assert.Equal(t, []string{"k8s:prod/deployment/web -> k8s:_cluster/ingressclass/nginx : 19 Deployment/web"}, inferredLines(f))
	})

	t.Run("wildcard-hostname", func(t *testing.T) {
		manifests := strings.Replace(gatewayAndRoute, "shop.example.com", `"*.example.com"`, 1) +
			webWith("          env: [{name: PUBLIC_URL, value: 'http://shop.example.com'}]\n")

		f, _ := discover(t, manifests)

		assert.Equal(t, []string{"k8s:prod/deployment/web -> k8s:infra/gateway/shared :http 27 Deployment/web"}, inferredLines(f))
	})

	t.Run("an-unserved-host-yields-nothing", func(t *testing.T) {
		f, _ := discover(t, gatewayAndRoute+webWith("          env: [{name: U, value: 'https://other.example.com/'}]\n"))

		assert.Empty(t, inferredLines(f))
	})
}

const gatewayAndRoute = `apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata: {name: shared, namespace: infra}
spec:
  gatewayClassName: envoy
  listeners:
    - {name: http, port: 80, protocol: HTTP}
    - {name: https, port: 443, protocol: HTTPS}
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata: {name: shop, namespace: infra}
spec:
  parentRefs: [{name: shared}]
  hostnames: [shop.example.com]
  rules: [{backendRefs: [{name: shop, port: 80}]}]
---
`
