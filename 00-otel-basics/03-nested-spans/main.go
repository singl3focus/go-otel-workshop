package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Span'ы связываются через context.Context:
//
//   tracer.Start(ctx, "parent")        кладет span в новый ctx;
//   передаем этот ctx в дочернюю функцию;
//   tracer.Start(childCtx, "child")    видит активный span и делает его parent.
//
// В выводе у parent'а и child'а одинаковый TraceID и разные SpanID;
// у child'а ParentSpanID совпадает со SpanID parent'а.
func main() {
	exp, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("00-otel-basics/03-nested-spans")

	ctx, parent := tracer.Start(context.Background(), "parent")
	doWork(ctx, tracer)
	parent.End()

	log.Println("done")
}

func doWork(ctx context.Context, tracer trace.Tracer) {
	_, span := tracer.Start(ctx, "child")
	defer span.End()
}
