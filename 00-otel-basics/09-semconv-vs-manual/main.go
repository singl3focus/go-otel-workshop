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

// EN: Semantic conventions define shared attribute names and types.
// Backends and instrumentation can use those conventions for consistent
// queries and views; backend features vary.
//
// go.opentelemetry.io/otel/semconv/vX.Y.Z provides generated keys and typed
// constructors. The import path selects a convention version; it is not
// the SDK version. A resource schema URL identifies that version.
//
// Handwritten strings can contain silent typos, obsolete names, or wrong
// types. Different schema URLs produce an error when resources are merged;
// that is separate from normal same-key attribute precedence.
//
// RU: Semantic conventions задают общие имена и типы атрибутов.
// Это помогает инструментаторам и бэкендам согласованно строить запросы
// и представления; доступные функции зависят от бэкенда.
//
// Пакет go.opentelemetry.io/otel/semconv/vX.Y.Z содержит сгенерированные ключи
// и типизированные конструкторы. Версия в импорте относится к соглашениям,
// а не к SDK. SchemaURL ресурса указывает версию схемы.
//
// Ручные строки допускают незаметные опечатки, устаревшие имена и неверные
// типы. Разные непустые SchemaURL дают ошибку при resource.Merge.
// Это отдельный механизм от приоритета значений атрибутов с одинаковым ключом.
func main() {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("00-otel-basics/09-semconv-vs-manual")

	// EN: Variant 1: handwritten strings.
	// Export still succeeds, despite a typo, obsolete attribute names,
	// and a string value where the convention requires an integer.
	//
	// RU: Вариант 1. Ручные строки.
	//
	// Работает, экспорт пройдет. Но:
	//   - "htttp.method" — опечатка, не словит ни компилятор, ни линтер;
	//   - "http.status_code" — устаревшее имя (актуальное — http.response.status_code);
	//   - status_code как строка — типичная ошибка, дашборды по int сломаются.
	_, manualSpan := tracer.Start(context.Background(), "GET /users/:id [manual]")
	manualSpan.SetAttributes(
		attribute.String("htttp.method", "GET"),        // EN: typo. RU: опечатка.
		attribute.String("http.status_code", "200"),    // EN: obsolete name and wrong type. RU: устаревшее имя и неверный тип.
		attribute.String("http.target", "/users/42"),   // EN: obsolete; now url.path. RU: устарело, теперь url.path.
		attribute.String("net.peer.name", "users-api"), // EN: obsolete; now server.address. RU: устарело, теперь server.address.
	)
	manualSpan.End()

	// EN: Variant 2: generated semantic conventions.
	// Names and constructor argument types are checked by the compiler.
	// The versioned import makes convention upgrades explicit.
	//
	// RU: Вариант 2. Через semconv.
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
