-- 000002_dynamic_templates.up.sql
-- Add dual template support to campaigns (template_ar, template_en)
ALTER TABLE campaigns ADD COLUMN template_ar TEXT;
ALTER TABLE campaigns ADD COLUMN template_en TEXT;

-- Add lang and dynamic field_values to campaign_cards
ALTER TABLE campaign_cards ADD COLUMN lang TEXT DEFAULT 'ar';
ALTER TABLE campaign_cards ADD COLUMN field_values TEXT;

CREATE INDEX IF NOT EXISTS idx_campaign_cards_lang ON campaign_cards(lang);
CREATE INDEX IF NOT EXISTS idx_campaign_cards_created ON campaign_cards(created_at DESC);
