<script lang="ts">
  import { t } from '../../i18n';

  interface Props {
    title?: string;
    message?: string;
    icon?: string;
  }

  let {
    title = '',
    message = '',
    icon = '📭'
  }: Props = $props();
</script>

<div class="empty-state">
  <span class="icon">{icon}</span>
  <h4>{title || $t('app.error')}</h4>
  {#if message}
    <p>{message}</p>
  {/if}
</div>

<style>
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 48px 24px;
    text-align: center;
    gap: 8px;
    color: var(--text-muted);
  }

  .icon {
    font-size: 40px;
    margin-bottom: 8px;
  }

  h4 {
    font-size: 17px;
    font-weight: 700;
    color: var(--text-main);
  }

  p {
    font-size: 14px;
    max-width: 320px;
    line-height: 1.5;
  }
</style>
