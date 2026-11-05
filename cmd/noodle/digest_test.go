package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fragmentHead = `apiVersion: noodle/v1alpha1
kind: Fragment
provenance: {adapter: k8s, source: m}
`

// writeFragments writes each body as one fragment file of a fresh directory.
func writeFragments(t *testing.T, bodies ...string) string {
	t.Helper()
	dir := t.TempDir()
	for i, body := range bodies {
		path := filepath.Join(dir, "k8s-"+string(rune('a'+i))+".yaml")
		require.NoError(t, os.WriteFile(path, []byte(fragmentHead+body), 0o600))
	}
	return dir
}

const apiBehindService = `elements:
  - {id: 'k8s:prod/deployment/api', kind: backend, tech: 'Deployment · api:1', ports: [{name: http, protocol: TCP, port: 8080}], src: {file: m.yaml, line: 1}}
  - {id: 'k8s:prod/service/api', kind: backend, tech: Service, ports: [{name: http, protocol: TCP, port: 80}], src: {file: m.yaml, line: 20}}
  - {id: 'k8s:prod/namespace/prod', kind: group, src: {file: m.yaml, line: 30}}
connections:
  - {id: 'k8s:c/1', from: 'k8s:prod/service/api', to: 'k8s:prod/deployment/api', port: http, kind: flow, src: {file: m.yaml, line: 20}}
`

func TestDigest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		frags []string
		want  string
	}{
		{"folds-service", []string{apiBehindService + `  - {id: 'k8s:c/2', from: 'k8s:prod/deployment/web', to: 'k8s:prod/service/api', port: http, kind: flow, inferred: true, src: {file: m.yaml, line: 40}}
`}, `# noodle digest of 1 fragments: 1 components, 1 connections, 0 secrets, 0 open
# components
prod/api  Deployment · api:1  ports http:8080  src m.yaml:1
# connections
prod/web -> prod/api  http inferred  src m.yaml:40
`},
		{"resolves-route", []string{strings.Replace(apiBehindService, "connections:\n", "  - {id: 'k8s:prod/httproute/api', kind: backend, tech: HTTPRoute, desc: api.example.com, src: {file: m.yaml, line: 50}}\nconnections:\n", 1) + `references:
  - {id: 'k8s:r/1', from: 'k8s:prod/httproute/api', to: 'k8s:prod/service/api', kind: backendRef, src: {file: m.yaml, line: 50}}
unresolved:
  - {about: ['k8s:prod/httproute/api'], kind: missing-backend, value: 'k8s:edge/gateway/shared', hint: gateway, src: {file: m.yaml, line: 52}}
`, `elements:
  - {id: 'k8s:edge/gateway/shared', kind: backend, tech: Gateway · envoy, src: {file: g.yaml, line: 3}}
`}, `# noodle digest of 2 fragments: 2 components, 1 connections, 0 secrets, 0 open
# components
edge/gateway/shared  Gateway · envoy  src g.yaml:3
prod/api  Deployment · api:1  ports http:8080  src m.yaml:1
# connections
edge/gateway/shared -> prod/api  host api.example.com  src m.yaml:52
`},
		{"keeps-externalsecret-path", []string{`elements:
  - {id: 'k8s:_external/onepassword', kind: external, tech: 1Password, src: {file: m.yaml, line: 9}}
references:
  - {id: 'k8s:r/1', from: 'k8s:prod/externalsecret/db', to: 'k8s:prod/secret/db', kind: target, src: {file: m.yaml, line: 12}}
  - {id: 'k8s:r/2', from: 'k8s:prod/externalsecret/db', to: 'k8s:_cluster/clustersecretstore/op', kind: store, src: {file: m.yaml, line: 13}}
  - {id: 'k8s:r/3', from: 'k8s:_cluster/clustersecretstore/op', to: 'k8s:_external/onepassword', kind: provider, src: {file: m.yaml, line: 9}}
  - {id: 'k8s:r/4', from: 'k8s:prod/externalsecret/s3', to: 'k8s:prod/secret/api-s3', kind: target, src: {file: m.yaml, line: 22}}
`}, `# noodle digest of 1 fragments: 1 components, 0 connections, 2 secrets, 0 open
# components
external/onepassword  1Password  src m.yaml:9
# secrets
prod/secret/api-s3 <- externalsecret s3  src m.yaml:22
prod/secret/db <- externalsecret <- clustersecretstore/op <- external/onepassword  src m.yaml:12
`},
		{"drops-config-refs", []string{apiBehindService + `references:
  - {id: 'k8s:r/1', from: 'k8s:prod/deployment/api', to: 'k8s:prod/configmap/settings', kind: env-from, src: {file: m.yaml, line: 5}}
  - {id: 'k8s:r/2', from: 'k8s:prod/deployment/api', to: 'k8s:prod/secret/db', kind: secret-key, src: {file: m.yaml, line: 6}}
  - {id: 'k8s:r/3', from: 'k8s:prod/deployment/api', to: 'k8s:prod/persistentvolumeclaim/data', kind: mounts, src: {file: m.yaml, line: 7}}
`}, `# noodle digest of 1 fragments: 1 components, 0 connections, 0 secrets, 0 open
# components
prod/api  Deployment · api:1  ports http:8080  src m.yaml:1
`},
		{"keeps-unresolved", []string{apiBehindService + `unresolved:
  - {about: ['k8s:prod/deployment/api', 'k8s:prod/deployment/worker'], kind: secret-endpoint, value: db/DB_URL, hint: secret-manager, src: {file: m.yaml, line: 6}}
  - {about: ['k8s:prod/service/api'], kind: unmatched-selector, value: 'app=ghost', src: {file: m.yaml, line: 20}}
  - {kind: unknown-kind, value: monitoring.coreos.com/Prometheus, count: 2, src: {file: m.yaml, line: 60}}
`}, `# noodle digest of 1 fragments: 1 components, 0 connections, 0 secrets, 3 open
# components
prod/api  Deployment · api:1  ports http:8080  src m.yaml:1
# open
secret-endpoint prod/db: DB_URL  about prod/api, prod/worker  src m.yaml:6
unknown-kind monitoring.coreos.com/Prometheus ×2  src m.yaml:60
unmatched-selector app=ghost  about prod/api  src m.yaml:20
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := writeFragments(t, tt.frags...)
			var stdout, stderr bytes.Buffer

			err := digestCommand([]string{dir}, &stdout, &stderr)

			require.NoError(t, err)
			assert.Equal(t, tt.want, stdout.String())
		})
	}
}

func TestDigest_FailsOnAnEmptyDirectory(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer

	err := digestCommand([]string{t.TempDir()}, &stdout, &stderr)

	require.ErrorContains(t, err, "no fragment")
}
