package baseissues

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// =========================
// В выводе ты увидишь только то, что было записано до End().
// Интерфейс Span в Go после End() больше не должен обновляться.
// =========================

func main3() {
	tp := newTracerProvider("example-mutate-after-end")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = tp.Shutdown(ctx)
	}()

	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("examples/04-mutate-after-end")

	ctx := context.Background()
	_, span := tracer.Start(ctx, "mutate.after.end")

	span.SetAttributes(attribute.String("phase", "before-end"))
	span.SetStatus(codes.Ok, "all good so far")
	span.End()

	// ПЛОХО:
	// После End() эти изменения уже не должны попадать в завершенный span.
	span.SetAttributes(attribute.String("phase", "after-end"))
	span.AddEvent("late-event")
	span.SetStatus(codes.Error, "too late")

	log.Println("span already ended; late mutations should be ignored")
	time.Sleep(500 * time.Millisecond)
}

