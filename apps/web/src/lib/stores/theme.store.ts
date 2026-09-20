import { writable } from 'svelte/store';

export type ThemePreference = 'system' | 'dark' | 'light';

function getInitialTheme(): ThemePreference {
  if (typeof window === 'undefined') return 'system';
  const saved = localStorage.getItem('user_theme') as ThemePreference;
  if (saved === 'dark' || saved === 'light' || saved === 'system') return saved;
  return 'system';
}

export const theme = writable<ThemePreference>(getInitialTheme());

export function applyTheme(val: ThemePreference) {
  if (typeof window === 'undefined') return;
  localStorage.setItem('user_theme', val);
  theme.set(val);

  let effective: 'dark' | 'light' = 'dark';
  if (val === 'system') {
    effective = window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
  } else {
    effective = val;
  }

  document.documentElement.setAttribute('data-theme', effective);
}

if (typeof window !== 'undefined') {
  // Listen for system theme changes when in 'system' mode
  const mql = window.matchMedia('(prefers-color-scheme: light)');
  mql.addEventListener('change', () => {
    let currentPref: ThemePreference = 'system';
    theme.subscribe((t) => (currentPref = t))();
    if (currentPref === 'system') {
      applyTheme('system');
    }
  });

  // Apply on startup
  applyTheme(getInitialTheme());
}
