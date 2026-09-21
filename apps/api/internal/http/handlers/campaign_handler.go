package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"cards-api/internal/http/response"
	"cards-api/internal/service"
)

type CampaignHandler struct {
	campaignService *service.CampaignService
}

func NewCampaignHandler(campaignService *service.CampaignService) *CampaignHandler {
	return &CampaignHandler{campaignService: campaignService}
}

// ListPublic returns all active campaigns.
func (h *CampaignHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	list, err := h.campaignService.GetAllPublicCampaigns(r.Context())
	if err != nil {
		writeServiceError(w, r, err, "LOAD_FAILED", "Failed to load campaigns")
		return
	}
	response.OKWithETag(w, r, map[string]any{"campaigns": list})
}

// GetPublic returns a single active campaign by slug.
func (h *CampaignHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	camp, err := h.campaignService.GetPublicCampaign(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		writeServiceError(w, r, err, "LOAD_FAILED", "Failed to load campaign")
		return
	}
	response.OKWithETag(w, r, camp)
}

// ListAdmin returns all campaigns with participation counters.
func (h *CampaignHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	list, err := h.campaignService.GetAllAdminCampaigns(r.Context())
	if err != nil {
		writeServiceError(w, r, err, "LOAD_FAILED", "Failed to load campaigns")
		return
	}
	response.OK(w, map[string]any{"campaigns": list})
}

// GetAdmin returns campaign details regardless of its active flag.
func (h *CampaignHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	camp, err := h.campaignService.GetCampaignDetails(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		writeServiceError(w, r, err, "LOAD_FAILED", "Failed to load campaign")
		return
	}
	response.OK(w, camp)
}

// Create adds a campaign; the slug is generated randomly when omitted.
func (h *CampaignHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.CreateCampaignInput
	if !decodeJSON(w, r, &input, maxCampaignBody) {
		return
	}

	camp, err := h.campaignService.CreateCampaign(r.Context(), input)
	if err != nil {
		writeServiceError(w, r, err, "CREATE_FAILED", "Failed to create campaign")
		return
	}
	response.Created(w, camp)
}

func (h *CampaignHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input service.UpdateCampaignInput
	if !decodeJSON(w, r, &input, maxCampaignBody) {
		return
	}

	camp, err := h.campaignService.UpdateCampaign(r.Context(), chi.URLParam(r, "slug"), input)
	if err != nil {
		writeServiceError(w, r, err, "UPDATE_FAILED", "Failed to update campaign")
		return
	}
	response.OK(w, camp)
}

func (h *CampaignHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.campaignService.DeleteCampaign(r.Context(), chi.URLParam(r, "slug")); err != nil {
		writeServiceError(w, r, err, "DELETE_FAILED", "Failed to delete campaign")
		return
	}
	response.OK(w, map[string]bool{"deleted": true})
}
