# 01 · HTTP-инструментация и экспорт OTLP

[English](README.md) | [Русский](README.ru.md)

Английская версия — оригинал. Русский перевод поддерживается рядом с ней.

Два HTTP-приложения используют общий [пакет observability](observability/otel.go).
Небольшой бизнес-слой намеренно продублирован для сравнения двух вариантов.
Общий пакет настраивает batch TracerProvider, resource, OTLP/gRPC-экспортёр
и propagators W3C TraceContext/Baggage.

| Задача | [app-manual](app-manual/README.ru.md) | [app-otelhttp](app-otelhttp/README.ru.md) |
| --- | --- | --- |
| Извлечение входящего контекста | `propagator.Extract` | `otelhttp.NewHandler` |
| Создание и завершение серверного спана | Ручной middleware | `otelhttp` |
| HTTP-атрибуты и статус при 5xx | Явные атрибуты и статус | Инструментация библиотеки |
| Имя маршрута | Свой formatter | `WithSpanNameFormatter` и `r.Pattern` |
| Связь логов с трассой | Чтение активного спана из контекста | Тот же подход |

В модуле есть закреплённый [стенд Collector + Jaeger](stack/README.ru.md):
запуск бэкендов, просмотр трасс и дополнительные интеграционные тесты.
Живая совместимость ещё не проверена: Compose и код тестов проверены без запуска
сервисов. Можно использовать и существующий receiver. Порт **4317** предназначен для gRPC;
4318 обычно принимает OTLP/HTTP и не взаимозаменяем с этим экспортёром.

## Запуск с receiver на localhost

Из этого каталога в Bash:

```sh
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
go run ./app-manual/cmd/server
```

В PowerShell:

```powershell
$env:OTEL_EXPORTER_OTLP_ENDPOINT = 'http://localhost:4317'
go run ./app-manual/cmd/server
```

В другом терминале (`curl.exe` в Windows PowerShell):

```sh
curl -i 'http://localhost:8080/work?delay_ms=200'
curl -i 'http://localhost:8080/error?status=500'
```

Остановите первое приложение, запустите `go run ./app-otelhttp/cmd/server`.
Сравните дочерние спаны `service.work` и поля `trace_id`/`span_id` в логах.
Для просмотра экспортированных спанов нужен реальный receiver/бэкенд.
Один HTTP 200 не подтверждает доставку; ошибки экспорта могут появиться позже.

| Переменная окружения | Значение по умолчанию |
| --- | --- |
| `APP_NAME` | `local-observability-app` |
| `HTTP_PORT` | `8080` |
| `LOG_LEVEL` | `info` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4317` |

`http://` выбирает gRPC без TLS, `https://` включает TLS. Адрес без схемы
считается небезопасным соединением. Эти настройки предназначены для локального
учебного окружения. Если спаны не приходят, см. [диагностику](../docs/troubleshooting.ru.md).
