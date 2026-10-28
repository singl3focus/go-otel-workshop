# 03 · Контекст внутри процесса

[English](README.md) | [Русский](README.ru.md)

Английская версия — оригинал. Русский перевод поддерживается рядом с ней.

Запускаемые HTTP-демонстрации показывают, как `context.Context` переносит
родительский спан и отмену. Общий [helper](helper.go) экспортирует спаны в stdout.
Первые три запускайте по одному: они занимают порт 8080.
Четвёртый — конечная демонстрация жизненного цикла с временными loopback-портами.

| Пример | Запрос | Что наблюдать |
| --- | --- | --- |
| [sync-request-context](sync-request-context/README.ru.md) | `/sync` | Handler → service → repository в одной трассе |
| [detached-background-context](detached-background-context/README.ru.md) | `/bad`, `/good` | Отмена запроса и отдельный таймаут фоновой задачи |
| [worker-pool-context](worker-pool-context/README.ru.md) | `/enqueue` | SpanContext передаётся в задании очереди |
| [worker-pool-lifecycle](worker-pool-lifecycle/README.ru.md) | CLI: `-mode=drain`, `timeout`, `cancel` | Лимит приёма, реальная отмена HTTP и выход воркеров до shutdown OTel |

```sh
go run ./sync-request-context
```

В другом терминале (`curl.exe` в Windows PowerShell):

```sh
curl -i http://localhost:8080/sync
```

Batch exporter выводит спаны с небольшой задержкой. Сравнивайте `TraceID`
и родительский `SpanID`, а не порядок вывода. Первые три — небольшие учебные серверы:
путь `log.Fatal(ListenAndServe(...))` не реализует graceful shutdown.
Дайте экспорту завершиться перед остановкой; сервер с shutdown есть в модуле 01.

До [межсервисной передачи](../05-cross-service-propagation/README.ru.md)
прочитайте [разбор отмены контекста](../docs/internal-context-propagation/detached-background-context.ru.md).
