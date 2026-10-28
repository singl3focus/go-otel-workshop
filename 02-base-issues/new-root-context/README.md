# Exercise · lost parent context

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

## Observe

```sh
go run .
```

This HTTP demo listens on port 8080. Request `/orders`: the business operation replaces the incoming context and starts a separate trace.

In another terminal: `curl -i http://localhost:8080/orders` (use `curl.exe` in Windows PowerShell).

## Repair

Edit `CreateOrder` in [exercise.go](exercise.go). Preserve the incoming context in `CreateOrder` and pass its returned context to `saveOrder`. All three test spans must share a TraceID with correct parent linkage.

```sh
go test .
go test -tags=solution .
```

The first command checks your answer and initially fails. The second selects
the [reference solution](solution.go) without overwriting your work.
Keep the broken [demonstration](problem.go) for comparison.
Return to [all exercises](../README.md).
