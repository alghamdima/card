# System Architecture: Cards Platform

## High-Level Topology

```
[ Browser / Client ]
       │
       ▼ (HTTP_PORT, default 8090 → TLS terminator in production)
┌───────────────────────────────────────────────┐
│           cards-web (nginx container)         │
│                                               │
│  ├── /api/*         ──► proxy to cards-api    │
│  ├── /_app/immutable ─► static, 1y cache      │
│  └── /*             ──► SvelteKit static SPA  │
│      (index.html revalidated on every load)   │
└───────────────────────┬───────────────────────┘
                        │ internal Docker network (API not published)
                        ▼
┌───────────────────────────────────────────────┐
│            cards-api (Go, chi)                │
│                                               │
│  ├── middleware: request-id, logger, recover, │
│  │   security headers, per-route rate limits  │
│  ├── handlers → services → repositories       │
│  └── SQLite (WAL, busy_timeout, 4 conns)      │
└───────────────────────┬───────────────────────┘
                        │ volume cards_db_data → /data
                        ▼
              /data/cards.db (campaigns, campaign_cards, schema_migrations)
```

## Key Characteristics

- **Bilingual & directional:** Arabic (RTL, default) and English (LTR); every UI string lives in
  `apps/web/src/lib/messages/{ar,en}.json` (key parity is enforced by review/tests), dates use the Gregorian calendar
  with Latin digits in both languages, and API error codes are translated client-side.
- **Client-side card rendering:** the card is drawn on a 1080×1350 canvas in the browser and exported as PNG; the server
  only stores anonymous participation records (no card image is ever uploaded).
- **Template model:** each occasion has an Arabic and an English *template* (artwork + positioned text fields). Artwork is
  optimised in the browser before upload; the API de-duplicates identical images in its responses.
- **Migrations:** plain SQL files applied at boot inside transactions; a failure aborts startup.
- **SQLite tuning:** WAL, `busy_timeout=5000`, `_txlock=immediate`, and a small connection pool so slow reads (e.g. CSV
  export, streamed row by row) do not block other requests.
- **Time handling:** card `date`/`time` are stored in `APP_TIMEZONE` (default `Asia/Riyadh`); "cards today" counters use
  the same timezone. `created_at` is UTC.
- **Security:** production refuses default credentials; login and card submission are rate limited per client IP (both in
  nginx and in the API); request bodies are size-limited and validated; CSV export neutralises formula injection; nginx
  applies CSP and other security headers.
- **Auth:** stateless HMAC-signed 24-hour bearer token (also set as an `HttpOnly` cookie); the SPA treats only a `401` as
  an expired session, never a transient network error.

## Repository Layout

```
apps/api   Go service (cmd/server, internal/{config,database,domain,http,repository,service}, migrations)
apps/web   SvelteKit SPA (src/routes, src/lib/{api,components,i18n,messages,stores,utils})
infra/     docker-compose.yml and nginx configuration
scripts/   backup / restore helpers
docs/      this documentation
```
