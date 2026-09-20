<script lang="ts">
  import { onMount } from 'svelte';
  import { t, locale } from '$lib/i18n';
  import { campaignsApi } from '$lib/api/campaigns';
  import type { CampaignSummary } from '$lib/types/campaign.types';
  import LanguageSwitcher from '$lib/components/ui/LanguageSwitcher.svelte';
  import ThemeSwitcher from '$lib/components/ui/ThemeSwitcher.svelte';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import EmptyState from '$lib/components/ui/EmptyState.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';

  let campaigns = $state<CampaignSummary[]>([]);
  let loading = $state(true);
  let errorMsg = $state('');

  function getCampaignTitle(camp: CampaignSummary, curLocale: string): string {
    if (curLocale === 'en') {
      return camp.titleEN || camp.title || '';
    }
    return camp.titleAR || camp.title || '';
  }

  async function loadData() {
    loading = true;
    errorMsg = '';
    try {
      const res = await campaignsApi.listPublic();
      campaigns = res.campaigns || [];
    } catch (e: any) {
      errorMsg = e.message || 'Failed to load occasions';
    } finally {
      loading = false;
    }
  }

  onMount(loadData);
</script>

<svelte:head>
  <title>{$t('app.title')}</title>
</svelte:head>

<div class="page-container">
  <header class="top-header">
    <div class="header-content">
      <div class="brand-row">
        <img
          src={$locale === 'en' ? '/images/brand/aljuf-en-tight.png' : '/images/brand/aljuf-ar-tight.png'}
          alt="Abdul Latif Jameel Finance"
          class="main-aljuf-logo light-only"
        />
        <img
          src={$locale === 'en' ? '/images/brand/aljuf-en-white-tight.png' : '/images/brand/aljuf-ar-white-tight.png'}
          alt="Abdul Latif Jameel Finance"
          class="main-aljuf-logo dark-only"
        />
      </div>
      <h1>{$t('app.productName')}</h1>
      <p>{$t('app.subtitle')}</p>
    </div>
    <div class="controls-wrap">
      <ThemeSwitcher />
      <LanguageSwitcher />
    </div>
  </header>

  <main class="main-content">
    {#if loading}
      <LoadingState />
    {:else if errorMsg}
      <ErrorState message={errorMsg} onretry={loadData} />
    {:else if campaigns.length === 0}
      <EmptyState
        icon="💌"
        title={$t('card.notFound')}
        message={$t('card.noCampaigns')}
      />
    {:else}
      <div class="campaign-grid">
        {#each campaigns as camp (camp.slug)}
          <a href="/cards/{camp.slug}" class="campaign-card">
            {#if camp.thumb}
              <div class="card-thumb">
                <img src={camp.thumb} alt={getCampaignTitle(camp, $locale)} loading="lazy" />
              </div>
            {/if}
            <div class="card-body">
              <h3 class="camp-title">{getCampaignTitle(camp, $locale)}</h3>
              <div class="card-meta">
                <span class="badge">{$t('app.active')}</span>
                <span class="total-cards">{camp.totalCards} {$t('admin.cardsCount')}</span>
              </div>
            </div>
          </a>
        {/each}
      </div>
    {/if}
  </main>

  <footer class="page-footer">
    <div class="footer-inner">
      <p>Abdul Latif Jameel Finance &middot; {$t('app.allRightsReserved')} 2026</p>
      <a href="/login" class="admin-link">{$t('nav.admin')}</a>
    </div>
  </footer>
</div>

<style>
  .page-container {
    max-width: 960px;
    margin: 0 auto;
    padding: 36px 20px 48px;
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  .top-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 40px;
    gap: 20px;
    flex-wrap: wrap;
  }

  .header-content {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .brand-row {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 6px;
  }

  .main-aljuf-logo {
    height: 60px;
    width: auto;
    object-fit: contain;
    transition: transform 0.2s ease;
  }

  :global([data-theme="light"]) .dark-only {
    display: none !important;
  }

  :global([data-theme="light"]) .light-only {
    display: block !important;
  }

  :global([data-theme="dark"]) .light-only,
  :global(:root:not([data-theme="light"])) .light-only {
    display: none !important;
  }

  :global([data-theme="dark"]) .dark-only,
  :global(:root:not([data-theme="light"])) .dark-only {
    display: block !important;
  }

  .controls-wrap {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  h1 {
    font-size: 28px;
    font-weight: 800;
    color: var(--text-main);
    letter-spacing: -0.5px;
  }

  p {
    font-size: 15px;
    color: var(--text-muted);
    max-width: 540px;
    line-height: 1.5;
  }

  .main-content {
    flex: 1;
  }

  .campaign-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 24px;
  }

  .campaign-card {
    background: var(--surface-card);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    overflow: hidden;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    display: flex;
    flex-direction: column;
    box-shadow: var(--shadow-sm);
  }

  .campaign-card:hover {
    transform: translateY(-4px);
    border-color: var(--color-accent);
    box-shadow: var(--shadow-card);
  }

  .card-thumb {
    width: 100%;
    aspect-ratio: 16 / 10;
    overflow: hidden;
    background: var(--surface-2);
  }

  .card-thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .card-body {
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    flex: 1;
    justify-content: space-between;
  }

  .camp-title {
    font-size: 17px;
    font-weight: 700;
    color: var(--text-main);
    line-height: 1.4;
  }

  .card-meta {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .badge {
    padding: 3px 8px;
    background: var(--color-success-wash);
    color: var(--color-success);
    border-radius: var(--radius-full);
    font-size: 11.5px;
    font-weight: 700;
  }

  .total-cards {
    font-size: 12.5px;
    color: var(--text-dim);
    font-weight: 600;
  }

  .page-footer {
    margin-top: 48px;
    padding-top: 24px;
    border-top: 1px solid var(--color-border);
  }

  .footer-inner {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 13px;
    color: var(--text-dim);
  }

  .admin-link {
    color: var(--text-muted);
    font-weight: 600;
    transition: color 0.15s ease;
  }

  .admin-link:hover {
    color: var(--color-accent);
  }
</style>
