# 01. Local observability stack

Два HTTP-приложения с одинаковыми эндпоинтами и идентичной бизнес-логикой,
но с двумя подходами к HTTP-инструментации OpenTelemetry. Общий модуль,
общий пакет [observability/](observability/) (OTel SDK setup — TracerProvider,
OTLP/gRPC exporter, propagator).

## Структура

```
01-local-observability-stack/
  go.mod                       # один общий модуль
  observability/               # общий пакет: Setup() + ShutdownFunc
  app-manual/                  # ручной middleware: span на запрос пишем руками
    internal/server/middleware/
      request_tracing.go       # propagator.Extract → tracer.Start → SetAttributes → End
      request_logging.go       # slog с trace_id из ctx
  app-otelhttp/                # auto-инструментация
    internal/server/
      router.go                # otelhttp.NewHandler(...) + WithSpanNameFormatter
      middleware/
        request_logging.go     # тот же логгер, span читаем из ctx
```

Бизнес-логика (`internal/service/`, `internal/server/handlers/`,
`internal/config/`) дублируется намеренно — каждый app можно читать сверху
вниз как самостоятельный пример.

## Что показывает сравнение

| Этап | app-manual | app-otelhttp |
|---|---|---|
| Extract propagator из HTTP-заголовков | руками: `otel.GetTextMapPropagator().Extract(...)` | внутри `otelhttp.NewHandler` |
| Создание server-span | `tracer.Start(..., trace.WithSpanKind(SpanKindServer))` | `otelhttp` |
| Атрибуты `http.request.method` / `url.path` / `server.address` / `http.response.status_code` | руками через `semconv` v1.26 | `otelhttp` (semconv внутри пакета) |
| Имя span (`METHOD /route`) | `span.SetName(r.Method + " " + httpRoute(r))` | `WithSpanNameFormatter` поверх `r.Pattern` (Go 1.22+ ServeMux) |
| `SetStatus(codes.Error, ...)` для 5xx | руками | `otelhttp` |
| `slog`-логгер с `trace_id`/`span_id` | общий `request_logging.go` | общий `request_logging.go` |

`request_logging` в обоих app одинаковый: ему всё равно, кто положил активный
span в `ctx` — он просто читает его через `trace.SpanContextFromContext`. Это
и иллюстрирует точку разрыва: span на запрос — единственное, что меняется.

## Как запускать

```bash
cd app-manual && make run
# в другом терминале
curl -i http://localhost:8080/work?delay_ms=200
```

То же самое для `app-otelhttp/`. Оба слушают `:8080` по умолчанию (`HTTP_PORT`).

Env vars: `APP_NAME`, `HTTP_PORT`, `LOG_LEVEL`, `OTEL_EXPORTER_OTLP_ENDPOINT`.
По умолчанию OTLP-эндпоинт — `http://otel-collector:4317`; если коллектора нет,
для локальных проверок поднимите `otel-collector` или поставьте
`OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317` и слушайте на хосте.
