const crypto = require('crypto');

// ─────────────────────────────────────────────
//  Secrets come from Netlify environment variables
// ─────────────────────────────────────────────
const ADMIN_PASSWORD = process.env.ADMIN_PASSWORD;
const SESSION_SECRET = process.env.SESSION_SECRET;
const SHEET_ID = process.env.SHEET_ID;
const SA_EMAIL = process.env.SA_EMAIL;
const PRIV_KEY = (process.env.SA_PRIVATE_KEY || '').replace(/\\n/g, '\n');

function missingEnv() {
  const need = { ADMIN_PASSWORD, SESSION_SECRET, SHEET_ID, SA_EMAIL, SA_PRIVATE_KEY: PRIV_KEY };
  return Object.keys(need).filter(k => !need[k]);
}

// Tab names — prefixed so they never collide with the greeting-card portal
const T_CONFIG  = 'GratitudeConfig';
const T_DESIGNS = 'GratitudeDesigns';
const T_CARDS   = 'GratitudeCards';
const imgTab = id => 'GImg_' + id;

const CHUNK = 40000;

// ── Google auth ──
let cachedToken = null, tokenExpiry = 0;

async function getToken() {
  if (cachedToken && Date.now() < tokenExpiry) return cachedToken;
  const now = Math.floor(Date.now() / 1000);
  const b64 = o => Buffer.from(JSON.stringify(o)).toString('base64url');
  const input = b64({ alg: 'RS256', typ: 'JWT' }) + '.' + b64({
    iss: SA_EMAIL,
    scope: 'https://www.googleapis.com/auth/spreadsheets',
    aud: 'https://oauth2.googleapis.com/token',
    exp: now + 3600, iat: now
  });
  const signer = crypto.createSign('RSA-SHA256');
  signer.update(input);
  const jwt = input + '.' + signer.sign(PRIV_KEY, 'base64url');
  const res = await fetch('https://oauth2.googleapis.com/token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: 'grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=' + jwt
  });
  const d = await res.json();
  if (!d.access_token) throw new Error('Google auth failed');
  cachedToken = d.access_token;
  tokenExpiry = Date.now() + 3400000;
  return cachedToken;
}

const API = 'https://sheets.googleapis.com/v4/spreadsheets/';

async function sGet(range) {
  const t = await getToken();
  const r = await fetch(API + SHEET_ID + '/values/' + encodeURIComponent(range),
    { headers: { Authorization: 'Bearer ' + t } });
  const d = await r.json();
  return d.values || [];
}

async function sSet(range, values) {
  const t = await getToken();
  const r = await fetch(API + SHEET_ID + '/values/' + encodeURIComponent(range) + '?valueInputOption=RAW', {
    method: 'PUT',
    headers: { Authorization: 'Bearer ' + t, 'Content-Type': 'application/json' },
    body: JSON.stringify({ values })
  });
  return r.json();
}

async function sAppend(tab, row) {
  const t = await getToken();
  const r = await fetch(API + SHEET_ID + '/values/' + encodeURIComponent(tab) + ':append?valueInputOption=USER_ENTERED', {
    method: 'POST',
    headers: { Authorization: 'Bearer ' + t, 'Content-Type': 'application/json' },
    body: JSON.stringify({ values: [row] })
  });
  return r.json();
}

async function sClear(range) {
  const t = await getToken();
  await fetch(API + SHEET_ID + '/values/' + encodeURIComponent(range) + ':clear',
    { method: 'POST', headers: { Authorization: 'Bearer ' + t } });
}

async function listTabs() {
  const t = await getToken();
  const r = await fetch(API + SHEET_ID + '?fields=sheets.properties.title',
    { headers: { Authorization: 'Bearer ' + t } });
  const d = await r.json();
  return (d.sheets || []).map(s => s.properties.title);
}

async function addTab(title) {
  const t = await getToken();
  await fetch(API + SHEET_ID + ':batchUpdate', {
    method: 'POST',
    headers: { Authorization: 'Bearer ' + t, 'Content-Type': 'application/json' },
    body: JSON.stringify({ requests: [{ addSheet: { properties: { title } } }] })
  });
}

