package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/singl3focus/go-otel-workshop/01-local-observability-stack/app-otelhttp/internal/service"
)

type Handlers struct {
	service *service.Service
}

func New(svc *service.Service) *Handlers {
	return &Handlers{service: svc}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.service.Health())
}

func (h *Handlers) Work(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.service.Work(r.Context(), r.URL.Query().Get("delay_ms")))
}

func (h *Handlers) Error(w http.ResponseWriter, r *http.Request) {
	status, payload := h.service.Error(r.Context(), r.URL.Query().Get("status"))
	writeJSON(w, status, payload)
}

func writeJSON(w http.ResponseWriter, status int, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}
