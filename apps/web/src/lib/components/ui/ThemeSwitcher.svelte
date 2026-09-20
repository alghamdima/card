<script lang="ts">
  import { theme, applyTheme, type ThemePreference } from '../../stores/theme.store';
  import { t } from '../../i18n';

  function cycleTheme() {
    let next: ThemePreference = 'dark';
    if ($theme === 'dark') next = 'light';
    else if ($theme === 'light') next = 'system';
    else next = 'dark';
    applyTheme(next);
  }
</script>

<button
  type="button"
  class="theme-toggle-btn"
  onclick={cycleTheme}
  title={`Theme: ${$theme} (click to toggle)`}
  aria-label="Toggle light dark theme"
>
  {#if $theme === 'dark'}
    <span class="icon">🌙</span>
    <span class="label">{$t('theme.dark') || 'داكن'}</span>
  {:else if $theme === 'light'}
    <span class="icon">☀️</span>
    <span class="label">{$t('theme.light') || 'فاتح'}</span>
  {:else}
    <span class="icon">💻</span>
    <span class="label">{$t('theme.system') || 'تلقائي'}</span>
  {/if}
</button>

<style>
  .theme-toggle-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border-radius: var(--radius-full);
    background: var(--surface-2);
    border: 1px solid var(--color-border);
    color: var(--text-main);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .theme-toggle-btn:hover {
    background: var(--surface-3);
    border-color: var(--color-accent);
    color: var(--color-accent);
  }

  .icon {
    font-size: 13.5px;
  }

  .label {
    font-size: 12px;
  }
</style>
