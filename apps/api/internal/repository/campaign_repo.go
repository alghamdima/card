package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"cards-api/internal/domain"
)

type CampaignRepository struct {
	db *sql.DB
}

func NewCampaignRepository(db *sql.DB) *CampaignRepository {
	return &CampaignRepository{db: db}
}

func (r *CampaignRepository) ExistsSlug(ctx context.Context, slug string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM campaigns WHERE slug = ?", slug).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CampaignRepository) GetPublicBySlug(ctx context.Context, slug string) (*domain.Campaign, error) {
	query := `
		SELECT id, slug, title, lang, text_color, head_color, boxes, image, thumb, template_ar, template_en, active, created_at
		FROM campaigns
		WHERE slug = ? AND active = 1
	`
	var c domain.Campaign
	var boxesRaw string
	var createdAtStr string
	var activeInt int
	var templateARRaw, templateENRaw sql.NullString

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&c.ID, &c.Slug, &c.Title, &c.Lang, &c.TextColor, &c.HeadColor,
		&boxesRaw, &c.Image, &c.Thumb, &templateARRaw, &templateENRaw, &activeInt, &createdAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query public campaign: %w", err)
	}

	c.Active = activeInt == 1
	c.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	if c.CreatedAt.IsZero() {
		c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	}

	_ = json.Unmarshal([]byte(boxesRaw), &c.Boxes)
	if templateARRaw.Valid && templateARRaw.String != "" {
		var tar domain.TemplateVariant
		if err := json.Unmarshal([]byte(templateARRaw.String), &tar); err == nil {
			c.TemplateAR = &tar
		}
	}
	if templateENRaw.Valid && templateENRaw.String != "" {
		var ten domain.TemplateVariant
		if err := json.Unmarshal([]byte(templateENRaw.String), &ten); err == nil {
			c.TemplateEN = &ten
		}
	}

	return &c, nil
}

func (r *CampaignRepository) GetAllPublic(ctx context.Context) ([]domain.CampaignSummary, error) {
	query := `
		SELECT c.slug, c.title, c.lang, c.text_color, c.head_color, c.thumb, c.active, c.created_at,
		       COUNT(k.id) as card_count
		FROM campaigns c
		LEFT JOIN campaign_cards k ON c.slug = k.campaign_slug
		WHERE c.active = 1
		GROUP BY c.slug
		ORDER BY c.id DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all public campaigns: %w", err)
	}
	defer rows.Close()

	var list []domain.CampaignSummary
	for rows.Next() {
		var s domain.CampaignSummary
		var createdAtStr string
		var activeInt int
		if err := rows.Scan(&s.Slug, &s.Title, &s.Lang, &s.TextColor, &s.HeadColor, &s.Thumb, &activeInt, &createdAtStr, &s.TotalCards); err != nil {
			return nil, err
		}
		s.Active = activeInt == 1
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		if s.CreatedAt.IsZero() {
			s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		}
		list = append(list, s)
	}

	return list, nil
}

func (r *CampaignRepository) GetAllAdmin(ctx context.Context) ([]domain.CampaignSummary, error) {
	query := `
		SELECT c.slug, c.title, c.lang, c.text_color, c.head_color, c.thumb, c.active, c.created_at,
		       COUNT(k.id) as card_count
		FROM campaigns c
		LEFT JOIN campaign_cards k ON c.slug = k.campaign_slug
		GROUP BY c.slug
		ORDER BY c.id DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all admin campaigns: %w", err)
	}
	defer rows.Close()

	var list []domain.CampaignSummary
	for rows.Next() {
		var s domain.CampaignSummary
		var createdAtStr string
		var activeInt int
		if err := rows.Scan(&s.Slug, &s.Title, &s.Lang, &s.TextColor, &s.HeadColor, &s.Thumb, &activeInt, &createdAtStr, &s.TotalCards); err != nil {
			return nil, err
		}
		s.Active = activeInt == 1
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		if s.CreatedAt.IsZero() {
			s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		}
		list = append(list, s)
	}

	return list, nil
}

