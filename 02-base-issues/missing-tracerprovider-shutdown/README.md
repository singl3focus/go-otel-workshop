# Exercise · missing TracerProvider.Shutdown()

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

## Observe

```sh
go run .
```

The demo creates 100 spans and exits immediately. A batch processor can lose queued spans before its background export runs.

## Repair

Edit `Run` in [exercise.go](exercise.go). Before returning, export the ended span and shut down the provider. The test checks both export and exporter shutdown. Use a bounded shutdown context in real applications.

```sh
go test .
go test -tags=solution .
```

The first command checks your answer and initially fails. The second selects
the [reference solution](solution.go) without overwriting your work.
Keep the broken [demonstration](problem.go) for comparison.
Return to [all exercises](../README.md).
