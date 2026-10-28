# Cache warmup with worker spans

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

## Run

```sh
go run .
```

## Expected trace

Four workers consume 200 keys from a shared channel. Expect five spans in one trace: `cache.warmup` and four `cache.worker` children. Workers overlap in time, and each records `worker.id`, `worker.warmed`, and `worker.miss`.

The deterministic demo rule is `len(key)%10 != 0`: all generated keys pass, so root totals are `keys.warmed=200` and `keys.miss=0`. It is not a random 10% failure simulation. Change the rule to explore miss events. Workers share the parent context safely because `wg.Wait()` keeps the run alive until they finish.

Spans use the stdout batch exporter; allow its export delay or wait for normal
process completion. See [all job examples](../README.md).
