# Gratitude E-Cards | بطاقات الامتنان الإلكترونية

Employees send a thank-you card to a colleague. Bilingual (English / العربية),
multiple card designs, optional anonymity.

## Pages

| Path | Access | Purpose |
|---|---|---|
| `/` | Public | Pick a card, write the message, download it |
| `/admin.html` | Password | Designs, text positions, presets, portal gate |
| `/dashboard.html` | Password | Totals and the full list of cards sent |

## Setup

### 1. Environment variables

Netlify → **Site configuration → Environment variables**:

- `ADMIN_PASSWORD`
- `SESSION_SECRET`
- `SHEET_ID`
- `SA_EMAIL`
- `SA_PRIVATE_KEY` — single line, `\n` between key lines

**Redeploy after adding them.** Env vars are read at deploy time.

### 2. Google Sheet

Can be the **same spreadsheet** as the greeting-card portal. All tabs used here
are prefixed so nothing collides:

- `GratitudeConfig` — settings
- `GratitudeDesigns` — one row per design (id, name, language, box positions, thumbnail)
- `GratitudeCards` — the submissions
- `GImg_<id>` — one tab per design holding its artwork as base64

Share the spreadsheet with `SA_EMAIL` as **Editor**.

### 3. First run

Open `/admin.html`, sign in, then **+ New design**. Upload artwork, drag the three
boxes onto the card, save. Repeat for each language — a design belongs to one
language and employees only see designs matching the language they choose.

## How the card is built

Three independent text boxes: **To**, **Message**, **From**. Each has its own
position, max font size, colour, and weight.

Text wraps to fit the box width; if the wrapped block is too tall, the font
shrinks and re-wraps until it fits. Short messages stay large, long ones stay
readable. Words longer than the box are hard-split.

Arabic text automatically uses Luma, English uses Karbon. Detection is per field,
so a card can mix languages.

If **From** is left blank the card draws nothing there. The dashboard shows
"Anonymous" so the table stays readable.

## Notes

- `render.js` is shared by the card page and the admin preview, so the preview is pixel-identical to the real output.
- The Google key never reaches the browser; all Sheets access goes through `netlify/functions/api.js`.
- Card canvas is 1080×1350. Uploaded artwork is resized to fit.
- Opening the link inside an in-app browser (Oracle, LinkedIn, etc.) shows an "Open in Browser" screen, because saving images doesn't work reliably in those.
