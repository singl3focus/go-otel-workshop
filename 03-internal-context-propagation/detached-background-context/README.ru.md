# Отделённая фоновая работа

[English](README.md) | [Русский](README.ru.md)

Английская версия — оригинал. Русский перевод поддерживается рядом с ней.

Запустите сервер, затем отправьте два запроса из другого терминала:

```sh
go run .
```

```sh
curl -i http://localhost:8080/bad
curl -i http://localhost:8080/good
```

В Windows PowerShell используйте `curl.exe`. Оба пути сразу возвращают 202.
После короткой задержки горутина создаёт спан и имитирует downstream-вызов,
который учитывает отмену контекста.

| Путь | Контекст фоновой работы | Результат |
| --- | --- | --- |
| `/bad` | Исходный контекст запроса | `background.bad` записывает отмену и Error |
| `/good` | `WithoutCancel` и таймаут две секунды | `background.good` завершается с Unset и `context.mode=detached` |

Оба фоновых спана сохраняют TraceID своего HTTP-родителя. Отмена влияет
на имитируемую работу, а не удаляет родительский SpanContext.
В хорошем варианте `cancel` вызывает сама горутина. Если поставить
`defer cancel()` в handler, новый контекст отменится сразу после ответа.

См. [подробный разбор](../../docs/internal-context-propagation/detached-background-context.ru.md)
и [вариант с очередью](../worker-pool-context/README.ru.md).
