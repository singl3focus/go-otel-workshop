# app-manual · Ручная инструментация

[English](README.md) | [Русский](README.ru.md)

Английская версия — оригинал. Русский перевод поддерживается рядом с ней.

[Middleware трассировки](internal/server/middleware/request_tracing.go) явно извлекает родителя, создаёт серверный спан, записывает HTTP-атрибуты и завершает его.

Сравните с [app-otelhttp](../app-otelhttp/README.ru.md). Оба варианта используют общую
[настройку SDK](../observability/otel.go). Конфигурация и требования к receiver
описаны в [руководстве модуля](../README.ru.md).

## Запуск из этого каталога

```sh
go run ./cmd/server
```

При установленном Make можно выполнить эквивалентный `make run`.

| Запрос | Поведение |
| --- | --- |
| `GET /health` | JSON с состоянием приложения |
| `GET /work?delay_ms=120` | Имитация работы и спан `service.work` |
| `GET /error?status=500` | Заданный ответ; при 5xx операция помечается Error |

Сервер слушает `:8080`, если порт не переопределён через `HTTP_PORT`.
При необходимости задайте `APP_NAME`, `LOG_LEVEL` и `OTEL_EXPORTER_OTLP_ENDPOINT`.
Receiver по умолчанию — `http://localhost:4317`; см. [локальный стенд](../stack/README.ru.md).
