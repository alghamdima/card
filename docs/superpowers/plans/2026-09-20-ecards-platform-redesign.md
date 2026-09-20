# E-Cards Platform Complete Redesign & Enhancement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Re-architect and redesign the Gratitude E-Cards Platform with ALJ Finance visual identity, dual-language template variants (Arabic & English), an interactive drag-and-resize visual template position editor for admins, flexible dynamic text fields, and a modern campaign analytics dashboard matching the provided reference designs.

**Architecture:** 
1. Database & Go Backend: Add migration `000002_dynamic_templates.up.sql` to support `template_ar`, `template_en`, and dynamic `field_values` with JSON fallback to ensure backward compatibility. Expose analytics and export endpoints.
2. Web Frontend (SvelteKit + TypeScript): Upgrade brand assets and tokens (`ALJUF` logos, colors `#3C1053`, `#FFCD00`, `#63656A`, `#47859F`).
3. Admin Visual Builder: Implement interactive dual-canvas with draggable/resizable bounding boxes, language tabs (AR/EN), dynamic field management (add/edit/delete, fontSize, color, align, order).
4. Employee Card Generation: Render dynamic fields on canvas (1080x1350) based on selected language variant (AR/EN) with real-time preview and one-click PNG export.
5. Analytics & Reporting: Implement tabbed campaign analytics dashboard with total card metrics, search, and CSV export.

**Tech Stack:** Go 1.23, Chi Router, SQLite (WAL mode), SvelteKit, TypeScript, HTML5 Canvas, ALJUF Brand Assets.

**Spec:** `docs/superpowers/specs/2026-09-20-ecards-platform-redesign-spec.md`

## Global Constraints
- Strictly adhere to ALJ Finance brand colors: Primary `#3C1053`, Dark `#2B0A3D`, Accent `#FFCD00`, Gray `#63656A`, Blue `#47859F`.
- Strictly use official ALJUF brand logos from `aljuf-brand` skill (`ALJUF_Logo-01` AR, `ALJUF_Logo-03` EN, `ALJUF_Logo-02` / `04` transparent white).
- No device breakdown counters in the analytics dashboard (All Time Total Cards only, per user explicit instruction).
- Maintain dual-language support (AR default RTL, EN LTR).

---

### Task 1: Brand Assets & Visual System Setup
**Files:**
- Create: `apps/web/static/images/brand/aljuf-ar.png`
- Create: `apps/web/static/images/brand/aljuf-en.png`
- Create: `apps/web/static/images/brand/aljuf-ar-white.png`
- Create: `apps/web/static/images/brand/aljuf-en-white.png`
- Modify: `apps/web/src/lib/styles/tokens.css`
- Modify: `apps/web/src/lib/styles/app.css`

- [ ] **Step 1: Copy ALJUF official logos to static directory**
- [ ] **Step 2: Update tokens.css with exact ALJ Finance brand tokens and palette**
- [ ] **Step 3: Verify static logo files and css build**

---

