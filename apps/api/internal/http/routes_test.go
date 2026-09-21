package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cards-api/internal/config"
	"cards-api/internal/database"
	internalhttp "cards-api/internal/http"
	"cards-api/internal/repository"
	"cards-api/internal/service"
)

const (
	testPassword = "test-admin-pw"
	tinyPNG      = "data:image/png;base64,iVBORw0KGgo="
)

type testEnv struct {
	t      *testing.T
	server *httptest.Server
	token  string
}

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	db, err := database.Connect(filepath.Join(t.TempDir(), "cards.db"))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := database.RunMigrations(db, "../../migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	campaignRepo := repository.NewCampaignRepository(db)
	cardRepo := repository.NewCardRepository(db)
	authService := service.NewAuthService(testPassword, "test-secret-key-123")

	router := internalhttp.SetupRoutes(internalhttp.RouterParams{
		Config:          &config.Config{AppEnv: "test"},
		DB:              db,
		AuthService:     authService,
		CampaignService: service.NewCampaignService(campaignRepo),
		CardService:     service.NewCardService(cardRepo, campaignRepo, time.UTC),
	})
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	env := &testEnv{t: t, server: server}
	token, err := authService.Login(testPassword)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	env.token = token
	return env
}

// do sends a JSON request; admin requests carry the bearer token.
func (e *testEnv) do(method, path string, body any, admin bool) (int, envelope, []byte) {
	e.t.Helper()

	var reader *bytes.Reader
	switch b := body.(type) {
	case nil:
		reader = bytes.NewReader(nil)
	case []byte:
		reader = bytes.NewReader(b)
	default:
		raw, err := json.Marshal(b)
		if err != nil {
			e.t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, e.server.URL+path, reader)
	if err != nil {
		e.t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if admin {
		req.Header.Set("Authorization", "Bearer "+e.token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	var raw bytes.Buffer
	_, _ = raw.ReadFrom(resp.Body)
	var env envelope
	_ = json.Unmarshal(raw.Bytes(), &env)
	return resp.StatusCode, env, raw.Bytes()
}

func (e *testEnv) mustStatus(want int, method, path string, body any, admin bool) envelope {
	e.t.Helper()
	status, env, raw := e.do(method, path, body, admin)
	if status != want {
		e.t.Fatalf("%s %s: status = %d, want %d (body: %.300s)", method, path, status, want, raw)
	}
	return env
}

func (e *testEnv) createCampaign(extra map[string]any) string {
	e.t.Helper()
	body := map[string]any{
		"titleAR":    "تهنئة العيد",
		"titleEN":    "Eid Greeting",
		"image":      tinyPNG,
		"thumb":      tinyPNG,
		"templateAR": map[string]any{"image": tinyPNG, "fields": []map[string]any{validField("emp_name")}},
		"templateEN": map[string]any{"image": tinyPNG, "fields": []map[string]any{validField("emp_name")}},
	}
	for k, v := range extra {
		body[k] = v
	}
	env := e.mustStatus(http.StatusCreated, "POST", "/api/v1/admin/campaigns", body, true)
	var camp struct {
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(env.Data, &camp); err != nil || camp.Slug == "" {
		e.t.Fatalf("create response has no slug: %s", env.Data)
	}
	return camp.Slug
}

func validField(id string) map[string]any {
	return map[string]any{
		"id": id, "name": id, "label": "Name", "x": 230, "y": 620, "width": 620, "height": 70,
		"fontSize": 47, "color": "#FFFFFF", "weight": "bold", "align": "center", "order": 1,
	}
}

func TestHealth(t *testing.T) {
	env := newTestEnv(t)
	env.mustStatus(http.StatusOK, "GET", "/api/v1/health", nil, false)
}

func TestUnknownRouteReturnsJSON404(t *testing.T) {
	env := newTestEnv(t)
	got := env.mustStatus(http.StatusNotFound, "GET", "/api/v1/nope", nil, false)
	if got.Error == nil || got.Error.Code != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND envelope, got %+v", got)
	}
}

func TestAdminRoutesRequireAuth(t *testing.T) {
	env := newTestEnv(t)
	env.mustStatus(http.StatusUnauthorized, "GET", "/api/v1/admin/campaigns", nil, false)
	env.mustStatus(http.StatusUnauthorized, "GET", "/api/v1/admin/dashboard", nil, false)
}

// Regression: the limiter used to key on RemoteAddr (IP:port), so every new
// connection got a fresh budget and brute-forcing the password was unthrottled.
func TestLoginRateLimitIgnoresSourcePort(t *testing.T) {
	env := newTestEnv(t)

	var lastStatus int
	for i := 0; i < 8; i++ {
		// A fresh client per attempt forces a new connection (new source port).
		client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
		resp, err := client.Post(env.server.URL+"/api/v1/auth/login", "application/json", strings.NewReader(`{"password":"wrong"}`))
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		lastStatus = resp.StatusCode
	}
	if lastStatus != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after repeated failed logins, got %d", lastStatus)
	}
}

func TestLoginSuccessAndToken(t *testing.T) {
	env := newTestEnv(t)
	got := env.mustStatus(http.StatusOK, "POST", "/api/v1/auth/login", map[string]string{"password": testPassword}, false)
	var data struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(got.Data, &data); err != nil || data.Token == "" {
		t.Fatalf("expected token in login response: %s", got.Data)
	}
	status, _, _ := env.do("GET", "/api/v1/auth/me", nil, true)
	if status != http.StatusOK {
		t.Fatalf("/auth/me with valid token: %d", status)
	}
}

// Regression: validation errors were wrapped with %w but compared with ==,
// so they surfaced as 500s carrying internal error text.
func TestCreateCampaignValidation(t *testing.T) {
	env := newTestEnv(t)

	cases := map[string]map[string]any{
		"missing title":  {"image": tinyPNG},
		"missing image":  {"titleAR": "x"},
		"bad image":      {"titleAR": "x", "image": "https://evil.example/x.png"},
		"bad color":      {"titleAR": "x", "image": tinyPNG, "textColor": "red"},
		"bad slug":       {"titleAR": "x", "image": tinyPNG, "slug": "a/b?c"},
		"bad lang":       {"titleAR": "x", "image": tinyPNG, "lang": "fr"},
		"bad field size": {"titleAR": "x", "image": tinyPNG, "templateAR": map[string]any{"fields": []map[string]any{{"id": "a", "width": 5, "height": 5}}}},
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			status, got, raw := env.do("POST", "/api/v1/admin/campaigns", body, true)
			if status != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (body: %s)", status, raw)
			}
			if got.Error == nil || got.Error.Code != "VALIDATION_ERROR" {
				t.Fatalf("expected VALIDATION_ERROR, got %s", raw)
			}
		})
	}
}

