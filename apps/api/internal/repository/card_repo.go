package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"cards-api/internal/domain"
)

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

// CardQuery filters the admin card listings. An empty Slug lists every campaign.
type CardQuery struct {
	Slug   string
	Search string
	Limit  int
	Offset int
}

const cardColumns = `id, campaign_slug, COALESCE(from_name, ''), COALESCE(to_name, ''), COALESCE(message, ''),
	COALESCE(heading, ''), COALESCE(lang, 'ar'), field_values, COALESCE(device, ''),
	COALESCE(date_str, ''), COALESCE(time_str, ''), created_at`

func (r *CardRepository) Save(ctx context.Context, card *domain.Card) error {
	var fieldValues sql.NullString
	if len(card.FieldValues) > 0 {
		b, err := json.Marshal(card.FieldValues)
		if err != nil {
			return fmt.Errorf("failed to marshal field values: %w", err)
		}
		fieldValues = sql.NullString{String: string(b), Valid: true}
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO campaign_cards (campaign_slug, from_name, to_name, message, heading, lang, field_values, device, date_str, time_str)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		card.CampaignSlug, card.FromName, card.ToName, card.Message,
		card.Heading, card.Lang, fieldValues, card.Device, card.DateStr, card.TimeStr,
	)
	if err != nil {
		return fmt.Errorf("failed to save card: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to read new card id: %w", err)
	}
	card.ID = id
	return nil
}

// List returns one page of cards (newest first) plus the total matching count.
func (r *CardRepository) List(ctx context.Context, q CardQuery) ([]domain.Card, int, error) {
	where, args := cardFilter(q)

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM campaign_cards"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count cards: %w", err)
	}

	pageArgs := append(append([]any{}, args...), q.Limit, q.Offset)
	rows, err := r.db.QueryContext(ctx,
		"SELECT "+cardColumns+" FROM campaign_cards"+where+" ORDER BY id DESC LIMIT ? OFFSET ?", pageArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list cards: %w", err)
	}
	defer rows.Close()

	list := []domain.Card{}
	for rows.Next() {
		c, err := scanCard(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, c)
	}
	return list, total, rows.Err()
}

// ForEachByCampaign streams every card of a campaign (newest first) so large
// exports never have to be held in memory.
func (r *CardRepository) ForEachByCampaign(ctx context.Context, slug string, fn func(domain.Card) error) error {
	rows, err := r.db.QueryContext(ctx,
		"SELECT "+cardColumns+" FROM campaign_cards WHERE campaign_slug = ? ORDER BY id DESC", slug)
	if err != nil {
		return fmt.Errorf("failed to stream cards: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c, err := scanCard(rows)
		if err != nil {
			return err
		}
		if err := fn(c); err != nil {
			return err
		}
	}
	return rows.Err()
}

// GetCampaignAnalyticsList reports per-campaign totals; today is the local
// "dd/mm/yyyy" date string stored in campaign_cards.date_str.
func (r *CardRepository) GetCampaignAnalyticsList(ctx context.Context, today string) ([]domain.CampaignAnalytics, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.slug, c.title,
		       COUNT(k.id),
		       COALESCE(SUM(CASE WHEN k.date_str = ? THEN 1 ELSE 0 END), 0),
		       COALESCE(MAX(k.created_at), '')
		FROM campaigns c
		LEFT JOIN campaign_cards k ON k.campaign_slug = c.slug
		GROUP BY c.slug
		ORDER BY COUNT(k.id) DESC, c.id DESC
	`, today)
	if err != nil {
		return nil, fmt.Errorf("failed to query campaign analytics: %w", err)
	}
	defer rows.Close()

	list := []domain.CampaignAnalytics{}
	for rows.Next() {
		var item domain.CampaignAnalytics
		var lastCardAt string
		if err := rows.Scan(&item.Slug, &item.Title, &item.TotalCards, &item.CardsToday, &lastCardAt); err != nil {
			return nil, fmt.Errorf("failed to scan campaign analytics: %w", err)
		}
		if t := parseDBTime(lastCardAt); !t.IsZero() {
			item.LastCardAt = t.Format(time.RFC3339)
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *CardRepository) GetDashboardStats(ctx context.Context, today string) (*domain.DashboardStats, error) {
	stats := &domain.DashboardStats{DeviceStats: make(map[string]int)}

	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*), COALESCE(SUM(active), 0) FROM campaigns",
	).Scan(&stats.TotalCampaigns, &stats.ActiveCampaigns); err != nil {
		return nil, fmt.Errorf("failed to count campaigns: %w", err)
	}

	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*), COALESCE(SUM(CASE WHEN date_str = ? THEN 1 ELSE 0 END), 0) FROM campaign_cards", today,
	).Scan(&stats.TotalCards, &stats.CardsToday); err != nil {
		return nil, fmt.Errorf("failed to count cards: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		"SELECT COALESCE(NULLIF(device, ''), 'Unknown'), COUNT(*) FROM campaign_cards GROUP BY 1")
	if err != nil {
		return nil, fmt.Errorf("failed to query device stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var device string
		var count int
		if err := rows.Scan(&device, &count); err != nil {
			return nil, fmt.Errorf("failed to scan device stats: %w", err)
		}
		stats.DeviceStats[device] = count
	}
	return stats, rows.Err()
}

func cardFilter(q CardQuery) (string, []any) {
	var conds []string
	var args []any

	if q.Slug != "" {
		conds = append(conds, "campaign_slug = ?")
		args = append(args, q.Slug)
	}
	if q.Search != "" {
		pattern := "%" + escapeLike(q.Search) + "%"
		conds = append(conds, `(to_name LIKE ? ESCAPE '\' OR from_name LIKE ? ESCAPE '\' OR message LIKE ? ESCAPE '\' OR field_values LIKE ? ESCAPE '\')`)
		args = append(args, pattern, pattern, pattern, pattern)
	}

	if len(conds) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func scanCard(row interface{ Scan(dest ...any) error }) (domain.Card, error) {
	var c domain.Card
	var createdAt string
	var fieldValues sql.NullString

	if err := row.Scan(
		&c.ID, &c.CampaignSlug, &c.FromName, &c.ToName, &c.Message, &c.Heading,
		&c.Lang, &fieldValues, &c.Device, &c.DateStr, &c.TimeStr, &createdAt,
	); err != nil {
		return c, fmt.Errorf("failed to scan card: %w", err)
	}

	if fieldValues.Valid && fieldValues.String != "" {
		if err := json.Unmarshal([]byte(fieldValues.String), &c.FieldValues); err != nil {
			log.Printf("card %d has invalid field_values JSON: %v", c.ID, err)
		}
	}
	c.CreatedAt = parseDBTime(createdAt)
	return c, nil
}
