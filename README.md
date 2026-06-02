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
│   ├── stress_test.py
│   └── stress_test_v2.py
├── cata.proto
├── pb/
└── agentic_scenario.py
```

## Current Status

- The Go middleware server compiles with the project-local Go cache command.
- Python scripts currently need dependency and protobuf runtime alignment before
  the full demo can be considered stable.
- `proto/middleware.proto` should become the only canonical protobuf contract.
- `cata.proto` and `pb/cata_*` are legacy scenario files and should be migrated or
  removed after the unified demo path is ready.

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

## Next Milestones

1. Initialize and connect the GitHub repository.
2. Unify protobuf definitions around `proto/middleware.proto`.
3. Add Python dependency management.
4. Regenerate protobuf files for Go and Python.
5. Build a deterministic demo script for the graduation presentation.
