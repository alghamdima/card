<script lang="ts">
  import { t } from '../../i18n';

  interface Props {
    text?: string;
  }

  let { text = '' }: Props = $props();
</script>

<div class="loading-state">
  <div class="spinner"></div>
  <p>{text || $t('app.loading')}</p>
</div>

<style>
  .loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 60px 24px;
    gap: 16px;
    color: var(--text-muted);
    font-size: 14.5px;
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid var(--color-border);
    border-top-color: var(--color-accent);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
