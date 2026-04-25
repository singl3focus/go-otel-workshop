# app-manual (Example 01 — manual instrumentation)

Минимальное HTTP-приложение, в котором вся HTTP-инструментация написана руками
в [internal/server/middleware/](internal/server/middleware/) — без `otelhttp`.

Парный пример с auto-инструментацией: [../app-otelhttp/](../app-otelhttp/).
Сравнение — в [../README.md](../README.md). OTel SDK setup общий: пакет
[../observability/](../observability/).

## Endpoints

- `GET /health` - health check
- `GET /work?delay_ms=120` - simulated work with delay
- `GET /error?status=500` - forced error response

## Run locally

```bash
make run
```

Env vars: `APP_NAME`, `HTTP_PORT`, `LOG_LEVEL`, `OTEL_EXPORTER_OTLP_ENDPOINT`.
