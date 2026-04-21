# Exercise: missing `span.End()`

## Демонстрация проблемы

```bash
go run .
```

В выводе `stdouttrace` будет только `good.span`; `broken.span` в экспорт не попадет. Код демонстрации: [problem.go](problem.go).

## Задача

Починить функцию `DoWork` в [exercise.go](exercise.go) так, чтобы прошел тест:

```bash
go test .
```

## Эталонное решение

Лежит в [solution.go](solution.go) под build-тегом `solution` (по умолчанию не компилируется, чтобы не конфликтовать с `exercise.go`). Открывать только после того, как попробовал сам.
