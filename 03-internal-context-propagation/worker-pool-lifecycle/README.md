# Worker-pool lifecycle: admission, cancellation, and drain

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

This finite example extends the [small worker pool](../worker-pool-context/README.md)
with lifecycle decisions: a bounded queue, explicit rejection, context-aware HTTP
I/O, per-job timeouts, and worker completion before tracing shutdown.
It starts its own loopback HTTP server on a temporary port and exits by itself.
No Docker, fixed port, or external service is required.

## Run three scenarios

From this directory:

```sh
go run . -mode=drain
go run . -mode=timeout
go run . -mode=cancel
```

Each run uses one worker and one waiting slot. The first job enters downstream
HTTP; the second waits; a third is rejected immediately as queue-full. Admission
then closes, and another submission is rejected as closed. The original request
span ends and its context is canceled before the pool shuts down.

| Mode | How work stops | Summary |
| --- | --- | --- |
| `drain` | Downstream is released; both accepted jobs finish | `accepted=2 completed=2 failed=0 queue_full=1` |
| `timeout` | Downstream waits for request cancellation; each job reaches its own timeout | `accepted=2 completed=0 failed=2 queue_full=1` |
| `cancel` | The drain budget is already canceled; pool cancellation interrupts active I/O and cancels the waiting job | `accepted=2 completed=0 failed=2 queue_full=1` |

Every summary also contains `after_close=closed`. Batch-exported JSON follows:
`demo.run`, `request.enqueue`, and two `worker.process` spans, all in one trace.
Worker spans are children of `request.enqueue`, not of a new background root.
`job.outcome` is `completed`, `deadline`, or `canceled`; failed jobs have an
exception event and Error status. Successful spans retain Unset status.
The third rejected job has no worker span because it was never accepted.

## What the code separates

- **Admission:** `Submit` checks request cancellation before admission and sends
  without waiting for queue capacity. Queue-full and pool-closed are distinct errors.
  Once accepted, later request cancellation does not withdraw the job.
- **Trace values:** a job stores its ID and `trace.SpanContext`, rather than the
  request context. The worker restores that parent onto the pool's lifetime context.
- **Job budget:** `WithTimeout` begins when a worker dequeues the job. Queue wait
  time is not included; use a separate deadline in the job if your contract needs it.
- **I/O:** `http.NewRequestWithContext` carries that budget into a real `Client.Do`.
  The downstream handler waits on its request context. No plain sleep simulates I/O.
- **Stopping:** `Close` stops acceptance and closes the queue. `Shutdown` first
  drains accepted jobs. If its budget expires, it cancels pool I/O and waits for
  all workers to exit. Only then does the demo end its root span and shut down OTel.

`Submit` and `Close` share a mutex. The send is nonblocking while the mutex is
held, so a concurrent close cannot race with a send to a closed channel.
That is also why the queue's full condition produces a clear overload result.

## Boundaries of this example

Cancellation is cooperative. `Shutdown` continues joining after its budget expires;
it does not promise a hard return deadline for a custom HTTP transport that ignores
context. The demonstrated transport observes cancellation. This chooses complete
worker cleanup before telemetry shutdown. A production service also needs an
overall process termination policy for broken dependencies.

Accepted jobs live only in memory: process crashes lose the queue. There are no
retries, persistence, HTTP enqueue endpoint, or outbound client spans here.
The loopback call demonstrates cancellation, while
[module 05](../../05-cross-service-propagation/README.md) teaches network tracing.
For an HTTP-facing pool, map overload to an appropriate response such as 503,
stop accepting requests, let handlers finish, then drain the pool.

## Verification

```sh
go test -race -count=1 -timeout=30s .
```

Tests coordinate with channels instead of sleep-based scheduling. They check
bounded admission, preserved parent IDs, request independence, actual HTTP
deadline cancellation, queued-job cancellation, concurrent Submit/Close,
and worker spans ending before provider shutdown. The deadline test necessarily
uses a real timer; other ordering is explicit.

References: [Request context](https://pkg.go.dev/net/http#Request.Context),
[CancelFunc](https://pkg.go.dev/context#CancelFunc),
[Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown).
