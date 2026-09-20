package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	"cards-api/internal/domain"
	"cards-api/internal/repository"
)

var (
	ErrCampaignNotFound      = errors.New("CAMPAIGN_NOT_FOUND")
	ErrCampaignAlreadyExists = errors.New("CAMPAIGN_ALREADY_EXISTS")
	ErrInvalidCampaignInput  = errors.New("INVALID_CAMPAIGN_INPUT")
)

type CampaignService struct {
	repo *repository.CampaignRepository
}

func NewCampaignService(repo *repository.CampaignRepository) *CampaignService {
	return &CampaignService{repo: repo}
}

// GenerateRandomSlug generates a URL-safe 8-character string using crypto/rand
func (s *CampaignService) GenerateRandomSlug(ctx context.Context) (string, error) {
	const charset = "abcdefghjkmnpqrstuvwxyz23456789" // URL-safe, no visually confusing chars (0, O, 1, l, i)
	const length = 8

	for attempts := 0; attempts < 10; attempts++ {
		bytes := make([]byte, length)
		if _, err := rand.Read(bytes); err != nil {
			return "", fmt.Errorf("failed to read random bytes: %w", err)
		}

		var sb strings.Builder
		for _, b := range bytes {
			sb.WriteByte(charset[int(b)%len(charset)])
		}
		candidate := sb.String()

		exists, err := s.repo.ExistsSlug(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}

	return "", errors.New("SLUG_GENERATION_FAILED")
}

func (s *CampaignService) GetPublicCampaign(ctx context.Context, slug string) (*domain.Campaign, error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug == "" {
		return nil, ErrCampaignNotFound
	}

	c, err := s.repo.GetPublicBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCampaignNotFound
	}

	return c, nil
}

func (s *CampaignService) GetAllPublicCampaigns(ctx context.Context) ([]domain.CampaignSummary, error) {
	return s.repo.GetAllPublic(ctx)
}

func (s *CampaignService) GetAllAdminCampaigns(ctx context.Context) ([]domain.CampaignSummary, error) {
	return s.repo.GetAllAdmin(ctx)
}

func (s *CampaignService) GetCampaignDetails(ctx context.Context, slug string) (*domain.Campaign, error) {
	c, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCampaignNotFound
	}
	return c, nil
}

type CreateCampaignInput struct {
	Title      string                  `json:"title"`
	TitleAR    string                  `json:"titleAR,omitempty"`
	TitleEN    string                  `json:"titleEN,omitempty"`
	Slug       string                  `json:"slug"` // Optional: if empty, will be auto-generated randomly!
	Lang       string                  `json:"lang"`
	TextColor  string                  `json:"textColor"`
	HeadColor  string                  `json:"headColor"`
	Boxes      *domain.BoxesConfig     `json:"boxes"`
	Image      string                  `json:"image"`
	Thumb      string                  `json:"thumb"`
	TemplateAR *domain.TemplateVariant `json:"templateAR,omitempty"`
	TemplateEN *domain.TemplateVariant `json:"templateEN,omitempty"`
}

