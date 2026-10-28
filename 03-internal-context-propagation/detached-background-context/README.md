# Detached background work

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

Run the server, then send the two requests from another terminal:

```sh
go run .
```

```sh
curl -i http://localhost:8080/bad
curl -i http://localhost:8080/good
```

Use `curl.exe` in Windows PowerShell. Both return 202 immediately. After a short
delay the worker starts its span and simulates a context-aware downstream call.

| Path | Background context | Result |
| --- | --- | --- |
| `/bad` | Original request context | `background.bad` records cancellation and Error |
| `/good` | `WithoutCancel` plus a two-second timeout | `background.good` completes with Unset status and `context.mode=detached` |

Both background spans retain their HTTP parent's TraceID. Cancellation affects
the simulated work, not the presence of the parent SpanContext.
The good goroutine owns its `cancel` call; placing `defer cancel()` in the handler
would cancel its new context as soon as the response returns.

See the [deep dive](../../docs/internal-context-propagation/detached-background-context.md)
and [worker queue alternative](../worker-pool-context/README.md).
