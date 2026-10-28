# 03 · Context inside one process

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

Runnable HTTP demonstrations show how parent span information and cancellation
travel through `context.Context`. They export spans to stdout using the shared
[helper](helper.go). Start the first three examples one at a time: they bind port
8080. The fourth is a finite lifecycle demo using temporary loopback ports.

| Example | Request | Observe |
| --- | --- | --- |
| [sync-request-context](sync-request-context/README.md) | `/sync` | Handler → service → repository in one trace |
| [detached-background-context](detached-background-context/README.md) | `/bad`, `/good` | Request cancellation versus an independently timed background task |
| [worker-pool-context](worker-pool-context/README.md) | `/enqueue` | SpanContext carried in a queued job |
| [worker-pool-lifecycle](worker-pool-lifecycle/README.md) | CLI: `-mode=drain`, `timeout`, `cancel` | Bounded admission, real HTTP cancellation, and worker drain before OTel shutdown |

```sh
go run ./sync-request-context
```

In another terminal (`curl.exe` in Windows PowerShell):

```sh
curl -i http://localhost:8080/sync
```

The batch exporter prints spans after a short delay. Match `TraceID` and parent
`SpanID`, not just output order. The first three small servers are teaching examples:
their `log.Fatal(ListenAndServe(...))` paths do not implement graceful shutdown.
Allow export time before stopping them; module 01 shows a shutdown-aware server.

Read the [cancellation deep dive](../docs/internal-context-propagation/detached-background-context.md)
before moving to [cross-service propagation](../05-cross-service-propagation/README.md).
