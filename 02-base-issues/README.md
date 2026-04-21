# 02. Base issues - типичные ошибки OpenTelemetry

Набор антипаттернов Go + OTel SDK: каждый кейс оформлен как отдельное упражнение, которое нужно пройти самостоятельно.

## Формат упражнения

Каждая подпапка - самостоятельный `package main` со следующей структурой:

- `problem.go` - запускаемая демонстрация антипаттерна (`go run .`);
- `exercise.go` - функция-стаб, которую нужно починить;
- `exercise_test.go` - тест на `tracetest.SpanRecorder`, который становится зеленым после корректного исправления (`go test .`);
- `solution.go` - эталонное решение под build-тегом `solution`, по умолчанию не компилируется;
- `README.md` - условие задачи.

## Как проходить

1. Зайти в подпапку нужного кейса.
2. Запустить `go run .` и посмотреть на поведение сломанного кода.
3. Открыть `exercise.go`, починить функцию.
4. Убедиться, что `go test .` проходит.
5. Опционально сверить свой вариант с `solution.go`.

## Список упражнений

- [missing-span-end/](missing-span-end/) - забыли `span.End()`.
- [missing-tracerprovider-shutdown/](missing-tracerprovider-shutdown/) - выход процесса без `tp.Shutdown()`.
- [mutate-after-end/](mutate-after-end/) - мутация span после `End()`.
- [new-root-context/](new-root-context/) - потеря входящего trace context через `context.Background()`.