func (r *CampaignRepository) GetBySlug(ctx context.Context, slug string) (*domain.Campaign, error) {
	query := `
		SELECT id, slug, title, lang, text_color, head_color, boxes, image, thumb, template_ar, template_en, active, created_at
		FROM campaigns
		WHERE slug = ?
	`
	var c domain.Campaign
	var boxesRaw string
	var createdAtStr string
	var activeInt int
	var templateARRaw, templateENRaw sql.NullString

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&c.ID, &c.Slug, &c.Title, &c.Lang, &c.TextColor, &c.HeadColor,
		&boxesRaw, &c.Image, &c.Thumb, &templateARRaw, &templateENRaw, &activeInt, &createdAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query campaign by slug: %w", err)
	}

	c.Active = activeInt == 1
	c.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	if c.CreatedAt.IsZero() {
		c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	}

	_ = json.Unmarshal([]byte(boxesRaw), &c.Boxes)
	if templateARRaw.Valid && templateARRaw.String != "" {
		var tar domain.TemplateVariant
		if err := json.Unmarshal([]byte(templateARRaw.String), &tar); err == nil {
			c.TemplateAR = &tar
		}
	}
	if templateENRaw.Valid && templateENRaw.String != "" {
		var ten domain.TemplateVariant
		if err := json.Unmarshal([]byte(templateENRaw.String), &ten); err == nil {
			c.TemplateEN = &ten
		}
	}

	return &c, nil
}

func (r *CampaignRepository) Create(ctx context.Context, c *domain.Campaign) error {
	boxesBytes, err := json.Marshal(c.Boxes)
	if err != nil {
		return fmt.Errorf("failed to marshal boxes: %w", err)
	}

	var templateARStr, templateENStr sql.NullString
	if c.TemplateAR != nil {
		b, err := json.Marshal(c.TemplateAR)
		if err == nil {
			templateARStr = sql.NullString{String: string(b), Valid: true}
		}
	}
	if c.TemplateEN != nil {
		b, err := json.Marshal(c.TemplateEN)
		if err == nil {
			templateENStr = sql.NullString{String: string(b), Valid: true}
		}
	}

	query := `
		INSERT INTO campaigns (slug, title, lang, text_color, head_color, boxes, image, thumb, template_ar, template_en, active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	activeInt := 0
	if c.Active {
		activeInt = 1
	}

	res, err := r.db.ExecContext(ctx, query,
		c.Slug, c.Title, c.Lang, c.TextColor, c.HeadColor,
		string(boxesBytes), c.Image, c.Thumb, templateARStr, templateENStr, activeInt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert campaign: %w", err)
	}

	id, _ := res.LastInsertId()
	c.ID = id
	return nil
}

func (r *CampaignRepository) Update(ctx context.Context, c *domain.Campaign) error {
	boxesBytes, err := json.Marshal(c.Boxes)
	if err != nil {
		return fmt.Errorf("failed to marshal boxes: %w", err)
	}

	activeInt := 0
	if c.Active {
		activeInt = 1
	}

	var templateARStr, templateENStr sql.NullString
	if c.TemplateAR != nil {
		b, err := json.Marshal(c.TemplateAR)
		if err == nil {
			templateARStr = sql.NullString{String: string(b), Valid: true}
		}
	}
	if c.TemplateEN != nil {
		b, err := json.Marshal(c.TemplateEN)
		if err == nil {
			templateENStr = sql.NullString{String: string(b), Valid: true}
		}
	}

	query := `
		UPDATE campaigns
		SET title = ?, lang = ?, text_color = ?, head_color = ?, boxes = ?, image = CASE WHEN ? != '' THEN ? ELSE image END, thumb = CASE WHEN ? != '' THEN ? ELSE thumb END, template_ar = ?, template_en = ?, active = ?
		WHERE slug = ?
	`
	_, err = r.db.ExecContext(ctx, query,
		c.Title, c.Lang, c.TextColor, c.HeadColor, string(boxesBytes), c.Image, c.Image, c.Thumb, c.Thumb, templateARStr, templateENStr, activeInt, c.Slug,
	)
	if err != nil {
		return fmt.Errorf("failed to update campaign: %w", err)
	}
	return nil
}

func (r *CampaignRepository) Delete(ctx context.Context, slug string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM campaign_cards WHERE campaign_slug = ?", slug); err != nil {
		return fmt.Errorf("failed to delete campaign cards: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM campaigns WHERE slug = ?", slug); err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}

	return tx.Commit()
}
