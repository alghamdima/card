# Cards Platform - API Documentation

Base URL: `/api/v1`  
Data format: `application/json`

All responses follow the structured envelope:
```json
{
  "success": true,
  "data": { ... },
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable description"
  }
}
```

Common error codes: `VALIDATION_ERROR` (400), `UNAUTHORIZED` (401), `NOT_FOUND` / `CAMPAIGN_NOT_FOUND` (404),
`CAMPAIGN_EXISTS` (409), `PAYLOAD_TOO_LARGE` (413), `RATE_LIMIT_EXCEEDED` (429), `INTERNAL_SERVER_ERROR` (500).
Unexpected server errors never include internal details in the message; they are logged with the `X-Request-ID`
returned in the response headers.

---

## 1. Public Endpoints

### Health Check (readiness)
- `GET /health` – pings the database; `503 UNHEALTHY` when it does not answer.
- Response: `{"success": true, "data": {"status": "healthy", "timestamp": "...", "service": "cards-api"}}`

### List Active Occasions
- `GET /campaigns`
- Response: `{"success": true, "data": {"campaigns": [{"slug", "title", "titleAR", "titleEN", "thumb", ...}]}}`
- Participation counters are **not** exposed publicly.

### Get Single Occasion Details
- `GET /campaigns/{slug}`
- Returns the templates (`templateAR` / `templateEN`) with their dynamic text `fields`. To keep the payload small, a
  template `image` identical to the occasion's `image` is returned as `""` – clients fall back to `image`.
- Both public GETs send `Cache-Control: no-cache` + `ETag`; send `If-None-Match` to get a cheap `304`.

### Submit Card Creation (analytics)
- `POST /campaigns/{slug}/cards` (rate limited: burst of 30, then 1 per 2 s per client; body limit 64 KB)
- Body:
```json
{
  "to": "Jane Smith",
  "from": "John Doe",
  "message": "Happy National Day!",
  "lang": "en",
  "fieldValues": { "emp_name": "Jane Smith", "job_title": "Engineer" },
  "device": "iPhone"
}
```
- At least one of `to`, `message` or a non-empty field value is required. Text is trimmed and length-limited;
  unknown `device` values are stored as `Other`; an empty `from` is stored as `Anonymous`.
- Response: `{"success": true, "data": {"saved": true, "id": 12}}`

---

## 2. Authentication

### Admin Login
- `POST /auth/login` (rate limited: 5 attempts, then 1 every 12 s per client)
- Body: `{"password": "your-password"}`
- Response: `{"success": true, "data": {"token": "<expiryMs>.<signature>", "user": {"role": "admin"}}}` and an
  `HttpOnly` `admin_token` cookie (`Secure` when served over HTTPS). Tokens are valid for 24 hours.

### Admin Logout
- `POST /auth/logout`

### Check Auth Status
- `GET /auth/me` (Requires `Authorization: Bearer <token>` or the cookie)

---

## 3. Admin Protected Endpoints
*Headers: `Authorization: Bearer <token>`. Responses are `Cache-Control: no-store`.*

### Dashboard Metrics
- `GET /admin/dashboard` – total/active occasions, total cards, cards today (in `APP_TIMEZONE`), device breakdown.

### Analytics per occasion
- `GET /admin/analytics/overview`

### Occasions
- `GET /admin/campaigns` – list with `totalCards`
- `GET /admin/campaigns/{slug}` – details (same image de-duplication as the public endpoint)
- `POST /admin/campaigns` – create (random 8-character `slug` when omitted)
- `PUT /admin/campaigns/{slug}` – partial update: empty strings / omitted fields keep the stored value
- `DELETE /admin/campaigns/{slug}` – `404` when it does not exist

Create body (body limit 30 MB, images must be PNG/JPEG/WebP/GIF data URLs, ≤ 8 MB; `thumb` ≤ 1.5 MB):
```json
{
  "titleAR": "تهنئة العيد",
  "titleEN": "Eid Greeting",
  "image": "data:image/jpeg;base64,...",
  "thumb": "data:image/jpeg;base64,...",
  "templateAR": { "image": "", "fields": [ { "id": "emp_name", "label": "الاسم", "x": 230, "y": 620, "width": 620, "height": 70, "fontSize": 47, "color": "#FFFFFF", "weight": "bold", "align": "center", "order": 1 } ] },
  "templateEN": { "image": "", "fields": [] }
}
```
Validation: colors are `#RRGGBB`; a field `id` is 1–64 of `[A-Za-z0-9_-]` and unique; positions must be inside the
1080×1350 card; `fontSize` 8–300; at most 20 fields per template; `slug` is lowercase letters, digits and hyphens.
Updating `titleAR` also updates the primary `title`. Replacing `templateAR.image` also becomes the occasion `image`.

### Cards
- `GET /admin/campaigns/{slug}/cards?limit=50&offset=0&q=term` – one page (`limit` ≤ 100), newest first, with `total`.
  `q` searches recipient, sender, message and field values.
- `GET /admin/cards?limit=50&offset=0&q=term` – same across all occasions.
- `GET /admin/campaigns/{slug}/export` – streamed CSV (UTF-8 with BOM for Excel). Cells that start with `= + - @` are
  prefixed with `'` to prevent spreadsheet formula injection.
