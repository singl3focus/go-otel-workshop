package main

import (
	"context"
	"log"
	"net/http"
	"time"

	baseissues "github.com/singl3focus/go-otel-workshop/02-base-issues"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// =========================
// При запросе на /orders у тебя будет входящий span от otelhttp,
// но service.create_order станет новым root span,
// потому что бизнес-логика перетерла входной ctx через context.Background().
// Контекст должен проходить через вызовы вниз по стеку, а otelhttp.NewHandler как раз создает span для входящего запроса.
// =========================

func main() {
	tp := baseissues.NewTracerProvider("example-new-root-context")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = tp.Shutdown(ctx)
	}()

	otel.SetTracerProvider(tp)

	svc := &Service{
		tracer: otel.Tracer("02-base-issues/new-root-context"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/orders", svc.HandleCreateOrder)

	handler := otelhttp.NewHandler(mux, "http.server")

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

type Service struct {
	tracer trace.Tracer
}

func (s *Service) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.CreateOrder(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte("order created\n"))
}

func (s *Service) CreateOrder(ctx context.Context) error {
	// ПЛОХО:
	// Мы теряем trace от входящего HTTP-запроса и создаем новый root span.
	ctx, span := s.tracer.Start(context.Background(), "service.create_order")
	defer span.End()

	span.SetAttributes(attribute.String("order.id", "123"))

	return s.saveToRepo(ctx)
}

func (s *Service) saveToRepo(ctx context.Context) error {
	_, span := s.tracer.Start(ctx, "repo.save_order")
	defer span.End()

	time.Sleep(50 * time.Millisecond)
	return nil
}
