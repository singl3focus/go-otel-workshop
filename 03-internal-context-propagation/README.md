# 03. Internal context propagation

Как `context.Context` переносит trace-состояние внутри одного процесса: через вызовы функций, через фоновые горутины, через очереди и воркер-пулы. Что ломается, если делать это «в лоб», и как делать правильно.

## Формат

В отличие от [02-base-issues/](../02-base-issues/), это не упражнения, а запускаемые демонстрации. Каждая подпапка — самостоятельный `package main` с одним или двумя HTTP-эндпоинтами; поведение сверяется по stdout-экспортеру трассировки (`stdouttrace` в [helper.go](helper.go)).

Общий каркас в каждом примере:

- `NewTracerProvider(...)` из корня модуля (shared helper);
- HTTP-сервер на `:8080` с одним-двумя путями (`/bad`, `/good`, `/sync`, `/enqueue` — см. ниже);
- по каждому запросу SDK печатает полные spans в stdout pretty-printed JSON.

## Как запускать

В подпапке интересующего примера:

```bash
go run .
```

В другом терминале — curl на соответствующий путь. Спаны видно в stdout процесса.

## Список примеров

### [sync-request-context/](sync-request-context/)

Базовый корректный паттерн: ctx пробрасывается через handler → service → repo, все spans попадают в **одну** трассу.

```bash
curl -i http://localhost:8080/sync
```

Нужен как baseline для сравнения с остальными кейсами — показывает, что должно получаться в нормальном случае.

### [detached-background-context/](detached-background-context/)

Фоновая горутина, переживающая handler. Два эндпоинта рядом:

- `/bad` — goroutine держит `r.Context()`. Net/http отменяет его по возврату handler'а, любой downstream-вызов внутри фона падает с `context.Canceled`. В трассе span `background.bad` имеет status=Error.
- `/good` — `context.WithoutCancel(r.Context())` + собственный `context.WithTimeout`. Trace linkage сохраняется (активный span остаётся в values), cancel родителя развязан, есть свой дедлайн.

Разбор: [docs/internal-context-propagation/detached-background-context.md](../docs/internal-context-propagation/detached-background-context.md).

```bash
curl -i http://localhost:8080/bad
curl -i http://localhost:8080/good
```

### [worker-pool-context/](worker-pool-context/)

Очередь + фоновые воркеры. Вместо того чтобы класть в job целый `ctx` (что унаследовало бы cancel request'а и значения middleware), кладётся только `trace.SpanContext` родителя. В воркере ctx собирается с нуля: `context.Background()` + `trace.ContextWithSpanContext(...)` для восстановления линковки трассы + `context.WithTimeout` для своего дедлайна.

```bash
curl -i http://localhost:8080/enqueue
```

Это один из вариантов решения проблемы из `detached-background-context/bad` применительно к worker-pool паттерну. Подробное сравнение альтернатив (`WithoutCancel` в handler'е vs в воркере, SpanContext only, `trace.Link` с отдельной трассой, propagator через MapCarrier) — см. тот же [docs/-документ](../docs/internal-context-propagation/detached-background-context.md#4-когда-withoutcancel-не-нужен-worker-pool-как-граничный-случай).