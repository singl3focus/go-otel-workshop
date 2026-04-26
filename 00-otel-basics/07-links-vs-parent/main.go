package main

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Parent и Link отвечают на разные вопросы.
//
// Parent-child — это одна причинная цепочка: child является частью операции parent.
// Link — это связь без иерархии: span связан с другим span'ом, но не становится его child.
func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("00-otel-basics/07-links-vs-parent")

	parentCtx, parent := tracer.Start(context.Background(), "request.parent")
	defer parent.End()

	_, child := tracer.Start(parentCtx, "request.child")
	log.Printf("parent-child trace_id=%s parent_span_id=%s child_span_id=%s",
		parent.SpanContext().TraceID(),
		parent.SpanContext().SpanID(),
		child.SpanContext().SpanID(),
	)
	child.End()

	sourceCtx, source := tracer.Start(context.Background(), "message.produced")
	source.End()

	_, linked := tracer.Start(context.Background(), "message.consumed",
		trace.WithLinks(trace.Link{
			SpanContext: trace.SpanContextFromContext(sourceCtx),
			Attributes: []attribute.KeyValue{
				attribute.String("link.reason", "async message"),
			},
		}),
	)
	log.Printf("linked span trace_id=%s linked_to_trace_id=%s linked_to_span_id=%s",
		linked.SpanContext().TraceID(),
		source.SpanContext().TraceID(),
		source.SpanContext().SpanID(),
	)
	defer linked.End()
}
