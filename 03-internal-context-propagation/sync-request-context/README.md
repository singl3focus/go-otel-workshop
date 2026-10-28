# Synchronous request context

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

From this directory, run the server and request `/sync` in another terminal:

```sh
go run .
```

```sh
curl -i http://localhost:8080/sync
```

Use `curl.exe` in Windows PowerShell. The response is `sync ok`.
The HTTP server span contains `handler.sync`, then `service.process`, then
`repo.save` as nested children. All four share one TraceID.
The service span includes `flow.type=sync`.

Each function receives the caller's context and passes the context returned by
`tracer.Start` to its child. Replacing it with `context.Background()` would break
the chain. Use this example as the baseline for [background work](../detached-background-context/README.md).
