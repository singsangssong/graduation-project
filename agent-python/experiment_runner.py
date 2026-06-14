import argparse
import csv
import json
import os
import random
import socket
import subprocess
import sys
import time
import urllib.parse
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime
from pathlib import Path

import grpc
import middleware_pb2
import middleware_pb2_grpc

from experiment_stats import percentile, reliability_summary, render_html, summarize_runs
from saga_demo import AGENTS as SAGA_AGENTS
from saga_demo import run_saga


ROOT = Path(__file__).resolve().parents[1]
GO_DIR = ROOT / "middleware-go"
PROFILES = {
    "live": {"agent_counts": [10, 50], "repetitions": 1, "qcfuse_ms": 40, "atcc_ms": 80},
    "paper": {"agent_counts": [10, 50, 100, 200], "repetitions": 5, "qcfuse_ms": 80, "atcc_ms": 150},
}
MODES = ("baseline", "qcfuse", "full")


def fetch_json(url, timeout=5):
    with urllib.request.urlopen(url, timeout=timeout) as response:
        return json.loads(response.read().decode("utf-8"))


def post_json(url):
    request = urllib.request.Request(url, data=b"", method="POST")
    return fetch_json(request)


def free_port():
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


class MiddlewareProcess:
    def __init__(self, mode, output_dir, stock=1, db_path=None, ports=None, profile=None):
        self.mode = mode
        self.output_dir = output_dir
        self.stock = stock
        self.db_path = db_path or output_dir / f"{mode}.db"
        self.grpc_port, self.http_port = ports or (free_port(), free_port())
        self.grpc_addr = f"127.0.0.1:{self.grpc_port}"
        self.metrics_url = f"http://127.0.0.1:{self.http_port}"
        self.profile = profile or PROFILES["live"]
        self.process = None
        self.log_file = None

    def start(self):
        env = os.environ.copy()
        env.update(
            {
                "EXPERIMENT_MODE": self.mode,
                "MIDDLEWARE_GRPC_ADDR": self.grpc_addr,
                "MIDDLEWARE_METRICS_ADDR": f"127.0.0.1:{self.http_port}",
                "QCFUSE_WINDOW_MS": str(self.profile["qcfuse_ms"]),
                "ATCC_WINDOW_MS": str(self.profile["atcc_ms"]),
                "TICKET_STOCK": str(self.stock),
                "SAGA_DB_PATH": str(self.db_path),
                "GOCACHE": "/private/tmp/agenic-middleware-gocache",
            }
        )
        self.output_dir.mkdir(parents=True, exist_ok=True)
        self.log_file = (self.output_dir / f"server-{self.mode}-{self.grpc_port}.log").open(
            "w", encoding="utf-8"
        )
        self.process = subprocess.Popen(
            ["go", "run", "."],
            cwd=GO_DIR,
            env=env,
            stdout=self.log_file,
            stderr=subprocess.STDOUT,
            text=True,
        )
        deadline = time.time() + 20
        while time.time() < deadline:
            if self.process.poll() is not None:
                self.stop()
                raise RuntimeError(f"{self.mode} server exited during startup; inspect {self.log_file.name}")
            try:
                fetch_json(f"{self.metrics_url}/metrics", timeout=1)
                return self
            except Exception:
                time.sleep(0.15)
        self.stop()
        raise RuntimeError(f"{self.mode} server readiness timeout; inspect {self.log_file.name}")

    def stop(self):
        if self.process and self.process.poll() is None:
            self.process.terminate()
            try:
                self.process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.process.kill()
                self.process.wait(timeout=5)
        if self.log_file:
            self.log_file.close()


def make_agents(agent_count, seed):
    rng = random.Random(seed)
    agents = []
    for index in range(agent_count):
        agents.append(
            {
                "agent_id": f"Agent-{index:03d}",
                "tokens": rng.randint(200, 5000),
                "latency": round(rng.uniform(0.5, 10.0), 3),
                "commit_delay": index * 0.0005,
            }
        )
    agents[-1]["tokens"] = 6000
    agents[-1]["latency"] = 12.0
    return agents


def sunk_cost(agent):
    return agent["tokens"] * 0.002 + agent["latency"] * 0.5


