<script lang="ts">
  import type { Snippet } from 'svelte';
  import { t } from '../../i18n';

  interface Props {
    open: boolean;
    title?: string;
    onclose: () => void;
    children?: Snippet;
  }

  let { open = false, title = '', onclose, children }: Props = $props();
</script>

{#if open}
  <div class="overlay" onclick={onclose} role="presentation">
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <div
      class="modal"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      tabindex="-1"
    >
      <div class="head">
        <h3>{title}</h3>
        <button class="close-btn" onclick={onclose} aria-label={$t('app.cancel')}>✕</button>
      </div>
      <div class="body">
        {#if children}
          {@render children()}
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px;
    z-index: 1000;
  }

  .modal {
    background: var(--surface-card);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    width: 100%;
    max-width: 600px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    box-shadow: var(--shadow-card);
    overflow: hidden;
  }

  .head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 18px 24px;
    border-bottom: 1px solid var(--color-border);
  }

  .head h3 {
    font-size: 18px;
    font-weight: 800;
    color: var(--text-main);
  }

  .close-btn {
    background: transparent;
    border: none;
    color: var(--text-muted);
    font-size: 18px;
    cursor: pointer;
    padding: 4px;
  }

  .close-btn:hover {
    color: var(--text-main);
  }

  .body {
    padding: 24px;
    overflow-y: auto;
  }
</style>
