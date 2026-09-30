package k8s_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/TheGostsniperfr/Noodle/internal/adapter"
	"github.com/TheGostsniperfr/Noodle/internal/adapter/k8s"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

func discover(t *testing.T, manifests string) (*model.Fragment, adapter.Stats) {
	t.Helper()
	f, stats, err := k8s.Adapter{}.Discover(context.Background(), adapter.Source{Reader: strings.NewReader(manifests)})
	require.NoError(t, err)
	return f, stats
}

func ids(f *model.Fragment) []string {
	var out []string
	for _, e := range f.Elements {
		out = append(out, e.ID)
	}
	return out
}

func element(t *testing.T, f *model.Fragment, id string) model.DiscoveredElement {
	t.Helper()
	for _, e := range f.Elements {
		if e.ID == id {
			return e
		}
	}
	require.Failf(t, "element not found", "%s in %v", id, ids(f))
	return model.DiscoveredElement{}
}

const deployment = `apiVersion: apps/v1
kind: Deployment
metadata: {name: api, namespace: shop}
spec:
  template:
    spec:
      containers:
        - {name: api, image: ghcr.io/acme/api:1.4.2, ports: [{name: http, containerPort: 8080}]}
`

func TestDiscover_ReadsEveryDocumentShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantIDs   []string
		wantLines []int
	}{
		{"multi-document with a leading separator, empty and comment-only documents",
			"---\n# rendered\n---\n" + deployment + "---\n---\n# end\n",
			[]string{"k8s:shop/deployment/api", "k8s:shop/namespace/shop"}, []int{4, 4}},
		{"a kind List keeps each item's line",
			"apiVersion: v1\nkind: List\nitems:\n  - apiVersion: v1\n    kind: Service\n    metadata: {name: a, namespace: n}\n  - apiVersion: v1\n    kind: Service\n    metadata: {name: b, namespace: n}\n",
			[]string{"k8s:n/service/a", "k8s:n/service/b", "k8s:n/namespace/n"}, []int{4, 7, 4}},
		{"a typed list such as ServiceList",
			"apiVersion: v1\nkind: ServiceList\nitems:\n  - {apiVersion: v1, kind: Service, metadata: {name: a, namespace: n}}\n",
			[]string{"k8s:n/service/a", "k8s:n/namespace/n"}, []int{4, 4}},
		{"JSON, as kubectl get -o json prints it",
			`{"apiVersion": "v1", "kind": "List", "items": [{"apiVersion": "v1", "kind": "Service", "metadata": {"name": "a", "namespace": "n"}}]}`,
			[]string{"k8s:n/service/a", "k8s:n/namespace/n"}, []int{1, 1}},
		{"anchors and merge keys resolve",
			"apiVersion: v1\nkind: Service\nmetadata: &m {name: a, namespace: n}\n---\napiVersion: v1\nkind: Service\nmetadata:\n  <<: *m\n  name: b\n",
			[]string{"k8s:n/service/a", "k8s:n/service/b", "k8s:n/namespace/n"}, []int{1, 5, 1}},
		{"a name written as a number",
			"apiVersion: v1\nkind: Service\nmetadata: {name: 404, namespace: n}\n",
			[]string{"k8s:n/service/404", "k8s:n/namespace/n"}, []int{1, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, _ := discover(t, tt.input)

			assert.Equal(t, tt.wantIDs, ids(f))
			var lines []int
			for _, e := range f.Elements {
				lines = append(lines, e.Src.Line)
			}
			assert.Equal(t, tt.wantLines, lines)
		})
	}
}

func TestDiscover_KeepsTheLastValue_OfAKeyARenderedChartRepeats(t *testing.T) {
	t.Parallel()
	input := "apiVersion: v1\nkind: Service\nmetadata: {name: a, namespace: n}\nspec:\n  type: ClusterIP\n  type: NodePort\n"

	f, _ := discover(t, input)

	assert.Equal(t, "Service · NodePort", element(t, f, "k8s:n/service/a").Tech)
}

