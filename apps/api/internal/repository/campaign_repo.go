package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"cards-api/internal/domain"
)

// ErrDuplicateSlug is returned when a campaign slug is already taken.
var ErrDuplicateSlug = errors.New("duplicate campaign slug")

type CampaignRepository struct {
	db *sql.DB
}

func NewCampaignRepository(db *sql.DB) *CampaignRepository {
	return &CampaignRepository{db: db}
}

const campaignColumns = `id, slug, title, COALESCE(title_ar, title), COALESCE(title_en, title), lang,
	text_color, head_color, boxes, image, COALESCE(thumb, ''), template_ar, template_en, active, created_at`

func (r *CampaignRepository) ExistsSlug(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM campaigns WHERE slug = ?)", slug).Scan(&exists)
	return exists, err
}

// GetPublicBySlug returns the campaign only when it is active.
func (r *CampaignRepository) GetPublicBySlug(ctx context.Context, slug string) (*domain.Campaign, error) {
	row := r.db.QueryRowContext(ctx, "SELECT "+campaignColumns+" FROM campaigns WHERE slug = ? AND active = 1", slug)
	return scanCampaignRow(row)
}

func (r *CampaignRepository) GetBySlug(ctx context.Context, slug string) (*domain.Campaign, error) {
	row := r.db.QueryRowContext(ctx, "SELECT "+campaignColumns+" FROM campaigns WHERE slug = ?", slug)
	return scanCampaignRow(row)
}

func (r *CampaignRepository) GetAllPublic(ctx context.Context) ([]domain.PublicCampaignSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT slug, title, COALESCE(title_ar, title), COALESCE(title_en, title), lang,
		       text_color, head_color, COALESCE(thumb, ''), created_at
		FROM campaigns
		WHERE active = 1
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query public campaigns: %w", err)
	}
	defer rows.Close()

	list := []domain.PublicCampaignSummary{}
	for rows.Next() {
		var s domain.PublicCampaignSummary
		var createdAt string
		if err := rows.Scan(&s.Slug, &s.Title, &s.TitleAR, &s.TitleEN, &s.Lang, &s.TextColor, &s.HeadColor, &s.Thumb, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan public campaign: %w", err)
		}
		s.CreatedAt = parseDBTime(createdAt)
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *CampaignRepository) GetAllAdmin(ctx context.Context) ([]domain.CampaignSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.slug, c.title, COALESCE(c.title_ar, c.title), COALESCE(c.title_en, c.title), c.lang,
		       c.text_color, c.head_color, COALESCE(c.thumb, ''), c.active, c.created_at,
		       (SELECT COUNT(*) FROM campaign_cards k WHERE k.campaign_slug = c.slug)
		FROM campaigns c
		ORDER BY c.id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query admin campaigns: %w", err)
	}
	defer rows.Close()

	list := []domain.CampaignSummary{}
	for rows.Next() {
		var s domain.CampaignSummary
		var createdAt string
		var active int
		if err := rows.Scan(&s.Slug, &s.Title, &s.TitleAR, &s.TitleEN, &s.Lang, &s.TextColor, &s.HeadColor, &s.Thumb, &active, &createdAt, &s.TotalCards); err != nil {
			return nil, fmt.Errorf("failed to scan admin campaign: %w", err)
		}
		s.Active = active == 1
		s.CreatedAt = parseDBTime(createdAt)
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *CampaignRepository) Create(ctx context.Context, c *domain.Campaign) error {
	boxes, templateAR, templateEN, err := encodeCampaignJSON(c)
	if err != nil {
		return err
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO campaigns (slug, title, title_ar, title_en, lang, text_color, head_color, boxes, image, thumb, template_ar, template_en, active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		c.Slug, c.Title, c.TitleAR, c.TitleEN, c.Lang, c.TextColor, c.HeadColor,
		boxes, c.Image, c.Thumb, templateAR, templateEN, boolToInt(c.Active),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateSlug
		}
		return fmt.Errorf("failed to insert campaign: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to read new campaign id: %w", err)
	}
	c.ID = id
	return nil
}

