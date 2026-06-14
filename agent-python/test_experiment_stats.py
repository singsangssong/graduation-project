import unittest

from experiment_stats import percentile, reliability_summary, render_html, summarize_runs


class ExperimentStatsTest(unittest.TestCase):
    def test_percentile_interpolates_sorted_values(self):
        self.assertEqual(percentile([1, 2, 3, 4], 50), 2.5)
        self.assertEqual(percentile([1, 2, 3, 4], 95), 3.85)

    def test_summarize_runs_groups_mode_and_agent_count(self):
        runs = [
            {
                "mode": "full",
                "agent_count": 10,
                "throughput_tps": 100.0,
                "p95_latency_ms": 20.0,
                "io_reduction_pct": 90.0,
                "winner_protection_pct": 100.0,
                "error_rate_pct": 0.0,
            },
            {
                "mode": "full",
                "agent_count": 10,
                "throughput_tps": 120.0,
                "p95_latency_ms": 24.0,
                "io_reduction_pct": 80.0,
                "winner_protection_pct": 100.0,
                "error_rate_pct": 0.0,
            },
        ]

        summary = summarize_runs(runs)

        self.assertEqual(len(summary), 1)
        self.assertEqual(summary[0]["throughput_tps_mean"], 110.0)
        self.assertAlmostEqual(summary[0]["throughput_tps_stdev"], 14.1421, places=4)
        self.assertEqual(summary[0]["io_reduction_pct_mean"], 85.0)

    def test_reliability_summary_reports_pass_rate(self):
        checks = [
            {"name": "recovery", "passed": True},
            {"name": "compensation", "passed": True},
            {"name": "failure_detection", "passed": False},
        ]

        result = reliability_summary(checks)

        self.assertEqual(result["total_checks"], 3)
        self.assertEqual(result["passed_checks"], 2)
        self.assertAlmostEqual(result["pass_rate_pct"], 66.6667, places=4)

    def test_render_html_contains_modes_and_reliability(self):
        html = render_html(
            [{"mode": "full", "agent_count": 10, "throughput_tps_mean": 100.0}],
            {"pass_rate_pct": 100.0, "checks": [{"name": "restart recovery", "passed": True}]},
            "live",
        )

        self.assertIn("Quantitative Experiment Report", html)
        self.assertIn("full", html)
        self.assertIn("restart recovery", html)


if __name__ == "__main__":
    unittest.main()
