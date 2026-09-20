-- 000002_dynamic_templates.down.sql
-- SQLite doesn't natively drop columns cleanly before 3.35, but migration placeholder kept for completeness
DROP INDEX IF EXISTS idx_campaign_cards_lang;
DROP INDEX IF EXISTS idx_campaign_cards_created;
