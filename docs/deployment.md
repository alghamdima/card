# Deployment & Operations Guide

## 1. Quick Start with Docker Compose

```bash
cp .env.example .env     # set ADMIN_PASSWORD and SESSION_SECRET (openssl rand -hex 32)
make up                  # builds the API and web images and starts the stack
```

The stack has two containers: `cards-api` (Go, not published) and `cards-web` (nginx serving the SPA and proxying
`/api`, published on `HTTP_PORT`, default 8090). The SPA is built **inside** the web image, so there is no host-side
build step and no bind-mounted `build/` folder.

- **Public Portal:** `http://<server-ip>:8090/`
- **Occasion Direct Link:** `http://<server-ip>:8090/cards/<slug>`
- **Admin Login:** `http://<server-ip>:8090/login`

`make up` fails early when `.env` is missing, and the API refuses to boot in production with the default admin
password or a weak `SESSION_SECRET` (see the README configuration table).

### Upgrading
`git pull && make up`. Database migrations run automatically at API start; **a failing migration stops the boot**
(the API never serves traffic on a half-migrated database). Take a backup first (`make backup-db`).
The API image fixes ownership of the `/data` volume on start, so volumes created by older root-running images keep working.

## 2. Stopping & Logs

```bash
make down    # stop and remove containers (the cards_db_data volume is kept)
make logs    # follow logs
```

## 3. Database Backups & Restorations

The SQLite database lives in the Docker volume `cards_db_data`, mounted at `/data/cards.db`.

```bash
make backup-db                                    # → ./backups/cards_backup_<timestamp>.db.gz
make restore-db FILE=./backups/cards_backup_YYYYMMDD_HHMMSS.db.gz
```

- Backups use SQLite's online `.backup` API (consistent, no downtime) and are integrity-checked before being kept.
  Backups older than `KEEP_DAYS` (default 30) are pruned; schedule the command with cron, e.g.
  `0 2 * * * cd /opt/cards && make backup-db`.
- Restore **stops the API**, verifies the file, saves a `pre_restore_*.db` copy of the current database in `./backups`,
  replaces the database (removing stale `-wal`/`-shm` files) and starts the API again.

## 4. Production Security Checklist

1. Strong, unique `ADMIN_PASSWORD` and a random 32+ character `SESSION_SECRET` (enforced by the API).
2. Terminate TLS in front of the stack (corporate reverse proxy / load balancer on 443 → `HTTP_PORT`). The API marks the
   session cookie `Secure` automatically when `X-Forwarded-Proto: https` is present.
3. **Client IPs and rate limiting**: nginx passes `$remote_addr` to the API as `X-Real-IP`. If another proxy sits in front
   of nginx, all users appear to come from that proxy and share one rate-limit bucket. In that case configure nginx's
   `real_ip_header` / `set_real_ip_from` for the proxy's address so `$remote_addr` is the real client.
4. CORS is disabled by default (same origin). Set `TRUSTED_ORIGINS` only when the SPA is hosted on another origin.
5. The API container is not published; only nginx is reachable. Keep it that way.

## 5. Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| `docker compose` says `Set ADMIN_PASSWORD in .env` | `.env` is missing or incomplete – copy `.env.example`. |
| API exits with `must be changed from the default value` | Default credentials in production – set real ones. |
| "Too many requests" on login | Login is limited to 5 attempts, then one per 12 s per client. |
| Card dates/"today" off by hours | Set `APP_TIMEZONE` (default `Asia/Riyadh`). |
| Cards render with a fallback font | Brand fonts are optional; see the README notes. |
