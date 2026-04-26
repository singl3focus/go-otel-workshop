package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// Зачем нужен semconv.
//
// OpenTelemetry-спецификация фиксирует имена и типы атрибутов, которые
// бэкенды (Jaeger/Tempo/Datadog/Honeycomb) умеют распознавать "из коробки":
// строят по ним готовые дашборды HTTP/DB/RPC, считают RED-метрики, делают
// фасеты в UI. Имя атрибута — это контракт между приложением и бэкендом.
//
// Пакет go.opentelemetry.io/otel/semconv/vX.Y.Z — кодген этого контракта:
//
//   - имена атрибутов как константы или конструкторы (semconv.HTTPRequestMethod);
//   - правильные типы (status_code → int, не string);
//   - версия спеки прошита в import path: апгрейд = осознанная замена импорта;
//   - SchemaURL для resource.Merge, чтобы версии не молча затирали друг друга.
//
// Что ломается без semconv:
//
//  1. Опечатки тихие. attribute.String("htttp.method", ...) — не ошибка
//     компиляции, но в бэкенде атрибут не подхватится.
//  2. Несовпадение с авто-инструментациями. otelhttp/otelgrpc/otelsql пишут
//     по semconv. Если твой ручной span назовёт то же поле иначе, в трассе
//     будут два разных атрибута на одно и то же.
//  3. Неправильные типы. http.response.status_code в спеке — int. Если
//     записать строкой, бэкенд может его не распарсить как число и
//     фильтр "5xx" сломается.
//  4. Расхождение версий спеки между сервисами. Без SchemaURL у ресурса
//     слияние через resource.Merge молча теряет конфликтующие атрибуты.
func main() {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("00-otel-basics/09-semconv-vs-manual")

	// Вариант 1. Ручные строки.
	//
	// Работает, экспорт пройдет. Но:
	//   - "htttp.method" — опечатка, не словит ни компилятор, ни линтер;
	//   - "http.status_code" — устаревшее имя (актуальное — http.response.status_code);
	//   - status_code как строка — типичная ошибка, дашборды по int сломаются.
	_, manualSpan := tracer.Start(context.Background(), "GET /users/:id [manual]")
	manualSpan.SetAttributes(
		attribute.String("htttp.method", "GET"),        // опечатка
		attribute.String("http.status_code", "200"),    // устаревшее имя + неверный тип
		attribute.String("http.target", "/users/42"),   // устаревшее имя (теперь url.path)
		attribute.String("net.peer.name", "users-api"), // устаревшее имя (теперь server.address)
	)
	manualSpan.End()

	// Вариант 2. Через semconv.
	//
	// Имена и типы заданы константами/конструкторами. Опечатка → ошибка
	// компиляции. Тип status_code зафиксирован как int. Версия спеки
	// прошита в импорте — апгрейд явный.
	_, semconvSpan := tracer.Start(context.Background(), "GET /users/:id [semconv]")
	semconvSpan.SetAttributes(
		semconv.HTTPRequestMethodGet,
		semconv.HTTPResponseStatusCode(200),
		semconv.URLPath("/users/42"),
		semconv.URLScheme("https"),
		semconv.ServerAddress("users-api"),
		semconv.ServerPort(443),
		semconv.HTTPRoute("/users/:id"),
	)
	semconvSpan.End()

	log.Println("done")
}
