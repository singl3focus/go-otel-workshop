package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

const (
	gatewayAddr   = ":8080"
	inventoryAddr = ":8081"

	inventoryBaseURL     = "http://localhost:8081"
	inventoryReservePath = "/reserve"
	tracerName           = "05-cross-service-propagation"
)

func NewTracerProvider(serviceName string) *sdktrace.TracerProvider {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)
}

func setupTracing() func(context.Context) error {
	tp := NewTracerProvider("example-cross-service-propagation")
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown
}

func shutdownTracing(shutdown func(context.Context) error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = shutdown(ctx)
}

func serveInventory(listener net.Listener) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST "+inventoryReservePath, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("inventory got traceparent=%q", r.Header.Get("traceparent"))
		w.WriteHeader(http.StatusNoContent)
	})

	serve(listener, "inventory", otelhttp.NewHandler(mux, "inventory.http"))
}

func serveGateway(handler http.HandlerFunc) {
	listener, err := net.Listen("tcp", gatewayAddr)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /checkout", handler)

	serve(listener, "gateway", otelhttp.NewHandler(mux, "gateway.http"))
}

func startInventory() {
	listener, err := net.Listen("tcp", inventoryAddr)
	if err != nil {
		log.Fatal(err)
	}
	go serveInventory(listener)
}

func serve(listener net.Listener, name string, handler http.Handler) {
	log.Printf("%s listening on %s", name, listener.Addr())
	if err := http.Serve(listener, handler); err != nil {
		log.Fatal(err)
	}
}

func defaultHTTPClient(client *http.Client) *http.Client {
	if client != nil {
		return client
	}
	return http.DefaultClient
}

func newReserveRequest(ctx context.Context, inventoryBaseURL string) (*http.Request, string, error) {
	url := reserveURL(inventoryBaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}

	return req, url, nil
}

func closeAndCheckReserveResponse(resp *http.Response) error {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("reserve inventory: status %d", resp.StatusCode)
	}

	return nil
}

func reserveURL(inventoryBaseURL string) string {
	return strings.TrimRight(inventoryBaseURL, "/") + inventoryReservePath
}
