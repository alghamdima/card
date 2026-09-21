package handlers

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"cards-api/internal/domain"
	"cards-api/internal/http/response"
	"cards-api/internal/service"
)

type CardHandler struct {
	cardService *service.CardService
}

func NewCardHandler(cardService *service.CardService) *CardHandler {
	return &CardHandler{cardService: cardService}
}

// Create records a generated card for a campaign (public, used for participation analytics).
func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.SaveCardInput
	if !decodeJSON(w, r, &input, maxSmallBody) {
		return
	}
	input.CampaignSlug = chi.URLParam(r, "slug")

	card, err := h.cardService.SaveCard(r.Context(), input)
	if err != nil {
		writeServiceError(w, r, err, "SAVE_FAILED", "Failed to save card")
		return
	}

	response.Created(w, map[string]any{"saved": true, "id": card.ID})
}

// ListByCampaign returns one page of a campaign's cards; supports ?limit, ?offset and ?q.
func (h *CardHandler) ListByCampaign(w http.ResponseWriter, r *http.Request) {
	h.writeCardPage(w, r, chi.URLParam(r, "slug"))
}

// ListAll returns one page of cards across all campaigns; supports ?limit, ?offset and ?q.
func (h *CardHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	h.writeCardPage(w, r, "")
}

func (h *CardHandler) writeCardPage(w http.ResponseWriter, r *http.Request, slug string) {
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	cards, total, err := h.cardService.CardPage(r.Context(), slug, query.Get("q"), limit, offset)
	if err != nil {
		writeServiceError(w, r, err, "LOAD_FAILED", "Failed to load cards")
		return
	}

	data := map[string]any{"cards": cards, "total": total, "limit": limit, "offset": offset}
	if slug != "" {
		data["slug"] = slug
	}
	response.OK(w, data)
}

func (h *CardHandler) DashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.cardService.GetDashboardStats(r.Context())
	if err != nil {
		writeServiceError(w, r, err, "LOAD_FAILED", "Failed to load dashboard stats")
		return
	}
	response.OK(w, stats)
}

func (h *CardHandler) AnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	list, err := h.cardService.GetCampaignAnalyticsList(r.Context())
	if err != nil {
		writeServiceError(w, r, err, "LOAD_FAILED", "Failed to load campaign analytics")
		return
	}
	response.OK(w, map[string]any{"campaigns": list})
}

var unsafeFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

// ExportCampaignCSV streams every card of a campaign as an Excel-friendly CSV.
func (h *CardHandler) ExportCampaignCSV(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-cards.csv"`, unsafeFilenameChars.ReplaceAllString(slug, "_")))
	// UTF-8 BOM so Excel detects Arabic text correctly.
	_, _ = w.Write([]byte{0xEF, 0xBB, 0xBF})

	out := csv.NewWriter(w)
	_ = out.Write([]string{"ID", "Full Name", "Job Title / Details", "Sender", "Language", "Date", "Time"})

	err := h.cardService.ExportCampaignCards(r.Context(), slug, func(c domain.Card) error {
		return out.Write([]string{
			strconv.FormatInt(c.ID, 10),
			csvSafe(c.ToName),
			csvSafe(cardDetails(c)),
			csvSafe(c.FromName),
			c.Lang,
			c.DateStr,
			c.TimeStr,
		})
	})
	out.Flush()
	if err != nil {
		// Headers are already sent, so the best we can do is log the truncated export.
		log.Printf("CSV export for %q aborted: %v", slug, err)
	}
}

func cardDetails(c domain.Card) string {
	if c.Message != "" {
		return c.Message
	}
	for _, k := range []string{"job_title", "title", "message", "details", "field_muaw8lag", "field_muazzyhs"} {
		if val, ok := c.FieldValues[k]; ok && val != nil {
			s := fmt.Sprint(val)
			if s != "" {
				return s
			}
		}
	}
	return ""
}

// csvSafe neutralizes spreadsheet formula injection: cells starting with
// = + - @ (or tab/CR) would be executed by Excel, so they are prefixed with a quote.
func csvSafe(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}
