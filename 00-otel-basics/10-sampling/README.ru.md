# Head sampling и решение родителя

[English](README.md) | [Русский](README.ru.md)

Английская версия — оригинал. Русский перевод поддерживается рядом с ней.

## Запуск

```sh
go run .
```

Детерминированный пример использует явный провайдер и in-memory экспортёр.
Каждая строка показывает, записывается ли спан, установлен ли флаг sampled
и сколько спанов экспортировано. Бэкенд и переменные окружения не нужны.

| Сценарий | Recording | Sampled | Экспортировано |
| --- | --- | --- | --- |
| Корень с AlwaysSample | true | true | 1 |
| Корень с NeverSample | false | false | 0 |
| Корень с ParentBased(NeverSample) | false | false | 0 |
| Тот же sampler, sampled remote parent | true | true | 1 |
| Тот же sampler, unsampled remote parent | false | false | 0 |

`ParentBased` использует вложенный sampler для корней, а для потомков
по умолчанию учитывает решение родителя. На практике вложенным sampler может
быть `TraceIDRatioBased(0.1)`. Короткий запуск не гарантирует ровно 10% трасс,
поэтому здесь решения фиксированы. У незаписываемого спана остаются валидные
TraceID/SpanID для propagation.

Head sampling принимает решение при создании спана. Поздний `RecordError`
не возвращает отброшенный спан. Сбор трасс с последующим решением по полным
данным — отдельная архитектура tail sampling, за пределами примера.

Источник: [API sampling OTel Go](https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#ParentBased).
