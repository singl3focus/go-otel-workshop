# Контекст в пуле воркеров

[English](README.md) | [Русский](README.ru.md)

Английская версия — оригинал. Русский перевод поддерживается рядом с ней.

Два воркера читают задания из канала в памяти. Задание содержит только
`trace.SpanContext` и ID заказа, а не весь контекст запроса.

```sh
go run .
```

В другом терминале (`curl.exe` в Windows PowerShell):

```sh
curl -i http://localhost:8080/enqueue
```

Ответ — 202. В трассе будут `http.enqueue`, дочерний `handler.enqueue`
и `worker.process`, чей родитель — `handler.enqueue`. У всех общий TraceID.
У `worker.process` есть `worker.id` и `order.id`, статус остаётся Unset.
Планировщик может запустить воркер до или после возврата handler.

Воркер восстанавливает контекст через `trace.ContextWithSpanContext` и задаёт
свой таймаут. В примере используется `time.Sleep`, который таймаут не учитывает;
реальные I/O-вызовы должны получать контекст и реагировать на его отмену.
Для отдельной трассы со ссылкой на производителя см.
[links и родители](../../00-otel-basics/07-links-vs-parent/).

Отменяемый I/O, отказ при перегрузке и остановку разбирает
[worker-pool-lifecycle](../worker-pool-lifecycle/README.ru.md).
