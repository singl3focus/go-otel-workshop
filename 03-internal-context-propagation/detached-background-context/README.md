# detached-background-context — фон, переживающий handler

Два эндпоинта рядом, чтобы сравнить поведение. Полный разбор (что именно делает net/http с request ctx, чем `context.WithoutCancel` отличается от `context.Background`, какие альтернативы есть для worker-pool) — в [docs/internal-context-propagation/detached-background-context.md](../../docs/internal-context-propagation/detached-background-context.md). Здесь — только запуск и что увидеть в трассе.

## Как запустить

```bash
go run .
```

В другом терминале:

```bash
curl -i http://localhost:8080/bad
curl -i http://localhost:8080/good
```

## Что увидеть в выводе

### `/bad`

Горутина уносит `r.Context()`. К моменту, когда фон делает `simulateDownstreamCall`, net/http уже отменил ctx (handler вернулся). В `stdouttrace`:

- span `http.bad` уезжает первым, сразу после ответа клиенту;
- span `background.bad` уезжает следом — у него `status=Error`, в events лежит `RecordError("context canceled")`;
- TraceID у обоих общий — линковка трассы сама по себе сохраняется, ломается именно downstream-вызов.

### `/good`

Перед запуском горутины в handler'е делается `bg := context.WithoutCancel(r.Context())` + `context.WithTimeout(bg, 2s)`. Cancel родителя на фон не распространяется, values (включая активный span) сохраняются.

- span `http.good` уезжает первым;
- span `background.good` — `status=Ok`, атрибут `context.mode=detached`;
- TraceID общий, downstream отрабатывает штатно.

## Что дальше

- Альтернатива для очередей и воркер-пулов (передавать только `SpanContext`, ctx собирать с нуля) — [../worker-pool-context/](../worker-pool-context/).
- Полный разбор и альтернативы — [docs/internal-context-propagation/detached-background-context.md](../../docs/internal-context-propagation/detached-background-context.md).