func (s *CampaignService) CreateCampaign(ctx context.Context, input CreateCampaignInput) (*domain.Campaign, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		if input.TitleAR != "" {
			title = strings.TrimSpace(input.TitleAR)
		} else if input.TitleEN != "" {
			title = strings.TrimSpace(input.TitleEN)
		} else {
			return nil, fmt.Errorf("%w: title is required", ErrInvalidCampaignInput)
		}
	}
	titleAR := strings.TrimSpace(input.TitleAR)
	if titleAR == "" {
		titleAR = title
	}
	titleEN := strings.TrimSpace(input.TitleEN)
	if titleEN == "" {
		titleEN = title
	}
	if input.Image == "" && input.TemplateAR == nil {
		return nil, fmt.Errorf("%w: image is required", ErrInvalidCampaignInput)
	}

	slug := strings.TrimSpace(strings.ToLower(input.Slug))
	if slug == "" {
		// Randomly generate slug automatically as requested!
		randSlug, err := s.GenerateRandomSlug(ctx)
		if err != nil {
			return nil, err
		}
		slug = randSlug
	} else {
		// sanitize slug
		slug = strings.ReplaceAll(slug, " ", "-")
		exists, err := s.repo.ExistsSlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrCampaignAlreadyExists
		}
	}

	lang := input.Lang
	if lang != "en" {
		lang = "ar"
	}
	textColor := input.TextColor
	if textColor == "" {
		textColor = "#FFFFFF"
	}
	headColor := input.HeadColor
	if headColor == "" {
		headColor = "#FFCD00"
	}

	image := input.Image
	if image == "" && input.TemplateAR != nil {
		image = input.TemplateAR.Image
	}
	thumb := input.Thumb
	if thumb == "" {
		thumb = image
	}

	boxes := input.Boxes
	if boxes == nil {
		boxes = &domain.BoxesConfig{
			To: domain.BoxItem{
				Top: 300, Bottom: 430, CenterX: 540, MaxW: 800, Size: 52,
				Color: textColor, Weight: "bold",
			},
			Message: domain.BoxItem{
				Top: 470, Bottom: 870, CenterX: 540, MaxW: 840, Size: 38,
				Color: textColor, HeadColor: headColor, BodyColor: textColor,
			},
			From: domain.BoxItem{
				Top: 900, Bottom: 1020, CenterX: 540, MaxW: 800, Size: 44,
				Color: textColor, Weight: "bold",
			},
		}
	}

	camp := &domain.Campaign{
		Slug:       slug,
		Title:      title,
		TitleAR:    titleAR,
		TitleEN:    titleEN,
		Lang:       lang,
		TextColor:  textColor,
		HeadColor:  headColor,
		Boxes:      *boxes,
		Image:      image,
		Thumb:      thumb,
		TemplateAR: input.TemplateAR,
		TemplateEN: input.TemplateEN,
		Active:     true,
	}

	if err := s.repo.Create(ctx, camp); err != nil {
		return nil, err
	}

	return camp, nil
}

type UpdateCampaignInput struct {
	Title      string                  `json:"title"`
	TitleAR    string                  `json:"titleAR,omitempty"`
	TitleEN    string                  `json:"titleEN,omitempty"`
	Lang       string                  `json:"lang"`
	TextColor  string                  `json:"textColor"`
	HeadColor  string                  `json:"headColor"`
	Boxes      *domain.BoxesConfig     `json:"boxes"`
	Image      string                  `json:"image,omitempty"`
	Thumb      string                  `json:"thumb,omitempty"`
	TemplateAR *domain.TemplateVariant `json:"templateAR,omitempty"`
	TemplateEN *domain.TemplateVariant `json:"templateEN,omitempty"`
	Active     *bool                   `json:"active"`
}

func (s *CampaignService) UpdateCampaign(ctx context.Context, slug string, input UpdateCampaignInput) (*domain.Campaign, error) {
	camp, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if camp == nil {
		return nil, ErrCampaignNotFound
	}

	if input.Title != "" {
		camp.Title = strings.TrimSpace(input.Title)
	}
	if input.TitleAR != "" {
		camp.TitleAR = strings.TrimSpace(input.TitleAR)
	}
	if input.TitleEN != "" {
		camp.TitleEN = strings.TrimSpace(input.TitleEN)
	}
	if input.Lang != "" {
		camp.Lang = input.Lang
	}
	if input.TextColor != "" {
		camp.TextColor = input.TextColor
	}
	if input.HeadColor != "" {
		camp.HeadColor = input.HeadColor
	}
	if input.Boxes != nil {
		camp.Boxes = *input.Boxes
	}
	if input.Image != "" {
		camp.Image = input.Image
	}
	if input.Thumb != "" {
		camp.Thumb = input.Thumb
	}
	if input.TemplateAR != nil {
		camp.TemplateAR = input.TemplateAR
	}
	if input.TemplateEN != nil {
		camp.TemplateEN = input.TemplateEN
	}
	if input.Active != nil {
		camp.Active = *input.Active
	}

	if err := s.repo.Update(ctx, camp); err != nil {
		return nil, err
	}

	return camp, nil
}

func (s *CampaignService) DeleteCampaign(ctx context.Context, slug string) error {
	return s.repo.Delete(ctx, slug)
}
