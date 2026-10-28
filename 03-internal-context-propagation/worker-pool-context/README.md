# Worker pool context

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

Two workers consume jobs from an in-memory channel. A job stores only a
`trace.SpanContext` and an order ID, not the complete request context.

```sh
go run .
```

In another terminal (`curl.exe` in Windows PowerShell):

```sh
curl -i http://localhost:8080/enqueue
```

The response is 202. Expect `http.enqueue`, its child `handler.enqueue`, and
`worker.process` whose parent is `handler.enqueue`. All three share a TraceID.
`worker.process` includes `worker.id` and `order.id`; its status stays Unset.
Scheduling can start the worker before or after the handler returns.

The worker reconstructs a context with `trace.ContextWithSpanContext`, then
sets its own timeout. The demo uses `time.Sleep`, which does not observe that
timeout; real I/O should receive and honor the context.
For a separate trace linked to its producer, see
[links versus parents](../../00-otel-basics/07-links-vs-parent/).

For cancellation-aware I/O, overload rejection, and shutdown, continue with
[worker-pool-lifecycle](../worker-pool-lifecycle/README.md).
