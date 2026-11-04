package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func service(name, namespace string) string {
	ns := ""
	if namespace != "" {
		ns = ", namespace: " + namespace
	}
	return "apiVersion: v1\nkind: Service\nmetadata: {name: " + name + ns + "}\nspec: {ports: [{port: 80}]}\n"
}

const noNamespaceLine = "3 objects name no namespace; pass -namespace"

// runDiscover runs the k8s adapter on stdin and returns stdout and stderr.
func runDiscover(t *testing.T, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	reg, regErr := adapters()
	require.NoError(t, regErr)
	var out, errOut bytes.Buffer
	err = discoverCommand(context.Background(), reg, append([]string{"k8s"}, args...), strings.NewReader(stdin), &out, &errOut)
	return out.String(), errOut.String(), err
}

func elementIDs(t *testing.T, fragment string) []string {
	t.Helper()
	var doc struct {
		Elements []struct{ ID string }
	}
	require.NoError(t, yaml.Unmarshal([]byte(fragment), &doc))
	var out []string
	for _, e := range doc.Elements {
		out = append(out, e.ID)
	}
	return out
}

func TestNamespace(t *testing.T) {
	mixed := service("a", "") + "---\n" + service("b", "") + "---\n" + service("c", "") + "---\n" + service("d", "prod") + "---\n" + service("e", "prod")
	allNamed := service("a", "web") + "---\n" + service("b", "web") + "---\n" + service("c", "web")
	app := filepath.Join(t.TempDir(), "app.yaml")
	require.NoError(t, os.WriteFile(app, []byte("apiVersion: argoproj.io/v1alpha1\nkind: Application\nmetadata: {name: shop, namespace: argocd}\nspec:\n  destination: {server: 'https://kubernetes.default.svc', namespace: web}\n"), 0o644))

	t.Run("mixed", func(t *testing.T) {
		stdout, stderr, err := runDiscover(t, mixed)

		require.NoError(t, err)
		assert.Contains(t, strings.Split(stderr, "\n"), noNamespaceLine)
		assert.ElementsMatch(t, []string{"k8s:default/service/a", "k8s:default/service/b", "k8s:default/service/c", "k8s:prod/service/d", "k8s:prod/service/e", "k8s:default/namespace/default", "k8s:prod/namespace/prod"}, elementIDs(t, stdout))
	})

	t.Run("all-named", func(t *testing.T) {
		_, stderr, err := runDiscover(t, allNamed)

		require.NoError(t, err)
		assert.NotContains(t, stderr, "name no namespace")
	})

	t.Run("none-named", func(t *testing.T) {
		_, stderr, err := runDiscover(t, service("a", "")+"---\n"+service("b", ""))

		require.NoError(t, err)
		assert.NotContains(t, stderr, "name no namespace")
	})

	t.Run("namespace-flag", func(t *testing.T) {
		stdout, stderr, err := runDiscover(t, mixed, "-namespace", "web")

		require.NoError(t, err)
		assert.NotContains(t, stderr, "name no namespace")
		assert.ElementsMatch(t, []string{"k8s:web/service/a", "k8s:web/service/b", "k8s:web/service/c", "k8s:prod/service/d", "k8s:prod/service/e", "k8s:web/namespace/web", "k8s:prod/namespace/prod"}, elementIDs(t, stdout))
	})

	t.Run("argocd-app", func(t *testing.T) {
		stdout, stderr, err := runDiscover(t, mixed, "-argocd-app", app)

		require.NoError(t, err)
		assert.NotContains(t, stderr, "name no namespace")
		assert.ElementsMatch(t, []string{"k8s:web/service/a", "k8s:web/service/b", "k8s:web/service/c", "k8s:prod/service/d", "k8s:prod/service/e", "k8s:web/namespace/web", "k8s:prod/namespace/prod"}, elementIDs(t, stdout))
	})

	t.Run("namespace-flag-wins-over-argocd-app", func(t *testing.T) {
		stdout, _, err := runDiscover(t, mixed, "-argocd-app", app, "-namespace", "other")

		require.NoError(t, err)
		assert.Contains(t, elementIDs(t, stdout), "k8s:other/service/a")
	})

	t.Run("argocd-app-without-destination-fails", func(t *testing.T) {
		bad := filepath.Join(t.TempDir(), "bad.yaml")
		require.NoError(t, os.WriteFile(bad, []byte("kind: Application\nspec: {}\n"), 0o644))

		_, _, err := runDiscover(t, mixed, "-argocd-app", bad)

		assert.ErrorContains(t, err, "no Application with spec.destination.namespace")
	})
}

const renderedShop = "apiVersion: v1\nkind: Service\nmetadata: {name: api, namespace: shop}\nspec: {ports: [{port: 80}]}\n"

func fakeKustomizeOnPath(t *testing.T) string {
	t.Helper()
	bin, argvFile, outFile := t.TempDir(), filepath.Join(t.TempDir(), "argv"), filepath.Join(t.TempDir(), "out.yaml")
	require.NoError(t, os.WriteFile(outFile, []byte(renderedShop), 0o644))
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"" + argvFile + "\"\ncat \"" + outFile + "\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(bin, "kustomize"), []byte(script), 0o755))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return argvFile
}

func TestKustomize(t *testing.T) {
	t.Run("fragment-name", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, "apps", "shop"), 0o755))
		argvFile := fakeKustomizeOnPath(t)
		t.Chdir(root)
		discovered := filepath.Join(root, "discovered") + "/"

		_, _, err := runDiscover(t, "", "-kustomize", "apps/shop", "-enable-helm", "-o", discovered)

		require.NoError(t, err)
		raw, err := os.ReadFile(filepath.Join(root, "discovered", "k8s-shop.yaml"))
		require.NoError(t, err)
		assert.Contains(t, string(raw), "k8s:shop/service/api")
		argv, err := os.ReadFile(argvFile)
		require.NoError(t, err)
		assert.Equal(t, "build\napps/shop\n--enable-helm\n", string(argv))
		entries, err := os.ReadDir(filepath.Join(root, "discovered"))
		require.NoError(t, err)
		assert.Len(t, entries, 1)
	})

	t.Run("missing-tool", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, "apps", "shop"), 0o755))
		t.Setenv("PATH", t.TempDir())
		t.Chdir(root)

		_, _, err := runDiscover(t, "", "-kustomize", "apps/shop", "-o", filepath.Join(root, "discovered")+"/")

		assert.ErrorContains(t, err, "kustomize")
		assert.ErrorContains(t, err, "nix shell nixpkgs#kustomize")
		assert.NoDirExists(t, filepath.Join(root, "discovered"))
	})

	t.Run("stdin-with-kustomize-fails", func(t *testing.T) {
		_, _, err := runDiscover(t, "", "-", "-kustomize", "apps/shop")

		assert.ErrorContains(t, err, "cannot be mixed with paths or -kustomize")
	})

	t.Run("enable-helm-without-kustomize-fails", func(t *testing.T) {
		_, _, err := runDiscover(t, "", "-enable-helm")

		assert.ErrorContains(t, err, "-enable-helm needs -kustomize")
	})
}
