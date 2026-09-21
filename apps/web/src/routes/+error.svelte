<script lang="ts">
  import { page } from '$app/state';
  import { t } from '$lib/i18n';
  import Button from '$lib/components/ui/Button.svelte';
  import { goto } from '$app/navigation';
  import { base } from '$app/paths';

  let isNotFound = $derived(page.status === 404);
</script>

<svelte:head>
  <title>{isNotFound ? $t('app.notFoundTitle') : $t('app.error')} | {$t('app.title')}</title>
</svelte:head>

<main class="error-page">
  <p class="code" aria-hidden="true">{page.status}</p>
  <h1>{isNotFound ? $t('app.notFoundTitle') : $t('app.error')}</h1>
  {#if isNotFound}
    <p class="text">{$t('app.notFoundText')}</p>
  {/if}
  <Button variant="primary" onclick={() => goto(`${base}/`)}>{$t('app.backHome')}</Button>
</main>

<style>
  .error-page {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 48px 24px;
    text-align: center;
  }

  .code {
    font-size: 72px;
    font-weight: 900;
    line-height: 1;
    color: var(--color-accent);
  }

  h1 {
    font-size: 22px;
    font-weight: 800;
    color: var(--text-main);
  }

  .text {
    max-width: 360px;
    font-size: 14.5px;
    line-height: 1.6;
    color: var(--text-muted);
    margin-bottom: 8px;
  }
</style>
