package k8s_test

import (
	"context"
	"crypto/sha256"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/adapter"
	"github.com/TheGostsniperfr/Noodle/internal/adapter/k8s"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

const renderedShop = `apiVersion: apps/v1
kind: Deployment
metadata: {name: api, namespace: shop}
spec:
  template:
    spec:
      containers: [{name: api, image: acme/api:1}]
---
apiVersion: v1
kind: Service
metadata: {name: api, namespace: shop}
spec: {selector: {app: api}, ports: [{port: 80}]}
`

// fakeKustomize puts a kustomize on PATH that records its argv, one argument per line,
// and prints rendered. It returns the file holding the argv.
func fakeKustomize(t *testing.T, rendered string) string {
	t.Helper()
	bin, argvFile, outFile := t.TempDir(), filepath.Join(t.TempDir(), "argv"), filepath.Join(t.TempDir(), "rendered.yaml")
	require.NoError(t, os.WriteFile(outFile, []byte(rendered), 0o644))
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"" + argvFile + "\"\ncat \"" + outFile + "\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(bin, "kustomize"), []byte(script), 0o755))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return argvFile
}

func argv(t *testing.T, file string) []string {
	t.Helper()
	raw, err := os.ReadFile(file)
	require.NoError(t, err)
	return strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
}

// shopDir makes apps/shop in a fresh working directory and returns its file hashes.
func shopDir(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "apps", "shop"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "apps", "shop", "kustomization.yaml"), []byte("resources: [deployment.yaml]\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "apps", "shop", "deployment.yaml"), []byte("kind: Deployment\n"), 0o644))
	t.Chdir(root)
}

func hashes(t *testing.T, dir string) map[string][32]byte {
	t.Helper()
	out := map[string][32]byte{}
	require.NoError(t, filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, err := os.ReadFile(path)
		out[path] = sha256.Sum256(raw)
		return err
	}))
	return out
}

func TestKustomize(t *testing.T) {
	t.Run("render", func(t *testing.T) {
		shopDir(t)
		argvFile := fakeKustomize(t, renderedShop)
		before := hashes(t, "apps/shop")

		f, _, err := k8s.Adapter{}.Discover(context.Background(), adapter.Source{Kustomize: "apps/shop"})

		require.NoError(t, err)
		assert.Equal(t, []string{"build", "apps/shop"}, argv(t, argvFile))
		assert.Equal(t, before, hashes(t, "apps/shop"))
		assert.ElementsMatch(t, []string{"k8s:shop/deployment/api", "k8s:shop/service/api", "k8s:shop/namespace/shop"}, ids(f))
		assert.Equal(t, model.Src{File: "apps/shop/kustomization.yaml"}, element(t, f, "k8s:shop/deployment/api").Src)
	})

	t.Run("enable-helm", func(t *testing.T) {
		shopDir(t)
		argvFile := fakeKustomize(t, renderedShop)

		_, _, err := k8s.Adapter{}.Discover(context.Background(), adapter.Source{Kustomize: "apps/shop", EnableHelm: true})

		require.NoError(t, err)
		assert.Equal(t, []string{"build", "apps/shop", "--enable-helm"}, argv(t, argvFile))
	})

	t.Run("missing-tool", func(t *testing.T) {
		shopDir(t)
		t.Setenv("PATH", t.TempDir())

		f, _, err := k8s.Adapter{}.Discover(context.Background(), adapter.Source{Kustomize: "apps/shop"})

		assert.Nil(t, f)
		assert.ErrorContains(t, err, "kustomize not found")
		assert.ErrorContains(t, err, "nix shell nixpkgs#kustomize")
	})

	t.Run("a-failing-build-is-reported", func(t *testing.T) {
		shopDir(t)
		bin := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(bin, "kustomize"), []byte("#!/bin/sh\necho 'bad patch' >&2\nexit 1\n"), 0o755))
		t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

		_, _, err := k8s.Adapter{}.Discover(context.Background(), adapter.Source{Kustomize: "apps/shop"})

		assert.ErrorContains(t, err, "bad patch")
	})
}

func TestSrcPaths(t *testing.T) {
	assertRelative := func(t *testing.T, f *model.Fragment, prefix string) {
		t.Helper()
		var files []string
		for _, e := range f.Elements {
			files = append(files, e.Src.File)
		}
		for _, c := range f.Connections {
			files = append(files, c.Src.File)
		}
		for _, r := range f.References {
			files = append(files, r.Src.File)
		}
		require.NotEmpty(t, files)
		for _, file := range files {
			assert.True(t, strings.HasPrefix(file, prefix), "%q does not start with %q", file, prefix)
			assert.False(t, strings.HasPrefix(file, "/"), file)
			assert.NotContains(t, file, os.TempDir())
		}
	}

	t.Run("manifests", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, "manifests"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(root, "manifests", "shop.yaml"), []byte(renderedShop), 0o644))
		t.Chdir(root)

		f, _, err := k8s.Adapter{}.Discover(context.Background(), adapter.Source{Paths: []string{filepath.Join(root, "manifests")}})

		require.NoError(t, err)
		assertRelative(t, f, "manifests/")
	})

	t.Run("kustomize", func(t *testing.T) {
		shopDir(t)
		fakeKustomize(t, renderedShop)
		abs, err := filepath.Abs("apps/shop")
		require.NoError(t, err)

		f, _, err := k8s.Adapter{}.Discover(context.Background(), adapter.Source{Kustomize: abs})

		require.NoError(t, err)
		assertRelative(t, f, "apps/shop/")
	})

	t.Run("unrendered-directories", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, "overlay"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(root, "overlay", "kustomization.yaml"), []byte("resources: []\n"), 0o644))
		t.Chdir(root)

		f, _, err := k8s.Adapter{}.Discover(context.Background(), adapter.Source{Paths: []string{root}})

		require.NoError(t, err)
		assert.Equal(t, []model.Unresolved{{Kind: "unrendered", Value: "overlay", Hint: "kustomize", Src: model.Src{File: "overlay/kustomization.yaml", Line: 1}}}, f.Unresolved)
	})
}