func TestDiscover_Fails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, input, wantErr string
	}{
		{"on an unrendered template, saying how to render it", "apiVersion: v1\nkind: Service\nmetadata:\n  name: {{ .Release.Name }}\n  labels: {{- include \"x\" . }}\n", "render it first"},
		{"on YAML that does not parse, naming the input", "kind: Service\n  metadata: [\n", "-: yaml:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := k8s.Adapter{}.Discover(context.Background(), adapter.Source{Reader: strings.NewReader(tt.input)})

			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestDiscover_Fails_OnAMissingPath(t *testing.T) {
	t.Parallel()

	_, _, err := k8s.Adapter{}.Discover(context.Background(), adapter.Source{Paths: []string{"testdata/nope"}})

	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestDiscover_LeavesOutWhatNoOneDeploys(t *testing.T) {
	t.Parallel()
	svc := func(name, extra string) string {
		return "---\napiVersion: v1\nkind: Service\nmetadata:\n  name: " + name + "\n  namespace: n\n" + extra
	}

	tests := []struct {
		name       string
		input      string
		wantIDs    []string
		wantReason string
	}{
		{"an object owned by a controller", svc("a", "  ownerReferences: [{kind: ReplicaSet, name: x, controller: true}]\n"), nil, "owned"},
		{"an owner reference that is not a controller keeps the object", svc("a", "  ownerReferences: [{kind: Foo, name: x}]\n"), []string{"k8s:n/service/a", "k8s:n/namespace/n"}, ""},
		{"an EndpointSlice", "apiVersion: discovery.k8s.io/v1\nkind: EndpointSlice\nmetadata: {name: a, namespace: n}\n", nil, "generated"},
		{"a Helm test hook", svc("a", "  annotations: {helm.sh/hook: test-success}\n"), nil, "helm-hook"},
		{"a Helm delete hook", svc("a", "  annotations: {helm.sh/hook: \"pre-delete,post-delete\"}\n"), nil, "helm-hook"},
		{"a Helm install hook such as a migration is kept", svc("a", "  annotations: {helm.sh/hook: \"pre-install, pre-upgrade\"}\n"), []string{"k8s:n/service/a", "k8s:n/namespace/n"}, ""},
		{"a hook that runs on install and on delete is kept", svc("a", "  annotations: {helm.sh/hook: \"post-install,pre-delete\"}\n"), []string{"k8s:n/service/a", "k8s:n/namespace/n"}, ""},
		{"a document without kind, such as values", "replicaCount: 2\nimage: {repository: x}\n", nil, "no-kind"},
		{"the same object twice keeps the first", svc("a", "spec: {type: NodePort}\n") + svc("a", ""), []string{"k8s:n/service/a", "k8s:n/namespace/n"}, "duplicate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, stats := discover(t, tt.input)

			assert.Equal(t, tt.wantIDs, ids(f))
			if tt.wantReason != "" {
				assert.Equal(t, 1, stats.Skipped[tt.wantReason], "skipped %v", stats.Skipped)
			}
		})
	}
}

func TestDiscover_CountsCRDSchemasAsNoise(t *testing.T) {
	t.Parallel()
	crd := "apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\nmetadata: {name: things.acme.io}\nspec: {group: acme.io}\n"
	input := deployment + "---\n" + crd

	f, stats := discover(t, input)

	assert.Equal(t, len(input), stats.InputBytes)
	assert.Equal(t, len(crd), stats.NoiseBytes)
	assert.Equal(t, 1, stats.Skipped["crd"])
	assert.Empty(t, f.Unresolved)
}

func TestDiscover_CountsACRDInsideAList_UpToTheNextItem(t *testing.T) {
	t.Parallel()
	crdItem := "  - apiVersion: apiextensions.k8s.io/v1\n    kind: CustomResourceDefinition\n    metadata: {name: things.acme.io}\n"
	input := "apiVersion: v1\nkind: List\nitems:\n" + crdItem + "  - {apiVersion: v1, kind: Service, metadata: {name: a, namespace: n}}\n"

	_, stats := discover(t, input)

	assert.Equal(t, len(crdItem), stats.NoiseBytes)
}

func TestDiscover_SameObjectsFromManifestsAndFromKubectl(t *testing.T) {
	t.Parallel()
	live := `apiVersion: v1
kind: List
items:
  - apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: api
      namespace: shop
      uid: 5c1f
      resourceVersion: "4411"
      managedFields: [{manager: kubectl, operation: Update}]
    spec:
      template:
        spec:
          containers:
            - {name: api, image: ghcr.io/acme/api:1.4.2, ports: [{name: http, containerPort: 8080}]}
    status: {replicas: 2, readyReplicas: 2}
  - apiVersion: apps/v1
    kind: ReplicaSet
    metadata:
      name: api-7d9f
      namespace: shop
      ownerReferences: [{apiVersion: apps/v1, kind: Deployment, name: api, controller: true}]
  - apiVersion: v1
    kind: Pod
    metadata:
      name: api-7d9f-x2k
      namespace: shop
      ownerReferences: [{apiVersion: apps/v1, kind: ReplicaSet, name: api-7d9f, controller: true}]
  - apiVersion: discovery.k8s.io/v1
    kind: EndpointSlice
    metadata: {name: api-abcde, namespace: shop}
`
	manifests, _ := discover(t, deployment)

	fromKubectl, stats := discover(t, live)

	strip := func(f *model.Fragment) []model.Element {
		var out []model.Element
		for _, e := range f.Elements {
			out = append(out, e.Element)
		}
		return out
	}
	assert.Equal(t, strip(manifests), strip(fromKubectl))
	assert.Equal(t, map[string]int{"owned": 2, "generated": 1}, stats.Skipped)
}

func TestDiscover_Workloads(t *testing.T) {
	t.Parallel()
	pod := func(kind, specPath string) string {
		spec := "containers:\n  - {name: main, image: \"registry.acme.io:5000/team/worker@sha256:abc\", ports: [{containerPort: 9000, protocol: UDP}]}\n"
		indent := strings.Repeat("  ", strings.Count(specPath, ".")+1)
		var b strings.Builder
		b.WriteString("apiVersion: v1\nkind: " + kind + "\nmetadata: {name: w, namespace: n}\n")
		parts := strings.Split(specPath, ".")
		for i, p := range parts {
			b.WriteString(strings.Repeat("  ", i) + p + ":\n")
		}
		for _, line := range strings.Split(strings.TrimSpace(spec), "\n") {
			b.WriteString(indent + line + "\n")
		}
		return b.String()
	}
	wantPorts := []model.Port{{Name: "9000", Protocol: "UDP", Number: 9000}}

	tests := []struct {
		kind, specPath, wantMultiplicity string
	}{
		{"Deployment", "spec.template.spec", ""},
		{"StatefulSet", "spec.template.spec", ""},
		{"DaemonSet", "spec.template.spec", "one per node"},
		{"Job", "spec.template.spec", ""},
		{"CronJob", "spec.jobTemplate.spec.template.spec", ""},
		{"ReplicaSet", "spec.template.spec", ""},
		{"Pod", "spec", ""},
	}
	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			t.Parallel()

			f, _ := discover(t, pod(tt.kind, tt.specPath))

			e := element(t, f, "k8s:n/"+strings.ToLower(tt.kind)+"/w")
			assert.Equal(t, tt.kind+" · worker", e.Tech)
			assert.Equal(t, wantPorts, e.Ports)
			assert.Equal(t, tt.wantMultiplicity, e.Multiplicity)
			assert.Equal(t, "k8s:n/namespace/n", e.Parent)
			assert.Empty(t, e.Title, "the id holds the name")
		})
	}
}

