# go-otel-workshop

Учебный воркшоп по OpenTelemetry в Go: от первого `tracer.Start` до трассировки фоновых джобов и bulk-операций. Каждый раздел — отдельный Go-модуль, который запускается из своей папки и не зависит от соседей.

## Кому адресовано

Go-разработчику, любого уровня. Достаточно знать `context.Context`, `net/http` и уметь запускать `go run .`. Опыт работы с Jaeger/Tempo/коллектором не требуется — большая часть примеров пишет span'ы прямо в stdout.

## Что внутри

| Раздел | О чём | Когда сюда идти |
|---|---|---|
| [00-otel-basics/](00-otel-basics/) | Пять минимальных `package main`: exporter → TracerProvider → tracer → `Start`/`End` → `Shutdown`, потом attributes/events/status, nested spans, resource, batcher vs syncer | Если первый раз видишь OTel SDK |
| [01-local-observability-stack/](01-local-observability-stack/) | HTTP-сервис с OTLP/gRPC экспортером в двух вариантах: ручной middleware (`app-manual/`) vs `otelhttp` (`app-otelhttp/`) | После 00, чтобы увидеть «как это выглядит в настоящем сервисе» |
| [02-base-issues/](02-base-issues/) | Упражнения по типовым антипаттернам: забытый `End()`, отсутствие `Shutdown()`, мутация после `End()`, потеря входящего trace context | Когда хочется проверить себя — `go test .` зеленеет, только если поправил правильно |
| [03-internal-context-propagation/](03-internal-context-propagation/) | Как `context.Context` переносит trace-состояние внутри процесса: handler → service → repo, фоновая горутина, worker-pool через очередь | Когда базовое HTTP-приложение работает, а отвалилась трассировка в фоновой работе |
| [04-non-standard-cases/](04-non-standard-cases/) | Трассировка без входящего HTTP-запроса: миграции, cron-реконсиляция, bulk reindex, прогрев кэша | Когда нужно инструментировать CLI/cron/джоб, а не handler |
| [05-cross-service-propagation/](05-cross-service-propagation/) | Упражнения по trace propagation между сервисами: входящий handler → outbound HTTP client → downstream handler | Когда внутри процесса всё связано, но в микросервисах trace распадается на несколько root |
| [docs/](docs/) | Глубокие разборы избранных кейсов: жизненный цикл span'а, `WithoutCancel`, ловушки propagation | Когда поверхностного объяснения мало и нужно «почему именно так» |

## Рекомендуемый порядок прохождения

Линейный, 00 → 04. Это естественная кривая сложности.

- **00 → 01.** Обязательно по очереди, если впервые видишь OTel.
- **01.** Можно скипнуть, если уже строил HTTP-сервис с OTel. Стоит зайти хотя бы в `app-manual/` и сравнить с `app-otelhttp/` — пара показывает, что именно `otelhttp` делает за тебя.
- **02.** Можно проходить параллельно с 01: упражнения короткие, по 1 файлу.
- **03.** Идти после 01 — там нужен HTTP-handler как контекст.
- **04.** Идти после 03 — здесь обыгрываются те же паттерны, но без входящего ctx.
- **05.** Идти после 03 — это та же propagation-идея, но уже через HTTP-границу между сервисами.

## Предусловия окружения

- Go 1.25.1 (см. `go.mod` в каждом разделе).
- Никаких docker/kubernetes для разделов 00, 02, 03, 04 — экспорт идёт в stdout через `stdouttrace`.
- Раздел 01 ждёт OTLP/gRPC коллектор на `OTEL_EXPORTER_OTLP_ENDPOINT` (по умолчанию `http://otel-collector:4317`). Без коллектора приложение запустится, но экспорт будет молча падать. Самый быстрый способ — поднять [otel-collector](https://opentelemetry.io/docs/collector/) локально и поставить `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317`.

Запуск любого примера:

```bash
cd <раздел>/<подпапка>
go run .
```

В разделе 02 — `go test .` для проверки решения упражнения.

## Как устроены упражнения в `02-base-issues/`

Каждая подпапка — самостоятельный `package main` из четырёх файлов:

- `problem.go` — запускаемая демонстрация антипаттерна. `go run .` показывает сломанное поведение.
- `exercise.go` — функция-стаб, которую нужно починить.
- `exercise_test.go` — тест на `tracetest.SpanRecorder`. Зеленеет после корректной правки. `go test .`.
- `solution.go` — эталон под build-тегом `solution`. По умолчанию **не компилируется**, чтобы не конфликтовать с `exercise.go`. Открывать имеет смысл только после собственной попытки.

Подробнее — в [02-base-issues/README.md](02-base-issues/README.md).

## Линтер

Конфиг — [.golangci.yml](.golangci.yml) (golangci v2, набор `govet`/`errcheck`/`staticcheck`/`unused`/`ineffassign`/`whitespace`/`gocyclo`).

```bash
golangci-lint run
```

Отчёт пишется в [reports/linters-out.txt](reports/linters-out.txt) (папка в `.gitignore`, генерится локально).
