# 01 · HTTP instrumentation and OTLP export

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

Two HTTP applications share one [observability package](observability/otel.go)
and intentionally duplicate their small business layer for side-by-side reading.
The package configures a batch TracerProvider, resource, OTLP/gRPC exporter,
and W3C TraceContext/Baggage propagators.

| Task | [app-manual](app-manual/README.md) | [app-otelhttp](app-otelhttp/README.md) |
| --- | --- | --- |
| Extract inbound context | `propagator.Extract` | `otelhttp.NewHandler` |
| Create/end server span | Custom tracing middleware | `otelhttp` |
| HTTP attributes and 5xx status | Explicit attributes/status | Library instrumentation |
| Route name | Custom formatter | `WithSpanNameFormatter` and `r.Pattern` |
| Log correlation | Read active span from context | Same approach |

A pinned [Collector + Jaeger stack](stack/README.md) is now included. Follow that
guide to start the backends, inspect traces, and run the opt-in integration tests.
Its live interoperability check is still pending; the Compose and test code have
been validated without starting the services. You can also use an existing receiver. Port **4317** is gRPC; port 4318
typically receives OTLP/HTTP and is not interchangeable with this exporter.

## Run with a receiver on localhost

From this directory, in Bash:

```sh
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
go run ./app-manual/cmd/server
```

In PowerShell:

```powershell
$env:OTEL_EXPORTER_OTLP_ENDPOINT = 'http://localhost:4317'
go run ./app-manual/cmd/server
```

In another terminal (`curl.exe` in Windows PowerShell):

```sh
curl -i 'http://localhost:8080/work?delay_ms=200'
curl -i 'http://localhost:8080/error?status=500'
```

Stop the first application, then run `go run ./app-otelhttp/cmd/server`.
Compare `service.work` child spans and the request logs' `trace_id`/`span_id`.
Use a real receiver/backend to inspect exported spans; an HTTP 200 alone does
not establish successful export. Export errors can appear asynchronously.

| Environment variable | Default |
| --- | --- |
| `APP_NAME` | `local-observability-app` |
| `HTTP_PORT` | `8080` |
| `LOG_LEVEL` | `info` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4317` |

`http://` selects insecure gRPC; `https://` enables TLS. Host-only endpoints
are treated as insecure. These choices are for a controlled local exercise.
See [troubleshooting](../docs/troubleshooting.md) if spans do not arrive.