func TestDiscover_MergesThePortsOfEveryContainer(t *testing.T) {
	t.Parallel()
	input := `apiVersion: apps/v1
kind: Deployment
metadata: {name: api, namespace: n}
spec:
  template:
    spec:
      initContainers:
        - {name: init, image: busybox, ports: [{containerPort: 1}]}
      containers:
        - {name: api, image: api:1, ports: [{name: http, containerPort: 8080}, {name: metrics, containerPort: 9090}]}
        - {name: proxy, image: envoy:1, ports: [{name: http, containerPort: 8080}, {name: admin, containerPort: 9901}]}
`

	f, _ := discover(t, input)

	e := element(t, f, "k8s:n/deployment/api")
	assert.Equal(t, []model.Port{
		{Name: "http", Protocol: "TCP", Number: 8080},
		{Name: "metrics", Protocol: "TCP", Number: 9090},
		{Name: "admin", Protocol: "TCP", Number: 9901},
	}, e.Ports)
	assert.Equal(t, "Deployment · api:1", e.Tech, "the first container is the main one")
}

func TestDiscover_AWorkloadWithoutContainers_HasNoPortsNorImage(t *testing.T) {
	t.Parallel()

	f, _ := discover(t, "apiVersion: apps/v1\nkind: Deployment\nmetadata: {name: api, namespace: n}\n")

	e := element(t, f, "k8s:n/deployment/api")
	assert.Equal(t, "Deployment", e.Tech)
	assert.Nil(t, e.Ports)
}

