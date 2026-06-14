#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
OUTPUT="$ROOT/outputs/experiments/latest-paper"

python3 "$ROOT/agent-python/experiment_runner.py" \
  --profile paper \
  --output-dir "$OUTPUT"

printf '\nPaper experiment report:\n%s\n' "$OUTPUT/report.html"
