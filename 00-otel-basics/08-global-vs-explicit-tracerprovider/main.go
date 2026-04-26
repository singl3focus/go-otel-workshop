package main

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// otel.SetTracerProvider задает global provider для otel.Tracer(...).
//
// Но TracerProvider можно использовать и явно через tp.Tracer(...). Это удобно
// в тестах и библиотеках: код не обязан зависеть от глобального состояния.
func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	globalTP := newTracerProvider("example-global-provider")
	defer func() { _ = globalTP.Shutdown(context.Background()) }()

	explicitTP := newTracerProvider("example-explicit-provider")
	defer func() { _ = explicitTP.Shutdown(context.Background()) }()

	otel.SetTracerProvider(globalTP)

	globalTracer := otel.Tracer("00-otel-basics/08/global")
	_, globalSpan := globalTracer.Start(context.Background(), "global.provider.span")
	log.Println("global.provider.span exported through otel.Tracer(...)")
	globalSpan.End()

	explicitTracer := explicitTP.Tracer("00-otel-basics/08/explicit")
	_, explicitSpan := explicitTracer.Start(context.Background(), "explicit.provider.span")
	log.Println("explicit.provider.span exported through explicitTP.Tracer(...)")
	explicitSpan.End()
}

func newTracerProvider(serviceName string) *sdktrace.TracerProvider {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)
}
