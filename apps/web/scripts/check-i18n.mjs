// Fails the build when the Arabic and English message files drift apart or the code references a missing key.
//   node scripts/check-i18n.mjs
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');

function flatten(tree, prefix = '') {
  return Object.entries(tree).flatMap(([key, value]) =>
    typeof value === 'object' && value !== null ? flatten(value, `${prefix}${key}.`) : [[`${prefix}${key}`, value]]
  );
}

const load = (locale) =>
  Object.fromEntries(flatten(JSON.parse(fs.readFileSync(path.join(root, 'src/lib/messages', `${locale}.json`), 'utf8'))));

const ar = load('ar');
const en = load('en');

function sourceFiles(dir) {
  return fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) return sourceFiles(full);
    return /\.(svelte|ts)$/.test(entry.name) ? [full] : [];
  });
}

const source = sourceFiles(path.join(root, 'src')).map((f) => fs.readFileSync(f, 'utf8')).join('\n');
const usedLiteralKeys = new Set([...source.matchAll(/\b(?:t|translate)\(\s*(?:'[a-z]{2}',\s*)?['"`]([\w.]+)['"`]/g)].map((m) => m[1]));

const errors = [];
const placeholders = (text) => (String(text).match(/\{\w+\}/g) ?? []).sort().join(',');

for (const key of Object.keys(en)) if (!(key in ar)) errors.push(`missing in ar.json: ${key}`);
for (const key of Object.keys(ar)) if (!(key in en)) errors.push(`missing in en.json: ${key}`);
for (const key of Object.keys(en)) {
  if (key in ar && placeholders(en[key]) !== placeholders(ar[key])) errors.push(`placeholder mismatch: ${key}`);
  if (!String(en[key]).trim() || (key in ar && !String(ar[key]).trim())) errors.push(`empty message: ${key}`);
}
for (const key of usedLiteralKeys) {
  if (/^(errors|theme)\.$/.test(key)) continue;
  if (!(key in en)) errors.push(`used in code but not defined: ${key}`);
}

if (errors.length) {
  console.error(`i18n check failed:\n  - ${errors.join('\n  - ')}`);
  process.exit(1);
}
console.log(`i18n check passed (${Object.keys(en).length} keys in both languages)`);
