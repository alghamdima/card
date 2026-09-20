# Deployment & Operations Guide

## 1. Quick Start with Docker Compose

To launch the complete application stack (Go API, Nginx, and SvelteKit frontend):

```bash
# Build web and start containers in detached mode
make up
```

Access the application:
- **Public Portal:** `http://<server-ip>:8090/`
- **Occasion Direct Link:** `http://<server-ip>:8090/cards/<slug>`
- **Admin Dashboard:** `http://<server-ip>:8090/admin/dashboard`
- **Admin Login:** `http://<server-ip>:8090/login`

---

## 2. Stopping & Cleaning

To stop and remove containers:
```bash
make down
```

---

## 3. Database Backups & Restorations

The SQLite database is kept inside the Docker volume `cards_db_data` mounted at `/data/cards.db`.

### Creating a Safe Online Backup:
```bash
make backup-db
# Or directly:
./scripts/backup-sqlite.sh
```
This runs SQLite's native `.backup` command in WAL mode (zero read locks, non-blocking) and stores a compressed timestamped file in `./backups/`.

### Restoring from Backup:
```bash
make restore-db FILE=./backups/cards_backup_YYYYMMDD_HHMMSS.db.gz
# Or directly:
./scripts/restore-sqlite.sh ./backups/cards_backup_YYYYMMDD_HHMMSS.db.gz
```

---

## 4. Production Security Recommendations
1. Modify `ADMIN_PASSWORD` in `.env` (or compose environment variables) to a strong unique value.
2. Update `SESSION_SECRET` with a 32+ character random secret.
3. Place Nginx behind your corporate reverse proxy or SSL terminator (Port 443 / HTTPS).
