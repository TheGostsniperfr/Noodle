The file `digest.txt` is the output of `noodle digest` on discovery fragments generated
statically from the rendered charts, one line per fact. It lists the workloads and the
connections that the manifests declare: routes from the gateway and ingress, Services to
their workloads, and addresses of known Services found in env and ConfigMaps. Each
`secret-endpoint` line names the Secret keys that hold an external address and the
workloads that read them; the values live in the secret manager, never in the
repositories. Its `src` paths point to the rendered charts, which are not here.

Static discovery never sees: connections implied by operator custom resources or by
chart values (datasources, alerting, scrape targets), image pulls, and anything that
runs outside the cluster. Start from the digest, and read the repositories for what it
does not answer.