async function delTab(title) {
  const t = await getToken();
  const r = await fetch(API + SHEET_ID + '?fields=sheets.properties',
    { headers: { Authorization: 'Bearer ' + t } });
  const d = await r.json();
  const sh = (d.sheets || []).find(s => s.properties.title === title);
  if (!sh) return;
  await fetch(API + SHEET_ID + ':batchUpdate', {
    method: 'POST',
    headers: { Authorization: 'Bearer ' + t, 'Content-Type': 'application/json' },
    body: JSON.stringify({ requests: [{ deleteSheet: { sheetId: sh.properties.sheetId } }] })
  });
}

async function ensureTabs() {
  const tabs = await listTabs();
  if (!tabs.includes(T_CONFIG))  await addTab(T_CONFIG);
  if (!tabs.includes(T_DESIGNS)) {
    await addTab(T_DESIGNS);
    await sSet(T_DESIGNS + '!A1:F1', [['id', 'name', 'lang', 'boxes', 'thumb', 'active']]);
  }
  if (!tabs.includes(T_CARDS)) {
    await addTab(T_CARDS);
    await sSet(T_CARDS + '!A1:I1', [['From', 'To', 'Message', 'Language', 'Design', 'Device', 'Date', 'Time', 'Heading']]);
  }
}

// ── Config ──
const DEFAULT_PRESETS_EN = JSON.stringify([
  { icon: '\u2B50', label: 'Thank You', text: '' },
  { icon: '\uD83D\uDCA1', label: 'You Inspire Me', text: '' },
  { icon: '\u2600\uFE0F', label: 'You Made My Day', text: '' }
]);

const DEFAULT_PRESETS_AR = JSON.stringify([
  { icon: '\u2B50', label: '\u0634\u0643\u0631\u0627\u064B \u0644\u0643', text: '' },
  { icon: '\uD83D\uDCA1', label: '\u0623\u0646\u062A \u062A\u0644\u0647\u0645\u0646\u064A', text: '' },
  { icon: '\u2600\uFE0F', label: '\u0644\u0642\u062F \u0623\u0633\u0639\u062F\u062A \u064A\u0648\u0645\u064A', text: '' }
]);

const DEFAULTS = {
  status: 'open',
  titleEn: 'Gratitude E-Cards',
  titleAr: '\u0628\u0637\u0627\u0642\u0627\u062A \u0627\u0644\u0627\u0645\u062A\u0646\u0627\u0646 \u0627\u0644\u0625\u0644\u0643\u062A\u0631\u0648\u0646\u064A\u0629',
  subtitleEn: 'Send a note of thanks to a colleague',
  subtitleAr: '\u0623\u0631\u0633\u0644 \u0643\u0644\u0645\u0629 \u0634\u0643\u0631 \u0644\u0632\u0645\u064A\u0644',
  closedTitleEn: 'Coming Soon',
  closedTitleAr: '\u0642\u0631\u064A\u0628\u0627\u064B',
  closedMsgEn: 'The gratitude card portal is currently closed.',
  closedMsgAr: '\u0628\u0648\u0627\u0628\u0629 \u0628\u0637\u0627\u0642\u0627\u062A \u0627\u0644\u0627\u0645\u062A\u0646\u0627\u0646 \u0645\u063A\u0644\u0642\u0629 \u062D\u0627\u0644\u064A\u0627\u064B.',
  presetsEn: DEFAULT_PRESETS_EN,
  presetsAr: DEFAULT_PRESETS_AR
};

async function readConfig() {
  const rows = await sGet(T_CONFIG + '!A:B');
  const cfg = Object.assign({}, DEFAULTS);
  rows.forEach(r => { if (r[0]) cfg[r[0]] = r[1] !== undefined ? r[1] : ''; });
  return cfg;
}

async function writeConfig(patch) {
  const merged = Object.assign(await readConfig(), patch);
  const rows = Object.keys(merged).map(k => [k, String(merged[k])]);
  await sClear(T_CONFIG + '!A:B');
  await sSet(T_CONFIG + '!A1:B' + rows.length, rows);
  return merged;
}

function parsePresets(s) {
  try { const a = JSON.parse(s); return Array.isArray(a) ? a : []; }
  catch (e) { return []; }
}

