# app-otelhttp (Example 01 — auto-instrumentation)

То же приложение, что [app-manual/](../app-manual/), но HTTP-инструментация
делегирована библиотеке `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`.

## Endpoints

- `GET /health` - health check
- `GET /work?delay_ms=120` - simulated work with delay
- `GET /error?status=500` - forced error response

## Run locally

```bash
make run
```

Env vars: `APP_NAME`, `HTTP_PORT`, `LOG_LEVEL`, `OTEL_EXPORTER_OTLP_ENDPOINT`.

## В чём разница с app-manual

Сравнение в [01-local-observability-stack/README.md](../README.md). Коротко:
ручной middleware из app-manual здесь заменён на `otelhttp.NewHandler` +
`otelhttp.WithRouteTag` на каждом маршруте. Общий пакет `observability/`
(OTel SDK setup) у обоих app один.
