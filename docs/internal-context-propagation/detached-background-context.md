# Detached background context

[English](detached-background-context.md) | [Русский](detached-background-context.ru.md)

English is the original; the Russian translation is maintained alongside it.

The [runnable example](../../03-internal-context-propagation/detached-background-context/README.md)
contrasts a request context with independently timed background work.

## Request lifetime and operation lifetime

For an incoming HTTP request, the context is canceled when the client connection
closes, the request is canceled (HTTP/2), or `ServeHTTP` returns. Do not treat every
server timeout setting as a generic context timeout: read/write deadlines and
handler cancellation have different behavior.

Returning a 202 does not keep the request context alive for queued work.
A context-aware HTTP/DB/RPC client can return cancellation when it receives that
old context. `time.Sleep`, by contrast, ignores cancellation and can conceal the
problem in a demo.

## Cancellation does not erase the trace parent

`SpanContext` is a value reachable through the context. Canceling the context
does not erase its TraceID/SpanID. A recording child can still be created,
annotated, ended, and exported; downstream work using that same context can fail.
Span lifetime is controlled by End, not by canceling its creation context.

Consequently `/bad` can show a correctly linked `background.bad` span with an
exception event and Error status. The failure is in the simulated downstream
operation, not evidence that the SDK lost its parent.

## Give independent work an owner

`context.WithoutCancel` preserves values while removing the parent's deadline
and cancellation. Add a bounded lifetime and let the goroutine that performs
the work own its cancellation:

```go
func (a *app) handleGood(w http.ResponseWriter, r *http.Request) {
    bg := context.WithoutCancel(r.Context())
    bg, cancel := context.WithTimeout(bg, 2*time.Second)
    go func() {
        defer cancel()
        ctx, span := a.tracer.Start(bg, "background.good")
        defer span.End()
        if err := simulateDownstreamCall(ctx); err != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, "background operation failed")
        }
    }()
    w.WriteHeader(http.StatusAccepted)
}
```

Placing `defer cancel()` in the handler would actively cancel `bg` on handler
return. It does more than release a timer. Moving the defer into the goroutine
lets the work finish or reach its own deadline. The context does not kill a
goroutine: downstream operations must observe its cancellation.
For production work, also connect worker shutdown to the application's lifetime
and drain it before shutting down tracing. WithoutCancel intentionally removes
the original parent's cancellation, including a shutdown signal carried there.

## Choose the values and trace relationship you need

| Approach | Trace relationship | Values and cancellation |
| --- | --- | --- |
| Keep request context | Same trace | Keeps values and request cancellation |
| WithoutCancel + own timeout | Same trace | Keeps all values; independent timeout |
| SpanContext in job + new context | Same trace by default | Only explicit job data and parent span IDs |
| New root with `trace.Link` | Separate linked trace | Explicitly model an independently owned operation |
| Inject/Extract carrier | Depends on consumer span start | Suitable for a serialized message or process boundary |

The [worker pool](../../03-internal-context-propagation/worker-pool-context/README.md)
uses only SpanContext in each job. Inside the worker it calls
`trace.ContextWithSpanContext(context.Background(), job.parent)`, then adds its
own timeout. This deliberately avoids retaining arbitrary request values.
When all workers finish before a synchronous parent returns, ordinary context
passing can be appropriate, as in [cache warmup](../../04-non-standard-cases/cache-warmup/README.md).

## Trace timing and propagation

A child can start or finish after its parent ends. Parent/child describes the
logical relationship; it does not force nested time intervals. Choose a link
when a separate trace better expresses operation ownership, rather than only to
make a timeline look tidy.

Preserving context values also preserves the active span for instrumentation
that injects HTTP headers. WithoutCancel alone is not network propagation:
an outbound transport or explicit propagator must still write `traceparent`.

## Review questions

- Does the task need to survive the HTTP response, and who drains it at shutdown?
- Is the worker using cancellation-aware operations or only a plain sleep?
- Does the background owner call cancel, and is its lifetime bounded?
- Are all request values needed, or would explicit job data and SpanContext suffice?
- Does the test cancel the original request after the response and verify both
  successful work and parent linkage?

## References

- [net/http Request.Context](https://pkg.go.dev/net/http#Request.Context)
- [context.WithoutCancel](https://pkg.go.dev/context#WithoutCancel)
- [context.WithTimeout](https://pkg.go.dev/context#WithTimeout)
- [OTel Go Span API](https://pkg.go.dev/go.opentelemetry.io/otel/trace#Span)

A complete finite [worker-lifecycle example](../../03-internal-context-propagation/worker-pool-lifecycle/README.md)
shows admission limits, a real cancellation-aware HTTP call, and worker draining.
