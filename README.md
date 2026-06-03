# Agentic Middleware Graduation Project

This project demonstrates an agentic middleware system for coordinating concurrent
AI-agent transactions over shared resources.

## What This Project Shows

The current prototype focuses on two middleware ideas:

- **QCFuse**: fuses many simultaneous read requests into one logical resource lookup.
- **ATCC**: chooses the winning commit request by agent-side sunk cost, using token
  usage and inference latency as arbitration signals.

The intended demo scenario is a high-contention ticket purchase:

1. Many agents read the same limited resource.
2. Agents spend different amounts of reasoning cost.
3. Agents concurrently request a commit.
4. Middleware approves the highest-cost request and rolls back the rest.

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
│   ├── stress_test.py
│   └── stress_test_v2.py
├── requirements.txt
└── agentic_scenario.py
```

## Current Status

- The Go middleware server compiles with the project-local Go cache command.
- `proto/middleware.proto` is the canonical protobuf contract.
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
python3 -m py_compile agent-python/agent_client.py agent-python/stress_test.py agent-python/stress_test_v2.py agentic_scenario.py
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

## Next Milestones

1. Record a demo video using `deterministic_demo.py` and `dashboard.html`.
2. Expand report figures with 10/50/100/200 agent stress-test results.
3. Stabilize the optional AutoGen scenario with real environment variables.
