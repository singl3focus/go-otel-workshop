# migrations — трассировка прогона миграций

CLI без HTTP. Один root span на весь прогон, child span на каждую миграцию. Итоговые счётчики (`migrations.total`, `migrations.applied`) оседают на root, чтобы по одному span'у было видно, где прогон прервался, без агрегаций по детям.

## Как запустить

```bash
go run .
```

## Что увидеть в выводе

`stdouttrace` напечатает четыре span'а одной трассы:

```
migrations.run               <- root, без HTTP-родителя
  migration.apply (001_create_users.sql)
  migration.apply (002_add_email_index.sql)
  migration.apply (003_backfill_status.sql)
```

На root — атрибуты `db.system=postgres`, `migrations.total=3`, `migrations.applied=3`. На каждом child — `migration.name`. В happy-path `status=Unset` у всех. Если бы `apply` вернул ошибку, `migrations.applied` на root отразил бы сколько успело пройти до падения, а сам root и упавший child получили бы `RecordError` + `status=Error`.

## Ключевые идеи

- **Нет входящего ctx → root создаём сами** от `context.Background()`.
- **`tp.Shutdown` обязателен:** CLI короткоживущий, без shutdown'а батчер не успеет вытолкнуть последние span'ы.
- **Счётчики на root** `migrations.applied` обновляются и в success-, и в error-ветке — диагностика без обхода детей.

## Что дальше

- Long-running цикл с тикером и graceful shutdown — [../batch-reconciliation/](../batch-reconciliation/).
- Bulk-операции (span на батч, не на элемент) — [../reindex/](../reindex/).
