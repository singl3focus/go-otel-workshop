package middleware

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

func RequestTracing(tracer trace.Tracer) func(http.Handler) http.Handler {
	if tracer == nil {
		tracer = otel.Tracer("local-observability-stack/http")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parentCtx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			ctx, span := tracer.Start(parentCtx, r.Method+" request", trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r.WithContext(ctx))

			route := httpRoute(r)
			span.SetName(r.Method + " " + route)

			span.SetAttributes(
				semconv.HTTPRequestMethodKey.String(r.Method),
				semconv.URLPath(r.URL.Path),
				semconv.HTTPRoute(route),
				semconv.HTTPResponseStatusCode(rec.status),
			)
			if r.Host != "" {
				span.SetAttributes(semconv.ServerAddress(r.Host))
			}
			if rec.status >= 500 {
				span.SetStatus(codes.Error, http.StatusText(rec.status))
			}
		})
	}
}

func httpRoute(r *http.Request) string {
	pattern := strings.TrimSpace(r.Pattern)
	if pattern == "" {
		if r.URL.Path == "" {
			return "/"
		}
		return r.URL.Path
	}

	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) == 2 {
		return parts[1]
	}

	return pattern
}
