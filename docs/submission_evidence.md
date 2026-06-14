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