func (r *CampaignRepository) Update(ctx context.Context, c *domain.Campaign) error {
	boxes, templateAR, templateEN, err := encodeCampaignJSON(c)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE campaigns
		SET title = ?, title_ar = ?, title_en = ?, lang = ?, text_color = ?, head_color = ?, boxes = ?,
		    image = ?, thumb = ?, template_ar = ?, template_en = ?, active = ?
		WHERE slug = ?
	`,
		c.Title, c.TitleAR, c.TitleEN, c.Lang, c.TextColor, c.HeadColor, boxes,
		c.Image, c.Thumb, templateAR, templateEN, boolToInt(c.Active), c.Slug,
	)
	if err != nil {
		return fmt.Errorf("failed to update campaign: %w", err)
	}
	return nil
}

// Delete removes the campaign and its cards. It reports whether a campaign existed.
func (r *CampaignRepository) Delete(ctx context.Context, slug string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, "DELETE FROM campaign_cards WHERE campaign_slug = ?", slug); err != nil {
		return false, fmt.Errorf("failed to delete campaign cards: %w", err)
	}
	res, err := tx.ExecContext(ctx, "DELETE FROM campaigns WHERE slug = ?", slug)
	if err != nil {
		return false, fmt.Errorf("failed to delete campaign: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return affected > 0, nil
}

// scanCampaignRow returns (nil, nil) when the row does not exist.
func scanCampaignRow(row *sql.Row) (*domain.Campaign, error) {
	var c domain.Campaign
	var boxesRaw, createdAt string
	var active int
	var templateAR, templateEN sql.NullString

	err := row.Scan(
		&c.ID, &c.Slug, &c.Title, &c.TitleAR, &c.TitleEN, &c.Lang, &c.TextColor, &c.HeadColor,
		&boxesRaw, &c.Image, &c.Thumb, &templateAR, &templateEN, &active, &createdAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan campaign: %w", err)
	}

	c.Active = active == 1
	c.CreatedAt = parseDBTime(createdAt)

	if err := json.Unmarshal([]byte(boxesRaw), &c.Boxes); err != nil {
		log.Printf("campaign %q has invalid boxes JSON: %v", c.Slug, err)
	}
	c.TemplateAR = decodeTemplate(c.Slug, "template_ar", templateAR)
	c.TemplateEN = decodeTemplate(c.Slug, "template_en", templateEN)
	return &c, nil
}

func decodeTemplate(slug, column string, raw sql.NullString) *domain.TemplateVariant {
	if !raw.Valid || raw.String == "" {
		return nil
	}
	var v domain.TemplateVariant
	if err := json.Unmarshal([]byte(raw.String), &v); err != nil {
		log.Printf("campaign %q has invalid %s JSON: %v", slug, column, err)
		return nil
	}
	return &v
}

func encodeCampaignJSON(c *domain.Campaign) (boxes string, templateAR, templateEN sql.NullString, err error) {
	boxesBytes, err := json.Marshal(c.Boxes)
	if err != nil {
		return "", templateAR, templateEN, fmt.Errorf("failed to marshal boxes: %w", err)
	}
	if templateAR, err = encodeTemplate(c.TemplateAR); err != nil {
		return "", templateAR, templateEN, fmt.Errorf("failed to marshal Arabic template: %w", err)
	}
	if templateEN, err = encodeTemplate(c.TemplateEN); err != nil {
		return "", templateAR, templateEN, fmt.Errorf("failed to marshal English template: %w", err)
	}
	return string(boxesBytes), templateAR, templateEN, nil
}

func encodeTemplate(v *domain.TemplateVariant) (sql.NullString, error) {
	if v == nil {
		return sql.NullString{}, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return sql.NullString{}, err
	}
	return sql.NullString{String: string(b), Valid: true}, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
