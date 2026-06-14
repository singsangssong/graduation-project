# Submission Evidence And Presentation Notes

## Research Implementation Claim

The submitted system is a SagaLLM-compatible transaction middleware rather than
a reproduction of the complete SagaLLM framework.

It connects Saga-style workflow reliability with shared-resource concurrency:

```text
Saga checkpoint and validation
  + QCFuse-style resource-aware read fusion
  + ATCC-style cost-aware commit arbitration
  + durable compensation and recovery
```

## Implemented Professional Features

- Durable SQLite Saga and checkpoint storage.
- Append-only Saga event timeline.
- Middleware restart recovery.
- Actual transactional ticket reserve/release handlers.
- Reverse-order compensation.
- Idempotent duplicate compensation.
- Explicit `COMPENSATION_FAILED` state for unsupported handlers.
- Resource-aware QCFuse grouping by `resource_id + intent`.
- ATCC winner commit and loser compensation integration.
- Concurrency checks with Go race detector.

## Evidence To Capture For Slides And Final Paper

Run the persistent Saga demo with three available tickets:

```sh
cd middleware-go
TICKET_STOCK=3 SAGA_DB_PATH=data/middleware.db \
  ATCC_WINDOW_MS=500 \
  GOCACHE=/private/tmp/agenic-middleware-gocache go run .
```

In another terminal:

```sh
cd agent-python
python3 saga_demo.py
```

Capture:

- resource stock before the concurrent workflow.
- resource stock after one commit and two compensations.
- one `COMMITTED` Saga and two `COMPENSATED` Sagas.
- event timelines for each Saga.
- QCFuse fused batch and saved DB reads.
- compensation action count.

Then restart the Go server with the same `SAGA_DB_PATH` and run:

```sh
cd agent-python
python3 recovery_check.py
```

Capture that all Saga states, steps, and event counts are recovered.

## Verified Persistent Demo Result

The following result was directly verified on June 14, 2026 using a fresh
SQLite database and three concurrent Sagas:

```text
Saga-Agent-A: COMMITTED
Saga-Agent-B: COMPENSATED
Saga-Agent-C: COMPENSATED

Resource stock: 3 -> 2
Sagas started: 3
Sagas validated: 3
Sagas compensated: 2
Compensation actions: 2
QCFuse read requests: 3 -> 1 fused batch
Saved DB reads: 2
```

Event timelines:

```text
Winner:
SAGA_STARTED -> STEP_COMPLETED -> VALIDATION_PASSED -> SAGA_COMMITTED

Losers:
SAGA_STARTED -> STEP_COMPLETED -> VALIDATION_PASSED
-> SAGA_ABORTED -> COMPENSATION_COMPLETED -> SAGA_COMPENSATED
```

After stopping and restarting the Go middleware with the same SQLite database:

```text
Saga-Agent-A: COMMITTED -> COMMITTED, events=4
Saga-Agent-B: COMPENSATED -> COMPENSATED, events=6
Saga-Agent-C: COMPENSATED -> COMPENSATED, events=6
Persistent resource stock: 2
```

## Suggested Final Paper Evaluation Sections

1. **Durability test**
   - Execute workflows, restart middleware, compare state before/after restart.
2. **Compensation correctness test**
   - Verify reserve decreases stock and compensation restores it exactly once.
3. **Concurrent arbitration test**
   - Verify one ATCC winner and compensation of losing Sagas.
4. **Read-fusion correctness test**
   - Verify same resource/intent requests fuse while different resources do not.
5. **Failure-handling test**
   - Verify unsupported compensation becomes `COMPENSATION_FAILED`.
6. **Concurrency-safety test**
   - Report `go test -race ./...` result.

## Honest Limitations For Presentation

- SQLite is used for a single-node submission-ready deployment.
- Compensation handlers currently cover typed ticket reserve/release actions.
- Deterministic validation is implemented; an LLM validator remains future work.
- ATCC uses a simplified sunk-cost score rather than an RL policy.

## Verified Automated Live Profile

The automated `live` profile was directly verified on June 14, 2026. It starts
fresh middleware processes for each comparison mode and runs the same seeded
workload.

| Mode | Agents | TPS | Logical I/O reduction | Winner protection |
| --- | ---: | ---: | ---: | ---: |
| baseline | 10 | 120.8 | 0.0% | 25.6% |
| baseline | 50 | 645.4 | 0.0% | 57.4% |
| qcfuse | 10 | 94.6 | 90.0% | 37.7% |
| qcfuse | 50 | 628.4 | 98.0% | 57.4% |
| full | 10 | 102.7 | 90.0% | 100.0% |
| full | 50 | 627.9 | 98.0% | 100.0% |

The live profile also passed all eight automated reliability checks:

- exactly one concurrent Saga committed;
- two losing Sagas compensated;
- compensation restored stock exactly once;
- duplicate compensation remained idempotent;
- unsupported compensation was detected;
- Saga states recovered after restart;
- event timelines recovered after restart;
- resource stock recovered after restart.

TPS values from the short live profile are presentation evidence, not the final
paper result. The paper must report the five-repetition `paper` profile with
means and sample standard deviations.

## Verified Five-Repetition Paper Profile

The final paper profile was re-verified on June 15, 2026 at 00:58:09 KST.
The experiment uses a commit barrier so that all agents in a run compete inside
the same ATCC arbitration window.

| Mode | Agents | Mean TPS ± SD | Mean p95 ms ± SD | I/O reduction | Winner protection |
| --- | ---: | ---: | ---: | ---: | ---: |
| baseline | 10 | 67.4 ± 2.1 | 146.7 ± 4.9 | 0.0% | 35.2% |
| baseline | 50 | 338.2 ± 7.0 | 141.6 ± 2.6 | 0.0% | 54.1% |
| baseline | 100 | 682.3 ± 21.1 | 139.1 ± 4.3 | 0.0% | 31.6% |
| baseline | 200 | 1357.8 ± 157.6 | 132.7 ± 15.5 | 0.0% | 57.1% |
| qcfuse | 10 | 70.4 ± 3.1 | 140.3 ± 6.2 | 90.0% | 48.3% |
| qcfuse | 50 | 340.3 ± 23.0 | 141.6 ± 9.7 | 98.0% | 50.6% |
| qcfuse | 100 | 677.0 ± 5.6 | 139.7 ± 1.0 | 98.6% | 43.3% |
| qcfuse | 200 | 1356.7 ± 9.6 | 134.3 ± 2.5 | 99.4% | 31.1% |
| full | 10 | 66.8 ± 4.2 | 148.6 ± 10.2 | 90.0% | 100.0% |
| full | 50 | 339.4 ± 8.0 | 142.2 ± 3.2 | 97.6% | 100.0% |
| full | 100 | 676.8 ± 3.4 | 141.5 ± 1.3 | 98.8% | 100.0% |
| full | 200 | 1353.7 ± 12.1 | 137.5 ± 1.7 | 99.3% | 100.0% |

All configurations recorded a 0% request error rate. In this controlled
simulation, read fusion substantially reduced logical I/O while throughput and
p95 latency remained similar across comparison modes. ATCC-style arbitration
consistently selected the maximum-cost candidate in the full mode.
