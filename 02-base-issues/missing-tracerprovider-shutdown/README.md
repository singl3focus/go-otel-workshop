# Exercise: missing `tp.Shutdown()`

## Демонстрация проблемы

```bash
go run .
```

Процесс создает 100 span'ов и завершается сразу после `span.End()` без вызова `tp.Shutdown()`. Поскольку экспорт идет через `WithBatcher`, spans уходят в буфер BSP и могут не успеть выгрузиться - в stdout часто будет видна только часть данных или вообще ничего. Код демонстрации: [problem.go](problem.go).

## Задача

Починить функцию `Run` в [exercise.go](exercise.go) так, чтобы прошел тест:

```bash
go test .
```

## Эталонное решение

[solution.go](solution.go) под build-тегом `solution`.
