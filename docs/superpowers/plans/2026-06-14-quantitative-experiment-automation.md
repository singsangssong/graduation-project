# Quantitative Experiment Automation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build reproducible baseline/QCFuse/full performance experiments, Saga reliability checks, a live presentation profile, and paper-ready result artifacts.

**Architecture:** The Go server exposes controlled experiment policies while preserving `full` as the default. A Python runner starts isolated server processes, executes seeded workloads, aggregates statistics, verifies persistent Saga recovery, and renders a standalone HTML report.

**Tech Stack:** Go, gRPC, SQLite, Python standard library, generated protobuf clients

---

### Task 1: Add Controlled Server Experiment Modes

**Files:**
- Modify: `middleware-go/main.go`
- Modify: `middleware-go/main_test.go`
- Modify: `middleware-go/qcfuse_test.go`

- [ ] Write failing tests for mode normalization, fusion policy, and commit ranking policy.
- [ ] Run focused Go tests and confirm they fail because policies do not exist.
- [ ] Implement `EXPERIMENT_MODE=baseline|qcfuse|full` with `full` as default.
- [ ] Add logical DB read and maximum candidate cost metrics.
- [ ] Run focused and complete Go tests.
- [ ] Commit the server policy change.

### Task 2: Add Experiment Statistics Library

**Files:**
- Create: `agent-python/experiment_stats.py`
- Create: `agent-python/test_experiment_stats.py`

- [ ] Write failing tests for percentiles, grouped mean/sample standard deviation, reliability rates, and HTML rendering.
- [ ] Run the Python unit tests and confirm expected failures.
- [ ] Implement minimal deterministic statistics and report rendering.
- [ ] Run Python tests and syntax checks.
- [ ] Commit the statistics library.

### Task 3: Build Automated Experiment Runner

**Files:**
- Create: `agent-python/experiment_runner.py`
- Modify: `.gitignore`
- Modify: `README.md`

- [ ] Implement live and paper profile definitions.
- [ ] Implement isolated Go server subprocess lifecycle and readiness checks.
- [ ] Implement seeded concurrent scalability workload.
- [ ] Write raw CSV, summary CSV/JSON, reliability JSON, and HTML report.
- [ ] Add concise usage documentation.
- [ ] Run the live profile end to end.
- [ ] Commit the runner.

### Task 4: Automate Saga Reliability Evaluation

**Files:**
- Modify: `agent-python/experiment_runner.py`
- Modify: `agent-python/experiment_stats.py`
- Modify: `agent-python/test_experiment_stats.py`

- [ ] Add reliability result tests.
- [ ] Run tests and confirm the new assertions fail.
- [ ] Execute persistent Saga competition, duplicate compensation checks, restart recovery, and unsupported-handler checks.
- [ ] Produce reliability rates and evidence in the HTML report.
- [ ] Run the live profile and verify all checks pass.
- [ ] Commit reliability automation.

### Task 5: Integrate Ten-Minute Presentation Scenario

**Files:**
- Create: `docs/presentation_10min_scenario.md`
- Modify: `docs/submission_evidence.md`
- Modify: `docs/final_presentation_strategy.md`

- [ ] Document the exact live command and fallback using precomputed results.
- [ ] Define a ten-minute narration timeline.
- [ ] Map each generated metric to a presentation claim and its limitation.
- [ ] Commit presentation integration.

### Task 6: Final Verification And Publication

**Files:**
- Modify only files requiring fixes discovered by verification.

- [ ] Run `go test ./...`, `go vet ./...`, and `go test -race ./...`.
- [ ] Run Python unit tests, syntax checks, and generated protobuf import checks.
- [ ] Run `experiment_runner.py --profile live`.
- [ ] Run `git diff --check`.
- [ ] Record verified results in `docs/submission_evidence.md`.
- [ ] Commit and push the completed experiment environment.
