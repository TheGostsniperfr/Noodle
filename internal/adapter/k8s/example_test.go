package k8s_test

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/adapter"
	"github.com/TheGostsniperfr/Noodle/internal/adapter/k8s"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

const (
	exampleDir     = "examples/discovery-k8s"
	exampleSources = exampleDir + "/manifests"
)

var exampleTime = time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)

// listFromManifests wraps the objects of manifests/platform.yaml as `kubectl get -o yaml`
// prints them, a List. Every item starts on the line its document starts on, so both
// inputs cite the same lines; the three header lines and the `---` separators are
// swapped for the List's own lines and a comment.
func listFromManifests(manifests string) string {
	lines := strings.Split(strings.TrimSuffix(manifests, "\n"), "\n")
	out := []string{"apiVersion: v1", "kind: List", "items:"}
	startsDocument := true
	for _, l := range lines[3:] {
		switch {
		case l == "---":
			out = append(out, "  # ---")
			startsDocument = true
		case startsDocument:
			out = append(out, "  - "+l)
			startsDocument = false
		case l == "":
			out = append(out, "")
		default:
			out = append(out, "    "+l)
		}
	}
	return strings.Join(out, "\n") + "\n"
}

func exampleFragment(t *testing.T) []byte {
	t.Helper()
	f, _, err := adapter.Run(context.Background(), k8s.Adapter{}, adapter.Source{Paths: []string{exampleSources}}, "", exampleTime)
	require.NoError(t, err)
	var buf bytes.Buffer
	require.NoError(t, model.EncodeFragment(&buf, f))
	return buf.Bytes()
}

func TestExample_ManifestsEqualList(t *testing.T) {
	root, err := filepath.Abs("../../..")
	require.NoError(t, err)
	manifests, err := os.ReadFile(filepath.Join(root, exampleSources, "platform.yaml"))
	require.NoError(t, err)
	listPath := filepath.Join(root, exampleDir, "list.yaml")
	if os.Getenv("UPDATE_LIST") != "" {
		require.NoError(t, os.WriteFile(listPath, []byte(listFromManifests(string(manifests))), 0o644))
	}
	list, err := os.ReadFile(listPath)
	require.NoError(t, err)
	require.Equal(t, listFromManifests(string(manifests)), string(list), "list.yaml is the manifests as a List: UPDATE_LIST=1 go test ./internal/adapter/k8s/ -run TestExample_ManifestsEqualList")
	t.Chdir(root)
	fromManifests := exampleFragment(t)

	staged := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(staged, exampleSources), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(staged, exampleSources, "platform.yaml"), list, 0o644))
	t.Chdir(staged)
	fromList := exampleFragment(t)

	assert.Equal(t, string(fromManifests), string(fromList))
	assert.Contains(t, string(fromList), "k8s:prod/deployment/worker")
}

// generators write one random object each, from names small enough to collide, so the
// property test meets selectors, duplicates and dangling references.
var generators = []func(r *rand.Rand, name, ns string) string{
	func(r *rand.Rand, name, ns string) string {
		return fmt.Sprintf("apiVersion: apps/v1\nkind: Deployment\nmetadata: {name: %s, namespace: %s}\nspec:\n  template:\n    metadata: {labels: {app: %s}}\n    spec:\n      containers:\n        - name: c\n          image: img\n          ports: [{name: http, containerPort: 8080}]\n          env:\n            - {name: A, value: 'svc%d.ns%d.svc:80'}\n            - {name: B, value: 'https://h%d.example.com/'}\n            - {name: S, valueFrom: {secretKeyRef: {name: s%d, key: k}}}\n          envFrom: [{configMapRef: {name: cm%d}}]\n",
			name, ns, name, r.Intn(3), r.Intn(3), r.Intn(3), r.Intn(3), r.Intn(3))
	},
	func(r *rand.Rand, name, ns string) string {
		return fmt.Sprintf("apiVersion: v1\nkind: Service\nmetadata: {name: %s, namespace: %s}\nspec:\n  selector: {app: %s}\n  ports: [{name: http, port: 80}, {port: %d}]\n", name, ns, name, 8000+r.Intn(3))
	},
	func(r *rand.Rand, name, ns string) string {
		return fmt.Sprintf("apiVersion: v1\nkind: ConfigMap\nmetadata: {name: cm%d, namespace: %s}\ndata: {u: 'svc%d.ns%d:80', v: 'x'}\n", r.Intn(3), ns, r.Intn(3), r.Intn(3))
	},
	func(r *rand.Rand, name, ns string) string {
		return fmt.Sprintf("apiVersion: gateway.networking.k8s.io/v1\nkind: HTTPRoute\nmetadata: {name: %s, namespace: %s}\nspec:\n  parentRefs: [{name: gw%d}]\n  hostnames: [h%d.example.com]\n  rules: [{backendRefs: [{name: svc%d, port: 80}]}]\n", name, ns, r.Intn(2), r.Intn(3), r.Intn(3))
	},
	func(r *rand.Rand, name, ns string) string {
		return fmt.Sprintf("apiVersion: gateway.networking.k8s.io/v1\nkind: Gateway\nmetadata: {name: gw%d, namespace: %s}\nspec:\n  listeners: [{name: http, port: 80, protocol: HTTP}]\n", r.Intn(2), ns)
	},
	func(r *rand.Rand, name, ns string) string {
		return fmt.Sprintf("apiVersion: external-secrets.io/v1beta1\nkind: ExternalSecret\nmetadata: {name: %s, namespace: %s}\nspec:\n  secretStoreRef: {name: st%d}\n", name, ns, r.Intn(2))
	},
	func(r *rand.Rand, name, ns string) string {
		return fmt.Sprintf("apiVersion: example.com/v1\nkind: Thing\nmetadata: {name: %s, namespace: %s}\nspec: {url: 'svc%d.ns%d.svc'}\n", name, ns, r.Intn(3), r.Intn(3))
	},
}

func TestDiscover_Property_Deterministic(t *testing.T) {
	deterministic := func(seed int64) bool {
		r := rand.New(rand.NewSource(seed))
		var docs []string
		for range 2 + r.Intn(12) {
			ns := fmt.Sprintf("ns%d", r.Intn(3))
			docs = append(docs, generators[r.Intn(len(generators))](r, fmt.Sprintf("svc%d", r.Intn(3)), ns))
		}
		r.Shuffle(len(docs), func(i, j int) { docs[i], docs[j] = docs[j], docs[i] })
		manifests := strings.Join(docs, "---\n")
		run := func() (string, bool) {
			f, _, err := adapter.Run(context.Background(), k8s.Adapter{}, adapter.Source{Reader: strings.NewReader(manifests)}, "", exampleTime)
			if err != nil {
				return err.Error(), false
			}
			return encode(f), true
		}
		first, ok := run()
		second, _ := run()
		return ok && first == second
	}

	err := quick.Check(deterministic, &quick.Config{MaxCount: 200})

	assert.NoError(t, err)
}

func encode(f *model.Fragment) string {
	var buf bytes.Buffer
	if err := model.EncodeFragment(&buf, f); err != nil {
		return err.Error()
	}
	return buf.String()
}
