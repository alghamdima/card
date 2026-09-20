package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cards-api/internal/domain"
)

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) Save(ctx context.Context, card *domain.Card) error {
	query := `
		INSERT INTO campaign_cards (campaign_slug, from_name, to_name, message, heading, device, date_str, time_str)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.ExecContext(ctx, query,
		card.CampaignSlug, card.FromName, card.ToName, card.Message,
		card.Heading, card.Device, card.DateStr, card.TimeStr,
	)
	if err != nil {
		return fmt.Errorf("failed to save card: %w", err)
	}

	id, _ := res.LastInsertId()
	card.ID = id
	return nil
}

func (r *CardRepository) GetByCampaignSlug(ctx context.Context, slug string) ([]domain.Card, error) {
	query := `
		SELECT id, campaign_slug, from_name, to_name, message, heading, device, date_str, time_str, created_at
		FROM campaign_cards
		WHERE campaign_slug = ?
		ORDER BY id DESC
	`
	rows, err := r.db.QueryContext(ctx, query, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get cards by slug: %w", err)
	}
	defer rows.Close()

	var list []domain.Card
	for rows.Next() {
		var c domain.Card
		var createdAtStr string
		if err := rows.Scan(
			&c.ID, &c.CampaignSlug, &c.FromName, &c.ToName, &c.Message,
			&c.Heading, &c.Device, &c.DateStr, &c.TimeStr, &createdAtStr,
		); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		if c.CreatedAt.IsZero() {
			c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		}
		list = append(list, c)
	}

	return list, nil
}

func (r *CardRepository) GetAll(ctx context.Context, limit, offset int) ([]domain.Card, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM campaign_cards").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, campaign_slug, from_name, to_name, message, heading, device, date_str, time_str, created_at
		FROM campaign_cards
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all cards: %w", err)
	}
	defer rows.Close()

	var list []domain.Card
	for rows.Next() {
		var c domain.Card
		var createdAtStr string
		if err := rows.Scan(
			&c.ID, &c.CampaignSlug, &c.FromName, &c.ToName, &c.Message,
			&c.Heading, &c.Device, &c.DateStr, &c.TimeStr, &createdAtStr,
		); err != nil {
			return nil, 0, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		if c.CreatedAt.IsZero() {
			c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		}
		list = append(list, c)
	}

	return list, total, nil
}

func (r *CardRepository) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	stats := &domain.DashboardStats{
		DeviceStats: make(map[string]int),
	}

	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM campaigns").Scan(&stats.TotalCampaigns)
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM campaigns WHERE active = 1").Scan(&stats.ActiveCampaigns)
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM campaign_cards").Scan(&stats.TotalCards)

	today := time.Now().Format("02/01/2006")
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM campaign_cards WHERE date_str = ? OR date(created_at) = date('now')", today).Scan(&stats.CardsToday)

	rows, err := r.db.QueryContext(ctx, "SELECT COALESCE(device, 'Unknown'), COUNT(*) FROM campaign_cards GROUP BY device")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var dev string
			var count int
			if err := rows.Scan(&dev, &count); err == nil {
				stats.DeviceStats[dev] = count
			}
		}
	}

	return stats, nil
}
