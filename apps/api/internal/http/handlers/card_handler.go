package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"cards-api/internal/service"
)

type CardHandler struct {
	cardService *service.CardService
}

func NewCardHandler(cardService *service.CardService) *CardHandler {
	return &CardHandler{cardService: cardService}
}

// Public: Submit a card for a campaign
func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var input service.SaveCardInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid JSON payload")
		return
	}

	input.CampaignSlug = slug

	card, err := h.cardService.SaveCard(r.Context(), input)
	if err != nil {
		if err == service.ErrCampaignNotFound {
			JSONError(w, http.StatusNotFound, "CAMPAIGN_NOT_FOUND", "Campaign does not exist or is inactive")
			return
		}
		if err == service.ErrInvalidCardInput {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid card submission data")
			return
		}
		JSONError(w, http.StatusInternalServerError, "SAVE_FAILED", "Failed to save card: "+err.Error())
		return
	}

	JSONCreated(w, map[string]interface{}{
		"saved": true,
		"id":    card.ID,
	})
}

// Admin: Get cards for a specific campaign
func (h *CardHandler) ListByCampaign(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	cards, err := h.cardService.GetCardsByCampaign(r.Context(), slug)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "LOAD_FAILED", "Failed to load cards")
		return
	}

	JSONOK(w, map[string]interface{}{
		"slug":  slug,
		"cards": cards,
		"total": len(cards),
	})
}

// Admin: Get all cards across all campaigns
func (h *CardHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	cards, total, err := h.cardService.GetAllCards(r.Context(), limit, offset)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "LOAD_FAILED", "Failed to load cards")
		return
	}

	JSONOK(w, map[string]interface{}{
		"cards":  cards,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Admin: Dashboard overall statistics
func (h *CardHandler) DashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.cardService.GetDashboardStats(r.Context())
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "LOAD_FAILED", "Failed to load dashboard stats")
		return
	}

	JSONOK(w, stats)
}
