# 00 · Основы OpenTelemetry

[English](README.md) | [Русский](README.ru.md)

Английская версия — оригинал. Русский перевод поддерживается рядом с ней.

Десять запускаемых программ вводят по одной идее трассировки. В каждой подпапке
свой `package main`, зависимости общие через `go.mod` модуля.
Бэкенд не нужен: спаны выводятся через `stdouttrace`.

| Пример | Идея | На что смотреть |
| --- | --- | --- |
| [01-minimal](01-minimal/) | Экспортёр → провайдер → трейсер → спан | Один завершённый спан `hello` |
| [02-attributes-events-status](02-attributes-events-status/) | Атрибуты, события, ошибки, статус | `ok.span` и `bad.span`; событие ошибки и статус Error задаются отдельно |
| [03-nested-spans](03-nested-spans/) | Родитель и потомок через контекст | Общий TraceID, родитель потомка равен SpanID родителя |
| [04-resource](04-resource/) | Метаданные ресурса | `service.name=basics-resource-demo` |
| [05-batcher-vs-syncer](05-batcher-vs-syncer/) | Буферизация экспорта | Без shutdown `buffered.span` может потеряться |
| [06-propagator-mapcarrier](06-propagator-mapcarrier/) | Inject/Extract | `traceparent` в carrier и совпадающие TraceID |
| [07-links-vs-parent](07-links-vs-parent/) | Ссылка и иерархия | Потомок в той же трассе, связанный спан — в другой |
| [08-global-vs-explicit-tracerprovider](08-global-vs-explicit-tracerprovider/) | Владение провайдером | Разные service resources у явного и глобального провайдеров |
| [09-semconv-vs-manual](09-semconv-vs-manual/) | Соглашения об атрибутах | Ручные ошибки проходят экспорт, типизированные конструкторы их предотвращают |
| [10-sampling](10-sampling/README.ru.md) | Head sampling и решение родителя | Запись, флаги sampling и число экспортированных спанов |

Из этого каталога:

```sh
go run ./01-minimal
go run ./10-sampling
```

Либо зайдите в каталог примера и выполните `go run .`. Порядок stdout соответствует
завершению спанов, а не дереву трассы: родитель может появиться после потомка.
Успешные спаны часто имеют **Unset**, а не Ok, если статус явно не задан.

Попробуйте изменить атрибут или убрать родительский контекст и сравнить ID.
Дальше: [HTTP-инструментация](../01-local-observability-stack/README.ru.md)
или [типичные ошибки](../02-base-issues/README.ru.md).
