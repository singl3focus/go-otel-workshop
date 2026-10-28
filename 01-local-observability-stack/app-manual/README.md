# app-manual · manual instrumentation

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

The [tracing middleware](internal/server/middleware/request_tracing.go) explicitly extracts the parent, starts a server span, records HTTP attributes, and ends it.

Compare with [app-otelhttp](../app-otelhttp/README.md). Both share the
[SDK setup](../observability/otel.go). Configuration and receiver requirements
are described in the [module guide](../README.md).

## Run from this directory

```sh
go run ./cmd/server
```

`make run` is an optional equivalent if Make is installed.

| Request | Behavior |
| --- | --- |
| `GET /health` | JSON health response |
| `GET /work?delay_ms=120` | Simulated work and `service.work` span |
| `GET /error?status=500` | Forced response; 5xx marks the operation as Error |

The server listens on `:8080` unless `HTTP_PORT` overrides it.
Set `APP_NAME`, `LOG_LEVEL`, and `OTEL_EXPORTER_OTLP_ENDPOINT` as needed.
The default receiver is `http://localhost:4317`; use the [local stack](../stack/README.md).
