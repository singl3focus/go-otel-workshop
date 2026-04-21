# Exercise: mutate after End()

## Демонстрация проблемы

```bash
go run .
```

В выводе `stdouttrace` у span `mutate.after.end` будут только те атрибуты и статус, что записаны **до** `End()`; все мутации после `End()` молча игнорируются. Код демонстрации: [problem.go](problem.go).

## Задача

Починить функцию `DoWork` в [exercise.go](exercise.go) так, чтобы прошел тест:

```bash
go test .
```

## Эталонное решение

[solution.go](solution.go) под build-тегом `solution`.
