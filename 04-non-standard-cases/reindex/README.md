# Batch reindexing

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

## Run

```sh
go run .
```

## Expected trace

Five hundred simulated documents are processed in five batches of 100. Expect six spans in one trace: `reindex.run` and five `reindex.batch` children. There is no span per document.

Documents at offsets 0, 137, 274, and 411 produce `doc.failed` events with `doc.id` and `reason=validation`. Root totals are `reindex.indexed=496` and `reindex.failed=4`. The simulated batch operation completes, so its status remains Unset despite individual failures.

Spans use the stdout batch exporter; allow its export delay or wait for normal
process completion. See [all job examples](../README.md).
