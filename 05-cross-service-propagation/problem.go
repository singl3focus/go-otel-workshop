package main

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
)

func main() {
	shutdown := setupTracing()
	defer shutdownTracing(shutdown)

	startInventory()
	serveGateway(checkoutHandler())
}

func checkoutHandler() http.HandlerFunc {
	tracer := otel.Tracer(tracerName + "/problem")

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "gateway.checkout")
		defer span.End()

		// ПЛОХО:
		// Обычный http.Client не создает client span и не инжектит traceparent.
		if err := reserveInventoryWithoutClientInstrumentation(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("checkout ok\n"))
	}
}

func reserveInventoryWithoutClientInstrumentation(ctx context.Context) error {
	req, _, err := newReserveRequest(ctx, inventoryBaseURL)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	return closeAndCheckReserveResponse(resp)
}
