<script lang="ts">
  import { toggleLocale, locale } from '../../i18n';
</script>

<button
  type="button"
  class="lang-btn"
  onclick={toggleLocale}
  aria-label="Switch Language"
>
  <span class="globe">🌐</span>
  <span>{$locale === 'ar' ? 'English' : 'العربية'}</span>
</button>

<style>
  .lang-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 14px;
    border-radius: var(--radius-full);
    background: var(--surface-2);
    border: 1px solid var(--color-border);
    color: var(--text-main);
    font-size: 13.5px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .lang-btn:hover {
    background: var(--surface-3);
    border-color: var(--color-border-strong);
  }

  .globe {
    font-size: 14px;
  }
</style>