def run_scalability_agent(agent, grpc_addr):
    started = time.perf_counter()
    channel = grpc.insecure_channel(grpc_addr)
    stub = middleware_pb2_grpc.TransactionMiddlewareStub(channel)
    try:
        stub.ReadResource(
            middleware_pb2.ReadRequest(
                agent_id=agent["agent_id"],
                resource_id="flight_ticket_A",
                intent="quantitative_experiment",
            ),
            timeout=5,
        )
        time.sleep(agent["commit_delay"])
        response = stub.CommitTransaction(
            middleware_pb2.CommitRequest(
                agent_id=agent["agent_id"],
                resource_id="flight_ticket_A",
                action_value=1,
                accumulated_tokens=agent["tokens"],
                inference_latency_sec=agent["latency"],
            ),
            timeout=5,
        )
        return {
            "agent_id": agent["agent_id"],
            "cost": sunk_cost(agent),
            "success": response.success,
            "rolled_back": response.is_rolled_back,
            "saved_cost": response.saved_cost_usd,
            "latency_ms": (time.perf_counter() - started) * 1000,
            "error": "",
        }
    except grpc.RpcError as error:
        return {
            "agent_id": agent["agent_id"],
            "cost": sunk_cost(agent),
            "success": False,
            "rolled_back": False,
            "saved_cost": 0.0,
            "latency_ms": (time.perf_counter() - started) * 1000,
            "error": error.details() or str(error),
        }
    finally:
        channel.close()


def execute_run(server, mode, agent_count, repetition):
    post_json(f"{server.metrics_url}/reset")
    agents = make_agents(agent_count, seed=agent_count * 1000 + repetition)
    started = time.perf_counter()
    with ThreadPoolExecutor(max_workers=agent_count) as executor:
        futures = [executor.submit(run_scalability_agent, agent, server.grpc_addr) for agent in agents]
        results = [future.result() for future in as_completed(futures)]
    elapsed = time.perf_counter() - started
    metrics = fetch_json(f"{server.metrics_url}/metrics")["metrics"]
    latencies = [result["latency_ms"] for result in results]
    errors = sum(1 for result in results if result["error"])
    winner = next((result for result in results if result["success"]), None)
    maximum_cost = max(result["cost"] for result in results)
    logical_reads = metrics["logical_db_reads"]
    return {
        "mode": mode,
        "agent_count": agent_count,
        "repetition": repetition,
        "elapsed_sec": round(elapsed, 6),
        "throughput_tps": round((agent_count - errors) / elapsed, 4),
        "mean_latency_ms": round(sum(latencies) / len(latencies), 4),
        "p50_latency_ms": percentile(latencies, 50),
        "p95_latency_ms": percentile(latencies, 95),
        "p99_latency_ms": percentile(latencies, 99),
        "read_requests": metrics["read_requests"],
        "logical_db_reads": logical_reads,
        "saved_db_reads": metrics["saved_db_reads"],
        "io_reduction_pct": round((agent_count - logical_reads) / agent_count * 100, 4),
        "approved_commits": metrics["approved_commits"],
        "rollbacks": metrics["rolled_back_commits"],
        "errors": errors,
        "error_rate_pct": round(errors / agent_count * 100, 4),
        "winner_agent_id": winner["agent_id"] if winner else "",
        "winner_cost": round(winner["cost"], 4) if winner else 0.0,
        "maximum_candidate_cost": round(maximum_cost, 4),
        "winner_protection_pct": round(winner["cost"] / maximum_cost * 100, 4) if winner else 0.0,
        "protected_cost": round(metrics["total_saved_cost_usd"], 4),
    }


def saga_state(stub, saga_id):
    return stub.GetSagaState(middleware_pb2.GetSagaStateRequest(saga_id=saga_id), timeout=5)


