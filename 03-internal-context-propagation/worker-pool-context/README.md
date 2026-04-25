# worker-pool-context — очередь, передаём SpanContext, не ctx

Альтернативный путь к проблеме из [../detached-background-context/](../detached-background-context/) применительно к воркер-пулам. Вместо того чтобы класть в job весь `ctx` (унаследовали бы cancel запроса и middleware-values, которые в воркере не нужны), кладётся только `trace.SpanContext` родителя. В воркере ctx собирается с нуля: `context.Background()` + `trace.ContextWithSpanContext(...)` для линковки трассы + `context.WithTimeout(...)` для своего дедлайна.

## Как запустить

```bash
go run .
```

В другом терминале:

```bash
curl -i http://localhost:8080/enqueue
```

## Что увидеть в выводе

После одного запроса `stdouttrace` напечатает три span'а в одной трассе:

```
http.enqueue        <- otelhttp, server-span
  handler.enqueue   <- a.tracer
worker.process      <- общий TraceID, parent = SpanContext из job
```

Ключевая деталь: `worker.process` стартует **после** того, как handler вернул ответ клиенту. Если бы воркер использовал `r.Context()`, span попал бы в трассу с `status=Error`, потому что ctx уже отменён. Здесь он `status=Ok`, у `worker.process` есть атрибуты `worker.id` и `order.id`. TraceID совпадает с `http.enqueue`/`handler.enqueue` — линковка через `SpanContext` сохраняется.

При нескольких запросах подряд: TraceID у каждой пары `handler.enqueue` / `worker.process` свой, но внутри пары совпадает.

## Что дальше

- Сравнение с другими альтернативами (`WithoutCancel` в handler'е vs в воркере, `trace.Link`, propagator через `MapCarrier`) — [docs/internal-context-propagation/detached-background-context.md](../../docs/internal-context-propagation/detached-background-context.md#4-когда-withoutcancel-не-нужен-worker-pool-как-граничный-случай).
