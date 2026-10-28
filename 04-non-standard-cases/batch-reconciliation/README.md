# Periodic reconciliation

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

## Run

```sh
go run .
```

## Expected trace

The process runs immediately, then every two seconds. Each tick has a different root TraceID. `reconcile.tick` has two `reconcile.fetch_page` children and six `reconcile.one` children directly under the root: fetch and item spans are siblings, not nested.

Each run records `job.name`, `job.page_size`, `reconcile.processed=6`, and `reconcile.failed=0`. A 1.5-second context is available to the run, but the simulated sleeps do not observe it. Press Ctrl+C to stop; signal handling exits the loop and flushes pending spans.

Spans use the stdout batch exporter; allow its export delay or wait for normal
process completion. See [all job examples](../README.md).
