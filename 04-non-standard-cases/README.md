# 04. Non-standard cases — трассировка без входящего запроса

Четыре сценария, где нет HTTP handler'а и нет готового `ctx` с trace-состоянием: процесс сам стартует root span, сам управляет его жизненным циклом и сам отвечает за flush перед выходом.

## Формат

Как в [03-internal-context-propagation/](../03-internal-context-propagation/) — подпапки самостоятельные, запускаются `go run .`, спаны идут в stdout через `stdouttrace` (см. [helper.go](helper.go)). Это эталонные версии, не упражнения.

## Общие принципы

1. **Нет входящего ctx → сами создаём root span.** Старт всегда от `context.Background()`. Никаких `r.Context()`, никаких наследуемых cancel'ов.
2. **Shutdown обязателен.** Батч-экспортер держит спаны в буфере; без `tp.Shutdown(ctx)` с таймаутом короткоживущие CLI теряют последние спаны.
3. **Периодический запуск → новый root на тик.** Один span на всю петлю = бесконечно растущая трасса. Каждый тик — `tracer.Start(..., trace.WithNewRoot())`.
4. **Bulk операции → span на батч, не на элемент.** 10M документов × span = span-storm. Используй счётчики-атрибуты и span events для отдельных элементов.
5. **Root-span получает итоговые счётчики.** `processed`, `failed`, `skipped` как атрибуты root. Диагностика по одному span'у без агрегаций.
6. **Ошибки: `RecordError` + `SetStatus(codes.Error, ...)`.** И то, и другое — RecordError кладёт event со стектрейсом, SetStatus влияет на UI/alerting.

## Список примеров

### [migrations/](migrations/) — трассировка миграций

Один root `migrations.run` на прогон, child span на каждую миграцию. На root'е оседают `migrations.total` и `migrations.applied`, чтобы по одному span'у было видно, где прогон прервался.

### [batch-reconciliation/](batch-reconciliation/) — cron/reconciliation job

Long-running процесс с тикером. **Каждый тик — отдельный root span с `trace.WithNewRoot()`.** Свой таймаут на тик через `context.WithTimeout`. Остановка по SIGINT/SIGTERM через `signal.NotifyContext`, flush в defer.

### [reindex/](reindex/) — bulk reindex

Span на батч, не на документ. Ошибки отдельных документов — через `span.AddEvent("doc.failed", ...)`. Итоговые `indexed`/`failed` на root-span.

### [cache-warmup/](cache-warmup/) — прогрев кэша после деплоя

Fan-out на N воркеров, общий root, span на воркера. В данном случае ctx пробрасывается напрямую (воркеры не переживают `run()`), поэтому trick'а с `SpanContext only` из [03/worker-pool-context](../03-internal-context-propagation/worker-pool-context/) здесь не нужно.

## Как запускать

```bash
cd 04-non-standard-cases/<case>
go run .
```

Для `batch-reconciliation` процесс крутится до Ctrl+C — остановка триггерит graceful shutdown и flush экспортера.
