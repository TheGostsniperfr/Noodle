package k8s_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// refLines prints every reference as "from -> to kind line", sorted, so a test states
// the whole list at once.
func refLines(f *model.Fragment) []string {
	out := []string{}
	for _, r := range f.References {
		out = append(out, fmt.Sprintf("%s -> %s %s %d", r.From, r.To, r.Kind, r.Src.Line))
	}
	sort.Strings(out)
	return out
}

// apiHead is the first nine lines of a Deployment api in prod whose container is open
// for more fields at line 10.
const apiHead = `apiVersion: apps/v1
kind: Deployment
metadata: {name: api, namespace: prod}
spec:
  template:
    spec:
      containers:
        - name: api
          image: acme/api:1
`

func TestReferences(t *testing.T) {
	const api = "k8s:prod/deployment/api"
	tests := []struct {
		name      string
		manifests string
		want      []string
		wantIDs   map[string]model.Element
	}{
		{"env-from", apiHead + `          envFrom:
            - configMapRef:
                name: settings
            - secretRef:
                name: db
`, []string{
			api + " -> k8s:prod/configmap/settings env-from 12",
			api + " -> k8s:prod/secret/db env-from 14",
		}, nil},
		{"secret-key", apiHead + `          env:
            - name: PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db
                  key: PASSWORD
`, []string{api + " -> k8s:prod/secret/db secret-key 14"}, nil},
		{"config-key", apiHead + `          env:
            - name: LEVEL
              valueFrom:
                configMapKeyRef: {name: settings, key: level}
`, []string{api + " -> k8s:prod/configmap/settings config-key 13"}, nil},
		{"mounts-secret", apiHead + `      volumes:
        - name: creds
          secret:
            secretName: db
`, []string{api + " -> k8s:prod/secret/db mounts 13"}, nil},
		{"mounts-configmap", apiHead + `      volumes:
        - name: conf
          configMap:
            name: settings
`, []string{api + " -> k8s:prod/configmap/settings mounts 13"}, nil},
		{"mounts-pvc", apiHead + `      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: api-data
`, []string{api + " -> k8s:prod/persistentvolumeclaim/api-data mounts 13"}, nil},
		{"mounts-projected-and-init-container", apiHead + `      initContainers:
        - name: init
          image: busybox
          env:
            - name: TOKEN
              valueFrom:
                secretKeyRef: {name: init-token, key: t}
      volumes:
        - name: all
          projected:
            sources:
              - secret:
                  name: db
              - configMap:
                  name: settings
`, []string{
			api + " -> k8s:prod/configmap/settings mounts 24",
			api + " -> k8s:prod/secret/db mounts 22",
			api + " -> k8s:prod/secret/init-token secret-key 16",
		}, nil},
		{"two-keys-of-one-secret-are-one-reference", apiHead + `          env:
            - name: A
              valueFrom: {secretKeyRef: {name: db, key: a}}
            - name: B
              valueFrom: {secretKeyRef: {name: db, key: b}}
`, []string{api + " -> k8s:prod/secret/db secret-key 12"}, nil},
		{"externalsecret-target", `apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata: {name: db, namespace: prod}
spec:
  secretStoreRef:
    name: onepassword
    kind: SecretStore
  target:
    name: db-credentials
`, []string{
			"k8s:prod/externalsecret/db -> k8s:prod/secret/db-credentials target 9",
			"k8s:prod/externalsecret/db -> k8s:prod/secretstore/onepassword store 6",
		}, nil},
		{"externalsecret-store", `apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata: {name: db, namespace: prod}
spec:
  secretStoreRef: {name: global, kind: ClusterSecretStore}
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata: {name: cache, namespace: prod}
spec:
  secretStoreRef: {name: local}
  target: {name: cache}
`, []string{
			"k8s:prod/externalsecret/cache -> k8s:prod/secret/cache target 12",
			"k8s:prod/externalsecret/cache -> k8s:prod/secretstore/local store 11",
			"k8s:prod/externalsecret/db -> k8s:_cluster/clustersecretstore/global store 5",
			"k8s:prod/externalsecret/db -> k8s:prod/secret/db target 1",
		}, nil},
		{"clusterstore-provider", `apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata: {name: global}
spec:
  provider:
    vault:
      server: https://vault.example.com
`, []string{"k8s:_cluster/clustersecretstore/global -> k8s:_external/vault provider 6"},
			map[string]model.Element{"k8s:_external/vault": {ID: "k8s:_external/vault", Kind: "external", Tech: "Vault"}}},
		{"store-provider-1password", storeWith("onepassword: {connectTimeout: 20}"),
			[]string{"k8s:prod/secretstore/store -> k8s:_external/onepassword provider 6"},
			map[string]model.Element{"k8s:_external/onepassword": {ID: "k8s:_external/onepassword", Kind: "external", Tech: "1Password"}}},
		{"store-provider-vault", storeWith("vault: {server: 'https://vault'}"),
			[]string{"k8s:prod/secretstore/store -> k8s:_external/vault provider 6"},
			map[string]model.Element{"k8s:_external/vault": {ID: "k8s:_external/vault", Kind: "external", Tech: "Vault"}}},
		{"store-provider-aws", storeWith("aws: {service: SecretsManager, region: eu-west-3}"),
			[]string{"k8s:prod/secretstore/store -> k8s:_external/aws provider 6"},
			map[string]model.Element{"k8s:_external/aws": {ID: "k8s:_external/aws", Kind: "external", Tech: "AWS"}}},
		{"argocd-syncs", `apiVersion: argoproj.io/v1alpha1
kind: Application
metadata: {name: shop, namespace: argocd}
spec:
  destination:
    server: https://kubernetes.default.svc
    namespace: prod
`, []string{"k8s:argocd/application/shop -> k8s:prod/namespace/prod syncs 7"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := discover(t, tt.manifests)

			assert.Equal(t, tt.want, refLines(f))
			if tt.wantIDs != nil {
				providers := map[string]model.Element{}
				for _, e := range f.Elements {
					if e.Kind == "external" {
						providers[e.ID] = e.Element
					}
				}
				assert.Equal(t, tt.wantIDs, providers)
			}
			for _, u := range f.Unresolved {
				assert.NotEqual(t, "unknown-kind", u.Kind, "%v", u)
			}
		})
	}
}

