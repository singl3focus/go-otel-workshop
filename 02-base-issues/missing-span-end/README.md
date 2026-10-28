# Exercise · missing span.End()

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

## Observe

```sh
go run .
```

Only `good.span` is exported; `broken.span` never calls End. Provider shutdown does not finish open spans.

## Repair

Edit `DoWork` in [exercise.go](exercise.go). Finish `demo.operation` and preserve its `demo=exercise` attribute. The test must see exactly one ended span.

```sh
go test .
go test -tags=solution .
```

The first command checks your answer and initially fails. The second selects
the [reference solution](solution.go) without overwriting your work.
Keep the broken [demonstration](problem.go) for comparison.
Return to [all exercises](../README.md).
