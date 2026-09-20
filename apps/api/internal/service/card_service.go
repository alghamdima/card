package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"cards-api/internal/domain"
	"cards-api/internal/repository"
)

var (
	ErrInvalidCardInput = errors.New("INVALID_CARD_INPUT")
)

type CardService struct {
	repo         *repository.CardRepository
	campaignRepo *repository.CampaignRepository
}

func NewCardService(repo *repository.CardRepository, campaignRepo *repository.CampaignRepository) *CardService {
	return &CardService{
		repo:         repo,
		campaignRepo: campaignRepo,
	}
}

type SaveCardInput struct {
	CampaignSlug string `json:"campaignSlug"`
	FromName     string `json:"from"`
	ToName       string `json:"to"`
	Message      string `json:"message"`
	Heading      string `json:"heading"`
	Device       string `json:"device"`
}

func (s *CardService) SaveCard(ctx context.Context, input SaveCardInput) (*domain.Card, error) {
	slug := strings.TrimSpace(strings.ToLower(input.CampaignSlug))
	if slug == "" {
		return nil, ErrInvalidCardInput
	}

	// Verify campaign exists and is active
	camp, err := s.campaignRepo.GetPublicBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if camp == nil || !camp.Active {
		return nil, ErrCampaignNotFound
	}

	from := strings.TrimSpace(input.FromName)
	if from == "" {
		from = "Anonymous"
	}
	to := strings.TrimSpace(input.ToName)
	msg := strings.TrimSpace(input.Message)
	heading := strings.TrimSpace(input.Heading)
	device := strings.TrimSpace(input.Device)
	if device == "" {
		device = "Web"
	}

	loc, _ := time.LoadLocation("Asia/Riyadh")
	if loc == nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	dateStr := now.Format("02/01/2006")
	timeStr := now.Format("15:04:05")

	card := &domain.Card{
		CampaignSlug: slug,
		FromName:     from,
		ToName:       to,
		Message:      msg,
		Heading:      heading,
		Device:       device,
		DateStr:      dateStr,
		TimeStr:      timeStr,
		CreatedAt:    now,
	}

	if err := s.repo.Save(ctx, card); err != nil {
		return nil, err
	}

	return card, nil
}

func (s *CardService) GetCardsByCampaign(ctx context.Context, slug string) ([]domain.Card, error) {
	return s.repo.GetByCampaignSlug(ctx, slug)
}

func (s *CardService) GetAllCards(ctx context.Context, limit, offset int) ([]domain.Card, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.GetAll(ctx, limit, offset)
}

func (s *CardService) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx)
}
