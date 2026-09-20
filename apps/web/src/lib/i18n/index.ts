import { writable, derived } from 'svelte/store';
import arMessages from '../messages/ar.json';
import enMessages from '../messages/en.json';

export type Locale = 'ar' | 'en';

const messagesMap: Record<Locale, any> = {
  ar: arMessages,
  en: enMessages
};

// Initialize locale from storage if in browser, default to 'ar'
const initialLocale: Locale = typeof window !== 'undefined' 
  ? ((localStorage.getItem('user_locale') as Locale) || 'ar')
  : 'ar';

export const locale = writable<Locale>(initialLocale);

export const dir = derived(locale, ($locale) => ($locale === 'ar' ? 'rtl' : 'ltr'));

// Update document lang and dir dynamically
if (typeof window !== 'undefined') {
  locale.subscribe(($locale) => {
    localStorage.setItem('user_locale', $locale);
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

// Translation helper t(path)
export const t = derived(locale, ($locale) => {
  return (path: string, params?: Record<string, string | number>): string => {
    const keys = path.split('.');
    let result: any = messagesMap[$locale];

    for (const key of keys) {
      if (result && typeof result === 'object' && key in result) {
        result = result[key];
      } else {
        // Fallback to English if not found in current locale
        let fallback: any = messagesMap['en'];
        for (const fk of keys) {
          if (fallback && typeof fallback === 'object' && fk in fallback) {
            fallback = fallback[fk];
          } else {
            return path;
          }
        }
        result = fallback;
      }
    }

    if (typeof result !== 'string') {
      return path;
    }

    if (params) {
      for (const [pk, pv] of Object.entries(params)) {
        result = result.replace(new RegExp(`{${pk}}`, 'g'), String(pv));
      }
    }

    return result;
  };
});
