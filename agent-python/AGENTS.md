# Python Agent Instructions

This folder contains Python clients and stress-test scripts.

Rules:

- Use generated files derived from `proto/middleware.proto`.
- Keep stress-test output concise and useful for live demos.
- Prefer deterministic or seedable test inputs when comparing behavior.
- Do not rely on `cata.proto` for the final middleware demo path.

Useful command:

```sh
python3 -m py_compile agent_client.py stress_test.py stress_test_v2.py
```
