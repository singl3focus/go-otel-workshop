# App Shell (Example 01)

Minimal Go HTTP app used in local observability stack.

## Endpoints

- `GET /health` - health check
- `GET /work?delay_ms=120` - simulated work with delay
- `GET /error?status=500` - forced error response

## Run locally

```bash
make run
```

Env vars are listed in `.env.example`.
