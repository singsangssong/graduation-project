# Go Middleware Instructions

This folder contains the gRPC middleware server.

Rules:

- Keep QCFuse read batching and ATCC commit arbitration easy to explain.
- Add focused tests when changing cost calculation or scheduler behavior.
- Use project-safe Go cache commands when running in sandboxed environments.
- Do not manually edit files under `middleware-go/pb/`; regenerate them from proto.

Useful command:

```sh
GOCACHE=/private/tmp/agenic-middleware-gocache go test ./...
```
