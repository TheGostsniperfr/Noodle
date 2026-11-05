# Discovery checkpoint, round 3 and last (spec 009 T29)

Run 2026-11-05 on the T09 snapshot, after revision 2 (ADR-0032). As in round 2, arm C gets
`noodle digest` of the fragments, and both arms are run again. Only C's paragraph changes
([`checkpoint-prompt-c3.md`](checkpoint-prompt-c3.md)): it no longer calls the digest
complete, and lists what discovery never sees. Rule of ADR-0029: `go` exactly when C
costs at most 70 % of A **and** finds at least as many of the 25 in-cluster edges.

| Arm | Cost | Edges | Transcript |
|---|---|---|---|
| A | $3.89 | 23/25 | docs/benchmarks/transcripts/checkpoint3-A.jsonl |
| C | $4.91 | 25/25 | docs/benchmarks/transcripts/checkpoint3-C.jsonl |

Decision: stop

C costs 126 % of A. It finds two edges more: Grafana → Prometheus and Prometheus →
Alertmanager, which A misses. Under ADR-0031 and ADR-0032 this is final: discovery stops
serving the agent, and phase 6a closes without its exit criterion.

## Measures

| | A | C |
|---|---|---|
| Turns | 34 | 45 |
| Wall time | 6m08s | 7m24s |
| Connections written | 86 | 157 |

Same snapshot as rounds 1 and 2 (`argocd` b9f9230, `k8s-majoutes` db49537, `webapp`
4387720, `talos` fff8518), same fragments and digest (~6.4k tokens).

## The three rounds

| Round | C gets | A | C | C / A | Edges A / C |
|---|---|---|---|---|---|
| 1 | fragments | $4.34 | $4.76 | 110 % | 24 / 25 |
| 2 | digest, "complete" | $4.82 | $3.04 | 63 % | 25 / 22 |
| 3 | digest, "never sees" list | $3.89 | $4.91 | 126 % | 23 / 25 |

- **No arm reaches both halves.** C is cheaper only when it trusts the digest, and then it
  misses what the digest cannot hold. Once told about those gaps, it reads the repositories
  as A does, and even more.
- **Recall is where discovery helps.** In two rounds out of three, C finds edges that A
  misses: api → Keycloak from a `secret-endpoint`, and the Prometheus edges once C knows
  to look. A's recall moves between 23 and 25 across identical runs.
- **The noise is large.** A's cost moves between $3.89 and $4.82, about ±11 %, on the same
  prompt and input. One run per arm cannot resolve a 70 % threshold better than that.
- **What would change the answer.** The missing edges come from operator custom resources
  and chart values (FR-018, wave 3). Discovery would have to emit those edges itself, so
  that the digest really is complete. That is the work this stop does not fund.
