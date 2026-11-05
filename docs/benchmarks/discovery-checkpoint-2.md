# Discovery checkpoint, round 2 (spec 009 T23)

Run 2026-11-05 on the T09 snapshot, after revision 1 (ADR-0031). Arm C gets
`noodle digest` of the fragments instead of the fragments; arm A is run again. Rule of
ADR-0029: `go` exactly when C costs at most 70 % of A **and** finds at least as many of
the 25 in-cluster edges.

Tokens: fragments ~35.0k, digest ~6.4k (18 %, bound 25 %, FR-025)

| Arm | Cost | Edges | Transcript |
|---|---|---|---|
| A | $4.82 | 25/25 | docs/benchmarks/transcripts/checkpoint2-A.jsonl |
| C | $3.04 | 22/25 | docs/benchmarks/transcripts/checkpoint2-C.jsonl |

Decision: stop

C costs 63 % of A, which passes the cost half, but finds three edges fewer.

## Protocol

As in [round 1](discovery-checkpoint.md), with [`run-arm.sh`](run-arm.sh) round `2`:
same snapshot, model, tools and prompt. C's paragraph is
[`checkpoint-prompt-c2.md`](checkpoint-prompt-c2.md). Fragments come from
[`render-pae.sh`](render-pae.sh), now with relative `src`. `result.json` is now written
outside the work directory, since both arms opened it in round 1. The CLI's own
`total_cost_usd`: A $3.66, C $2.39.

| | A | C |
|---|---|---|
| Turns | 39 | 30 |
| Wall time | 6m39s | 6m22s |
| Connections written | 133 | 92 |

## The three edges C misses

| # | Edge | Why the digest has no line for it |
|---|---|---|
| 17 | Grafana → Prometheus | a datasource in the chart values, not a manifest field the adapter reads |
| 18 | Prometheus → Alertmanager | implied by the `Prometheus` custom resource: an operator rule (FR-018, wave 3) |
| 21 | nodes → Docker Hub | image pulls are not connections in a fragment |

These are the operator and controller blind spot of 2026-09-30. The digest lists the
`Prometheus` and `Alertmanager` resources as components and as `unknown-kind`, but
draws no edge between them. C did not go looking, because its paragraph says the
digest "is complete for the connections between components inside the cluster". That
claim was wrong for these edges. The prompt is part of the arm, so the run counts as it is.

## What changed from round 1

- The digest works on cost: C went from 110 % of A to 63 %, with 9 fewer turns than A.
- `secret-endpoint` works on recall. C took api → Keycloak and every database edge from
  the Secret key names, without reading `application.properties` for them.
- A went from $4.34 to $4.82 between two identical runs: about 10 % of noise per run.