func TestCampaignLifecycle(t *testing.T) {
	env := newTestEnv(t)
	slug := env.createCampaign(map[string]any{"slug": "Eid 2026"})
	if slug != "eid-2026" {
		t.Fatalf("slug should be normalized, got %q", slug)
	}

	env.mustStatus(http.StatusConflict, "POST", "/api/v1/admin/campaigns", map[string]any{"titleAR": "x", "image": tinyPNG, "slug": "eid-2026"}, true)

	// Public detail: duplicated artwork and the thumbnail are stripped.
	got := env.mustStatus(http.StatusOK, "GET", "/api/v1/campaigns/"+slug, nil, false)
	var detail struct {
		Image      string `json:"image"`
		Thumb      string `json:"thumb"`
		TemplateAR struct {
			Image string `json:"image"`
		} `json:"templateAR"`
	}
	if err := json.Unmarshal(got.Data, &detail); err != nil {
		t.Fatal(err)
	}
	if detail.Image != tinyPNG || detail.TemplateAR.Image != "" || detail.Thumb != "" {
		t.Fatalf("unexpected image payload: %+v", detail)
	}

	// Public listing never exposes participation counters.
	_, _, raw := env.do("GET", "/api/v1/campaigns", nil, false)
	if bytes.Contains(raw, []byte("totalCards")) {
		t.Fatalf("public listing leaks totalCards: %s", raw)
	}

	// Updating the Arabic title also updates the primary title shown in the admin list.
	env.mustStatus(http.StatusOK, "PUT", "/api/v1/admin/campaigns/"+slug, map[string]any{"titleAR": "عنوان جديد"}, true)
	_, _, raw = env.do("GET", "/api/v1/admin/campaigns", nil, true)
	if !bytes.Contains(raw, []byte("عنوان جديد")) {
		t.Fatalf("admin listing should show the new title: %s", raw)
	}

	// Deactivating hides the campaign from the public API.
	env.mustStatus(http.StatusOK, "PUT", "/api/v1/admin/campaigns/"+slug, map[string]any{"active": false}, true)
	env.mustStatus(http.StatusNotFound, "GET", "/api/v1/campaigns/"+slug, nil, false)

	env.mustStatus(http.StatusOK, "DELETE", "/api/v1/admin/campaigns/"+slug, nil, true)
	env.mustStatus(http.StatusNotFound, "DELETE", "/api/v1/admin/campaigns/"+slug, nil, true)
}

