# 05. Cross-service propagation

Как trace context переходит через границу процесса: HTTP handler → business span → outbound HTTP client → downstream HTTP handler. Это следующий шаг после [03-internal-context-propagation/](../03-internal-context-propagation/): внутри процесса достаточно передавать `context.Context`, но между сервисами trace-id/span-id должны уехать в сетевые заголовки.

## Демонстрация проблемы

```bash
go run .
```

Запускаются два HTTP-сервера:

- `gateway` на `:8080`;
- `inventory` на `:8081`.

В другом терминале:

```bash
curl -i http://localhost:8080/checkout
```

Оба сервиса обернуты в `otelhttp.NewHandler`, поэтому входящие server spans создаются. Но `gateway` вызывает `inventory` обычным `http.Client`, без `otelhttp.NewTransport`. Из-за этого в исходящий HTTP-запрос не попадает `traceparent`, и `inventory` начинает новый trace.

Код демонстрации: [problem.go](problem.go).

## Формат упражнения

Как в [02-base-issues/](../02-base-issues/), раздел оформлен как самостоятельный `package main`:

- `problem.go` — запускаемая демонстрация антипаттерна (`go run .`);
- `exercise.go` — функция-стаб, которую нужно починить;
- `exercise_test.go` — тест на `tracetest.SpanRecorder` + `httptest.Server`, который становится зеленым после корректного исправления (`go test .`);
- `solution.go` — эталонное решение под build-тегом `solution`, по умолчанию не компилируется;
- `helper.go` — общий `stdouttrace` TracerProvider для демонстрации.

## Задача

Починить функцию `ReserveInventory` в [exercise.go](exercise.go) так, чтобы прошел тест:

```bash
go test .
```

Тест проверяет три вещи:

- downstream handler получил тот же TraceID, что и родительский span;
- появился outbound client span;
- downstream server span стал child span'ом outbound client span.

Эталонное решение: [solution.go](solution.go) под build-тегом `solution`.

## Главное правило

`context.Context` сам по себе не пересекает сеть. Он только хранит активный span внутри процесса. На HTTP-границе нужен propagator:

- server side: `otelhttp.NewHandler(...)` извлекает `traceparent` из входящих заголовков;
- client side: `otelhttp.NewTransport(...)` создает client span и инжектит `traceparent` в исходящий запрос.

Если инструментирована только server side, каждый сервис будет честно писать spans, но trace развалится на несколько независимых root.
