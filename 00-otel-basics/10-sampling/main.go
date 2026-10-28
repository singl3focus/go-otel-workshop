package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// EN: A head sampler decides at Start, before the operation's outcome is known.
// RU: Head sampler решает при Start, когда итог операции ещё неизвестен.
func main() {
	cases := []struct {
		name    string
		sampler sdktrace.Sampler
		parent  context.Context
	}{
		{"always/root", sdktrace.AlwaysSample(), context.Background()},
		{"never/root", sdktrace.NeverSample(), context.Background()},
		{"parent-based/root", sdktrace.ParentBased(sdktrace.NeverSample()), context.Background()},
		{"parent-based/sampled-parent", sdktrace.ParentBased(sdktrace.NeverSample()), remoteParent(true)},
		{"parent-based/unsampled-parent", sdktrace.ParentBased(sdktrace.NeverSample()), remoteParent(false)},
	}
	for _, c := range cases {
		result, err := observe(c.parent, c.sampler)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%-30s recording=%t sampled=%t exported=%d\n",
			c.name, result.recording, result.spanContext.IsSampled(), result.exported)
	}
}

type observation struct {
	recording   bool
	spanContext trace.SpanContext
	exported    int
}

func observe(parent context.Context, sampler sdktrace.Sampler) (observation, error) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp), sdktrace.WithSampler(sampler))
	_, span := tp.Tracer("00-otel-basics/10-sampling").Start(parent, "sampling.operation")
	result := observation{recording: span.IsRecording(), spanContext: span.SpanContext()}

	// EN: A late error cannot change the head sampling decision.
	// RU: Поздняя ошибка не меняет решение head sampling.
	span.RecordError(errors.New("simulated late failure"))
	span.SetStatus(codes.Error, "simulated late failure")
	span.End()

	// EN: Read before Shutdown: InMemoryExporter clears its storage on shutdown.
	// RU: Читаем до Shutdown: InMemoryExporter очищает хранилище при остановке.
	result.exported = len(exp.GetSpans())
	return result, tp.Shutdown(context.Background())
}

func remoteParent(sampled bool) context.Context {
	var flags trace.TraceFlags
	if sampled {
		flags = trace.FlagsSampled
	}
	parent := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1},
		SpanID:     trace.SpanID{2},
		TraceFlags: flags,
		Remote:     true,
	})
	return trace.ContextWithRemoteSpanContext(context.Background(), parent)
}