func TestDiscover_Services(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, spec, wantTech string
		wantPorts            []model.Port
	}{
		{"ClusterIP by default", "ports: [{name: http, port: 80, targetPort: 8080}]", "Service", []model.Port{{Name: "http", Protocol: "TCP", Number: 80}}},
		{"headless", "clusterIP: None\nports: [{port: 5432}]", "Service · headless", []model.Port{{Name: "5432", Protocol: "TCP", Number: 5432}}},
		{"LoadBalancer", "type: LoadBalancer\nports: [{name: dns, port: 53, protocol: UDP}]", "Service · LoadBalancer", []model.Port{{Name: "dns", Protocol: "UDP", Number: 53}}},
		{"ExternalName", "type: ExternalName\nexternalName: db.acme.io", "Service · ExternalName", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spec := "spec:\n  " + strings.ReplaceAll(tt.spec, "\n", "\n  ") + "\n"

			f, _ := discover(t, "apiVersion: v1\nkind: Service\nmetadata: {name: s, namespace: n}\n"+spec)

			e := element(t, f, "k8s:n/service/s")
			assert.Equal(t, tt.wantTech, e.Tech)
			assert.Equal(t, tt.wantPorts, e.Ports)
		})
	}
}

func TestDiscover_Namespaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, input string
		wantZones   map[string]int
	}{
		{"a declared namespace points at its object",
			"apiVersion: v1\nkind: Service\nmetadata: {name: a, namespace: shop}\n---\napiVersion: v1\nkind: Namespace\nmetadata: {name: shop}\n",
			map[string]int{"k8s:shop/namespace/shop": 5}},
		{"an implicit namespace points at its first object",
			"apiVersion: v1\nkind: ConfigMap\nmetadata: {name: c, namespace: shop}\n---\napiVersion: v1\nkind: Service\nmetadata: {name: a, namespace: shop}\n",
			map[string]int{"k8s:shop/namespace/shop": 1}},
		{"an object without namespace lands in default, as kubectl apply does",
			"apiVersion: v1\nkind: Service\nmetadata: {name: a}\n",
			map[string]int{"k8s:default/namespace/default": 1}},
		{"a declared namespace with nothing in it is still a zone",
			"apiVersion: v1\nkind: Namespace\nmetadata: {name: empty}\n",
			map[string]int{"k8s:empty/namespace/empty": 1}},
		{"a namespace holding only config is not a zone",
			"apiVersion: v1\nkind: ConfigMap\nmetadata: {name: c, namespace: cfg}\n",
			map[string]int{}},
		{"a cluster-scoped object creates no default zone",
			"apiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRole\nmetadata: {name: r}\n",
			map[string]int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, _ := discover(t, tt.input)

			zones := map[string]int{}
			for _, e := range f.Elements {
				if e.Kind == "group" {
					zones[e.ID] = e.Src.Line
				}
			}
			assert.Equal(t, tt.wantZones, zones)
		})
	}
}

