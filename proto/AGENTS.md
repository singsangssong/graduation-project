# proto Instructions

`middleware.proto` is the canonical middleware API contract.

Rules:

- Preserve existing field numbers for backward compatibility.
- Do not reuse removed field numbers.
- Regenerate both Go and Python protobuf files after schema changes.
- Keep RPC names aligned with the Go server and Python clients.
- Prefer adding explicit fields over overloading existing fields with new meaning.
