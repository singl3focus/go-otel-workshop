# sync-request-context — baseline propagation через ctx

Корректный синхронный путь: `handler → service → repo`, ctx идёт по цепочке без подмен. Все span'ы попадают в **одну** трассу с правильной иерархией parent/child. Используется как точка отсчёта для остальных кейсов раздела.

## Как запустить

```bash
go run .
```

В другом терминале:

```bash
curl -i http://localhost:8080/sync
```

## Что увидеть в выводе

`stdouttrace` напечатает четыре span'а одной трассы (один TraceID, цепочка ParentSpanID):

```
http.sync          <- otelhttp, server-span
  handler.sync     <- a.tracer
    service.process
      repo.save
```

`service.process` несёт атрибут `flow.type=sync`. `repo.save` — самый глубокий, у него нет потомков. Все четыре span'а имеют одинаковый TraceID; ParentSpanID каждого следующего совпадает со SpanID предыдущего.

## Что дальше

- Если фон/очереди/воркеры обрезают трассу — [../detached-background-context/](../detached-background-context/) и [../worker-pool-context/](../worker-pool-context/).
- Если входящий trace context теряется уже на уровне handler'а — упражнение [../../02-base-issues/new-root-context/](../../02-base-issues/new-root-context/).
