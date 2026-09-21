<script lang="ts">
  import type { Snippet } from 'svelte';
  import { t } from '../../i18n';

  interface Props {
    open: boolean;
    title?: string;
    maxWidth?: string;
    onclose: () => void;
    children?: Snippet;
  }

  let { open = false, title = '', maxWidth = '600px', onclose, children }: Props = $props();

  const titleId = `modal-title-${Math.random().toString(36).slice(2, 9)}`;
  const FOCUSABLE = 'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

  let dialogEl = $state<HTMLDivElement>();

  // While open: lock page scroll, move focus into the dialog and restore it on close.
  $effect(() => {
    if (!open) return;

    const previouslyFocused = document.activeElement as HTMLElement | null;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    dialogEl?.focus();

    return () => {
      document.body.style.overflow = previousOverflow;
      previouslyFocused?.focus?.();
    };
  });

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.stopPropagation();
      onclose();
      return;
    }

    // Keep Tab / Shift+Tab inside the dialog.
    if (e.key === 'Tab' && dialogEl) {
      const items = [...dialogEl.querySelectorAll<HTMLElement>(FOCUSABLE)];
      if (items.length === 0) {
        e.preventDefault();
        return;
      }
      const first = items[0];
      const last = items[items.length - 1];
      if (e.shiftKey && (document.activeElement === first || document.activeElement === dialogEl)) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    }
  }

  // Only a click that starts and ends on the backdrop closes the dialog, so selecting text and releasing outside does not.
  let pressStartedOnOverlay = false;
</script>

{#if open}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="overlay"
    role="presentation"
    onpointerdown={(e) => (pressStartedOnOverlay = e.target === e.currentTarget)}
    onclick={(e) => {
      if (pressStartedOnOverlay && e.target === e.currentTarget) onclose();
    }}
    onkeydown={handleKeydown}
  >
    <div
      class="modal"
      style:max-width={maxWidth}
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      tabindex="-1"
      bind:this={dialogEl}
    >
      <div class="head">
        <h3 id={titleId}>{title}</h3>
        <button type="button" class="close-btn" onclick={onclose} aria-label={$t('app.close')}>✕</button>
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
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    box-shadow: var(--shadow-card);
    overflow: hidden;
    outline: none;
  }

  .head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
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
    padding: 4px 8px;
    border-radius: var(--radius-sm);
  }

  .close-btn:hover {
    color: var(--text-main);
  }

  .body {
    padding: 24px;
    overflow-y: auto;
  }
</style>
