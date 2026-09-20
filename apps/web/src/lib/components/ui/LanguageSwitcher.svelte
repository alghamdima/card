<script lang="ts">
  import { toggleLocale, locale } from '../../i18n';
</script>

<button
  type="button"
  class="icon-btn lang-btn"
  onclick={toggleLocale}
  title={$locale === 'ar' ? 'Switch to English' : 'التحويل إلى العربية'}
  aria-label="Switch Language"
>
  <span class="lang-code">{$locale === 'ar' ? 'EN' : 'ع'}</span>
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

  .lang-code {
    font-size: 13px;
    font-weight: 800;
    letter-spacing: -0.2px;
  }
</style>
