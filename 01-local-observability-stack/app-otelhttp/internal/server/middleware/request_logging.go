package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// RequestLogging пишет slog по каждому запросу и подмешивает trace_id/span_id
// активного span'а из ctx. Сам span создаёт outer-обёртка — в этом app это
// otelhttp.NewHandler, который проставляет server-span до того, как запрос
// дойдёт до этого middleware.
func RequestLogging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)
			spanContext := trace.SpanContextFromContext(r.Context())
			traceID := ""
			spanID := ""
			if spanContext.IsValid() {
				traceID = spanContext.TraceID().String()
				spanID = spanContext.SpanID().String()
			}

			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"status", rec.status,
				"duration_ms", time.Since(started).Milliseconds(),
				"trace_id", traceID,
				"span_id", spanID,
				"traceparent", r.Header.Get("traceparent"),
			)
		})
	}
}
