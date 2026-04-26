package main

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Propagator переносит trace context через key/value carrier
// (carrier — это объект, куда propagator умеет записать и откуда умеет прочитать
// ключи propagation; в HTTP таким carrier'ом являются headers).
//
// HTTP-инструментация делает это через заголовки автоматически, но механика та же:
//   - Inject читает активный span из ctx и пишет traceparent в carrier;
//   - Extract читает traceparent из carrier и кладет remote parent в новый ctx;
//   - tracer.Start(extractedCtx, ...) создает child span в той же трассе.
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
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := otel.Tracer("00-otel-basics/06-propagator-mapcarrier")

	producerCtx, producerSpan := tracer.Start(context.Background(), "producer.send")
	defer producerSpan.End()

	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(producerCtx, carrier)

	log.Printf("carrier traceparent=%s", carrier.Get("traceparent"))
	log.Printf("producer trace_id=%s span_id=%s",
		producerSpan.SpanContext().TraceID(),
		producerSpan.SpanContext().SpanID(),
	)

	consumerCtx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)
	_, consumerSpan := tracer.Start(consumerCtx, "consumer.receive")
	defer consumerSpan.End()

	log.Printf("consumer trace_id=%s expected_parent_span_id=%s",
		consumerSpan.SpanContext().TraceID(),
		producerSpan.SpanContext().SpanID(),
	)
}
