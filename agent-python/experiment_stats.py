import html
import math
import statistics
from collections import defaultdict


SUMMARY_METRICS = (
    "throughput_tps",
    "mean_latency_ms",
    "p95_latency_ms",
    "p99_latency_ms",
    "io_reduction_pct",
    "winner_protection_pct",
    "protected_cost",
    "error_rate_pct",
)


def rounded(value):
    return round(float(value), 4)


def percentile(values, percentile_value):
    if not values:
        return 0.0
    ordered = sorted(values)
    position = (len(ordered) - 1) * percentile_value / 100
    lower = math.floor(position)
    upper = math.ceil(position)
    if lower == upper:
        return rounded(ordered[lower])
    weight = position - lower
    return rounded(ordered[lower] * (1 - weight) + ordered[upper] * weight)


def summarize_runs(runs):
    groups = defaultdict(list)
    for run in runs:
        groups[(run["mode"], int(run["agent_count"]))].append(run)

    summaries = []
    for (mode, agent_count), grouped_runs in sorted(groups.items()):
        item = {
            "mode": mode,
            "agent_count": agent_count,
            "repetitions": len(grouped_runs),
        }
        for metric in SUMMARY_METRICS:
            values = [float(run.get(metric, 0.0)) for run in grouped_runs]
            item[f"{metric}_mean"] = rounded(statistics.mean(values))
            item[f"{metric}_stdev"] = rounded(statistics.stdev(values)) if len(values) > 1 else 0.0
        summaries.append(item)
    return summaries


def reliability_summary(checks):
    passed = sum(1 for check in checks if check["passed"])
    total = len(checks)
    return {
        "total_checks": total,
        "passed_checks": passed,
        "pass_rate_pct": rounded(passed / total * 100) if total else 0.0,
        "checks": checks,
    }


def render_html(summary, reliability, profile):
    rows = []
    for item in summary:
        rows.append(
            "<tr>"
            f"<td><strong>{html.escape(str(item['mode']))}</strong></td>"
            f"<td>{item['agent_count']}</td>"
            f"<td>{item.get('throughput_tps_mean', 0):.2f}</td>"
            f"<td>{item.get('p95_latency_ms_mean', 0):.2f}</td>"
            f"<td>{item.get('io_reduction_pct_mean', 0):.2f}%</td>"
            f"<td>{item.get('winner_protection_pct_mean', 0):.2f}%</td>"
            f"<td>{item.get('error_rate_pct_mean', 0):.2f}%</td>"
            "</tr>"
        )
    checks = []
    for check in reliability.get("checks", []):
        state = "PASS" if check["passed"] else "FAIL"
        checks.append(
            f"<li class=\"{state.lower()}\"><span>{state}</span> {html.escape(check['name'])}</li>"
        )
    return f"""<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Quantitative Experiment Report</title>
<style>
body {{ margin: 0; font: 16px/1.5 -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; color: #172033; background: #f4f6f9; }}
header {{ padding: 42px 6vw 30px; color: white; background: #153e75; }}
main {{ max-width: 1180px; margin: 28px auto; padding: 0 24px 48px; }}
h1, h2 {{ letter-spacing: 0; }}
.metrics {{ display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; margin: 20px 0; }}
.metric {{ padding: 18px; border-left: 4px solid #16a085; background: white; }}
.metric strong {{ display: block; font-size: 28px; }}
section {{ margin: 22px 0; padding: 24px; background: white; border: 1px solid #dbe2ea; }}
table {{ width: 100%; border-collapse: collapse; }}
th, td {{ padding: 12px; text-align: right; border-bottom: 1px solid #e6ebf0; }}
th:first-child, td:first-child {{ text-align: left; }}
ul {{ padding: 0; list-style: none; }}
li {{ margin: 8px 0; }}
li span {{ display: inline-block; width: 48px; font-weight: 700; }}
.pass span {{ color: #087f5b; }} .fail span {{ color: #c92a2a; }}
@media (max-width: 760px) {{ .metrics {{ grid-template-columns: 1fr; }} section {{ overflow-x: auto; }} }}
</style>
</head>
<body>
<header><h1>Quantitative Experiment Report</h1><p>Profile: {html.escape(profile)}</p></header>
<main>
<div class="metrics">
<div class="metric"><span>Comparison modes</span><strong>{len(set(item['mode'] for item in summary))}</strong></div>
<div class="metric"><span>Measured configurations</span><strong>{len(summary)}</strong></div>
<div class="metric"><span>Reliability pass rate</span><strong>{reliability.get('pass_rate_pct', 0):.1f}%</strong></div>
</div>
<section><h2>Scalability Comparison</h2><table>
<thead><tr><th>Mode</th><th>Agents</th><th>TPS</th><th>p95 ms</th><th>I/O reduction</th><th>Winner protection</th><th>Error rate</th></tr></thead>
<tbody>{''.join(rows)}</tbody></table></section>
<section><h2>Saga Reliability</h2><ul>{''.join(checks)}</ul></section>
<section><h2>Interpretation Boundary</h2><p>Results measure a controlled shared-ticket simulation. Logical DB reads are middleware counters, not physical storage-engine I/O.</p></section>
</main></body></html>"""
