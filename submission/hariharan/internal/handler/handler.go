package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/hariharandr/config-service/internal/domain"
	"github.com/hariharandr/config-service/internal/service"
)

// Handler holds dependencies for HTTP layer.
// Only depends on service — NOT repository.
type Handler struct {
	svc service.ConfigService
}

// New creates a new Handler.
func New(svc service.ConfigService) *Handler {
	return &Handler{svc: svc}
}

// Router sets up all routes.
func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/ping", h.ping)
	r.Get("/configs/{id}", h.getConfig)
	r.Post("/configs", h.upsertConfig)

	return r
}

// --- Handlers ---

// ping is used for liveness/readiness probes.
func (h *Handler) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// getConfig handles GET /configs/{id}
func (h *Handler) getConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cfg, err := h.svc.GetConfig(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, cfg)
}

// upsertConfig handles POST /configs
func (h *Handler) upsertConfig(w http.ResponseWriter, r *http.Request) {
	var req domain.UpsertRequest

	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON body",
		})
		return
	}

	cfg, err := h.svc.UpsertConfig(r.Context(), &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, cfg)
}

// --- Helpers ---

// handleError maps domain/service errors → HTTP responses
func (h *Handler) handleError(w http.ResponseWriter, err error) {
	var validationErr *domain.ValidationError

	switch {
	case errors.As(err, &validationErr):
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": validationErr.Error(),
		})

	case errors.Is(err, domain.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "config not found",
		})

	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
	}
}

// writeJSON is a helper to standardize JSON responses
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
