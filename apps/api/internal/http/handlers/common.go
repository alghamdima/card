package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	appmiddleware "cards-api/internal/http/middleware"
	"cards-api/internal/http/response"
	"cards-api/internal/service"
)

const (
	// maxSmallBody bounds JSON payloads that carry no artwork (login, card submissions).
	maxSmallBody = 64 << 10
	// maxCampaignBody bounds campaign payloads, which embed base64 artwork; keep in sync with the proxy's client_max_body_size.
	maxCampaignBody = 30 << 20
)

// decodeJSON reads a size-limited JSON body into dst. On failure it writes the
// error response itself and returns false.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any, limit int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			response.Error(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "Request body is too large")
			return false
		}
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid JSON payload")
		return false
	}
	return true
}

// writeServiceError maps service-layer errors to HTTP responses. Unexpected
// errors are logged with the request ID and never leaked to the client.
func writeServiceError(w http.ResponseWriter, r *http.Request, err error, failCode, failMessage string) {
	var validation *service.ValidationError
	switch {
	case errors.Is(err, service.ErrCampaignNotFound):
		response.Error(w, http.StatusNotFound, "CAMPAIGN_NOT_FOUND", "Campaign not found or inactive")
	case errors.Is(err, service.ErrCampaignAlreadyExists):
		response.Error(w, http.StatusConflict, "CAMPAIGN_EXISTS", "A campaign with this link already exists")
	case errors.As(err, &validation):
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", validation.Message)
	default:
		log.Printf("[%s] %s %s: %v", appmiddleware.RequestIDFrom(r.Context()), r.Method, r.URL.Path, err)
		response.Error(w, http.StatusInternalServerError, failCode, failMessage)
	}
}

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health reports readiness: the API is only healthy when the database answers.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		log.Printf("health check failed: %v", err)
		response.Error(w, http.StatusServiceUnavailable, "UNHEALTHY", "Database unavailable")
		return
	}

	response.OK(w, map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "cards-api",
	})
}
