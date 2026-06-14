#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
OUTPUT="$ROOT/outputs/experiments/latest-live"

python3 "$ROOT/agent-python/experiment_runner.py" \
  --profile live \
  --output-dir "$OUTPUT"

printf '\nPresentation report:\n%s\n' "$OUTPUT/report.html"
