<script lang="ts">
  import { t } from '../../i18n';

  interface Props {
    offset: number;
    total: number;
    pageSize: number;
    disabled?: boolean;
    onchange: (newOffset: number) => void;
  }

  let { offset, total, pageSize, disabled = false, onchange }: Props = $props();

  let pageCount = $derived(Math.max(1, Math.ceil(total / pageSize)));
  let currentPage = $derived(Math.floor(offset / pageSize) + 1);
</script>

{#if pageCount > 1}
  <nav class="pager" aria-label={$t('app.pageOf', { page: currentPage, pages: pageCount })}>
    <button
      type="button"
      class="pager-btn"
      disabled={disabled || offset <= 0}
      onclick={() => onchange(Math.max(0, offset - pageSize))}
    >
      {$t('app.previous')}
    </button>
    <span class="pager-info">{$t('app.pageOf', { page: currentPage, pages: pageCount })}</span>
    <button
      type="button"
      class="pager-btn"
      disabled={disabled || currentPage >= pageCount}
      onclick={() => onchange(offset + pageSize)}
    >
      {$t('app.next')}
    </button>
  </nav>
{/if}

<style>
  .pager {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 16px;
    padding: 16px;
    border-top: 1px solid var(--color-border);
  }

  .pager-btn {
    padding: 7px 16px;
    background: var(--surface-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-full);
    color: var(--text-main);
    font-size: 13px;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .pager-btn:hover:not(:disabled) {
    border-color: var(--color-accent);
    color: var(--color-accent);
  }

  .pager-btn:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .pager-info {
    font-size: 13px;
    color: var(--text-muted);
    font-variant-numeric: tabular-nums;
  }
</style>