func TestDiscover_ReportsUnknownKindsOncePerKind(t *testing.T) {
	t.Parallel()
	cluster := func(name string) string {
		return "---\napiVersion: postgresql.cnpg.io/v1\nkind: Cluster\nmetadata: {name: " + name + ", namespace: n}\n"
	}
	input := cluster("a") + cluster("b") + cluster("c") +
		"---\napiVersion: v1\nkind: Widget\nmetadata: {name: w, namespace: n}\n" +
		"---\napiVersion: monitoring.coreos.com/v1\nkind: ServiceMonitor\nmetadata: {name: m, namespace: n}\n" +
		"---\napiVersion: gateway.envoyproxy.io/v1alpha1\nkind: ClientTrafficPolicy\nmetadata: {name: p, namespace: n}\n"

	f, _ := discover(t, input)

	assert.Equal(t, []model.Unresolved{
		{Kind: "unknown-kind", Value: "postgresql.cnpg.io/Cluster", Count: 3, Src: model.Src{File: "-", Line: 2}},
		{Kind: "unknown-kind", Value: "core/Widget", Src: model.Src{File: "-", Line: 14}},
	}, f.Unresolved)
}

func TestDiscover_WalksADirectory(t *testing.T) {
	t.Parallel()

	f, stats, err := k8s.Adapter{}.Discover(context.Background(), adapter.Source{Paths: []string{"testdata/repo"}})

	require.NoError(t, err)
	assert.Equal(t, []string{"k8s:shop/deployment/api", "k8s:shop/statefulset/db", "k8s:shop/namespace/shop"}, ids(f),
		"hidden directories, non-YAML files and the chart's templates are not read")
	assert.Equal(t, "testdata/repo/apps/api.yaml", element(t, f, "k8s:shop/deployment/api").Src.File)
	assert.Equal(t, []model.Unresolved{
		{Kind: "unrendered", Value: "testdata/repo/chart", Hint: "helm", Src: model.Src{File: "testdata/repo/chart/Chart.yaml", Line: 1}},
		{Kind: "unrendered", Value: "testdata/repo/overlay", Hint: "kustomize", Src: model.Src{File: "testdata/repo/overlay/kustomization.yaml", Line: 1}},
	}, f.Unresolved)
	assert.Equal(t, 2, stats.Skipped["no-kind"], "values.yaml and docs/notes.yaml")
}

func TestDiscover_ReadsAFileGivenByName_WhateverItsExtension(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "rendered.txt")
	require.NoError(t, os.WriteFile(path, []byte(deployment), 0o644))

	f, _, err := k8s.Adapter{}.Discover(context.Background(), adapter.Source{Paths: []string{path}})

	require.NoError(t, err)
	assert.Contains(t, ids(f), "k8s:shop/deployment/api")
}

func TestRun_GivesTheSameBytes_OnTheSameInput(t *testing.T) {
	t.Parallel()
	encode := func() string {
		f, _, err := adapter.Run(context.Background(), k8s.Adapter{}, adapter.Source{Paths: []string{"testdata/repo"}}, "", time.Unix(0, 0))
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, model.EncodeFragment(&buf, f))
		return buf.String()
	}

	first := encode()

	assert.Equal(t, first, encode())
}

func TestRun_WritesAFragmentTheSchemaAccepts(t *testing.T) {
	t.Parallel()
	schema, err := jsonschema.NewCompiler().Compile("../../../schemas/v1alpha1/fragment.schema.json")
	require.NoError(t, err)
	f, _, err := adapter.Run(context.Background(), k8s.Adapter{}, adapter.Source{Paths: []string{"testdata/repo"}}, "abc", time.Unix(0, 0))
	require.NoError(t, err)
	var buf bytes.Buffer
	require.NoError(t, model.EncodeFragment(&buf, f))
	var doc any
	require.NoError(t, yaml.Unmarshal(buf.Bytes(), &doc))
	raw, err := json.Marshal(doc)
	require.NoError(t, err)
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	require.NoError(t, err)

	err = schema.Validate(inst)

	assert.NoError(t, err)
}
