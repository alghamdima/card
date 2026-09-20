<script lang="ts">
  import { toasts } from './toast.store';
</script>

<div class="toast-container" aria-live="polite">
  {#each $toasts as item (item.id)}
    <div class="toast toast-{item.type || 'success'}">
      <span>{item.text}</span>
    </div>
  {/each}
</div>

<style>
  .toast-container {
    position: fixed;
    bottom: 24px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    flex-direction: column;
    gap: 8px;
    z-index: 9999;
    pointer-events: none;
  }

  .toast {
    padding: 12px 24px;
    border-radius: var(--radius-full);
    font-size: 14px;
    font-weight: 700;
    box-shadow: var(--shadow-card);
    animation: toastIn 0.25s cubic-bezier(0.16, 1, 0.3, 1);
    pointer-events: auto;
  }

  .toast-success {
    background: #1E2022;
    color: #FFFFFF;
    border: 1px solid var(--color-success);
  }

  .toast-error {
    background: #2D1412;
    color: var(--color-danger);
    border: 1px solid var(--color-danger);
  }

  .toast-info {
    background: var(--surface-card);
    color: var(--text-main);
    border: 1px solid var(--color-border);
  }

  @keyframes toastIn {
    from {
      opacity: 0;
      transform: translateY(12px) scale(0.96);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
</style>
