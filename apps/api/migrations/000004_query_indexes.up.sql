-- 000004_query_indexes.up.sql
-- Serve the admin card listings (per campaign, newest first) and the "cards today" counters from indexes.
CREATE INDEX IF NOT EXISTS idx_campaign_cards_slug_id ON campaign_cards(campaign_slug, id DESC);
CREATE INDEX IF NOT EXISTS idx_campaign_cards_date_str ON campaign_cards(date_str);
