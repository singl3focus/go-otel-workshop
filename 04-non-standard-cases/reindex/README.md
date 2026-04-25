# reindex — bulk reindex, span на батч

CLI прогоняет 500 «документов» батчами по 100 в имитируемый поисковый движок. Главная идея: **span на батч, не на документ**. При 10M документов span-per-doc ложит экспортер и стоимость хранения. Уровень батча — естественная граница observability: latency bulk-запроса, retries, размер. Отдельные документы — через `span.AddEvent`.

## Как запустить

```bash
go run .
```

## Что увидеть в выводе

`stdouttrace` напечатает шесть span'ов одной трассы:

```
reindex.run                  <- root
  reindex.batch (offset=0,   size=100)
  reindex.batch (offset=100, size=100)
  reindex.batch (offset=200, size=100)
  reindex.batch (offset=300, size=100)
  reindex.batch (offset=400, size=100)
```

На root — `index.name=orders_v3`, `source.total=500`, `batch.size=100`, плюс итоговые `reindex.indexed` и `reindex.failed`. На каждом `reindex.batch` — `batch.offset`, `batch.size`, `batch.indexed`, `batch.failed`.

Имитация валидации роняет каждый 137-й документ. На таких документах span-батча содержит **events** `doc.failed` с атрибутами `doc.id` и `reason=validation`. Сам span — `status=Unset` (батч прошёл, отдельные документы — отдельная история).

## Ключевые идеи

- **Span на батч, не на документ.** Иначе spans-storm.
- **Документ-failures как events**, а не как отдельные span'ы. Events дешевле и логически правильнее: они привязаны к моменту времени внутри батча.
- **Агрегаты — атрибутами на root.** `reindex.indexed` и `reindex.failed` обновляются и в success-, и в error-ветке (см. defer на root). По одному span'у видно итог прогона.

## Что дальше

- Миграции (один root, span на миграцию) — [../migrations/](../migrations/).
- Fan-out на воркеры — [../cache-warmup/](../cache-warmup/).
