# 04 · Jobs, batches, and CLI programs

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

Four complete demonstrations start traces without an incoming HTTP request.
They use the shared [stdout exporter helper](helper.go) and flush on normal exit.

| Scenario | Trace boundary | Aggregates |
| --- | --- | --- |
| [migrations](migrations/README.md) | Root per migration run; child per migration | Applied and total migrations |
| [batch-reconciliation](batch-reconciliation/README.md) | New root per periodic tick | Processed and failed items |
| [reindex](reindex/README.md) | Root per run; child per document batch | Indexed/failed counts, failure events |
| [cache-warmup](cache-warmup/README.md) | Root per run; child per worker | Warmed/missing keys |

From this directory:

```sh
go run ./migrations
```

Choose trace boundaries that describe a finite operation. Batch/worker spans
avoid creating a span per item. Keep a bounded shutdown context to flush the
exporter after work completes. When an operation fails, `RecordError` adds an
exception event; `SetStatus(codes.Error, ...)` sets its error status separately.

The simulated delays in these examples use `time.Sleep`; declaring a timeout
does not interrupt a sleep. Pass context to real downstream clients and check
cancellation in loops. These examples illustrate instrumentation boundaries,
not a complete resilient job runner.
