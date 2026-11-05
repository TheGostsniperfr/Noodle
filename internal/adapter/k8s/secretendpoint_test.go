package k8s_test

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

const endpointValue = "postgres://app:hunter2@db.internal:5432/app"

func prodWorkload(name, env string) string {
	return fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata: {name: %s, namespace: prod}
spec:
  template:
    spec:
      containers:
        - name: %s
          image: acme/%s:1
%s---
`, name, name, name, env)
}

const dbURLFromSecret = `          env:
            - name: DATABASE
              valueFrom:
                secretKeyRef: {name: db, key: DB_URL}
            - name: DB_PASS
              valueFrom:
                secretKeyRef: {name: db, key: PASSWORD}
`

// dbSecret holds values the fragment must never show, in data and in stringData.
func dbSecret(ns, encoded string) string {
	return `apiVersion: v1
kind: Secret
metadata: {name: db, namespace: ` + ns + `}
data:
  DB_URL: ` + encoded + `
  PASSWORD: ` + encoded + `
stringData:
  SMTP_SERVER_PASSWORD: "` + endpointValue + `"
---
`
}

const envFromDB = `          envFrom:
            - secretRef: {name: db}
`

func secretEndpointEntries(f *model.Fragment) []model.Unresolved {
	out := []model.Unresolved{}
	for _, u := range f.Unresolved {
		if u.Kind == "secret-endpoint" {
			out = append(out, u)
		}
	}
	return out
}

func TestSecretEndpoint(t *testing.T) {
	t.Parallel()
	encoded := base64.StdEncoding.EncodeToString([]byte(endpointValue))
	tests := []struct {
		name      string
		manifests string
		want      []model.Unresolved
	}{
		{"secret-endpoint-env", prodWorkload("api", dbURLFromSecret) + prodWorkload("worker", dbURLFromSecret) + dbSecret("prod", encoded),
			[]model.Unresolved{{Kind: "secret-endpoint", Value: "db/DB_URL", Hint: "secret-manager",
				About: []string{"k8s:prod/deployment/api", "k8s:prod/deployment/worker"}, Src: model.Src{Line: 13}}}},
		{"secret-endpoint-envfrom", prodWorkload("api", envFromDB) + `apiVersion: v1
kind: Secret
metadata: {name: db, namespace: prod}
data:
  DB_HOST: ` + encoded + `
  PASSWORD: ` + encoded + `
stringData:
  SMTP_SERVER: "` + endpointValue + `"
`, []model.Unresolved{{Kind: "secret-endpoint", Value: "db/DB_HOST,SMTP_SERVER", Hint: "secret-manager",
			About: []string{"k8s:prod/deployment/api"}, Src: model.Src{Line: 11}}}},
		{"secret-endpoint-externalsecret", prodWorkload("api", envFromDB) + dbSecret("prod", encoded) + `apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata: {name: db, namespace: prod}
spec:
  secretStoreRef: {kind: ClusterSecretStore, name: onepassword}
  target:
    name: db
    template:
      data:
        JDBC_URL: "jdbc:postgresql://{{ .host }}/app"
  data:
    - secretKey: S3_ENDPOINT
      remoteRef: {key: app/s3}
    - secretKey: API_TOKEN
      remoteRef: {key: app/token}
`, []model.Unresolved{{Kind: "secret-endpoint", Value: "db/DB_URL,JDBC_URL,S3_ENDPOINT", Hint: "secret-manager",
			About: []string{"k8s:prod/deployment/api"}, Src: model.Src{Line: 11}}}},
		{"secret-endpoint-ignores-password", prodWorkload("api", `          env:
            - name: DB_PASS
              valueFrom:
                secretKeyRef: {name: db, key: PASSWORD}
            - name: SMTP
              valueFrom:
                secretKeyRef: {name: db, key: SMTP_SERVER_PASSWORD}
`) + dbSecret("prod", encoded), []model.Unresolved{}},
		{"secret-endpoint-namespaces", prodWorkload("api", envFromDB) + dbSecret("prod", encoded) +
			strings.Replace(prodWorkload("api", dbURLFromSecret), "namespace: prod", "namespace: staging", 1),
			[]model.Unresolved{
				{Kind: "secret-endpoint", Value: "db/DB_URL", Hint: "secret-manager", About: []string{"k8s:prod/deployment/api"}, Src: model.Src{Line: 11}},
				{Kind: "secret-endpoint", Value: "db/DB_URL", Hint: "secret-manager", About: []string{"k8s:staging/deployment/api"}, Src: model.Src{Line: 34}},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, _ := discover(t, tt.manifests)

			got := secretEndpointEntries(f)
			for i := range got {
				got[i].Src.File = ""
			}
			assert.Equal(t, tt.want, got)
			out, err := yaml.Marshal(f)
			require.NoError(t, err)
			for _, leak := range []string{endpointValue, encoded, "hunter2", "jdbc:postgresql"} {
				assert.NotContains(t, string(out), leak)
			}
		})
	}
}
