package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// EN: A resource describes the entity producing telemetry: service.name,
// service.version, host, container, and Kubernetes metadata.
// Backends can group spans by service.name.
//
// NewSchemaless omits a schema URL, avoiding a schema-URL conflict when
// merging with another resource. It does not migrate attribute names.
//
// RU: Resource описывает источник телеметрии: service.name,
// service.version, хост, контейнер и метаданные Kubernetes.
// Бэкенд может группировать спаны по service.name.
//
// NewSchemaless не задаёт SchemaURL и позволяет избежать конфликта URL схем
// при resource.Merge. Имена атрибутов он не преобразует.
func main() {
	exp, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())

	res := resource.NewSchemaless(
		semconv.ServiceName("basics-resource-demo"),
		semconv.ServiceVersion("0.1.0"),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithResource(res),
	)
	defer func() { _ = tp.Shutdown(context.Background()) }()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("00-otel-basics/04-resource")
	_, span := tracer.Start(context.Background(), "with.resource")
	span.End()

	log.Println("done")
}
