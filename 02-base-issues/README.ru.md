# 02 · Типичные ошибки трассировки

[English](README.md) | [Русский](README.ru.md)

Английская версия — оригинал. Русский перевод поддерживается рядом с ней.

Четыре упражнения выделяют типичные ошибки SDK. Зависимости общие в `go.mod`
модуля; каждая подпапка — отдельный `package main`.

| Упражнение | Что исправить | Что проверяет тест |
| --- | --- | --- |
| [missing-span-end](missing-span-end/README.ru.md) | Завершить операцию | Один завершённый спан с атрибутом |
| [missing-tracerprovider-shutdown](missing-tracerprovider-shutdown/README.ru.md) | Выгрузить очередь и остановить провайдер | Один экспортированный спан и shutdown экспортёра |
| [mutate-after-end](mutate-after-end/README.ru.md) | Записать итог до End | `result=success` и статус Ok |
| [new-root-context](new-root-context/README.ru.md) | Сохранить входящий контекст | TraceID и связь с родителем |

## Порядок работы

Из этого каталога:

```sh
cd missing-span-end
go run .
go test .
go test -tags=solution .
```

`problem.go` демонстрирует ошибку, `exercise.go` содержит ваш ответ.
Обычный тест сначала падает. Исправьте упражнение и сравните с `solution.go`.
Тег `solution` исключает ваш ответ и запускает **те же тесты** на эталоне.
Он не превращает `problem.go` в исправленную демонстрацию.

Проверка эталонов всего модуля: `go test -tags=solution ./...`.
В тесте shutdown специальный экспортёр сохраняет выгруженные спаны после
остановки: стандартный in-memory экспортёр SDK очищает их при shutdown.
