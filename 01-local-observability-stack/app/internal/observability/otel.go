package observability

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const defaultServiceName = "local-observability-app"

type ShutdownFunc func(context.Context) error

func Setup(ctx context.Context, serviceName string, otlpEndpoint string) (ShutdownFunc, error) {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	exporter, err := newTraceExporter(ctx, otlpEndpoint)
	if err != nil {
		return nil, err
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

func newTraceExporter(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
	host, insecure, err := normalizeGRPCEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(host)}
	if insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	return exporter, nil
}

func normalizeGRPCEndpoint(raw string) (host string, insecure bool, err error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "localhost:4317", true, nil
	}

	if !strings.Contains(value, "://") {
		return value, true, nil
	}

	u, err := url.Parse(value)
	if err != nil {
		return "", false, fmt.Errorf("parse OTLP endpoint %q: %w", value, err)
	}
	if u.Host == "" {
		return "", false, fmt.Errorf("OTLP endpoint %q has empty host", value)
	}

	switch u.Scheme {
	case "http":
		return u.Host, true, nil
	case "https":
		return u.Host, false, nil
	default:
		return "", false, fmt.Errorf("unsupported OTLP endpoint scheme %q", u.Scheme)
	}
}
