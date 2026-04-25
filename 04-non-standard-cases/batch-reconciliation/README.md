# batch-reconciliation — long-running цикл с тикером

Демон-процесс с `time.Ticker` (2s) и graceful shutdown по `SIGINT`/`SIGTERM`. Главная идея: **каждый тик — новый root span** с `trace.WithNewRoot()`. Если обернуть всю for-петлю одним span'ом, трасса растёт бесконечно, экспортер не может её закрыть, в UI это выглядит как зависшая операция.

## Как запустить

```bash
go run .
```

Процесс крутится, пока не получит сигнал. Остановка — `Ctrl+C`. По сигналу `signal.NotifyContext` отменяет rootCtx, цикл выходит из `for-select`, defer-блок зовёт `tp.Shutdown` с таймаутом, чтобы вытолкнуть последний батч.

## Что увидеть в выводе

На каждом тике (раз в 2 секунды) — независимая трасса с тремя уровнями:

```
reconcile.tick                 <- ROOT, новый TraceID на тик
  reconcile.fetch_page (page=0)
    reconcile.one (entity.id=a)
    reconcile.one (entity.id=b)
    reconcile.one (entity.id=c)
  reconcile.fetch_page (page=1)
    ...
```

На `reconcile.tick` — атрибуты `job.name=reconciliation`, `job.page_size=3`, плюс итоговые `reconcile.processed` и `reconcile.failed`. Между тиками TraceID **разный** — это и значит «новый root».

В лог пишется `tick done: processed=N failed=M err=...` после каждого тика.

## Ключевые идеи

- **`trace.WithNewRoot()`** делает семантику явной: даже если в parent ctx случайно окажется активный span, мы его не наследуем. Защита от того, чтобы периодический процесс случайно подцепился к чужой трассе.
- **Свой таймаут на тик** через `context.WithTimeout(parent, tickTimeout)`. Если тик не успевает в `tickTimeout=1.5s`, ctx отменяется, span получает `status=Error`, следующий тик стартует штатно.
- **`signal.NotifyContext` + defer Shutdown** — единственный способ не потерять последние спаны при graceful shutdown.

## Что дальше

- Bulk reindex (span на батч документов) — [../reindex/](../reindex/).
- Fan-out на воркеры с общим root — [../cache-warmup/](../cache-warmup/).
