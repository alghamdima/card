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

---

## 1. Public Endpoints

### Health Check
- `GET /health`
- Response: `{"success": true, "data": {"status": "healthy", "timestamp": "...", "service": "cards-api"}}`

### List Active Occasions
- `GET /campaigns`
- Response: `{"success": true, "data": {"campaigns": [...]}}`

### Get Single Occasion Details
- `GET /campaigns/{slug}`
- Response: `{"success": true, "data": { "slug": "...", "title": "...", "image": "...", "boxes": {...} }}`

### Submit Card Creation (Export telemetry)
- `POST /campaigns/{slug}/cards`
- Body:
```json
{
  "from": "John Doe",
  "to": "Jane Smith",
  "message": "Happy National Day!",
  "heading": "",
  "device": "iPhone"
}
```
- Response: `{"success": true, "data": {"saved": true, "id": 12}}`

---

## 2. Authentication

### Admin Login
- `POST /auth/login` (Rate limited)
- Body: `{"password": "your-password"}`
- Response: `{"success": true, "data": {"token": "timestamp.signature", "user": {"role": "admin"}}}`

### Admin Logout
- `POST /auth/logout`

### Check Auth Status
- `GET /auth/me` (Requires Bearer token or Cookie)

---

## 3. Admin Protected Endpoints
*Headers: `Authorization: Bearer <token>`*

### Dashboard Metrics
- `GET /admin/dashboard`
- Returns total occasions, active occasions, total cards, cards today, and device breakdown.

### List All Campaigns (Admin)
- `GET /admin/campaigns`

### Create Campaign (Auto-generated Random Slug)
- `POST /admin/campaigns`
- Body:
```json
{
  "title": "Eid 2026",
  "textColor": "#FFFFFF",
  "headColor": "#FFCD00",
  "image": "data:image/png;base64,...",
  "thumb": "data:image/png;base64,..."
}
```
*Note: If `slug` is omitted, the API automatically generates an 8-character URL-safe random string!*

### Update Campaign
- `PUT /admin/campaigns/{slug}`

### Delete Campaign
- `DELETE /admin/campaigns/{slug}`

### View Campaign Submissions
- `GET /admin/campaigns/{slug}/cards`

### View All Cards Submissions
- `GET /admin/cards?limit=50&offset=0`