func storeWith(provider string) string {
	return `apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata: {name: store, namespace: prod}
spec:
  provider:
    ` + provider + "\n"
}

func TestReferences_UnknownProviderIsReported_NotDrawn(t *testing.T) {
	f, _ := discover(t, storeWith("gcpsm: {projectID: p}"))

	assert.Empty(t, f.References)
	assert.Equal(t, []model.Unresolved{{About: []string{"k8s:prod/secretstore/store"}, Kind: "unknown-provider", Value: "gcpsm", Hint: "secretstore",
		Src: model.Src{File: "-", Line: 6}}}, f.Unresolved)
}

const secretFixture = `apiVersion: v1
kind: Secret
metadata: {name: db, namespace: prod}
type: Opaque
data:
  PASSWORD: %s
stringData:
  TOKEN: %s
---
` + apiHead + `          env:
            - name: PASSWORD
              valueFrom:
                secretKeyRef: {name: db, key: PASSWORD}
      volumes:
        - name: creds
          secret: {secretName: db}
`

func encodeFragmentBytes(t *testing.T, f *model.Fragment) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, model.EncodeFragment(&buf, f))
	return buf.Bytes()
}

func TestReferences_NoSecretValue(t *testing.T) {
	const password, token = "hunter2-correct-horse", "tok-9f8e7d6c5b4a"
	manifests := fmt.Sprintf(secretFixture, base64.StdEncoding.EncodeToString([]byte(password)), token)

	f, _ := discover(t, manifests)
	out := string(encodeFragmentBytes(t, f))

	assert.Equal(t, []string{
		"k8s:prod/deployment/api -> k8s:prod/secret/db mounts 25",
		"k8s:prod/deployment/api -> k8s:prod/secret/db secret-key 22",
	}, refLines(f))
	for _, leak := range []string{password, base64.StdEncoding.EncodeToString([]byte(password)), token, base64.StdEncoding.EncodeToString([]byte(token))} {
		assert.NotContains(t, out, leak)
	}
}

type secretValue string

func (secretValue) Generate(r *rand.Rand, _ int) reflect.Value {
	var b strings.Builder
	b.WriteString("sv")
	for range 24 {
		b.WriteByte("abcdefghijklmnopqrstuvwxyz0123456789"[r.Intn(36)])
	}
	return reflect.ValueOf(secretValue(b.String()))
}

func TestReferences_Property_NoSecretValue(t *testing.T) {
	noLeak := func(data, stringData secretValue) bool {
		encoded := base64.StdEncoding.EncodeToString([]byte(data))
		f, _ := discover(t, fmt.Sprintf(secretFixture, encoded, stringData))
		out := string(encodeFragmentBytes(t, f))
		for _, leak := range []string{string(data), encoded, string(stringData), base64.StdEncoding.EncodeToString([]byte(stringData))} {
			if strings.Contains(out, leak) {
				return false
			}
		}
		return len(f.References) == 2
	}

	err := quick.Check(noLeak, &quick.Config{MaxCount: 200})

	assert.NoError(t, err)
}
