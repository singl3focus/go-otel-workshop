package main

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// ReserveInventory открывает business span и вызывает inventory-сервис.
//
// TODO: downstream HTTP handler должен попасть в тот же trace, что и caller.
// Добейся того, чтобы exercise_test.go прошел зеленым.
func ReserveInventory(ctx context.Context, client *http.Client, inventoryBaseURL string) error {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, "checkout.reserve_inventory")
	defer span.End()

	req, url, err := newReserveRequest(ctx, inventoryBaseURL)
	if err != nil {
		return err
	}
	span.SetAttributes(attribute.String("inventory.url", url))

	resp, err := defaultHTTPClient(client).Do(req)
	if err != nil {
		return fmt.Errorf("reserve inventory: %w", err)
	}

	return closeAndCheckReserveResponse(resp)
}
