# Exercise: new root context

## Демонстрация проблемы

```bash
go run .
```

Стартует HTTP-сервер на `:8080`. В другом терминале:

```bash
curl -i http://localhost:8080/orders
```

В stdout будут видны два трейса вместо одного: отдельно `http.server` и отдельно `service.create_order` + `repo.save_order` — бизнес-логика обрезала входящий trace context. Код демонстрации: [problem.go](problem.go).

## Задача

Починить функцию `CreateOrder` в [exercise.go](exercise.go) так, чтобы прошел тест:

```bash
go test .
```

## Эталонное решение

[solution.go](solution.go) под build-тегом `solution`.
