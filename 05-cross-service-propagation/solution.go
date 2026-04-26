//go:build solution

package main

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func ReserveInventorySolution(ctx context.Context, client *http.Client, inventoryBaseURL string) error {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, "checkout.reserve_inventory")
	defer span.End()

	req, url, err := newReserveRequest(ctx, inventoryBaseURL)
	if err != nil {
		return err
	}
	span.SetAttributes(attribute.String("inventory.url", url))

	instrumented := instrumentedClient(client)
	resp, err := instrumented.Do(req)
	if err != nil {
		return fmt.Errorf("reserve inventory: %w", err)
	}

	return closeAndCheckReserveResponse(resp)
}

func instrumentedClient(client *http.Client) *http.Client {
	base := defaultHTTPClient(client).Transport
	if base == nil {
		base = http.DefaultTransport
	}

	copy := *defaultHTTPClient(client)
	copy.Transport = otelhttp.NewTransport(base)
	return &copy
}
