<script lang="ts">
  import { onMount } from 'svelte';
  import { t } from '$lib/i18n';
  import { campaignsApi } from '$lib/api/campaigns';
  import type { CampaignSummary } from '$lib/types/campaign.types';
  import LanguageSwitcher from '$lib/components/ui/LanguageSwitcher.svelte';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import EmptyState from '$lib/components/ui/EmptyState.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';

  let campaigns = $state<CampaignSummary[]>([]);
  let loading = $state(true);
  let errorMsg = $state('');

  async function loadData() {
    loading = true;
    errorMsg = '';
    try {
      const res = await campaignsApi.listPublic();
      campaigns = res.campaigns || [];
    } catch (e: any) {
      errorMsg = e.message || 'Failed to load campaigns';
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
        <img src="/images/brand/aljuf-ar.png" alt="Abdul Latif Jameel Finance" class="main-aljuf-logo" />
      </div>
      <h1>{$t('app.title')}</h1>
      <p>{$t('app.subtitle')}</p>
    </div>
    <div class="lang-wrap">
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
                <img src={camp.thumb} alt={camp.title} loading="lazy" />
              </div>
            {/if}
            <div class="card-body">
              <h3 class="camp-title">{camp.title}</h3>
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
    <p>Abdul Latif Jameel Finance &middot; {$t('app.allRightsReserved')} 2026</p>
    <a href="/login" class="admin-link">{$t('nav.admin')}</a>
  </footer>
</div>

<style>
  .page-container {
    max-width: 900px;
    margin: 0 auto;
    padding: 32px 16px 48px;
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  .top-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 36px;
    gap: 16px;
  }

  .header-content {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .brand-row {
    margin-bottom: 4px;
  }

  .main-aljuf-logo {
    height: 48px;
    object-fit: contain;
  }

  h1 {
    font-size: 26px;
    font-weight: 800;
    color: var(--text-main);
  }

  p {
    font-size: 15px;
    color: var(--text-muted);
  }

  .main-content {
    flex: 1;
  }

  .campaign-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 20px;
  }

  .campaign-card {
    background: var(--surface-card);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    overflow: hidden;
    transition: transform 0.15s ease, border-color 0.15s ease;
    display: flex;
    flex-direction: column;
  }

  .campaign-card:hover {
    transform: translateY(-2px);
    border-color: var(--color-accent);
  }

  .card-thumb {
    width: 100%;
    aspect-ratio: 1080 / 1350;
    overflow: hidden;
    background: var(--surface-2);
  }

  .card-thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .card-body {
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .camp-title {
    font-size: 16px;
    font-weight: 700;
    color: var(--text-main);
  }

  .card-meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .badge {
    background: var(--color-success-wash);
    color: var(--color-success);
    font-size: 11px;
    font-weight: 700;
    padding: 3px 8px;
    border-radius: var(--radius-full);
  }

  .total-cards {
    font-size: 12.5px;
    color: var(--text-dim);
  }

  .page-footer {
    margin-top: 48px;
    padding-top: 24px;
    border-top: 1px solid var(--color-border);
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 13px;
    color: var(--text-dim);
  }

  .admin-link {
    color: var(--text-muted);
    font-weight: 600;
  }

  .admin-link:hover {
    color: var(--color-accent);
  }

  @media (max-width: 600px) {
    .top-header {
      flex-direction: column-reverse;
      align-items: flex-start;
    }
    .page-footer {
      flex-direction: column;
      gap: 12px;
    }
  }
</style>
