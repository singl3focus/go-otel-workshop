# Missing span.End(): lifecycle, memory, and export

[English](missing-span-end.md) | [Русский](missing-span-end.ru.md)

English is the original; the Russian translation is maintained alongside it.

The [exercise](../../02-base-issues/missing-span-end/README.md) shows an ended
`good.span` and an unended `broken.span`. Only the first reaches the ordinary
SDK export pipeline.

## Lifecycle of a recording span

1. `tracer.Start` creates a recording span when the sampler permits it, returns
   a context containing the span, and notifies processors through `OnStart`.
2. Attributes, events, errors, and status can be updated while it records.
3. `End` records the end time and invokes `OnEnd` with a read-only view.
4. The simple processor exports synchronously; the batch processor queues the
   completed span for later export. A recording span is not necessarily exported
   by every processor: custom sampling/processor choices can change that behavior.

An omitted End skips `OnEnd`. Normal simple/batch processors therefore do not
export that span. `Shutdown` flushes completed spans and stops the pipeline;
it does not finish the application's open operations.

| Mistake | State | Repair |
| --- | --- | --- |
| Missing End | Span has not entered the ordinary completed-span export path | End the span at the operation boundary |
| Missing Shutdown | Ended spans may remain queued when the process exits | Drain work, then shut down with a bounded context |

## Garbage collection is not span completion

Go collects objects based on reachability. An unreachable unended span can be
collected without being exported. GC does not call End, so lack of increasing
memory does not prove correct instrumentation.

```go
func handle(ctx context.Context, tracer trace.Tracer) {
    _, span := tracer.Start(ctx, "operation")
    _ = span
}
```

If nothing retains the derived context or span, this function loses telemetry.
If a long-lived field, queue, or goroutine retains that context, it can also
retain the span and its data. Standard processors do not maintain a registry of
every active span; custom processors may retain spans in `OnStart`.
SDK attribute/event limits also mean that one retained span need not grow without
bound. Retaining an ever-growing number of contexts is a separate memory problem.

Do not use finalizers to supply operation boundaries. Their timing is unrelated
to the operation's end and is not guaranteed at process exit. A fabricated late
end time would distort latency even if it happened to export successfully.
Observe actual heap/goroutine profiles when investigating retention. Process RSS
includes more than Go heap allocations; it is not a one-to-one count of spans.

## Missing parents and orphan spans

A child inherits its parent's TraceID and SpanID values. It can end and export
even if the parent never ends. The stored trace then has a child referring to
a parent absent from the backend, often called an orphan span. UI rendering
varies by backend.

The child's SpanContext contains IDs, not an owning reference to the parent
object. A derived context may still retain its ancestry. Other causes of missing
parents include export loss and inconsistent sampling. Missing propagation is
different: a downstream service often starts an unrelated trace with a new ID,
rather than an orphan referring to the old parent.

## Provider shutdown and operation order

Shut down the provider after the work you intend to trace has completed.
New spans from a shut-down SDK provider are non-recording; ending an old span
after processors stop cannot be relied upon for delivery. Repeated shutdown
does not restart the provider. Cancellation and telemetry shutdown have separate
responsibilities: stop accepting work, drain workers, end operations, then flush.

## References

- [OTel Go SDK span implementation](https://github.com/open-telemetry/opentelemetry-go/blob/v1.43.0/sdk/trace/span.go)
- [Span processor implementation](https://github.com/open-telemetry/opentelemetry-go/blob/v1.43.0/sdk/trace/batch_span_processor.go)
- [TracerProvider.Shutdown API](https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#TracerProvider.Shutdown)
