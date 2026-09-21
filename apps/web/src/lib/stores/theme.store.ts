import { get, writable } from 'svelte/store';
import { getItem, setItem } from '../utils/storage';

export type ThemePreference = 'system' | 'dark' | 'light';
export type EffectiveTheme = 'dark' | 'light';

const THEME_KEY = 'user_theme';

function readPreference(): ThemePreference {
  const saved = getItem(THEME_KEY);
  return saved === 'dark' || saved === 'light' || saved === 'system' ? saved : 'system';
}

function resolve(pref: ThemePreference): EffectiveTheme {
  if (pref !== 'system') return pref;
  if (typeof window === 'undefined') return 'dark';
  return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
}

export const theme = writable<ThemePreference>(readPreference());

/** The theme actually shown (system preference resolved), e.g. to pick the matching logo. */
export const effectiveTheme = writable<EffectiveTheme>(resolve(readPreference()));

export function applyTheme(pref: ThemePreference) {
  if (typeof window === 'undefined') return;
  setItem(THEME_KEY, pref);
  theme.set(pref);

  const effective = resolve(pref);
  effectiveTheme.set(effective);
  document.documentElement.setAttribute('data-theme', effective);
}

if (typeof window !== 'undefined') {
  // Follow OS theme changes while in "system" mode.
  window.matchMedia('(prefers-color-scheme: light)').addEventListener('change', () => {
    if (get(theme) === 'system') applyTheme('system');
  });

  applyTheme(readPreference());
}
