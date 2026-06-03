# AGENTS.md

## Project Goal

This repository is a graduation project for building an agentic middleware system.
The final goal is a demo-ready system that clearly shows how middleware can protect
shared resources when multiple AI agents read, reason, and commit concurrently.

## Architecture

- `proto/`: canonical Protocol Buffers definitions.
- `middleware-go/`: Go gRPC middleware server.
- `middleware-go/pb/`: generated Go protobuf files.
- `agent-python/`: Python clients and stress-test scripts.
- `agentic_scenario.py`: optional AutoGen demo scenario that uses the canonical
  Python protobuf files under `agent-python/`.

## Core Concepts

- QCFuse: batches concurrent read requests into a single fused resource lookup.
- ATCC: arbitrates commit requests by accumulated agent cost, such as token usage
  and inference latency.
- Saga-style rollback: rejects losing commit requests and reports protected cost.

## Development Rules

- Treat `proto/middleware.proto` as the single source of truth for the middleware API.
- Do not manually edit generated protobuf files. Regenerate them from the proto file.
- Keep demo output easy to understand in Korean when it helps the presentation.
- Prefer deterministic scripts and clear summaries for graduation demo scenarios.
- Avoid unrelated refactors while fixing execution or demo stability.

## Useful Commands

Go middleware:

```sh
cd middleware-go
GOCACHE=/private/tmp/agenic-middleware-gocache go test ./...
go run .
```

Python syntax check:

```sh
python3 -m py_compile agent-python/agent_client.py agent-python/stress_test.py agent-python/stress_test_v2.py agentic_scenario.py
```

## Verification Rules

Before claiming work is complete:

- Run Go build/test checks.
- Run Python syntax/import checks for touched scripts.
- If protobuf changes, regenerate Go and Python generated files.
- Confirm the demo flow still works end to end:
  Go server -> Python agents -> QCFuse read batching -> ATCC commit arbitration.

## Presentation Priority

When choosing between a clever implementation and a clear demo, prefer the clear demo.
The system should be easy to explain, easy to run, and convincing during a live
graduation-project presentation.
