package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"cards-api/internal/domain"
	"cards-api/internal/repository"
)

var (
	ErrCampaignNotFound      = errors.New("CAMPAIGN_NOT_FOUND")
	ErrCampaignAlreadyExists = errors.New("CAMPAIGN_ALREADY_EXISTS")
	ErrInvalidCampaignInput  = errors.New("INVALID_CAMPAIGN_INPUT")
)

const (
	defaultTextColor = "#FFFFFF"
	defaultHeadColor = "#FFCD00"

	maxTitleRunes     = 120
	maxLabelRunes     = 100
	maxImageChars     = 8 << 20 // base64 data URL of the card artwork
	maxThumbChars     = 3 << 19 // 1.5 MB: thumbnails are inlined in the public listing
	maxTemplateFields = 20
	canvasWidth       = 1080
	canvasHeight      = 1350
)

type CampaignService struct {
	repo *repository.CampaignRepository
}

func NewCampaignService(repo *repository.CampaignRepository) *CampaignService {
	return &CampaignService{repo: repo}
}

// GenerateRandomSlug generates a URL-safe 8-character string using crypto/rand.
func (s *CampaignService) GenerateRandomSlug(ctx context.Context) (string, error) {
	const charset = "abcdefghjkmnpqrstuvwxyz23456789" // no visually confusing chars (0, o, 1, l, i)
	const length = 8

	max := big.NewInt(int64(len(charset)))
	for attempts := 0; attempts < 10; attempts++ {
		var sb strings.Builder
		for i := 0; i < length; i++ {
			// rand.Int avoids the modulo bias of "byte % len(charset)".
			n, err := rand.Int(rand.Reader, max)
			if err != nil {
				return "", fmt.Errorf("failed to read random bytes: %w", err)
			}
			sb.WriteByte(charset[n.Int64()])
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
	return presentCampaign(c), nil
}

func (s *CampaignService) GetAllPublicCampaigns(ctx context.Context) ([]domain.PublicCampaignSummary, error) {
	return s.repo.GetAllPublic(ctx)
}

func (s *CampaignService) GetAllAdminCampaigns(ctx context.Context) ([]domain.CampaignSummary, error) {
	return s.repo.GetAllAdmin(ctx)
}

func (s *CampaignService) GetCampaignDetails(ctx context.Context, slug string) (*domain.Campaign, error) {
	c, err := s.repo.GetBySlug(ctx, strings.TrimSpace(strings.ToLower(slug)))
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCampaignNotFound
	}
	return presentCampaign(c), nil
}

type CreateCampaignInput struct {
	Title      string                  `json:"title"`
	TitleAR    string                  `json:"titleAR,omitempty"`
	TitleEN    string                  `json:"titleEN,omitempty"`
	Slug       string                  `json:"slug"` // optional: generated randomly when empty
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
	titleAR, err := cleanTitle(input.TitleAR)
	if err != nil {
		return nil, err
	}
	titleEN, err := cleanTitle(input.TitleEN)
	if err != nil {
		return nil, err
	}
	title, err := cleanTitle(input.Title)
	if err != nil {
		return nil, err
	}
	if title == "" {
		title = firstNonEmpty(titleAR, titleEN)
	}
	if title == "" {
		return nil, invalidf(ErrInvalidCampaignInput, "title is required")
	}
	titleAR = firstNonEmpty(titleAR, title)
	titleEN = firstNonEmpty(titleEN, title)

	lang := input.Lang
	if lang == "" {
		lang = "ar"
	}
	if lang != "ar" && lang != "en" {
		return nil, invalidf(ErrInvalidCampaignInput, "lang must be \"ar\" or \"en\"")
	}
	textColor, err := colorOrDefault("textColor", input.TextColor, defaultTextColor)
	if err != nil {
		return nil, err
	}
	headColor, err := colorOrDefault("headColor", input.HeadColor, defaultHeadColor)
	if err != nil {
		return nil, err
	}

	if err := validateTemplate("templateAR", input.TemplateAR); err != nil {
		return nil, err
	}
	if err := validateTemplate("templateEN", input.TemplateEN); err != nil {
		return nil, err
	}

	image := input.Image
	if image == "" && input.TemplateAR != nil {
		image = input.TemplateAR.Image
	}
	if image == "" {
		return nil, invalidf(ErrInvalidCampaignInput, "image is required")
	}
	if err := validateImage("image", image, maxImageChars); err != nil {
		return nil, err
	}

	thumb := input.Thumb
	if thumb == "" && len(image) <= maxThumbChars {
		thumb = image
	}
	if thumb != "" {
		if err := validateImage("thumb", thumb, maxThumbChars); err != nil {
			return nil, err
		}
	}

	slug, err := s.resolveSlug(ctx, input.Slug)
	if err != nil {
		return nil, err
	}

	boxes := input.Boxes
	if boxes == nil {
		boxes = defaultBoxes(textColor, headColor)
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
		if errors.Is(err, repository.ErrDuplicateSlug) {
			return nil, ErrCampaignAlreadyExists
		}
		return nil, err
	}

	return presentCampaign(camp), nil
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

// UpdateCampaign applies a partial update: empty strings and nil pointers leave the stored value untouched.
func (s *CampaignService) UpdateCampaign(ctx context.Context, slug string, input UpdateCampaignInput) (*domain.Campaign, error) {
	camp, err := s.repo.GetBySlug(ctx, strings.TrimSpace(strings.ToLower(slug)))
	if err != nil {
		return nil, err
	}
	if camp == nil {
		return nil, ErrCampaignNotFound
	}

	title, err := cleanTitle(input.Title)
	if err != nil {
		return nil, err
	}
	titleAR, err := cleanTitle(input.TitleAR)
	if err != nil {
		return nil, err
	}
	titleEN, err := cleanTitle(input.TitleEN)
	if err != nil {
		return nil, err
	}
	if titleAR != "" {
		camp.TitleAR = titleAR
		camp.Title = titleAR // the primary title follows the Arabic one, as on creation
	}
	if titleEN != "" {
		camp.TitleEN = titleEN
	}
	if title != "" {
		camp.Title = title
	}

	if input.Lang != "" {
		if input.Lang != "ar" && input.Lang != "en" {
			return nil, invalidf(ErrInvalidCampaignInput, "lang must be \"ar\" or \"en\"")
		}
		camp.Lang = input.Lang
	}
	if input.TextColor != "" {
		if !isHexColor(input.TextColor) {
			return nil, invalidf(ErrInvalidCampaignInput, "textColor must be a #RRGGBB color")
		}
		camp.TextColor = input.TextColor
	}
	if input.HeadColor != "" {
		if !isHexColor(input.HeadColor) {
			return nil, invalidf(ErrInvalidCampaignInput, "headColor must be a #RRGGBB color")
		}
		camp.HeadColor = input.HeadColor
	}
	if input.Boxes != nil {
		camp.Boxes = *input.Boxes
	}

	if input.Image != "" {
		if err := validateImage("image", input.Image, maxImageChars); err != nil {
			return nil, err
		}
		camp.Image = input.Image
	}
	if input.Thumb != "" {
		if err := validateImage("thumb", input.Thumb, maxThumbChars); err != nil {
			return nil, err
		}
		camp.Thumb = input.Thumb
	}

	if input.TemplateAR != nil {
		if err := validateTemplate("templateAR", input.TemplateAR); err != nil {
			return nil, err
		}
		camp.TemplateAR = input.TemplateAR
		// The Arabic artwork is the campaign's canonical image; keep them in sync.
		if input.TemplateAR.Image != "" && input.Image == "" {
			camp.Image = input.TemplateAR.Image
		}
	}
	if input.TemplateEN != nil {
		if err := validateTemplate("templateEN", input.TemplateEN); err != nil {
			return nil, err
		}
		camp.TemplateEN = input.TemplateEN
	}
	if input.Active != nil {
		camp.Active = *input.Active
	}

	if err := s.repo.Update(ctx, camp); err != nil {
		return nil, err
	}

	return presentCampaign(camp), nil
}

func (s *CampaignService) DeleteCampaign(ctx context.Context, slug string) error {
	deleted, err := s.repo.Delete(ctx, strings.TrimSpace(strings.ToLower(slug)))
	if err != nil {
		return err
	}
	if !deleted {
		return ErrCampaignNotFound
	}
	return nil
}

func (s *CampaignService) resolveSlug(ctx context.Context, requested string) (string, error) {
	slug := strings.ReplaceAll(strings.TrimSpace(strings.ToLower(requested)), " ", "-")
	if slug == "" {
		return s.GenerateRandomSlug(ctx)
	}
	if !slugPattern.MatchString(slug) {
		return "", invalidf(ErrInvalidCampaignInput, "slug may only contain lowercase letters, digits and hyphens (max 64 characters)")
	}

	exists, err := s.repo.ExistsSlug(ctx, slug)
	if err != nil {
		return "", err
	}
	if exists {
		return "", ErrCampaignAlreadyExists
	}
	return slug, nil
}

// presentCampaign trims duplicated artwork from the response. The same image
// is stored as campaign.image, templateAR.image and templateEN.image, and the
// admin/public clients fall back to campaign.image when a variant image is
// empty, so sending it three times only triples the payload.
func presentCampaign(c *domain.Campaign) *domain.Campaign {
	c.Thumb = "" // only the listing endpoints use thumbnails
	if c.TemplateAR != nil && c.TemplateAR.Image == c.Image {
		c.TemplateAR.Image = ""
	}
	if c.TemplateEN != nil && c.TemplateEN.Image == c.Image {
		c.TemplateEN.Image = ""
	}
	return c
}

func defaultBoxes(textColor, headColor string) *domain.BoxesConfig {
	return &domain.BoxesConfig{
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

func cleanTitle(s string) (string, error) {
	title, ok := cleanText(s, maxTitleRunes)
	if !ok {
		return "", invalidf(ErrInvalidCampaignInput, "title must be at most %d characters", maxTitleRunes)
	}
	return title, nil
}

func colorOrDefault(name, value, fallback string) (string, error) {
	if value == "" {
		return fallback, nil
	}
	if !isHexColor(value) {
		return "", invalidf(ErrInvalidCampaignInput, "%s must be a #RRGGBB color", name)
	}
	return value, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func validateImage(name, dataURL string, maxChars int) error {
	if len(dataURL) > maxChars {
		return invalidf(ErrInvalidCampaignInput, "%s is too large (max %g MB)", name, float64(maxChars)/(1<<20))
	}
	if !isImageDataURL(dataURL) {
		return invalidf(ErrInvalidCampaignInput, "%s must be a PNG, JPEG, WebP or GIF data URL", name)
	}
	return nil
}

// validateTemplate checks a language variant. An empty image is allowed and
// means "use the campaign image".
func validateTemplate(name string, v *domain.TemplateVariant) error {
	if v == nil {
		return nil
	}
	if v.Image != "" {
		if err := validateImage(name+".image", v.Image, maxImageChars); err != nil {
			return err
		}
	}
	if len(v.Fields) > maxTemplateFields {
		return invalidf(ErrInvalidCampaignInput, "%s allows at most %d fields", name, maxTemplateFields)
	}

	seen := make(map[string]bool, len(v.Fields))
	for i := range v.Fields {
		f := &v.Fields[i]
		if !fieldIDPattern.MatchString(f.ID) {
			return invalidf(ErrInvalidCampaignInput, "%s.fields[%d].id must be 1-64 letters, digits, '_' or '-'", name, i)
		}
		if seen[f.ID] {
			return invalidf(ErrInvalidCampaignInput, "%s has a duplicate field id %q", name, f.ID)
		}
		seen[f.ID] = true

		if f.Name == "" {
			f.Name = f.ID
		}
		var ok bool
		if f.Label, ok = cleanText(f.Label, maxLabelRunes); !ok {
			return invalidf(ErrInvalidCampaignInput, "%s.fields[%d].label is too long", name, i)
		}
		if f.Placeholder, ok = cleanText(f.Placeholder, maxLabelRunes); !ok {
			return invalidf(ErrInvalidCampaignInput, "%s.fields[%d].placeholder is too long", name, i)
		}

		switch {
		case f.X < 0 || f.X > canvasWidth || f.Y < 0 || f.Y > canvasHeight:
			return invalidf(ErrInvalidCampaignInput, "%s.fields[%d] position is outside the card", name, i)
		case f.Width < 20 || f.Width > canvasWidth || f.Height < 20 || f.Height > canvasHeight:
			return invalidf(ErrInvalidCampaignInput, "%s.fields[%d] has an invalid size", name, i)
		case f.FontSize < 8 || f.FontSize > 300:
			return invalidf(ErrInvalidCampaignInput, "%s.fields[%d].fontSize must be between 8 and 300", name, i)
		case !isHexColor(f.Color):
			return invalidf(ErrInvalidCampaignInput, "%s.fields[%d].color must be a #RRGGBB color", name, i)
		case f.Weight != "" && f.Weight != "bold" && f.Weight != "regular":
			return invalidf(ErrInvalidCampaignInput, "%s.fields[%d].weight must be \"bold\" or \"regular\"", name, i)
		case f.Align != "" && f.Align != "center" && f.Align != "left" && f.Align != "right":
			return invalidf(ErrInvalidCampaignInput, "%s.fields[%d].align is invalid", name, i)
		case f.MaxChars < 0 || f.MaxChars > 500:
			return invalidf(ErrInvalidCampaignInput, "%s.fields[%d].maxChars must be between 0 and 500", name, i)
		}
	}
	return nil
}
