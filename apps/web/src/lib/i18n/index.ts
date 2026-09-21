import { derived, get, writable } from 'svelte/store';
import arMessages from '../messages/ar.json';
import enMessages from '../messages/en.json';
import { getItem, setItem } from '../utils/storage';

export type Locale = 'ar' | 'en';

type MessageTree = { [key: string]: string | MessageTree };

const messagesMap: Record<Locale, MessageTree> = {
  ar: arMessages,
  en: enMessages
};

const LOCALE_KEY = 'user_locale';

function isLocale(value: unknown): value is Locale {
  return value === 'ar' || value === 'en';
}

function readStoredLocale(): Locale {
  const stored = getItem(LOCALE_KEY);
  return isLocale(stored) ? stored : 'ar';
}

export const locale = writable<Locale>(readStoredLocale());

export const dir = derived(locale, ($locale) => ($locale === 'ar' ? 'rtl' : 'ltr'));

// Keep <html lang/dir> and the stored preference in sync with the active locale.
if (typeof window !== 'undefined') {
  locale.subscribe(($locale) => {
    setItem(LOCALE_KEY, $locale);
    document.documentElement.lang = $locale;
    document.documentElement.dir = $locale === 'ar' ? 'rtl' : 'ltr';
  });
}

export function setLocale(newLocale: Locale) {
  locale.set(newLocale);
}

export function toggleLocale() {
  locale.update((current) => (current === 'ar' ? 'en' : 'ar'));
}

function lookup(tree: MessageTree, path: string): string | undefined {
  let node: string | MessageTree | undefined = tree;
  for (const key of path.split('.')) {
    if (node === undefined || typeof node !== 'object') return undefined;
    node = node[key];
  }
  return typeof node === 'string' ? node : undefined;
}

/** Resolve a dotted key for a locale, falling back to English and finally to the key itself. */
export function translate(loc: Locale, path: string, params?: Record<string, string | number>): string {
  let text = lookup(messagesMap[loc], path) ?? lookup(messagesMap.en, path) ?? path;
  if (params) {
    for (const [name, value] of Object.entries(params)) {
      text = text.split(`{${name}}`).join(String(value));
    }
  }
  return text;
}

export const t = derived(locale, ($locale) => {
  return (path: string, params?: Record<string, string | number>): string => translate($locale, path, params);
});

/** Translate an API/network error into a message in the active language. */
export function translateError(err: unknown, fallbackKey = 'app.error'): string {
  const loc = get(locale);
  const code = typeof err === 'object' && err !== null ? (err as { code?: unknown }).code : undefined;
  if (typeof code === 'string') {
    const key = `errors.${code}`;
    const message = lookup(messagesMap[loc], key) ?? lookup(messagesMap.en, key);
    if (message) return message;
  }
  return translate(loc, fallbackKey);
}

// Gregorian calendar and Latin digits in both locales: Arabic locales otherwise default to the Hijri calendar.
const INTL_LOCALE: Record<Locale, string> = {
  ar: 'ar-SA-u-ca-gregory-nu-latn',
  en: 'en-GB'
};

export function formatDate(value: string | number | Date, loc: Locale = get(locale)): string {
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return new Intl.DateTimeFormat(INTL_LOCALE[loc], { year: 'numeric', month: 'short', day: 'numeric' }).format(date);
}

export function formatTime(value: Date, loc: Locale = get(locale)): string {
  return new Intl.DateTimeFormat(INTL_LOCALE[loc], { hour: '2-digit', minute: '2-digit', hour12: false }).format(value);
}
