# Discovery checkpoint (spec 009 T09)

Run 2026-11-05 on PAE as of 2026-09-29, rule of ADR-0029: `go` exactly when arm C costs
at most 70 % of arm A **and** finds at least as many of the 25 in-cluster edges.

| Arm | Cost | Edges | Transcript |
|---|---|---|---|
| A | $4.34 | 24/25 | docs/benchmarks/transcripts/checkpoint-A.jsonl |
| C | $4.76 | 25/25 | docs/benchmarks/transcripts/checkpoint-C.jsonl |

Decision: stop

C costs 110 % of A, above the 70 % threshold, although it finds one more edge.

## Protocol

- Inputs: `argocd` b9f9230, `k8s-majoutes` db49537, `webapp` 4387720 (all 2026-09-29) and
  `talos` fff8518, exported without `.git`. The ground truth is
  [`platform-overview-edges.md`](platform-overview-edges.md).
- Fragments for C: [`render-pae.sh`](render-pae.sh) renders each chart as its
  ApplicationSet does and runs `noodle discover k8s -namespace` per Application:
  24 fragments, about 150 KB.
- Each arm: [`run-arm.sh`](run-arm.sh), one fresh `claude -p` with `claude-opus-5-5`,
  no user settings, hooks, plugins, MCP or skills, no subagent and no web. Tools: Read,
  Glob, Grep, Write and read-only shell commands. Same prompt
  ([`checkpoint-prompt.md`](checkpoint-prompt.md)); C gets one more paragraph
  ([`checkpoint-prompt-c.md`](checkpoint-prompt-c.md)). Output: `checkpoint-<arm>.facts.yaml`.
- Cost: `go run ./tools/tokencost` on the transcript (its prices are an assumption per
  model family). The CLI's own `total_cost_usd` agrees on the order: A $3.37, C $3.47.

## Measures

| | A | C |
|---|---|---|
| Turns | 35 | 48 |
| Wall time | 7m02s | 7m54s |
| Connections written | 109 | 112 |
| Cost by tool, bash / grep / read / write | $2.68 / $0.40 / $0.46 / $0.65 | $2.98 / $0.79 / $0.26 / $0.52 |

The edge A misses is #2, majoutes-api prod to Keycloak prod. A writes it as
unresolved because the URL is in 1Password, although it found the same edge in
preprod. C draws it from a fragment.

## Why C is not cheaper

- C did not trust the fragments enough to stop reading. It still opened the value files
  and the application config for the external services (S3, SMTP, Discord, the database
  VMs), which no fragment holds: they come from Secret keys and `application.properties`.
  That is the credential-implied blind spot of 2026-09-30.
- The fragments themselves are not small. About 38k tokens over 24 files, mostly
  configuration references to ConfigMaps and Secrets that a diagram does not need.
  Every later turn re-reads them from the cache.
- Turn count drives cost: C took 13 more turns, mostly `grep` and `cat` across
  `discovered/`.
- One run per arm is noisy (ADR-0029, Consequences). The gap is still too far from
  70 % for noise to explain it.
