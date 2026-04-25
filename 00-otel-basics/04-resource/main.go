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

// Resource — атрибуты, общие для ВСЕХ span'ов из этого процесса:
// service.name, service.version, host, container, k8s-метки. Бэкенды по
// service.name группируют сервисы, без него span'ы попадают в "unknown_service".
//
// NewSchemaless удобно использовать когда не хочешь думать про SchemaURL —
// такой ресурс корректно сливается через resource.Merge с любым другим.
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
