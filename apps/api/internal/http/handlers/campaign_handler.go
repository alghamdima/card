package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"cards-api/internal/service"
)

type CampaignHandler struct {
	campaignService *service.CampaignService
}

func NewCampaignHandler(campaignService *service.CampaignService) *CampaignHandler {
	return &CampaignHandler{campaignService: campaignService}
}

// Public: Get all active campaigns
func (h *CampaignHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	list, err := h.campaignService.GetAllPublicCampaigns(r.Context())
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "LOAD_FAILED", "Failed to load campaigns")
		return
	}
	JSONOK(w, map[string]interface{}{"campaigns": list})
}

// Public: Get single active campaign by slug
func (h *CampaignHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	camp, err := h.campaignService.GetPublicCampaign(r.Context(), slug)
	if err != nil {
		if err == service.ErrCampaignNotFound {
			JSONError(w, http.StatusNotFound, "CAMPAIGN_NOT_FOUND", "Campaign not found or inactive")
			return
		}
		JSONError(w, http.StatusInternalServerError, "LOAD_FAILED", "Failed to load campaign")
		return
	}

	JSONOK(w, camp)
}

// Admin: Get all campaigns
func (h *CampaignHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	list, err := h.campaignService.GetAllAdminCampaigns(r.Context())
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "LOAD_FAILED", "Failed to load campaigns")
		return
	}
	JSONOK(w, map[string]interface{}{"campaigns": list})
}

// Admin: Get campaign details
func (h *CampaignHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	camp, err := h.campaignService.GetCampaignDetails(r.Context(), slug)
	if err != nil {
		if err == service.ErrCampaignNotFound {
			JSONError(w, http.StatusNotFound, "CAMPAIGN_NOT_FOUND", "Campaign not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "LOAD_FAILED", "Failed to load campaign")
		return
	}
	JSONOK(w, camp)
}

// Admin: Create campaign (random slug if not provided!)
func (h *CampaignHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.CreateCampaignInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid JSON payload")
		return
	}

	camp, err := h.campaignService.CreateCampaign(r.Context(), input)
	if err != nil {
		if err == service.ErrCampaignAlreadyExists {
			JSONError(w, http.StatusConflict, "CAMPAIGN_EXISTS", "A campaign with this link already exists")
			return
		}
		if err == service.ErrInvalidCampaignInput {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		JSONError(w, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create campaign: "+err.Error())
		return
	}

	JSONCreated(w, camp)
}

// Admin: Update campaign
func (h *CampaignHandler) Update(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	var input service.UpdateCampaignInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid JSON payload")
		return
	}

	camp, err := h.campaignService.UpdateCampaign(r.Context(), slug, input)
	if err != nil {
		if err == service.ErrCampaignNotFound {
			JSONError(w, http.StatusNotFound, "CAMPAIGN_NOT_FOUND", "Campaign not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update campaign: "+err.Error())
		return
	}

	JSONOK(w, camp)
}

// Admin: Delete campaign
func (h *CampaignHandler) Delete(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if err := h.campaignService.DeleteCampaign(r.Context(), slug); err != nil {
		JSONError(w, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete campaign")
		return
	}

	JSONOK(w, map[string]bool{"deleted": true})
}
