<script lang="ts">
  import { page } from '$app/state';
  import { onMount } from 'svelte';
  import { t } from '$lib/i18n';
  import { campaignsApi } from '$lib/api/campaigns';
  import type { Campaign } from '$lib/types/campaign.types';
  import CardPreview from '$lib/components/cards/CardPreview.svelte';
  import CardEditor from '$lib/components/cards/CardEditor.svelte';
  import Button from '$lib/components/ui/Button.svelte';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';
  import LanguageSwitcher from '$lib/components/ui/LanguageSwitcher.svelte';
  import { showToast } from '$lib/components/ui/toast.store';

  let slug = $derived(page.params.slug);
  let campaign = $state<Campaign | null>(null);
  let loading = $state(true);
  let errorMsg = $state('');

  let toName = $state('');
  let messageText = $state('');
  let fromName = $state('');
  let isAnonymous = $state(false);

  let isGenerating = $state(false);
  let downloadedSuccess = $state(false);
  let cardPreviewComponent = $state<CardPreview | null>(null);

  async function loadCampaign() {
    if (!slug) return;
    loading = true;
    errorMsg = '';
    try {
      campaign = await campaignsApi.getPublic(slug);
    } catch (e: any) {
      errorMsg = e.message || $t('card.notFound');
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    if (slug) {
      loadCampaign();
    }
  });

  async function handleDownload() {
    if (!campaign || !cardPreviewComponent) return;

    isGenerating = true;
    try {
      const canvas = cardPreviewComponent.getCanvas();
      if (!canvas) throw new Error('Canvas not initialized');

      // 1. Trigger PNG download
      const dataUrl = canvas.toDataURL('image/png');
      const link = document.createElement('a');
      link.download = `${campaign.slug}-card.png`;
      link.href = dataUrl;
      link.click();

      // 2. Record card stats asynchronously in API
      const isMobile = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent);
      const device = /iPhone|iPad|iPod/i.test(navigator.userAgent)
        ? 'iPhone'
        : /Android/i.test(navigator.userAgent)
        ? 'Android'
        : 'Desktop';

      campaignsApi.submitCard(campaign.slug, {
        from: isAnonymous ? 'Anonymous' : fromName || 'Anonymous',
        to: toName,
        message: messageText,
        device
      }).catch(() => {
        // Non-blocking telemetry
      });

      downloadedSuccess = true;
      showToast($t('card.successToast'), 'success');

      // Reset fields for privacy after saving
      toName = '';
      messageText = '';
      fromName = '';
      isAnonymous = false;
    } catch (e: any) {
      showToast(e.message || 'Export error', 'error');
    } finally {
      isGenerating = false;
    }
  }

  function makeAnother() {
    downloadedSuccess = false;
  }
</script>

<svelte:head>
  <title>{campaign ? `${campaign.title} | ${$t('app.title')}` : $t('app.title')}</title>
</svelte:head>

<div class="card-page">
  <div class="page-wrap">
    <div class="top-nav">
      <a href="/" class="back-link">
        <span class="arrow">←</span>
        <span>{$t('app.back')}</span>
      </a>
      <LanguageSwitcher />
    </div>

    {#if loading}
      <LoadingState />
    {:else if errorMsg}
      <ErrorState message={errorMsg} onretry={loadCampaign} />
    {:else if campaign}
      <header class="header">
        <h1>{campaign.title}</h1>
        <p>{$t('card.livePreview')}</p>
      </header>

      <div class="card-box">
        <CardPreview
          bind:this={cardPreviewComponent}
          imageSrc={campaign.image}
          boxes={campaign.boxes}
          to={toName}
          from={isAnonymous ? '' : fromName}
          message={messageText}
        />
      </div>

      {#if !downloadedSuccess}
        <div class="form-card">
          <CardEditor
            to={toName}
            message={messageText}
            from={fromName}
            {isAnonymous}
            ontochange={(v) => (toName = v)}
            onmessagechange={(v) => (messageText = v)}
            onfromchange={(v) => (fromName = v)}
            onanonchange={(v) => (isAnonymous = v)}
          />

          <div class="submit-wrap">
            <Button
              variant="primary"
              loading={isGenerating}
              disabled={isGenerating}
              onclick={handleDownload}
            >
              <span>📥</span>
              <span>{isGenerating ? $t('card.generating') : $t('card.download')}</span>
            </Button>
          </div>
        </div>
      {:else}
        <div class="success-box">
          <div class="how-to-save">
            <h3>{$t('card.howToSave')}</h3>
            <p class="guide-item">📱 {$t('card.iosInstructions')}</p>
            <p class="guide-item">🤖 {$t('card.androidInstructions')}</p>
          </div>

          <Button variant="secondary" onclick={makeAnother}>
            <span>🔄</span>
            <span>{$t('card.makeAnother')}</span>
          </Button>
        </div>
      {/if}
    {/if}
  </div>
</div>

<style>
  .card-page {
    min-height: 100vh;
    display: flex;
    justify-content: center;
    padding: 20px 16px 48px;
    background: var(--bg-app);
  }

  .page-wrap {
    width: 100%;
    max-width: 480px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .top-nav {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .back-link {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 14px;
    font-weight: 700;
    color: var(--text-muted);
  }

  .back-link:hover {
    color: var(--color-accent);
  }

  :global([dir="rtl"]) .arrow {
    transform: rotate(180deg);
  }

  .header {
    text-align: center;
  }

  .header h1 {
    font-size: 22px;
    font-weight: 800;
    color: var(--text-main);
  }

  .header p {
    font-size: 13.5px;
    color: var(--text-muted);
    margin-top: 4px;
  }

  .card-box {
    width: 100%;
  }

  .form-card {
    background: var(--surface-card);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 20px;
    box-shadow: var(--shadow-sm);
  }

  .submit-wrap {
    width: 100%;
  }

  .submit-wrap :global(button) {
    width: 100%;
    min-height: 52px;
  }

  .success-box {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .how-to-save {
    background: var(--surface-card);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .how-to-save h3 {
    font-size: 15px;
    font-weight: 700;
    color: var(--text-main);
  }

  .guide-item {
    font-size: 13.5px;
    color: var(--text-muted);
    line-height: 1.6;
  }

  .success-box :global(button) {
    width: 100%;
    min-height: 50px;
  }
</style>
