# Agentic Middleware Graduation Project

This project demonstrates an agentic middleware system for coordinating concurrent
AI-agent transactions over shared resources.

## What This Project Shows

The current prototype combines three middleware ideas:

- **QCFuse**: fuses many simultaneous read requests into one logical resource lookup.
- **ATCC**: chooses the winning commit request by agent-side sunk cost, using token
  usage and inference latency as arbitration signals.
- **SagaLLM-compatible lifecycle**: records workflow checkpoints, validates a saga
  before commit, and compensates completed steps when ATCC rejects a transaction.

The intended demo scenario is a high-contention ticket purchase:

1. Many agents read the same limited resource.
2. Agents spend different amounts of reasoning cost.
3. Each agent registers a Saga checkpoint and passes deterministic validation.
4. Agents concurrently request a commit.
5. Middleware commits the highest-cost request and compensates the losing Sagas.

## Repository Structure

```text
.
├── AGENTS.md
├── README.md
├── proto/
│   └── middleware.proto
├── middleware-go/
│   ├── main.go
│   └── pb/
├── agent-python/
│   ├── agent_client.py
│   ├── middleware_pb2.py
│   ├── middleware_pb2_grpc.py
│   ├── saga_demo.py
│   ├── stress_test.py
│   └── stress_test_v2.py
├── requirements.txt
└── agentic_scenario.py
```

## Current Status

- The Go middleware server compiles with the project-local Go cache command.
- `proto/middleware.proto` is the canonical protobuf contract.
- Saga lifecycle state is maintained by the Go `SagaCoordinator`.
- QCFuse batches are isolated by resource ID and intent to avoid cross-resource fusion.
- Python generated files under `agent-python/` are regenerated from
  `proto/middleware.proto`.
- `agentic_scenario.py` is an optional AutoGen-based scenario and requires the
  Python dependencies in `requirements.txt`.

## Run Checks

Go:

```sh
cd middleware-go
GOCACHE=/private/tmp/agenic-middleware-gocache go test ./...
```

Python syntax:

```sh
python3 -m py_compile agent-python/agent_client.py agent-python/deterministic_demo.py agent-python/saga_demo.py agent-python/stress_test.py agent-python/stress_test_v2.py agentic_scenario.py
```

Python generated-file import check:

```sh
python3 -c "import sys; sys.path.insert(0, 'agent-python'); import middleware_pb2, middleware_pb2_grpc"
```

## Regenerate Python Protobuf Files

```sh
python3 -m grpc_tools.protoc -I proto \
  --python_out=agent-python \
  --grpc_python_out=agent-python \
  --pyi_out=agent-python \
  proto/middleware.proto
```

## Run The Deterministic Demo

Start the middleware:

```sh
cd middleware-go
GOCACHE=/private/tmp/agenic-middleware-gocache go run .
```

In another terminal:

```sh
cd agent-python
python3 deterministic_demo.py
```

The demo writes JSON and CSV outputs to `outputs/demo-results/`.

## Run Quantitative Experiments

The experiment runner starts isolated Go middleware processes, executes the
baseline/QCFuse/full comparison, verifies persistent Saga reliability, and
generates CSV, JSON, and HTML evidence.

Presentation-friendly live profile:

```sh
./scripts/run_live_experiment.sh
```

Repeated paper profile:

```sh
./scripts/run_paper_experiment.sh
```

Results are written under `outputs/experiments/<timestamp>/`. Open
`report.html` to show the comparison table and reliability checks.

Experiment modes can also be selected manually:

```sh
EXPERIMENT_MODE=baseline  # individual logical reads + arrival-order commit
EXPERIMENT_MODE=qcfuse    # fused reads + arrival-order commit
EXPERIMENT_MODE=full      # fused reads + ATCC cost-aware commit
```

## Run The SagaLLM-Compatible Demo

Start the middleware, then run:

```sh
cd middleware-go
TICKET_STOCK=3 SAGA_DB_PATH=data/middleware.db \
  GOCACHE=/private/tmp/agenic-middleware-gocache go run .

# another terminal
cd agent-python
python3 saga_demo.py
```

The demo executes this lifecycle for each concurrent agent:

```text
BeginSaga
  -> RegisterSagaStep(checkpoint + compensation)
  -> ReadResource(QCFuse)
  -> ValidateSaga
  -> CommitTransaction(ATCC)
  -> COMMITTED or COMPENSATED
```

The result is written to `outputs/saga-demo-result.json`.

Stop and restart the middleware with the same `SAGA_DB_PATH`, then verify
durable recovery:

```sh
cd agent-python
python3 recovery_check.py
```

## Saga API

The canonical proto exposes:

- `BeginSaga`: creates an agent workflow transaction.
- `RegisterSagaStep`: records a completed checkpoint and its compensation action.
- `ValidateSaga`: performs deterministic pre-commit validation.
- `AbortSaga`: aborts the workflow and compensates completed steps in reverse order.
- `GetSagaState`: returns the current workflow state and step statuses.
- `CommitTransaction`: commits a validated Saga winner or compensates ATCC losers.

The Go server persists Saga state, steps, events, resource stock, and reservations
to SQLite. The `/events?saga_id=...` and `/resource?resource_id=...` HTTP endpoints
expose the durable event timeline and current resource state.

## Metrics Dashboard

The Go server exposes a metrics API at:

```text
http://localhost:8080/metrics
```

Open `dashboard.html` in a browser while the Go server is running to watch:

- QCFuse saved DB reads
- ATCC commit batches
- winner agent
- rollback count
- total saved cost
- Saga started/validated/compensated counts
- compensation action count

## Next Milestones

1. Record a demo video using `deterministic_demo.py` and `dashboard.html`.
2. Expand report figures with 10/50/100/200 agent stress-test results.
3. Stabilize the optional AutoGen scenario with real environment variables.
