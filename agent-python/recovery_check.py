import argparse
import json
import urllib.parse
import urllib.request
from pathlib import Path

import grpc
import middleware_pb2
import middleware_pb2_grpc


def fetch_json(url: str):
    with urllib.request.urlopen(url, timeout=3) as response:
        return json.loads(response.read().decode("utf-8"))


def main():
    parser = argparse.ArgumentParser(description="Verify persisted Saga recovery after middleware restart")
    parser.add_argument("--grpc-addr", default="localhost:50051")
    parser.add_argument("--metrics-url", default="http://localhost:8080")
    parser.add_argument("--input", default="outputs/saga-demo-result.json")
    parser.add_argument("--resource-id", default="flight_ticket_A")
    args = parser.parse_args()

    payload = json.loads(Path(args.input).read_text(encoding="utf-8"))
    channel = grpc.insecure_channel(args.grpc_addr)
    stub = middleware_pb2_grpc.TransactionMiddlewareStub(channel)

    recovered = []
    for previous in payload["results"]:
        state = stub.GetSagaState(
            middleware_pb2.GetSagaStateRequest(saga_id=previous["saga_id"])
        )
        query = urllib.parse.urlencode({"saga_id": previous["saga_id"]})
        events = fetch_json(f"{args.metrics_url}/events?{query}")
        recovered.append(
            {
                "agent_id": previous["agent_id"],
                "saga_id": state.saga_id,
                "expected_status": previous["saga_status"],
                "recovered_status": state.status,
                "step_statuses": [step.status for step in state.steps],
                "event_count": len(events),
            }
        )

    channel.close()
    failures = [
        item for item in recovered
        if item["expected_status"] != item["recovered_status"]
    ]
    resource_query = urllib.parse.urlencode({"resource_id": args.resource_id})
    resource = fetch_json(f"{args.metrics_url}/resource?{resource_query}")
    expected_stock = payload["resource_after"]["available_stock"]

    print("\n=== Persistent Saga Recovery Check ===")
    for item in recovered:
        print(
            f"{item['agent_id']}: {item['expected_status']} "
            f"-> {item['recovered_status']} events={item['event_count']}"
        )
    if failures:
        raise SystemExit(f"recovery mismatch: {failures}")
    if resource["available_stock"] != expected_stock:
        raise SystemExit(
            f"resource stock mismatch: {resource['available_stock']} != {expected_stock}"
        )
    print(f"Persistent resource stock recovered: {resource['available_stock']}")
    print("All Saga states and event timelines recovered after restart.")


if __name__ == "__main__":
    main()
