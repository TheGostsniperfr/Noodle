# ADR-0018 · Credential-implied connections and operator rules

- Status: Accepted
- Date: 2026-10-01
- Extends: ADR-0015 (fragments), ADR-0017 (`unresolved`)

## Context

A recall check against a hand-made diagram of a real platform (25 in-cluster edges)
found 5 edges after spec 003 T04, about 11 expected after T05 and T06, and two blind
spots no task covered:

- **Credential-implied connections**, about 9 edges: a database, S3 or SMTP address
  lives in a secret manager. Git only holds the secret's name and its keys
  (`pae-prod-postgresql-majoutes/POSTGRES_JDBC_URL`, `KC_DB_URL_HOST`).
- **Operator and controller configuration**, 4 edges: a Prometheus CR names its
  Alertmanagers, external-dns runs with `--provider=aws`, a ClusterIssuer names its ACME
  server and DNS solver, a CloudNativePG `Cluster` is a database whose Services the
  operator creates.

Static discovery that misses half the edges is still useful only if the agent knows it
did, and the token savings it promises are not measured yet.

## Decision

- **`secret-endpoint`**, a new `unresolved` kind. A workload reading a Secret key whose
  name says it holds an address (`HOST`, `URL`, `URI`, `ENDPOINT`, `DSN`, `JDBC`,
  `ADDR`, `SERVER`) gets one entry per Secret: value `<secret>/<key>[,<key>…]` with the
  matching keys only, every reading workload in `about`. Values are never read (FR-008). An agent
  answers it once and the answer is remembered (ADR-0017).
- **Operator rules**: one small deterministic rule per operator or controller, each with
  a fixture: CloudNativePG `Cluster` (a database element and its `-rw`, `-ro`, `-r`
  Services), Prometheus operator (Prometheus to Alertmanager), cert-manager issuers
  (ACME server, DNS-01 provider), external-dns (`--provider`). A rule that infers from
  naming marks its connection `inferred`. New rules follow measured gaps, not lists of
  popular operators.
- **The fragment is a floor, not a ceiling.** The skill tells the agent which edge
  families discovery never sees (outside the cluster, CI, secret values) and to check
  them; the benchmark keeps failing a run whose recall drops below the agent alone.
- **A checkpoint before more rules.** After T05 and T06, the same diagram is built twice,
  agent alone and agent with the fragment, cost and recall compared. Discovery continues
  only if the fragment run costs clearly less and finds as many edges.

## Consequences

- `unresolved` kinds: ADR-0017's list plus `secret-endpoint`; its token guards apply.
- Image registries stay out: one element per registry adds lines for an edge a reader
  rarely needs.
- The benchmark tool (T12) moves before the checkpoint.
