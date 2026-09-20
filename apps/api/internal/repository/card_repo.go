package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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
	var fieldValuesStr sql.NullString
	if card.FieldValues != nil {
		b, err := json.Marshal(card.FieldValues)
		if err == nil {
			fieldValuesStr = sql.NullString{String: string(b), Valid: true}
		}
	}

	query := `
		INSERT INTO campaign_cards (campaign_slug, from_name, to_name, message, heading, lang, field_values, device, date_str, time_str)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.ExecContext(ctx, query,
		card.CampaignSlug, card.FromName, card.ToName, card.Message,
		card.Heading, card.Lang, fieldValuesStr, card.Device, card.DateStr, card.TimeStr,
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
		SELECT id, campaign_slug, from_name, to_name, message, heading, COALESCE(lang, 'ar'), field_values, device, date_str, time_str, created_at
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
		var fieldValuesRaw sql.NullString
		if err := rows.Scan(
			&c.ID, &c.CampaignSlug, &c.FromName, &c.ToName, &c.Message,
			&c.Heading, &c.Lang, &fieldValuesRaw, &c.Device, &c.DateStr, &c.TimeStr, &createdAtStr,
		); err != nil {
			return nil, err
		}
		if fieldValuesRaw.Valid && fieldValuesRaw.String != "" {
			_ = json.Unmarshal([]byte(fieldValuesRaw.String), &c.FieldValues)
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
		SELECT id, campaign_slug, from_name, to_name, message, heading, COALESCE(lang, 'ar'), field_values, device, date_str, time_str, created_at
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
		var fieldValuesRaw sql.NullString
		if err := rows.Scan(
			&c.ID, &c.CampaignSlug, &c.FromName, &c.ToName, &c.Message,
			&c.Heading, &c.Lang, &fieldValuesRaw, &c.Device, &c.DateStr, &c.TimeStr, &createdAtStr,
		); err != nil {
			return nil, 0, err
		}
		if fieldValuesRaw.Valid && fieldValuesRaw.String != "" {
			_ = json.Unmarshal([]byte(fieldValuesRaw.String), &c.FieldValues)
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		if c.CreatedAt.IsZero() {
			c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		}
		list = append(list, c)
	}

	return list, total, nil
}

func (r *CardRepository) GetCampaignAnalyticsList(ctx context.Context) ([]domain.CampaignAnalytics, error) {
	query := `
		SELECT c.slug, c.title,
		       COUNT(k.id) as total_cards,
		       COALESCE(SUM(CASE WHEN k.date_str = ? OR date(k.created_at) = date('now') THEN 1 ELSE 0 END), 0) as cards_today,
		       COALESCE(MAX(k.created_at), '') as last_card_at
		FROM campaigns c
		LEFT JOIN campaign_cards k ON c.slug = k.campaign_slug
		GROUP BY c.slug
		ORDER BY total_cards DESC, c.id DESC
	`
	today := time.Now().Format("02/01/2006")
	rows, err := r.db.QueryContext(ctx, query, today)
	if err != nil {
		return nil, fmt.Errorf("failed to query campaign analytics: %w", err)
	}
	defer rows.Close()

	var list []domain.CampaignAnalytics
	for rows.Next() {
		var item domain.CampaignAnalytics
		if err := rows.Scan(&item.Slug, &item.Title, &item.TotalCards, &item.CardsToday, &item.LastCardAt); err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, nil
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
