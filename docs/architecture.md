# System Architecture: Cards Platform

## High-Level Topology

```
[ Browser / Client ]
       │
       ▼ (Port 8090 / 80 / 443)
┌───────────────────────────────────────────────┐
│              Nginx Container                  │
│                                               │
│  ├── /api/v1/*   ──────► Proxy to Go API:8080 │
│  ├── /fonts/*    ──────► Cached Static Fonts  │
│  └── /*          ──────► SvelteKit Static SPA │
└───────────────────────┬───────────────────────┘
                        │
                        ▼ (Internal Docker Network)
┌───────────────────────────────────────────────┐
│            Go Chi API Container               │
│                                               │
│  ├── Chi Router + Middlewares (CORS, Limiter) │
│  ├── Clean Layered Services                   │
│  └── SQLite DB Driver (WAL Mode, zero CGO)   │
└───────────────────────┬───────────────────────┘
                        │
                        ▼ (Volume mount /data)
┌───────────────────────────────────────────────┐
│         SQLite Database: /data/cards.db       │
│  ├── campaigns                                │
│  ├── campaign_cards                           │
│  └── schema_migrations                        │
└───────────────────────────────────────────────┘
```

## Key Characteristics
- **Bilingual & Directional:** Built-in Arabic (RTL, default) and English (LTR). No hardcoded strings.
- **Auto-generated Random Slugs:** URL-safe 8-character cryptographic random links for occasions (`/cards/[slug]`).
- **Zero-Friction Privacy:** Card creations are downloaded locally as PNG and immediately wiped from browser state.
- **SQLite with WAL:** Non-blocking concurrent reads with `busy_timeout=5000` and transactional consistency.
- **Secure Reverse Proxy:** Nginx acts as the single exposed gateway applying security headers (`X-Frame-Options`, `X-Content-Type-Options`, `X-XSS-Protection`).
