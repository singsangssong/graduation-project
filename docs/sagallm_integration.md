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

## Persistent Store And Compensation

The middleware now persists the following SQLite tables:

- `sagas`: workflow state and validation/abort metadata.
- `saga_steps`: ordered checkpoints and compensation actions.
- `saga_events`: append-only workflow transition timeline.
- `resources`: durable shared-resource stock.
- `resource_reservations`: idempotent Saga reservation state.

The built-in typed resource handlers support:

```text
reserve <resource_id>
release <resource_id>
```

Registering a reserve step decreases durable stock. When ATCC rejects the Saga,
the reverse compensation handler releases the reservation and restores stock.
Duplicate release attempts are idempotent and do not restore stock twice.

Unsupported compensation handlers move the Saga to `COMPENSATION_FAILED` and
record a failure event instead of silently succeeding.

After middleware restart, `SQLiteSagaStore` reconstructs Saga records and steps.
The event timeline remains queryable through `/events`.

## Current Limitations

- Built-in compensation execution currently supports ticket-style
  `reserve/release` resource handlers. External API handlers are future work.
- Validation is deterministic; an independent LLM validation agent is not yet
  connected.
- ATCC uses a simplified token/latency score instead of the paper's full
  adaptive/RL concurrency-control policy.

## Next Production-Oriented Extensions

1. Replace SQLite with PostgreSQL for multi-node deployment.
2. Add explicit client-supplied idempotency keys to checkpoint execution.
3. Execute typed compensation handlers against external payment/resource APIs.
4. Add an optional LLM validator after deterministic validation.
5. Automatically resume incomplete Sagas after middleware restart.
6. Measure real database lock wait time and isolation-level behavior.
