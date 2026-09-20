<script lang="ts">
  import { t } from '../../i18n';
  import Button from './Button.svelte';

  interface Props {
    message?: string;
    onretry?: () => void;
  }

  let { message = '', onretry }: Props = $props();
</script>

<div class="error-state">
  <span class="icon">⚠️</span>
  <h4>{$t('app.error')}</h4>
  <p>{message}</p>
  {#if onretry}
    <Button variant="secondary" onclick={onretry}>
      {$t('app.retry')}
    </Button>
  {/if}
</div>

<style>
  .error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 48px 24px;
    text-align: center;
    gap: 12px;
  }

  .icon {
    font-size: 36px;
  }

  h4 {
    font-size: 17px;
    font-weight: 700;
    color: var(--color-danger);
  }

  p {
    font-size: 14px;
    color: var(--text-muted);
    max-width: 340px;
    line-height: 1.5;
  }
</style>
