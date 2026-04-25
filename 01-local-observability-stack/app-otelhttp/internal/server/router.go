package server

import (
	"log/slog"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/singl3focus/go-otel-workshop/01-local-observability-stack/app-otelhttp/internal/server/handlers"
	"github.com/singl3focus/go-otel-workshop/01-local-observability-stack/app-otelhttp/internal/server/middleware"
	"github.com/singl3focus/go-otel-workshop/01-local-observability-stack/app-otelhttp/internal/service"
)

// NewRouter собирает HTTP-обработчик с auto-инструментацией через otelhttp.
//
// Сравни с app-manual: там middleware/request_tracing.go руками тащит
// propagator.Extract, открывает server-span, проставляет semconv-атрибуты
// и закрывает span. Здесь весь этот код заменяет otelhttp.NewHandler:
//   - propagator.Extract из заголовков — внутри otelhttp;
//   - server-span со SpanKind=Server — создаётся otelhttp;
//   - http.request.method / http.response.status_code / url.path / server.address —
//     проставляет otelhttp по semconv.
//
// Имя span формируется через WithSpanNameFormatter. Тонкость: otelhttp вызывает
// formatter дважды — до dispatch (когда r.Pattern ещё пуст; имя берётся как
// operation) и после dispatch (когда ServeMux уже проставил r.Pattern). Второй
// вызов и даёт финальное имя вида "GET /work".
//
// RequestLogging оставлен как в app-manual: ему достаточно прочитать активный
// span из ctx — кто его создал, otelhttp или ручной middleware, не важно.
func NewRouter(appName string, logger *slog.Logger) http.Handler {
	svc := service.New(appName)
	h := handlers.New(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /work", h.Work)
	mux.HandleFunc("GET /error", h.Error)

	handler := middleware.RequestLogging(logger)(mux)

	return otelhttp.NewHandler(handler, "http.server",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			if r.Pattern != "" {
				return r.Pattern
			}
			return operation
		}),
	)
}
