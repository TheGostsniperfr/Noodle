# PAE platform-overview: the 25 in-cluster edges

Ground truth of the discovery checkpoint (spec 009 FR-014, ADR-0029). Source:
`gitlab.sre.paris/pae/documentation`, `architecture/systems/platform/model.yaml` at
`5e56671`, the connections of the overview as of 2026-09-29, before the observability
proposal. They are 31; the 25 below are those whose source runs inside the Majoutes
cluster (`z-cluster`). The six left out start outside it: user to edge, edge to Envoy,
ops to the apiserver, developer to GitLab, and GitLab CI to Docker Hub and to the repos.

An arm finds an edge when one of its connections joins the same two components in the
same direction; port and protocol are not required.

| # | Id | From | To |
|---|---|---|---|
| 1 | e-3-envoy-api | Envoy shared-gateway | majoutes-api (prod) |
| 2 | e-4-api-keycloak | majoutes-api (prod) | Keycloak (prod) |
| 3 | e-5-api-pg | majoutes-api (prod) | PostgreSQL prod VM |
| 4 | e-api-s3 | majoutes-api (prod) | AWS S3 |
| 5 | e-api-smtp | majoutes-api (prod) | Gmail SMTP |
| 6 | e-api-discord | majoutes-api (prod) | Discord |
| 7 | e-envoy-front | Envoy shared-gateway | majoutes-front (prod) |
| 8 | e-envoy-keycloak | Envoy shared-gateway | Keycloak (prod) |
| 9 | e-keycloak-pg | Keycloak (prod) | PostgreSQL prod VM |
| 10 | e-gatus-pg | Gatus (prod) | PostgreSQL prod VM |
| 11 | e-envoy-pp-front | Envoy shared-gateway | majoutes-front (preprod) |
| 12 | e-envoy-pp-api | Envoy shared-gateway | majoutes-api (preprod) |
| 13 | e-pp-api-keycloak | majoutes-api (preprod) | Keycloak (preprod) |
| 14 | e-pp-api-pg | majoutes-api (preprod) | PostgreSQL preprod VM |
| 15 | e-pp-keycloak-pg | Keycloak (preprod) | PostgreSQL preprod VM |
| 16 | e-pp-gatus-pg | Gatus (preprod) | PostgreSQL preprod VM |
| 17 | e-grafana-prometheus | Grafana | Prometheus |
| 18 | e-prometheus-alertmanager | Prometheus | Alertmanager |
| 19 | e-oidc-envoy-gitlab | Envoy shared-gateway | GitLab OIDC |
| 20 | e-d-argocd-repos | Argo CD | GitLab repositories |
| 21 | e-e-nodes-dockerhub | cluster nodes (kubelet) | Docker Hub |
| 22 | e-f-eso-onepassword | External Secrets | 1Password |
| 23 | e-g-certmanager-route53 | cert-manager | AWS Route 53 |
| 24 | e-h-certmanager-letsencrypt | cert-manager | Let's Encrypt |
| 25 | e-i-externaldns-route53 | external-dns | AWS Route 53 |
