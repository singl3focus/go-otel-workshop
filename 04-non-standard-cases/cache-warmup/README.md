# cache-warmup — fan-out на воркеры, span на воркера

CLI прогревает кэш: 200 ключей, 4 воркера, общий канал задач. Главная идея: **span на воркера, не на ключ**. Отдельные мисс-события идут через `span.AddEvent`, агрегаты — через атрибуты воркера. На root поднимаются итоговые `keys.warmed` / `keys.miss`.

В отличие от [../../03-internal-context-propagation/worker-pool-context/](../../03-internal-context-propagation/worker-pool-context/), здесь **ctx пробрасывается напрямую** — воркеры живут только внутри `run()`, ждать их завершения мы и так должны через `wg.Wait()`. Трюк с «передаём только `SpanContext`» здесь не нужен: нет cancel'а, от которого нужно отвязаться, нет переживания родителя.

## Как запустить

```bash
go run .
```

## Что увидеть в выводе

`stdouttrace` напечатает пять span'ов одной трассы:

```
cache.warmup            <- root
  cache.worker (id=0)
  cache.worker (id=1)
  cache.worker (id=2)
  cache.worker (id=3)
```

На root — `keys.total=200`, `workers=4`, плюс итоговые `keys.warmed` и `keys.miss`. На каждом `cache.worker` — `worker.id`, `worker.warmed`, `worker.miss`, плюс events `warm.miss` для каждого промаха (имитация: ~10% ключей).

Воркеры стартуют параллельно, их `cache.worker` span'ы перекрываются по времени — это видно по их `startTime`/`endTime`.

## Ключевые идеи

- **Span на воркера**, не на ключ. 1M ключей × span = взрыв трассировки.
- **Промахи как events** на span'е воркера. Дёшево и достаточно для разбора, что именно не нашлось.
- **Прямой проброс ctx безопасен**, потому что воркеры не переживают `run()` — `wg.Wait()` гарантирует, что мы не выйдем раньше них. Этот случай — антипод проблемы `detached-background-context`.

## Что дальше

- Worker-pool, в котором воркеры **переживают** handler — [../../03-internal-context-propagation/worker-pool-context/](../../03-internal-context-propagation/worker-pool-context/). Сравнение моделей.
- Bulk reindex (span на батч) — [../reindex/](../reindex/).
