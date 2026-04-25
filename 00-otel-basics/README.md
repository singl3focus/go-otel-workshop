# 00. OpenTelemetry basics

Пять минимальных программ. Каждая — самостоятельный `package main` на 30-40
строк, запускается `go run .`, печатает span'ы в stdout через `stdouttrace`.
Каждый шаг добавляет ровно одну новую идею поверх предыдущего.

Если первый раз видишь OTel — иди по порядку. Если ищешь что-то конкретное —
таблица ниже.

| Шаг | Что появляется |
|---|---|
| [01-minimal/](01-minimal/) | exporter → TracerProvider → tracer → `Start`/`End` → `Shutdown` |
| [02-attributes-events-status/](02-attributes-events-status/) | `SetAttributes`, `AddEvent`, `SetStatus`, `RecordError` |
| [03-nested-spans/](03-nested-spans/) | parent/child через `context.Context` |
| [04-resource/](04-resource/) | `service.name` и почему он важен |
| [05-batcher-vs-syncer/](05-batcher-vs-syncer/) | `WithBatcher` без `Shutdown` теряет span'ы |

## Как запускать

```bash
cd 01-minimal && go run .
```

В каждой подпапке свой `main.go`. Зависимости одни на всю секцию (один `go.mod`).

## Что дальше

- Реальный HTTP-сервер с OTLP-экспортером и propagator из заголовков —
  [01-local-observability-stack/](../01-local-observability-stack/).
- Типичные ошибки и упражнения — [02-base-issues/](../02-base-issues/).
- Контекст-пропагация внутри процесса — [03-internal-context-propagation/](../03-internal-context-propagation/).
- Трассировка без HTTP (миграции, реконсиляция, bulk) —
  [04-non-standard-cases/](../04-non-standard-cases/).
