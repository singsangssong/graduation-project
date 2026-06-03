import argparse
import csv
import json
import time
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

import grpc
import middleware_pb2
import middleware_pb2_grpc


DEFAULT_AGENTS = [
    {"agent_id": "Agent-A", "tokens": 4500, "latency": 8.5},
    {"agent_id": "Agent-B", "tokens": 200, "latency": 1.2},
    {"agent_id": "Agent-C", "tokens": 1800, "latency": 3.0},
    {"agent_id": "Agent-D", "tokens": 3200, "latency": 4.7},
    {"agent_id": "Agent-E", "tokens": 900, "latency": 2.1},
]


def fetch_json(url: str):
    with urllib.request.urlopen(url, timeout=3) as response:
        return json.loads(response.read().decode("utf-8"))


def reset_metrics(metrics_url: str):
    request = urllib.request.Request(
        f"{metrics_url}/reset",
        data=b"",
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=3) as response:
        return json.loads(response.read().decode("utf-8"))


def run_agent(agent, grpc_addr: str, resource_id: str, reasoning_sleep_sec: float):
    start = time.time()
    channel = grpc.insecure_channel(grpc_addr)
    stub = middleware_pb2_grpc.TransactionMiddlewareStub(channel)

    read_request = middleware_pb2.ReadRequest(
        agent_id=agent["agent_id"],
        resource_id=resource_id,
        intent="purchase_ticket_demo",
    )
    read_response = stub.ReadResource(read_request)

    # Actual sleep is kept short and deterministic for live demos. The reported
    # latency is still sent to ATCC as the agent-side sunk cost signal.
    time.sleep(reasoning_sleep_sec)

    commit_request = middleware_pb2.CommitRequest(
        agent_id=agent["agent_id"],
        resource_id=resource_id,
        action_value=1,
        accumulated_tokens=agent["tokens"],
        inference_latency_sec=agent["latency"],
    )
    commit_response = stub.CommitTransaction(commit_request)
    channel.close()

    return {
        "agent_id": agent["agent_id"],
        "tokens": agent["tokens"],
        "latency": agent["latency"],
        "read_message": read_response.message,
        "success": commit_response.success,
        "rolled_back": commit_response.is_rolled_back,
        "commit_message": commit_response.message,
        "saved_cost_usd": round(commit_response.saved_cost_usd, 4),
        "elapsed_sec": round(time.time() - start, 3),
    }


def write_outputs(results, metrics, output_dir: Path):
    output_dir.mkdir(parents=True, exist_ok=True)
    payload = {
        "results": results,
        "metrics": metrics,
    }

    json_path = output_dir / "deterministic_demo_result.json"
    csv_path = output_dir / "deterministic_demo_result.csv"

    json_path.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")

    with csv_path.open("w", newline="", encoding="utf-8") as csv_file:
        writer = csv.DictWriter(csv_file, fieldnames=list(results[0].keys()))
        writer.writeheader()
        writer.writerows(results)

    return json_path, csv_path


def main():
    parser = argparse.ArgumentParser(description="Deterministic ATCC/QCFuse demo runner")
    parser.add_argument("--grpc-addr", default="localhost:50051")
    parser.add_argument("--metrics-url", default="http://localhost:8080")
    parser.add_argument("--resource-id", default="flight_ticket_A")
    parser.add_argument("--reasoning-sleep-sec", type=float, default=0.2)
    parser.add_argument("--output-dir", default="outputs/demo-results")
    parser.add_argument("--skip-reset", action="store_true")
    args = parser.parse_args()

    if not args.skip_reset:
        try:
            reset_metrics(args.metrics_url)
        except (urllib.error.URLError, TimeoutError) as error:
            raise SystemExit(f"metrics reset failed: {error}") from error

    with ThreadPoolExecutor(max_workers=len(DEFAULT_AGENTS)) as executor:
        futures = [
            executor.submit(
                run_agent,
                agent,
                args.grpc_addr,
                args.resource_id,
                args.reasoning_sleep_sec,
            )
            for agent in DEFAULT_AGENTS
        ]
        results = [future.result() for future in as_completed(futures)]

    results.sort(key=lambda item: item["agent_id"])
    metrics = fetch_json(f"{args.metrics_url}/metrics")
    json_path, csv_path = write_outputs(results, metrics, Path(args.output_dir))

    winner = next((item for item in results if item["success"]), None)
    rollbacks = sum(1 for item in results if item["rolled_back"])
    saved_db_reads = metrics["metrics"]["saved_db_reads"]
    total_saved_cost = metrics["metrics"]["total_saved_cost_usd"]

    print("\n=== Deterministic Agentic Middleware Demo ===")
    print(f"Winner: {winner['agent_id'] if winner else 'none'}")
    print(f"Rollbacks: {rollbacks}")
    print(f"Saved DB reads by QCFuse: {saved_db_reads}")
    print(f"Saved cost by ATCC rollback: ${total_saved_cost:.2f}")
    print(f"JSON: {json_path}")
    print(f"CSV : {csv_path}")


if __name__ == "__main__":
    main()