// ── Designs ──
async function readDesigns() {
  const rows = await sGet(T_DESIGNS + '!A2:F');
  return rows.filter(r => r[0]).map(r => ({
    id: r[0],
    name: r[1] || '',
    lang: (r[2] === 'ar' ? 'ar' : 'en'),
    boxes: (function () { try { return JSON.parse(r[3] || '{}'); } catch (e) { return {}; } })(),
    thumb: r[4] || '',
    active: String(r[5] || 'yes') !== 'no'
  }));
}

async function writeDesigns(list) {
  await sClear(T_DESIGNS + '!A2:F');
  if (!list.length) return;
  const rows = list.map(d => [d.id, d.name, d.lang, JSON.stringify(d.boxes || {}), d.thumb || '', d.active === false ? 'no' : 'yes']);
  await sSet(T_DESIGNS + '!A2:F' + (rows.length + 1), rows);
}

async function readImage(id) {
  const tabs = await listTabs();
  if (!tabs.includes(imgTab(id))) return '';
  const rows = await sGet(imgTab(id) + '!A:A');
  return rows.map(r => r[0] || '').join('');
}

async function writeImage(id, b64) {
  const tabs = await listTabs();
  if (!tabs.includes(imgTab(id))) await addTab(imgTab(id));
  await sClear(imgTab(id) + '!A:A');
  const chunks = [];
  for (let i = 0; i < b64.length; i += CHUNK) chunks.push([b64.slice(i, i + CHUNK)]);
  if (!chunks.length) return;
  await sSet(imgTab(id) + '!A1:A' + chunks.length, chunks);
}

// ── Sessions ──
function makeToken() {
  const exp = Date.now() + 8 * 3600 * 1000;
  return exp + '.' + crypto.createHmac('sha256', SESSION_SECRET).update(String(exp)).digest('hex');
}
function checkToken(tok) {
  if (!tok || typeof tok !== 'string') return false;
  const p = tok.split('.');
  if (p.length !== 2) return false;
  const exp = Number(p[0]);
  if (!exp || Date.now() > exp) return false;
  const want = crypto.createHmac('sha256', SESSION_SECRET).update(String(exp)).digest('hex');
  try { return crypto.timingSafeEqual(Buffer.from(p[1]), Buffer.from(want)); }
  catch (e) { return false; }
}

const H = { 'Content-Type': 'application/json', 'Cache-Control': 'no-store' };
const ok = b => ({ statusCode: 200, headers: H, body: JSON.stringify(b) });
const bad = (c, m) => ({ statusCode: c, headers: H, body: JSON.stringify({ error: m }) });

