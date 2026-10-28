# 05 · Trace context across HTTP services

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

The gateway calls inventory: inbound server span → business span → outbound
client span → downstream server span. Context in memory does not cross the
network automatically; the HTTP transport must inject propagation headers.

## Observe the broken demo

```sh
go run .
```

The gateway listens on 8080 and inventory on 8081. In another terminal
(`curl.exe` in Windows PowerShell):

```sh
curl -i http://localhost:8080/checkout
```

Both servers use `otelhttp.NewHandler`, but the gateway uses a plain `http.Client`.
Inventory logs an empty `traceparent` and creates a different TraceID.
The two servers run in one process to keep setup small; they still communicate
over HTTP, so the propagation boundary is real.

## Repair the exercise

Edit `ReserveInventory` in [exercise.go](exercise.go). Keep its business span,
preserve the caller's context, and instrument the outbound transport.

```sh
go test .
go test -tags=solution .
```

The first command initially fails; the second selects [solution.go](solution.go).
The test uses `httptest.Server`, so it does not require ports 8080/8081.
It checks the shared TraceID, presence of a client span, and the downstream
server's parent being that client span. Read the [request helper](helper.go)
for client copying, request context, and response body cleanup.

The original `go run .` demonstration remains broken after the exercise is fixed.
Revisit [in-process context](../03-internal-context-propagation/README.md)
to distinguish context ownership from network propagation.
