package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type completionRecorder struct {
	*tracetest.SpanRecorder
	finished chan sdktrace.ReadOnlySpan
}

func (r *completionRecorder) OnEnd(span sdktrace.ReadOnlySpan) {
	r.SpanRecorder.OnEnd(span)
	if strings.HasPrefix(span.Name(), "background.") {
		r.finished <- span
	}
}

// EN: Return the response, then cancel the request; detached work must still finish.
// RU: Сначала возвращаем ответ, затем отменяем запрос: отделённый фон должен завершиться.
func TestBackgroundLifetimeAfterResponse(t *testing.T) {
	for _, mode := range []string{"bad", "good"} {
		t.Run(mode, func(t *testing.T) {
			recorder := &completionRecorder{SpanRecorder: tracetest.NewSpanRecorder(), finished: make(chan sdktrace.ReadOnlySpan, 1)}
			tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
			t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
			tracer := tp.Tracer("test")
			parentCtx, parent := tracer.Start(context.Background(), "http.parent")
			requestCtx, cancel := context.WithCancel(parentCtx)
			defer cancel()
			request := httptest.NewRequest(http.MethodGet, "/"+mode, nil).WithContext(requestCtx)
			response := httptest.NewRecorder()
			app := &app{tracer: tracer}
			if mode == "good" {
				app.handleGood(response, request)
			} else {
				app.handleBad(response, request)
			}
			parent.End()
			cancel()
			if response.Code != http.StatusAccepted {
				t.Fatalf("response / ответ: %d", response.Code)
			}
			select {
			case span := <-recorder.finished:
				if span.SpanContext().TraceID() != parent.SpanContext().TraceID() || span.Parent().SpanID() != parent.SpanContext().SpanID() {
					t.Fatal("background lost parent / фон потерял родителя")
				}
				want := codes.Unset
				if mode == "bad" {
					want = codes.Error
				}
				if span.Status().Code != want {
					t.Fatalf("background status / статус фона: got %v want %v", span.Status().Code, want)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("background did not finish / фон не завершился")
			}
		})
	}
}