exports.handler = async function (event) {
  if (event.httpMethod !== 'POST') return bad(405, 'Method not allowed');

  const miss = missingEnv();
  if (miss.length) return bad(500, 'Server not configured. Missing environment variables: ' + miss.join(', '));

  let body;
  try { body = JSON.parse(event.body || '{}'); } catch (e) { return bad(400, 'Bad JSON'); }
  const action = body.action;

  try {
    // ---------- PUBLIC ----------
    if (action === 'getPublicConfig') {
      const cfg = await readConfig();
      const base = {
        status: cfg.status,
        text: {
          titleEn: cfg.titleEn, titleAr: cfg.titleAr,
          subtitleEn: cfg.subtitleEn, subtitleAr: cfg.subtitleAr,
          closedTitleEn: cfg.closedTitleEn, closedTitleAr: cfg.closedTitleAr,
          closedMsgEn: cfg.closedMsgEn, closedMsgAr: cfg.closedMsgAr
        }
      };
      if (cfg.status !== 'open') return ok(base);
      const designs = (await readDesigns()).filter(d => d.active);
      base.presets = { en: parsePresets(cfg.presetsEn), ar: parsePresets(cfg.presetsAr) };
      base.designs = designs.map(d => ({ id: d.id, name: d.name, lang: d.lang, thumb: d.thumb }));
      return ok(base);
    }

    if (action === 'getDesign') {
      const cfg = await readConfig();
      if (cfg.status !== 'open') return bad(403, 'Portal is closed');
      const id = String(body.id || '');
      const all = await readDesigns();
      const d = all.find(x => x.id === id && x.active);
      if (!d) return bad(404, 'Design not found');
      const image = await readImage(id);
      return ok({ id: d.id, lang: d.lang, boxes: d.boxes, image: image });
    }

    if (action === 'saveCard') {
      const cfg = await readConfig();
      if (cfg.status !== 'open') return bad(403, 'Portal is closed');
      const from = String(body.from || '').slice(0, 60);
      const to = String(body.to || '').slice(0, 60);
      const msg = String(body.message || '').slice(0, 400);
      const heading = String(body.heading || '').slice(0, 80);
      const lang = body.lang === 'ar' ? 'ar' : 'en';
      const design = String(body.designId || '').slice(0, 20);
      const device = body.device === 'iPhone' ? 'iPhone' : 'Android';
      const now = new Date();
      const o = { timeZone: 'Asia/Riyadh' };
      await sAppend(T_CARDS, [
        from || 'Anonymous', to, msg, lang, design, device,
        now.toLocaleDateString('en-GB', o), now.toLocaleTimeString('en-GB', o), heading
      ]);
      return ok({ saved: true });
    }

    // ---------- LOGIN ----------
    if (action === 'login') {
      const pw = String(body.password || '');
      const a = Buffer.from(pw.padEnd(64).slice(0, 64));
      const b = Buffer.from(ADMIN_PASSWORD.padEnd(64).slice(0, 64));
      if (!crypto.timingSafeEqual(a, b)) {
        await new Promise(r => setTimeout(r, 700));
        return bad(401, 'Wrong password');
      }
      return ok({ token: makeToken() });
    }

    // ---------- PROTECTED ----------
    if (!checkToken(body.token)) return bad(401, 'Not authorised');

    if (action === 'getAdminConfig') {
      await ensureTabs();
      const cfg = await readConfig();
      const designs = await readDesigns();
      return ok({
        config: cfg,
        presets: { en: parsePresets(cfg.presetsEn), ar: parsePresets(cfg.presetsAr) },
        designs: designs.map(d => ({ id: d.id, name: d.name, lang: d.lang, boxes: d.boxes, thumb: d.thumb, active: d.active }))
      });
    }

    if (action === 'getDesignFull') {
      const id = String(body.id || '');
      const all = await readDesigns();
      const d = all.find(x => x.id === id);
      if (!d) return bad(404, 'Design not found');
      const image = await readImage(id);
      return ok({ design: d, image: image });
    }

    if (action === 'saveDesign') {
      await ensureTabs();
      const list = await readDesigns();
      let id = String(body.id || '').trim();
      if (!id) {
        let n = 1;
        while (list.some(x => x.id === 'd' + n)) n++;
        id = 'd' + n;
      }
      const rec = {
        id: id,
        name: String(body.name || 'Design').slice(0, 60),
        lang: body.lang === 'ar' ? 'ar' : 'en',
        boxes: body.boxes || {},
        thumb: String(body.thumb || ''),
        active: body.active !== false
      };
      const idx = list.findIndex(x => x.id === id);
      if (idx >= 0) {
        if (!rec.thumb) rec.thumb = list[idx].thumb;
        list[idx] = rec;
      } else {
        list.push(rec);
      }
      await writeDesigns(list);
      if (body.image) await writeImage(id, String(body.image));
      return ok({ saved: true, id: id });
    }

    if (action === 'deleteDesign') {
      const id = String(body.id || '');
      const list = await readDesigns();
      await writeDesigns(list.filter(x => x.id !== id));
      await delTab(imgTab(id));
      return ok({ deleted: true });
    }

    if (action === 'saveConfig') {
      const allowed = ['status', 'titleEn', 'titleAr', 'subtitleEn', 'subtitleAr',
                       'closedTitleEn', 'closedTitleAr', 'closedMsgEn', 'closedMsgAr',
                       'presetsEn', 'presetsAr'];
      const patch = {};
      allowed.forEach(k => { if (body[k] !== undefined) patch[k] = body[k]; });
      const merged = await writeConfig(patch);
      return ok({ config: merged });
    }

    if (action === 'getDashboard') {
      await ensureTabs();
      const rows = await sGet(T_CARDS + '!A:I');
      const data = rows.filter((r, i) => i > 0 && (r[0] || r[1]));
      return ok({ rows: data });
    }

    return bad(400, 'Unknown action');
  } catch (err) {
    return bad(500, err.message || 'Server error');
  }
};
