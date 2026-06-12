# SagaLLM-Compatible Transaction Middleware

## Purpose

SagaLLM focuses on reliable multi-agent workflows through checkpointing,
validation, compensation, and recovery. This project provides the shared-resource
transaction layer that a SagaLLM-style workflow can call when agents read and
commit external resources concurrently.

The implementation does not reproduce the complete SagaLLM framework. Instead,
it exposes a Saga-compatible lifecycle and integrates it with QCFuse-style read
fusion and ATCC-style commit arbitration.

## Lifecycle

```text
BeginSaga
  -> RegisterSagaStep
  -> ReadResource
  -> Agent reasoning
  -> ValidateSaga
  -> CommitTransaction
       -> winner: COMMITTED
       -> loser: ABORTED -> reverse compensation -> COMPENSATED
```

## State Model

Saga states:

- `ACTIVE`: accepts new checkpoints.
- `VALIDATED`: deterministic pre-commit validation passed.
- `VALIDATION_FAILED`: required checkpoint validation failed.
- `COMMITTED`: ATCC approved the transaction.
- `ABORTED`: the Saga was rejected or explicitly aborted.
- `COMPENSATED`: all completed steps were compensated in reverse order.

Step states:

- `COMPLETED`: action completed and may need compensation.
- `COMPENSATED`: registered compensation action was processed.

## API Mapping

| API | SagaLLM-compatible role |
|---|---|
| `BeginSaga` | Creates a workflow transaction and stable `saga_id` |
| `RegisterSagaStep` | Stores a completed checkpoint and compensation action |
| `ValidateSaga` | Runs deterministic pre-commit validation |
| `ReadResource` | Performs resource-aware QCFuse read fusion |
| `CommitTransaction` | Submits a validated Saga to ATCC arbitration |
| `AbortSaga` | Aborts and compensates completed steps in reverse order |
| `GetSagaState` | Exposes workflow state for agents, dashboards, and recovery |

## Integration With ATCC

`CommitTransaction` accepts a `saga_id`.

Before entering ATCC arbitration, the Saga must be `VALIDATED`. When the commit
batch is processed:

- The highest sunk-cost request is approved and its Saga becomes `COMMITTED`.
- Losing requests are rejected and their Sagas become `COMPENSATED`.
- Compensation actions are processed in reverse checkpoint order.

This turns the previous rollback response into a workflow-level state transition.

## Integration With QCFuse

QCFuse batching is isolated by:

```text
resource_id + intent
```

Requests for different resources or intents are never fused into the same
logical read. This corrects the original demo behavior, which could mix all
requests arriving in the same scheduling window.

## Metrics

The `/metrics` endpoint now reports:

- `sagas_started`
- `saga_steps_registered`
- `sagas_validated`
- `saga_validation_failures`
- `sagas_compensated`
- `compensation_actions`

These metrics are displayed by `dashboard.html`.

## Current Limitations

- Saga state is stored in memory and is not recovered after middleware restart.
- Compensation actions are recorded and transitioned, but external compensation
  APIs are not executed.
- Validation is deterministic; an independent LLM validation agent is not yet
  connected.
- The resource layer still uses the controlled ticket-stock simulation.
- ATCC uses a simplified token/latency score instead of the paper's full
  adaptive/RL concurrency-control policy.

## Next Production-Oriented Extensions

1. Persist Saga state and steps to PostgreSQL.
2. Add idempotency keys to checkpoint and compensation execution.
3. Execute typed compensation handlers against real external resources.
4. Add an optional LLM validator after deterministic validation.
5. Recover incomplete Sagas after middleware restart.
6. Replace the simulated stock with real database transactions and lock metrics.