### Task 2: Database Schema & Backend Domain Models
**Files:**
- Create: `apps/api/migrations/000002_dynamic_templates.up.sql`
- Create: `apps/api/migrations/000002_dynamic_templates.down.sql`
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/database/sqlite.go`

- [ ] **Step 1: Write SQL migration for template_ar, template_en, lang, and field_values**
- [ ] **Step 2: Update Go domain models (TextFieldConfig, TemplateVariant, DynamicCard)**
- [ ] **Step 3: Update SQLite schema migration runner in sqlite.go**
- [ ] **Step 4: Verify migration execution with go test / build**

---

### Task 3: Backend Repository, Service & Handlers for Dynamic Templates & Analytics
**Files:**
- Modify: `apps/api/internal/repository/campaign_repo.go`
- Modify: `apps/api/internal/repository/card_repo.go`
- Modify: `apps/api/internal/service/campaign_service.go`
- Modify: `apps/api/internal/service/card_service.go`
- Modify: `apps/api/internal/http/handlers/campaign_handler.go`
- Modify: `apps/api/internal/http/handlers/card_handler.go`
- Modify: `apps/api/internal/http/routes.go`

- [ ] **Step 1: Update campaign repository to read/write template_ar & template_en**
- [ ] **Step 2: Add analytics and CSV export queries in card repository**
- [ ] **Step 3: Update campaign & card services to handle dynamic template variants and card submissions**
- [ ] **Step 4: Expose routes `/api/v1/admin/analytics/overview` and `/api/v1/admin/campaigns/{slug}/export`**
- [ ] **Step 5: Run tests on Go API and verify clean build**

---

### Task 4: Frontend Types, API Client & Translation Dictionaries
**Files:**
- Modify: `apps/web/src/lib/types/campaign.types.ts`
- Modify: `apps/web/src/lib/api/campaigns.ts`
- Modify: `apps/web/src/lib/messages/ar.json`
- Modify: `apps/web/src/lib/messages/en.json`

- [ ] **Step 1: Add TypeScript interfaces for TextFieldConfig, TemplateVariant, and Analytics**
- [ ] **Step 2: Update API methods in campaigns.ts for dynamic templates and analytics export**
- [ ] **Step 3: Enrich Arabic & English translation keys for new editor, tabs, and analytics terms**

---

### Task 5: Dynamic Canvas Renderer & Card Generator Engine
**Files:**
- Modify: `apps/web/src/lib/components/cards/canvas-renderer.ts`
- Modify: `apps/web/src/lib/components/cards/CardPreview.svelte`
- Modify: `apps/web/src/lib/components/cards/CardEditor.svelte`

- [ ] **Step 1: Refactor canvas-renderer.ts to support dynamic array of TextFieldConfig**
- [ ] **Step 2: Support automatic font scaling, multiline wrapping, alignment, and custom colors per field**
- [ ] **Step 3: Update CardPreview.svelte to render dynamic fields**
- [ ] **Step 4: Update CardEditor.svelte to dynamically generate inputs based on active template fields**

---

### Task 6: Visual Template & Position Builder for Admin (Interactive Drag & Resize)
**Files:**
- Create: `apps/web/src/lib/components/admin/TemplateCanvasEditor.svelte`
- Create: `apps/web/src/lib/components/admin/FieldSettingsModal.svelte`
- Modify: `apps/web/src/routes/admin/cards/+page.svelte`

- [ ] **Step 1: Implement TemplateCanvasEditor with draggable and resizable bounding box handles**
- [ ] **Step 2: Implement Language Switcher (AR/EN tabs) with distinct artwork upload per language**
- [ ] **Step 3: Implement dynamic field management (Add, Delete, Reorder, Font Size, Color, Text Align)**
- [ ] **Step 4: Wire Dual Canvas Layout: Interactive Editor on the left, Live Preview on the right**
- [ ] **Step 5: Test save & load of template position configuration**

---

### Task 7: Employee Public Card Generation Page Enhancement
**Files:**
- Modify: `apps/web/src/routes/cards/[slug]/+page.svelte`
- Modify: `apps/web/src/routes/+page.svelte`
- Modify: `apps/web/src/routes/+layout.svelte`

- [ ] **Step 1: Add language toggle (العربية / English) for the employee on the card page**
- [ ] **Step 2: Switch template artwork and input fields dynamically on language change**
- [ ] **Step 3: Update header/footer with ALJUF official logo and refined brand styling**
- [ ] **Step 4: Verify high-resolution PNG generation and background telemetry logging**

---

### Task 8: Card Generating Analytics & Reports Module
**Files:**
- Create: `apps/web/src/lib/components/admin/AnalyticsTabs.svelte`
- Create: `apps/web/src/lib/components/admin/EmployeeCardsTable.svelte`
- Modify: `apps/web/src/routes/admin/dashboard/+page.svelte`
- Modify: `apps/web/src/routes/admin/users/+page.svelte`
- Modify: `apps/web/src/lib/components/admin/AdminHeader.svelte`

- [ ] **Step 1: Implement horizontal campaign navigation tabs matching the reference image**
- [ ] **Step 2: Implement All Time Stats card with prominent Total Cards count**
- [ ] **Step 3: Implement Employee Cards list with search filter and CSV export button**
- [ ] **Step 4: Refactor AdminHeader with official ALJUF branding**

---

### Task 9: End-to-End Build, Integration & Verification
**Files:**
- All modified and created files
- [ ] **Step 1: Run Go API unit tests and verify backend compilation**
- [ ] **Step 2: Build SvelteKit static frontend (`npm run build`)**
- [ ] **Step 3: Test docker container deployment and verify live on browser**
- [ ] **Step 4: Validate all scenarios: create bilingual campaign, position fields, generate cards, check analytics**