def run_reliability(output_dir, profile):
    checks = []
    db_path = output_dir / "reliability.db"
    ports = (free_port(), free_port())
    server = MiddlewareProcess("full", output_dir, stock=3, db_path=db_path, ports=ports, profile=profile).start()
    try:
        with ThreadPoolExecutor(max_workers=len(SAGA_AGENTS)) as executor:
            futures = [executor.submit(run_saga, agent, server.grpc_addr, "flight_ticket_A") for agent in SAGA_AGENTS]
            saga_results = [future.result() for future in as_completed(futures)]
        committed = [item for item in saga_results if item["saga_status"] == "COMMITTED"]
        compensated = [item for item in saga_results if item["saga_status"] == "COMPENSATED"]
        stock_after = fetch_json(f"{server.metrics_url}/resource?resource_id=flight_ticket_A")["available_stock"]
        checks.append({"name": "one Saga committed", "passed": len(committed) == 1, "observed": len(committed)})
        checks.append({"name": "losing Sagas compensated", "passed": len(compensated) == 2, "observed": len(compensated)})
        checks.append({"name": "compensation restored stock exactly once", "passed": stock_after == 2, "observed": stock_after})

        loser = compensated[0]
        channel = grpc.insecure_channel(server.grpc_addr)
        stub = middleware_pb2_grpc.TransactionMiddlewareStub(channel)
        stub.AbortSaga(middleware_pb2.AbortSagaRequest(saga_id=loser["saga_id"], reason="duplicate abort"), timeout=5)
        stock_duplicate = fetch_json(f"{server.metrics_url}/resource?resource_id=flight_ticket_A")["available_stock"]
        checks.append({"name": "duplicate compensation is idempotent", "passed": stock_duplicate == 2, "observed": stock_duplicate})

        unsupported = stub.BeginSaga(middleware_pb2.BeginSagaRequest(agent_id="Unsupported-Agent", goal="external call"), timeout=5)
        stub.RegisterSagaStep(
            middleware_pb2.RegisterSagaStepRequest(
                saga_id=unsupported.saga.saga_id,
                step_id="external",
                action="call external-api",
                result="called",
                compensation_action="undo external-api",
            ),
            timeout=5,
        )
        failed = stub.AbortSaga(middleware_pb2.AbortSagaRequest(saga_id=unsupported.saga.saga_id, reason="failure test"), timeout=5)
        checks.append({"name": "unsupported compensation is detected", "passed": failed.saga.status == "COMPENSATION_FAILED", "observed": failed.saga.status})
        channel.close()
        expected = {item["saga_id"]: item["saga_status"] for item in saga_results}
    finally:
        server.stop()

    recovered_server = MiddlewareProcess("full", output_dir, stock=3, db_path=db_path, ports=ports, profile=profile).start()
    try:
        channel = grpc.insecure_channel(recovered_server.grpc_addr)
        stub = middleware_pb2_grpc.TransactionMiddlewareStub(channel)
        recovered = {saga_id: saga_state(stub, saga_id).status for saga_id in expected}
        events_ok = all(
            len(fetch_json(f"{recovered_server.metrics_url}/events?{urllib.parse.urlencode({'saga_id': saga_id})}")) > 0
            for saga_id in expected
        )
        recovered_stock = fetch_json(f"{recovered_server.metrics_url}/resource?resource_id=flight_ticket_A")["available_stock"]
        checks.append({"name": "Saga states recover after restart", "passed": recovered == expected, "observed": recovered})
        checks.append({"name": "event timelines recover after restart", "passed": events_ok, "observed": events_ok})
        checks.append({"name": "resource stock recovers after restart", "passed": recovered_stock == 2, "observed": recovered_stock})
        channel.close()
    finally:
        recovered_server.stop()
    return reliability_summary(checks)


def write_csv(path, rows):
    if not rows:
        return
    with path.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(handle, fieldnames=list(rows[0].keys()))
        writer.writeheader()
        writer.writerows(rows)


def main():
    parser = argparse.ArgumentParser(description="Automated quantitative experiment runner")
    parser.add_argument("--profile", choices=PROFILES, default="live")
    parser.add_argument("--output-dir")
    parser.add_argument("--skip-reliability", action="store_true")
    args = parser.parse_args()

    profile = PROFILES[args.profile]
    timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
    output_dir = Path(args.output_dir or ROOT / "outputs" / "experiments" / timestamp).resolve()
    output_dir.mkdir(parents=True, exist_ok=True)
    raw_runs = []

    print(f"Experiment profile={args.profile} output={output_dir}")
    for mode in MODES:
        server = MiddlewareProcess(mode, output_dir, profile=profile).start()
        try:
            for agent_count in profile["agent_counts"]:
                for repetition in range(1, profile["repetitions"] + 1):
                    run = execute_run(server, mode, agent_count, repetition)
                    raw_runs.append(run)
                    print(
                        f"[{mode:8}] agents={agent_count:3} run={repetition} "
                        f"TPS={run['throughput_tps']:.1f} IO-reduction={run['io_reduction_pct']:.1f}% "
                        f"winner-protection={run['winner_protection_pct']:.1f}%"
                    )
        finally:
            server.stop()

    reliability = {"total_checks": 0, "passed_checks": 0, "pass_rate_pct": 0.0, "checks": []}
    if not args.skip_reliability:
        print("Running Saga reliability and restart recovery checks...")
        reliability = run_reliability(output_dir, profile)
        print(f"Reliability: {reliability['passed_checks']}/{reliability['total_checks']} passed")

    summary = summarize_runs(raw_runs)
    write_csv(output_dir / "raw_runs.csv", raw_runs)
    write_csv(output_dir / "summary.csv", summary)
    (output_dir / "summary.json").write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding="utf-8")
    (output_dir / "reliability.json").write_text(json.dumps(reliability, ensure_ascii=False, indent=2), encoding="utf-8")
    (output_dir / "report.html").write_text(render_html(summary, reliability, args.profile), encoding="utf-8")
    print(f"Report: {output_dir / 'report.html'}")
    if reliability["checks"] and reliability["passed_checks"] != reliability["total_checks"]:
        raise SystemExit("reliability checks failed")


if __name__ == "__main__":
    main()
