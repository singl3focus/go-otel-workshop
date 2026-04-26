package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Span'ы связываются через context.Context:
//
//	tracer.Start(ctx, "parent")        кладет span в новый ctx;
//	передаем этот ctx в дочернюю функцию;
//	tracer.Start(childCtx, "child")    видит активный span и делает его parent.
//
// В выводе у parent'а и child'а одинаковый TraceID и разные SpanID;
// у child'а ParentSpanID совпадает со SpanID parent'а.
func main() {
	exp, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("00-otel-basics/03-nested-spans")

	ctx, parentSpan := tracer.Start(context.Background(), "parent")
	defer parentSpan.End()

	doWork(ctx, tracer)

	log.Println("done")
}

func doWork(ctx context.Context, tracer trace.Tracer) {
	_, childSpan := tracer.Start(ctx, "child")
	defer childSpan.End()

	time.Sleep(100 * time.Millisecond)
}
