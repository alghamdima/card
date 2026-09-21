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
  class="icon-btn theme-toggle-btn"
  onclick={cycleTheme}
  title={`${$t(`theme.${$theme}`)} — ${$t('app.themeToggle')}`}
  aria-label={$t('app.themeToggle')}
>
  {#if $theme === 'dark'}
    <!-- Moon icon -->
    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/>
    </svg>
  {:else if $theme === 'light'}
    <!-- Sun icon -->
    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <circle cx="12" cy="12" r="4"/>
      <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"/>
    </svg>
  {:else}
    <!-- Auto / Monitor icon with dot -->
    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <rect width="20" height="14" x="2" y="3" rx="2"/>
      <line x1="8" x2="16" y1="21" y2="21"/>
      <line x1="12" x2="12" y1="17" y2="21"/>
      <circle cx="12" cy="10" r="1.5" fill="currentColor"/>
    </svg>
  {/if}
</button>

<style>
  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 38px;
    height: 38px;
    border-radius: var(--radius-full);
    background: var(--surface-1);
    border: 1px solid var(--color-border);
    color: var(--text-muted);
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    box-shadow: var(--shadow-sm);
  }

  .icon-btn:hover {
    background: var(--surface-2);
    border-color: var(--color-accent);
    color: var(--color-accent);
    transform: translateY(-1px);
  }

  .icon-btn:active {
    transform: translateY(0);
  }
</style>
