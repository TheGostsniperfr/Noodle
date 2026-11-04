# How to discover a system from Kubernetes manifests

Discovery reads infrastructure files and writes a fragment: the elements, connections and
references they define, each with the file and line it came from. It does not edit your
model. You link your curated elements to the discovered ones with `matches`.

## Run it

```bash
noodle discover k8s manifests/ -o architecture/platform/discovered/k8s-manifests.yaml
```

With no path, or with `-`, noodle reads standard input. Helm charts and Kustomize overlays
are not read directly: render them first, then pipe the result in.

```bash
helm template my-release chart/ | noodle discover k8s - -o architecture/platform/discovered/k8s-helm.yaml
kustomize build overlays/prod | noodle discover k8s - -o architecture/platform/discovered/k8s-prod.yaml
```

Objects that name no namespace get `default`, as with `kubectl apply`. Set another with
`-namespace NS`. `-ref` records the version of the source, for example a Git commit.

The command prints one summary line on stderr: how many objects it read, what it skipped
and why, how many elements, connections and references it found, and how many items it
could not resolve. Run it twice on the same input and the fragment is identical, byte
for byte.

## Link your model to it

Discovered ids look like `k8s:<namespace>/<kind>/<name>`. In `model.yaml`, an element lists
the ids it stands for:

```yaml
- id: api
  kind: backend
  title: API
  matches: [k8s:prod/deployment/api, k8s:prod/service/api]
```

`noodle render` checks that a discovered id belongs to at most one element. A curated
element with no `matches` stays valid: discovery is optional. Merging discovered fields into
the model and reporting drift between the two (`noodle drift`) are not shipped yet; until
then, read the fragment, or have the `fact-finder` agent read it, and write the curated
part yourself.

## What it finds, and what it does not

Static discovery reads manifests. It knows workloads, services, ingress and gateway
routes, and the references between them, such as a secret mounted by a pod. It does not
know what is only decided at runtime, and it does not read live clusters. What it cannot
resolve is listed as unresolved in the fragment, so you or an agent decide.

Fragments are generated: do not commit them, and do not edit them.
