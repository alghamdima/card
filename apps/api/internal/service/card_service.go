package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cards-api/internal/domain"
	"cards-api/internal/repository"
)

var ErrInvalidCardInput = errors.New("INVALID_CARD_INPUT")

const (
	maxNameRunes      = 120
	maxMessageRunes   = 1000
	maxFieldValues    = 30
	maxFieldKeyRunes  = 64
	maxFieldValueRune = 500
	defaultCardLimit  = 50
	maxCardLimit      = 100
	anonymousSender   = "Anonymous"
)

// knownDevices are the labels the web client reports; anything else is stored as "Other".
var knownDevices = map[string]bool{"iPhone": true, "Android": true, "Desktop": true, "Web": true}

type CardService struct {
	repo         *repository.CardRepository
	campaignRepo *repository.CampaignRepository
	loc          *time.Location
	now          func() time.Time
}

// NewCardService builds the card service. Card dates ("dd/mm/yyyy") are
// recorded and aggregated in loc, so "cards today" matches the business day.
func NewCardService(repo *repository.CardRepository, campaignRepo *repository.CampaignRepository, loc *time.Location) *CardService {
	if loc == nil {
		loc = time.UTC
	}
	return &CardService{repo: repo, campaignRepo: campaignRepo, loc: loc, now: time.Now}
}

type SaveCardInput struct {
	CampaignSlug string                 `json:"campaignSlug"`
	FromName     string                 `json:"from"`
	ToName       string                 `json:"to"`
	Message      string                 `json:"message"`
	Heading      string                 `json:"heading"`
	Lang         string                 `json:"lang"`
	FieldValues  map[string]interface{} `json:"fieldValues"`
	Device       string                 `json:"device"`
}

func (s *CardService) SaveCard(ctx context.Context, input SaveCardInput) (*domain.Card, error) {
	slug := strings.TrimSpace(strings.ToLower(input.CampaignSlug))
	if slug == "" {
		return nil, invalidf(ErrInvalidCardInput, "campaign is required")
	}

	camp, err := s.campaignRepo.GetPublicBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if camp == nil || !camp.Active {
		return nil, ErrCampaignNotFound
	}

	fieldValues, err := cleanFieldValues(input.FieldValues)
	if err != nil {
		return nil, err
	}

	from, err := cleanCardText("from", input.FromName, maxNameRunes)
	if err != nil {
		return nil, err
	}
	if from == "" {
		from = anonymousSender
	}
	to, err := cleanCardText("to", input.ToName, maxNameRunes)
	if err != nil {
		return nil, err
	}
	if to == "" {
		// Dynamic templates carry the recipient inside the field values.
		to = firstNonEmpty(fieldValues["emp_name"], fieldValues["name"])
	}
	message, err := cleanCardText("message", input.Message, maxMessageRunes)
	if err != nil {
		return nil, err
	}
	heading, err := cleanCardText("heading", input.Heading, maxNameRunes)
	if err != nil {
		return nil, err
	}

	if to == "" && message == "" && len(fieldValues) == 0 {
		return nil, invalidf(ErrInvalidCardInput, "card content is empty")
	}

	lang := input.Lang
	if lang != "en" {
		lang = "ar"
	}
	device := strings.TrimSpace(input.Device)
	if device == "" {
		device = "Web"
	} else if !knownDevices[device] {
		device = "Other"
	}

	now := s.now().In(s.loc)
	card := &domain.Card{
		CampaignSlug: slug,
		FromName:     from,
		ToName:       to,
		Message:      message,
		Heading:      heading,
		Lang:         lang,
		FieldValues:  toAnyMap(fieldValues),
		Device:       device,
		DateStr:      now.Format("02/01/2006"),
		TimeStr:      now.Format("15:04:05"),
		CreatedAt:    now,
	}

	if err := s.repo.Save(ctx, card); err != nil {
		return nil, err
	}
	return card, nil
}

func (s *CardService) GetCampaignAnalyticsList(ctx context.Context) ([]domain.CampaignAnalytics, error) {
	return s.repo.GetCampaignAnalyticsList(ctx, s.today())
}

// CardPage lists cards (newest first), optionally restricted to one campaign and/or a search term.
func (s *CardService) CardPage(ctx context.Context, slug, search string, limit, offset int) ([]domain.Card, int, error) {
	if limit <= 0 || limit > maxCardLimit {
		limit = defaultCardLimit
	}
	if offset < 0 {
		offset = 0
	}
	search, _ = cleanText(search, maxNameRunes)

	return s.repo.List(ctx, repository.CardQuery{
		Slug:   strings.TrimSpace(strings.ToLower(slug)),
		Search: search,
		Limit:  limit,
		Offset: offset,
	})
}

// ExportCampaignCards streams every card of a campaign to fn.
func (s *CardService) ExportCampaignCards(ctx context.Context, slug string, fn func(domain.Card) error) error {
	return s.repo.ForEachByCampaign(ctx, strings.TrimSpace(strings.ToLower(slug)), fn)
}

func (s *CardService) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx, s.today())
}

func (s *CardService) today() string {
	return s.now().In(s.loc).Format("02/01/2006")
}

func cleanCardText(name, value string, maxRunes int) (string, error) {
	text, ok := cleanText(value, maxRunes)
	if !ok {
		return "", invalidf(ErrInvalidCardInput, "%s must be at most %d characters", name, maxRunes)
	}
	return text, nil
}

// cleanFieldValues keeps only non-empty string/number values under sane key names.
func cleanFieldValues(in map[string]interface{}) (map[string]string, error) {
	if len(in) > maxFieldValues {
		return nil, invalidf(ErrInvalidCardInput, "too many field values (max %d)", maxFieldValues)
	}

	out := make(map[string]string, len(in))
	for key, raw := range in {
		if key == "" || len([]rune(key)) > maxFieldKeyRunes {
			return nil, invalidf(ErrInvalidCardInput, "invalid field name")
		}
		var text string
		switch v := raw.(type) {
		case nil:
			continue
		case string:
			text = v
		case float64, bool:
			text = fmt.Sprint(v)
		default:
			return nil, invalidf(ErrInvalidCardInput, "field %q must be text", key)
		}

		text, ok := cleanText(text, maxFieldValueRune)
		if !ok {
			return nil, invalidf(ErrInvalidCardInput, "field %q is too long (max %d characters)", key, maxFieldValueRune)
		}
		if text != "" {
			out[key] = text
		}
	}
	return out, nil
}

func toAnyMap(in map[string]string) map[string]interface{} {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
