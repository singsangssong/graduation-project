# Quantitative Experiment Automation Design

## Goal

Build a reproducible experiment environment that compares middleware behavior,
generates paper-ready quantitative results, verifies Saga reliability, and
provides a short live profile suitable for a ten-minute graduation presentation.

## Comparison Modes

The same Go middleware binary exposes three controlled experiment modes through
`EXPERIMENT_MODE`.

| Mode | Read path | Commit arbitration | Saga |
| --- | --- | --- | --- |
| `baseline` | each request counts as one logical DB read | arrival-order winner | disabled in scalability workload |
| `qcfuse` | requests with equal resource and intent are fused | arrival-order winner | disabled in scalability workload |
| `full` | QCFuse-style read fusion | ATCC sunk-cost ranking | reliability workload enabled |

The modes intentionally share the same gRPC transport, scheduling windows, and
agent workload. This isolates the effect of read fusion and cost-aware
arbitration without claiming a comparison against a production DBMS.

## Profiles

### Live Profile

- Modes: `baseline`, `qcfuse`, `full`
- Agent counts: 10 and 50
- Repetitions: 1
- Short scheduling windows
- Target duration: under two minutes
- Produces an HTML summary that can be opened during the presentation

### Paper Profile

- Modes: `baseline`, `qcfuse`, `full`
- Agent counts: 10, 50, 100, and 200
- Repetitions: 5
- Fixed random seeds per repetition
- Runs Saga reliability checks after scalability runs

## Metrics And Formulas

Each raw run records:

- elapsed time
- throughput: `completed requests / elapsed seconds`
- mean, p50, p95, and p99 end-to-end latency
- read requests and logical DB reads
- saved DB reads
- I/O reduction: `(read requests - logical DB reads) / read requests * 100`
- approved commits, rollbacks, and errors
- winner cost
- maximum candidate cost
- winner protection ratio: `winner cost / maximum candidate cost * 100`
- protected loser cost reported by the middleware

The summary groups runs by mode and agent count, then reports mean and sample
standard deviation. Raw observations remain available so reported values are
auditable.

## Reliability Evaluation

The reliability suite uses the persistent full mode and measures:

1. compensation success rate;
2. duplicate compensation idempotency;
3. middleware restart state recovery;
4. event timeline recovery;
5. resource stock recovery;
6. unsupported compensation failure detection.

Reliability results are emitted separately from scalability results because they
measure correctness rather than throughput.

## Components

### Go Middleware

- Add `ExperimentMode` to configuration and `/metrics`.
- Baseline read mode records every request as a logical DB read.
- Baseline and QCFuse-only commit modes select the first queued request.
- Full mode retains ATCC cost ranking.
- Add metrics for logical DB reads, candidate maximum cost, and winner cost.
- Preserve existing default behavior as `full`.

### Python Experiment Runner

`agent-python/experiment_runner.py` owns the experiment lifecycle:

1. start a Go server subprocess with isolated ports and SQLite DB;
2. wait for readiness;
3. execute deterministic concurrent agent workloads;
4. collect middleware metrics and client latency observations;
5. stop the server;
6. aggregate CSV and JSON outputs;
7. run the reliability and restart-recovery suite;
8. render a standalone HTML report.

The runner fails loudly when a server cannot start, a request errors, or a
reliability assertion fails.

### Output Artifacts

Outputs are written under `outputs/experiments/<timestamp>/`:

- `raw_runs.csv`
- `summary.csv`
- `summary.json`
- `reliability.json`
- `report.html`

The generated output directory remains ignored by Git. Documentation records
only verified headline results.

## Presentation Integration

The live presentation uses:

```sh
python3 agent-python/experiment_runner.py --profile live
```

During the demo, the presenter shows the terminal progress and then opens the
generated `report.html`. The live experiment demonstrates that the same workload
produces fewer logical DB reads under QCFuse and selects the maximum-cost winner
only in full mode. The precomputed paper profile supplies the detailed graphs
and repeated measurements shown after the live demo.

## Error Handling

- Each server receives isolated gRPC, HTTP, and SQLite paths.
- Server startup has a readiness timeout and includes captured logs on failure.
- Every run resets metrics and resource state.
- Workload errors are counted and make reliability checks fail.
- Server processes are terminated in `finally` blocks.

## Testing

- Go unit tests verify mode parsing, read accounting, and arbitration policy.
- Python unit tests verify percentiles, aggregation, reliability scoring, and
  HTML rendering.
- End-to-end verification runs the live profile.
- Existing Go tests, race detector, Python syntax checks, and diff checks remain
  required before completion.

## Honest Scope

The experiment measures a controlled shared-ticket simulation. Logical DB reads
are middleware counters, not physical storage-engine I/O. The final paper and
presentation must use the terms `QCFuse-style`, `ATCC-style`, and
`SagaLLM-compatible` rather than claiming complete reproductions of the papers.
