# Occasion Cards (Gratitude E-Cards Platform)

A bilingual (Arabic RTL / English LTR) platform where employees design personalised greeting cards for company
occasions and download them as high-resolution PNGs, and where administrators create occasions, position text on the
artwork with a visual editor, and follow participation analytics.

- **Web** – SvelteKit 5 static SPA (`apps/web`), rendered entirely in the browser (canvas card renderer).
- **API** – Go + chi + SQLite (WAL, pure-Go driver, no CGO) (`apps/api`).
- **Edge** – nginx serves the SPA, proxies `/api`, applies security headers and rate limits (`infra/`).

## Quick start (Docker)

```bash
cp .env.example .env        # then set ADMIN_PASSWORD and SESSION_SECRET (openssl rand -hex 32)
make up                     # builds both images and starts the stack
```

The site is served on `http://<host>:8090` (change with `HTTP_PORT`). The API **refuses to start in production with the
default admin password or a short/default `SESSION_SECRET`** – this is intentional.

| URL | What |
| --- | --- |
| `/` | Public list of active occasions |
| `/cards/<slug>` | Card designer for one occasion |
| `/login` → `/admin/dashboard` | Admin analytics, occasions, template builder, card log |

## Local development

```bash
make dev-api    # Go API on :8080 (APP_ENV=development allows the built-in dev credentials: password "admin")
make dev-web    # Vite dev server on :5173, proxies /api to :8080
```

Requirements: Go 1.23+, Node 24+.

## Configuration

| Variable | Default | Notes |
| --- | --- | --- |
| `ADMIN_PASSWORD` | – | **Required** in production. |
| `SESSION_SECRET` | – | **Required** in production, 32+ random characters. |
| `APP_ENV` | `production` | `development` relaxes the credential checks. |
| `HTTP_PORT` | `8090` | Public port of the web container. |
| `APP_TIMEZONE` | `Asia/Riyadh` | Timezone for card dates and the "cards today" counters. |
| `TRUSTED_ORIGINS` | *(empty)* | Enables CORS for these origins. Not needed when SPA and API share an origin. |
| `TRUST_PROXY` | `false` (`true` in compose) | Read the client IP from `X-Real-IP` for rate limiting. Only enable behind a proxy that overwrites the header. |
| `DATABASE_PATH`, `MIGRATIONS_PATH`, `PORT` | see `.env.example` | API internals. |

## Testing & quality

```bash
make test    # Go unit + integration tests (real router, real SQLite, no mocks)
make lint    # gofmt, go vet, svelte-check, i18n key parity
```

CI (`.github/workflows/ci.yml`) also builds both Docker images and validates the nginx configuration.

## Operations

```bash
make backup-db                                   # verified online backup → ./backups (30-day retention)
make restore-db FILE=backups/cards_backup_*.db.gz  # stops the API, keeps a safety copy, restores, restarts
```

See [docs/deployment.md](docs/deployment.md), [docs/architecture.md](docs/architecture.md) and [docs/api.md](docs/api.md).

## Notes

- **Brand fonts**: the card renderer prefers the licensed Luma/Karbon fonts. They are not in the repository; drop
  `Luma-SemiBold.ttf`, `Luma-Regular.ttf`, `Karbon-Bold.ttf` and `Karbon-Regular.ttf` into `apps/web/static/fonts/`.
  Without them cards use Tajawal (Arabic) and Instrument Sans (Latin).
- **Artwork** is optimised in the browser before upload (max 1080×1350, JPEG unless it uses transparency) and a small
  thumbnail is generated for the public listing.