func TestCardSubmissionAndListing(t *testing.T) {
	env := newTestEnv(t)
	slug := env.createCampaign(nil)
	cardsPath := fmt.Sprintf("/api/v1/campaigns/%s/cards", slug)

	env.mustStatus(http.StatusNotFound, "POST", "/api/v1/campaigns/missing/cards", map[string]any{"to": "x"}, false)
	env.mustStatus(http.StatusBadRequest, "POST", cardsPath, map[string]any{}, false)
	env.mustStatus(http.StatusBadRequest, "POST", cardsPath, map[string]any{"to": strings.Repeat("a", 500)}, false)

	for _, name := range []string{"Sara", "Omar", "منى"} {
		env.mustStatus(http.StatusCreated, "POST", cardsPath, map[string]any{
			"fieldValues": map[string]any{"emp_name": name, "job_title": "Engineer"},
			"lang":        "en",
			"device":      "SomethingUnexpected",
		}, false)
	}

	got := env.mustStatus(http.StatusOK, "GET", "/api/v1/admin/campaigns/"+slug+"/cards?limit=2&offset=0", nil, true)
	var page struct {
		Cards []struct {
			To     string `json:"to"`
			Device string `json:"device"`
		} `json:"cards"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(got.Data, &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 3 || len(page.Cards) != 2 {
		t.Fatalf("pagination: total=%d cards=%d", page.Total, len(page.Cards))
	}
	if page.Cards[0].To != "منى" || page.Cards[0].Device != "Other" {
		t.Fatalf("expected newest card first with sanitized device, got %+v", page.Cards[0])
	}

	got = env.mustStatus(http.StatusOK, "GET", "/api/v1/admin/campaigns/"+slug+"/cards?q=omar", nil, true)
	if err := json.Unmarshal(got.Data, &page); err != nil || page.Total != 1 {
		t.Fatalf("search should match one card, got total=%d (%v)", page.Total, err)
	}

	// "Cards today" must be computed in the same timezone the card date was recorded in.
	got = env.mustStatus(http.StatusOK, "GET", "/api/v1/admin/dashboard", nil, true)
	var stats struct {
		TotalCards int `json:"totalCards"`
		CardsToday int `json:"cardsToday"`
	}
	if err := json.Unmarshal(got.Data, &stats); err != nil || stats.TotalCards != 3 || stats.CardsToday != 3 {
		t.Fatalf("dashboard stats: %+v (%v)", stats, err)
	}
}

func TestCardBodyIsSizeLimited(t *testing.T) {
	env := newTestEnv(t)
	slug := env.createCampaign(nil)

	huge := []byte(`{"to":"` + strings.Repeat("a", 200<<10) + `"}`)
	status, got, _ := env.do("POST", fmt.Sprintf("/api/v1/campaigns/%s/cards", slug), huge, false)
	if status != http.StatusRequestEntityTooLarge || got.Error == nil || got.Error.Code != "PAYLOAD_TOO_LARGE" {
		t.Fatalf("expected 413 PAYLOAD_TOO_LARGE, got %d %+v", status, got.Error)
	}
}

func TestCSVExportEscapesAndNeutralizesFormulas(t *testing.T) {
	env := newTestEnv(t)
	slug := env.createCampaign(nil)

	env.mustStatus(http.StatusCreated, "POST", "/api/v1/campaigns/"+slug+"/cards", map[string]any{
		"to":      "=HYPERLINK(\"http://evil\")",
		"message": "Hello, \"friend\"\nsecond line",
		"from":    "منى",
	}, false)

	status, _, raw := env.do("GET", "/api/v1/admin/campaigns/"+slug+"/export", nil, true)
	if status != http.StatusOK {
		t.Fatalf("export status %d", status)
	}
	if !bytes.HasPrefix(raw, []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatal("CSV should start with a UTF-8 BOM for Excel")
	}
	csvText := string(raw)
	// Formula is defused with a leading quote and the field is CSV-quoted (RFC 4180 doubles inner quotes).
	if !strings.Contains(csvText, `"'=HYPERLINK(""http://evil"")"`) {
		t.Fatalf("formula not neutralized / quotes not doubled:\n%s", csvText)
	}
	if !strings.Contains(csvText, `"Hello, ""friend""`) {
		t.Fatalf("message not escaped per RFC 4180:\n%s", csvText)
	}
	if !strings.Contains(csvText, "منى") {
		t.Fatalf("Arabic text lost:\n%s", csvText)
	}
}

// Public campaign data must never be served stale (an admin who just created an occasion expects to see it),
// but unchanged data should cost a 304 instead of re-sending base64 artwork.
func TestPublicResponsesRevalidateWithETag(t *testing.T) {
	env := newTestEnv(t)
	slug := env.createCampaign(nil)
	url := env.server.URL + "/api/v1/campaigns/" + slug

	get := func(ifNoneMatch string) *http.Response {
		req, _ := http.NewRequest("GET", url, nil)
		if ifNoneMatch != "" {
			req.Header.Set("If-None-Match", ifNoneMatch)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		return resp
	}

	first := get("")
	etag := first.Header.Get("ETag")
	if first.StatusCode != http.StatusOK || etag == "" || first.Header.Get("Cache-Control") != "no-cache" {
		t.Fatalf("expected 200 with ETag and no-cache, got %d etag=%q cc=%q", first.StatusCode, etag, first.Header.Get("Cache-Control"))
	}
	if got := get(etag).StatusCode; got != http.StatusNotModified {
		t.Fatalf("matching ETag should give 304, got %d", got)
	}
	if got := get("W/" + etag).StatusCode; got != http.StatusNotModified {
		t.Fatalf("weak ETag (added by a compressing proxy) should give 304, got %d", got)
	}

	// After an edit the old ETag no longer matches, so the change is visible immediately.
	env.mustStatus(http.StatusOK, "PUT", "/api/v1/admin/campaigns/"+slug, map[string]any{"titleEN": "Changed"}, true)
	if got := get(etag).StatusCode; got != http.StatusOK {
		t.Fatalf("stale ETag after an edit should give 200, got %d", got)
	}
}
