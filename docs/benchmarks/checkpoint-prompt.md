The directories here are the code of the Majoutes platform as of 2026-09-29:
`argocd` (nfra/argocd: Argo CD ApplicationSets and the platform charts under
`tenants/sre/<app>/<env>`), `k8s-majoutes` (the application charts under
`majoutes/<app>/<env>`), `webapp` (the application source) and `talos` (the cluster).
Argo CD deploys each `<app>/<env>/Chart.yaml` as one Application: release name
`<app>`, namespace `<app>` for sre (or `destinationNamespace` from
`.argocd-override.yaml`) and `<env>-majoutes` for majoutes.

Collect the facts for an architecture diagram of the whole platform: every network
connection that a component running inside the Kubernetes cluster opens, to another
component of the cluster or to a service outside it (databases, SaaS, registries, Git,
DNS, identity providers). Include prod and preprod.

Write them to `facts.yaml`, nothing else, in this shape:

```yaml
connections:
  - from: <component>
    to: <component>
    port: <number or name, if known>
    protocol: <if known>
    evidence: <file:line or file path that shows it>
```

Use the same component name for the same thing everywhere. Only list what the files
support. When you are done, reply with the number of connections written.
