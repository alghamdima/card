-- 000001_init_schema.up.sql
CREATE TABLE IF NOT EXISTS campaigns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    lang TEXT NOT NULL DEFAULT 'ar',
    text_color TEXT NOT NULL DEFAULT '#FFFFFF',
    head_color TEXT NOT NULL DEFAULT '#FFCD00',
    boxes TEXT NOT NULL,
    image MEDIUMTEXT NOT NULL,
    thumb MEDIUMTEXT,
    active INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS campaign_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_slug TEXT NOT NULL,
    from_name TEXT,
    to_name TEXT,
    message TEXT,
    heading TEXT,
    device TEXT,
    date_str TEXT,
    time_str TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(campaign_slug) REFERENCES campaigns(slug) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_campaign_cards_slug ON campaign_cards(campaign_slug);
CREATE INDEX IF NOT EXISTS idx_campaigns_active ON campaigns(active);
CREATE INDEX IF NOT EXISTS idx_campaigns_created ON campaigns(created_at DESC);
