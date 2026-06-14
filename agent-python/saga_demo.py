import argparse
import json
import time
import urllib.request
import urllib.parse
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

import grpc
import middleware_pb2
import middleware_pb2_grpc


AGENTS = [
    {"agent_id": "Saga-Agent-A", "tokens": 4500, "latency": 8.5},
    {"agent_id": "Saga-Agent-B", "tokens": 1200, "latency": 2.0},
    {"agent_id": "Saga-Agent-C", "tokens": 2800, "latency": 4.0},
]


def reset_server(metrics_url: str):
    request = urllib.request.Request(f"{metrics_url}/reset", data=b"", method="POST")
    with urllib.request.urlopen(request, timeout=3) as response:
        return json.loads(response.read().decode("utf-8"))


def fetch_json(url: str):
    with urllib.request.urlopen(url, timeout=3) as response:
        return json.loads(response.read().decode("utf-8"))


def fetch_resource(metrics_url: str, resource_id: str):
    query = urllib.parse.urlencode({"resource_id": resource_id})
    return fetch_json(f"{metrics_url}/resource?{query}")


def fetch_events(metrics_url: str, saga_id: str):
    query = urllib.parse.urlencode({"saga_id": saga_id})
    return fetch_json(f"{metrics_url}/events?{query}")


def run_saga(agent, grpc_addr: str, resource_id: str):
    channel = grpc.insecure_channel(grpc_addr)
    stub = middleware_pb2_grpc.TransactionMiddlewareStub(channel)

    begin = stub.BeginSaga(
        middleware_pb2.BeginSagaRequest(
            agent_id=agent["agent_id"],
            goal=f"Purchase {resource_id}",
            context={"framework": "SagaLLM-compatible-demo"},
        )
    )
    if not begin.success:
        raise RuntimeError(begin.message)
    saga_id = begin.saga.saga_id

    checkpoint = stub.RegisterSagaStep(
        middleware_pb2.RegisterSagaStepRequest(
            saga_id=saga_id,
            step_id="reserve-ticket",
            action=f"reserve {resource_id}",
            result="reservation prepared",
            compensation_action=f"release {resource_id}",
        )
    )
    if not checkpoint.success:
        raise RuntimeError(checkpoint.message)

    read = stub.ReadResource(
        middleware_pb2.ReadRequest(
            agent_id=agent["agent_id"],
            resource_id=resource_id,
            intent="purchase_ticket",
            saga_id=saga_id,
        )
    )

    validated = stub.ValidateSaga(middleware_pb2.ValidateSagaRequest(saga_id=saga_id))
    if not validated.success:
        raise RuntimeError(validated.message)

    commit = stub.CommitTransaction(
        middleware_pb2.CommitRequest(
            agent_id=agent["agent_id"],
            resource_id=resource_id,
            action_value=1,
            accumulated_tokens=agent["tokens"],
            inference_latency_sec=agent["latency"],
            saga_id=saga_id,
        )
    )
    final_state = stub.GetSagaState(middleware_pb2.GetSagaStateRequest(saga_id=saga_id))
    channel.close()

    return {
        "agent_id": agent["agent_id"],
        "saga_id": saga_id,
        "tokens": agent["tokens"],
        "latency": agent["latency"],
        "read_message": read.message,
        "commit_success": commit.success,
        "rolled_back": commit.is_rolled_back,
        "saga_status": final_state.status,
        "step_statuses": [step.status for step in final_state.steps],
        "compensation_actions": [
            step.compensation_action
            for step in final_state.steps
            if step.status == "COMPENSATED"
        ],
    }


def main():
    parser = argparse.ArgumentParser(description="SagaLLM-compatible middleware demo")
    parser.add_argument("--grpc-addr", default="localhost:50051")
    parser.add_argument("--metrics-url", default="http://localhost:8080")
    parser.add_argument("--resource-id", default="flight_ticket_A")
    parser.add_argument("--output", default="outputs/saga-demo-result.json")
    parser.add_argument("--skip-reset", action="store_true")
    args = parser.parse_args()

    if not args.skip_reset:
        reset_server(args.metrics_url)
    stock_before = fetch_resource(args.metrics_url, args.resource_id)
    started = time.time()
    with ThreadPoolExecutor(max_workers=len(AGENTS)) as executor:
        futures = [
            executor.submit(run_saga, agent, args.grpc_addr, args.resource_id)
            for agent in AGENTS
        ]
        results = [future.result() for future in as_completed(futures)]

    results.sort(key=lambda item: item["agent_id"])
    metrics = fetch_json(f"{args.metrics_url}/metrics")
    stock_after = fetch_resource(args.metrics_url, args.resource_id)
    event_timelines = {
        result["saga_id"]: fetch_events(args.metrics_url, result["saga_id"])
        for result in results
    }
    payload = {
        "elapsed_sec": round(time.time() - started, 3),
        "resource_before": stock_before,
        "resource_after": stock_after,
        "results": results,
        "event_timelines": event_timelines,
        "metrics": metrics,
    }

    output = Path(args.output)
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")

    print("\n=== SagaLLM-Compatible Transaction Demo ===")
    for result in results:
        print(
            f"{result['agent_id']}: commit={result['commit_success']} "
            f"saga={result['saga_status']} steps={result['step_statuses']}"
        )
    print(
        "Saga metrics: "
        f"started={metrics['metrics']['sagas_started']}, "
        f"validated={metrics['metrics']['sagas_validated']}, "
        f"compensated={metrics['metrics']['sagas_compensated']}, "
        f"actions={metrics['metrics']['compensation_actions']}"
    )
    print(
        f"Resource stock: {stock_before['available_stock']} "
        f"-> {stock_after['available_stock']}"
    )
    for result in results:
        timeline = [event["type"] for event in event_timelines[result["saga_id"]]]
        print(f"{result['agent_id']} events: {' -> '.join(timeline)}")
    print(f"JSON: {output}")


if __name__ == "__main__":
    main()
