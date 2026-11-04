package k8s_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

func TestIdentity(t *testing.T) {
	t.Run("scenario", func(t *testing.T) {
		manifests := `apiVersion: apps/v1
kind: Deployment
metadata: {name: api, namespace: prod}
spec:
  template:
    spec:
      containers: [{name: api, image: acme/api:1}]
`

		f, _ := discover(t, manifests)

		var got []model.DiscoveredElement
		for _, e := range f.Elements {
			got = append(got, model.DiscoveredElement{Element: model.Element{ID: e.ID}, Object: e.Object})
		}
		assert.Equal(t, []model.DiscoveredElement{
			{Element: model.Element{ID: "k8s:prod/deployment/api"},
				Object: &model.Object{APIVersion: "apps/v1", Kind: "Deployment", Namespace: "prod", Name: "api"}},
			{Element: model.Element{ID: "k8s:prod/namespace/prod"}},
		}, got)
	})

	t.Run("declared-namespace-and-cluster-scoped", func(t *testing.T) {
		manifests := `apiVersion: v1
kind: Namespace
metadata: {name: prod}
---
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata: {name: nginx}
spec: {controller: k8s.io/ingress-nginx}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata: {name: web, namespace: prod}
spec:
  ingressClassName: nginx
  defaultBackend: {service: {name: web, port: {number: 80}}}
---
apiVersion: v1
kind: Service
metadata: {name: web, namespace: prod}
spec: {ports: [{port: 80}]}
`

		f, _ := discover(t, manifests)

		objects := map[string]*model.Object{}
		for _, e := range f.Elements {
			objects[e.ID] = e.Object
		}
		assert.Equal(t, map[string]*model.Object{
			"k8s:prod/namespace/prod":         {APIVersion: "v1", Kind: "Namespace", Name: "prod"},
			"k8s:_cluster/ingressclass/nginx": {APIVersion: "networking.k8s.io/v1", Kind: "IngressClass", Name: "nginx"},
			"k8s:prod/ingress/web":            {APIVersion: "networking.k8s.io/v1", Kind: "Ingress", Namespace: "prod", Name: "web"},
			"k8s:prod/service/web":            {APIVersion: "v1", Kind: "Service", Namespace: "prod", Name: "web"},
		}, objects)
	})

	t.Run("edges-carry-the-object-that-created-them", func(t *testing.T) {
		manifests := `apiVersion: v1
kind: Service
metadata: {name: web, namespace: prod}
spec: {selector: {app: web}, ports: [{port: 80}]}
---
apiVersion: apps/v1
kind: Deployment
metadata: {name: web, namespace: prod}
spec:
  template:
    metadata: {labels: {app: web}}
    spec: {containers: [{name: web, image: web}]}
`

		f, _ := discover(t, manifests)

		var objects []*model.Object
		for _, c := range f.Connections {
			objects = append(objects, c.Object)
		}
		assert.Equal(t, []*model.Object{{APIVersion: "v1", Kind: "Service", Namespace: "prod", Name: "web"}}, objects)
	})

	t.Run("connection-object", func(t *testing.T) {
		h := holders["configmap"]

		f, _ := discover(t, strings.ReplaceAll(h.manifests, "VALUE", "postgres.data.svc:5432")+postgresService)

		var objects []*model.Object
		for _, c := range f.Connections {
			if c.Inferred {
				objects = append(objects, c.Object)
			}
		}
		assert.Equal(t, []*model.Object{{APIVersion: "v1", Kind: "ConfigMap", Namespace: "prod", Name: "settings"}}, objects)
	})
}
