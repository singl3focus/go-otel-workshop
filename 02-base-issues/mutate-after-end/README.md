# Exercise · mutations after End

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

## Observe

```sh
go run .
```

`phase=before-end` and Ok survive; `phase=after-end`, `late-event`, and the late Error status do not change the completed span.

## Repair

Edit `DoWork` in [exercise.go](exercise.go). Record `result=success` and status Ok while the span is recording, then finish it. The test checks the exported final state.

```sh
go test .
go test -tags=solution .
```

The first command checks your answer and initially fails. The second selects
the [reference solution](solution.go) without overwriting your work.
Keep the broken [demonstration](problem.go) for comparison.
Return to [all exercises](../README.md).
