<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { t, locale, translateError } from '$lib/i18n';
  import { campaignsApi } from '$lib/api/campaigns';
  import type { PublicCampaignSummary } from '$lib/types/campaign.types';
  import BrandLogo from '$lib/components/ui/BrandLogo.svelte';
  import LanguageSwitcher from '$lib/components/ui/LanguageSwitcher.svelte';
  import ThemeSwitcher from '$lib/components/ui/ThemeSwitcher.svelte';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import EmptyState from '$lib/components/ui/EmptyState.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';

  let campaigns = $state<PublicCampaignSummary[]>([]);
  let loading = $state(true);
  let errorMsg = $state('');

  const year = new Date().getFullYear();

  function getCampaignTitle(camp: PublicCampaignSummary, curLocale: string): string {
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
    } catch (e) {
      errorMsg = translateError(e);
    } finally {
      loading = false;
    }
  }

  onMount(loadData);
</script>

<svelte:head>
  <title>{$t('app.productName')} | {$t('app.title')}</title>
</svelte:head>

<div class="page-container">
  <!-- Minimal Top Navigation Bar -->
  <header class="top-nav-bar">
    <div class="brand-group">
      <a href="{base}/" class="brand-link" title={$t('app.companyName')}>
        <BrandLogo />
      </a>
    </div>

    <div class="controls-group">
      <ThemeSwitcher />
      <LanguageSwitcher />
    </div>
  </header>

  <!-- Hero Section -->
  <section class="hero-section">
    <h1 class="hero-title">{$t('app.productName')}</h1>
    <p class="hero-subtitle">{$t('app.subtitle')}</p>
  </section>

  <!-- Occasions Grid -->
  <main class="main-content">
    {#if loading}
      <LoadingState />
    {:else if errorMsg}
      <ErrorState message={errorMsg} onretry={loadData} />
    {:else if campaigns.length === 0}
      <EmptyState icon="💌" title={$t('card.noCampaigns')} />
    {:else}
      <div class="campaign-grid">
        {#each campaigns as camp (camp.slug)}
          <a href="{base}/cards/{camp.slug}" class="campaign-card">
            {#if camp.thumb}
              <div class="card-thumb-wrap">
                <img src={camp.thumb} alt={getCampaignTitle(camp, $locale)} loading="lazy" decoding="async" />
                <div class="card-overlay">
                  <span class="preview-action-chip">
                    {$t('card.createCard')}
                    <span class="chip-arrow flip-rtl" aria-hidden="true">→</span>
                  </span>
                </div>
              </div>
            {/if}
            <div class="card-body">
              <h3 class="camp-title">{getCampaignTitle(camp, $locale)}</h3>
              <div class="card-footer-meta">
                <span class="status-indicator">
                  <span class="status-dot"></span>
                  {$t('app.active')}
                </span>
              </div>
            </div>
          </a>
        {/each}
      </div>
    {/if}
  </main>

  <!-- Clean Footer -->
  <footer class="page-footer">
    <div class="footer-inner">
      <p>{$t('app.companyName')} &middot; {$t('app.allRightsReserved')} {year}</p>
      <a href="{base}/login" class="admin-link">{$t('nav.admin')}</a>
    </div>
  </footer>
</div>

<style>
  .page-container {
    max-width: 1040px;
    margin: 0 auto;
    padding: 24px 20px 48px;
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  /* Clean Top Nav */
  .top-nav-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0 20px;
    border-bottom: 1px solid var(--color-border);
    margin-bottom: 32px;
  }

  .brand-group {
    display: flex;
    align-items: center;
  }

  .brand-link {
    display: flex;
    align-items: center;
    text-decoration: none;
  }

  .controls-group {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  /* Hero Section */
  .hero-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 8px;
    margin-bottom: 36px;
    padding: 8px 0;
  }

  .hero-title {
    font-size: 24px;
    font-weight: 800;
    color: var(--text-main);
    letter-spacing: -0.3px;
    margin: 0;
  }

  .hero-subtitle {
    font-size: 14.5px;
    color: var(--text-muted);
    margin: 0;
    line-height: 1.5;
  }

  .main-content {
    flex: 1;
  }

  /* Occasion Cards Grid */
  .campaign-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(290px, 1fr));
    gap: 24px;
  }

  .campaign-card {
    background: var(--surface-card);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    overflow: hidden;
    transition: all 0.22s cubic-bezier(0.16, 1, 0.3, 1);
    display: flex;
    flex-direction: column;
    box-shadow: var(--shadow-sm);
    text-decoration: none;
    position: relative;
  }

  .campaign-card:hover {
    transform: translateY(-4px);
    border-color: var(--color-accent);
    box-shadow: var(--shadow-card);
  }

  .card-thumb-wrap {
    width: 100%;
    aspect-ratio: 16 / 10;
    overflow: hidden;
    background: var(--surface-2);
    position: relative;
  }

  .card-thumb-wrap img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: transform 0.35s ease;
  }

  .campaign-card:hover .card-thumb-wrap img {
    transform: scale(1.04);
  }

  .card-overlay {
    position: absolute;
    inset: 0;
    background: linear-gradient(to top, rgba(0, 0, 0, 0.45) 0%, transparent 60%);
    display: flex;
    align-items: flex-end;
    justify-content: flex-end;
    padding: 14px;
    opacity: 0;
    transition: opacity 0.2s ease;
  }

  .campaign-card:hover .card-overlay {
    opacity: 1;
  }

  .preview-action-chip {
    padding: 5px 12px;
    background: var(--surface-1);
    color: var(--text-main);
    border-radius: var(--radius-full);
    font-size: 12px;
    font-weight: 700;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
    border: 1px solid var(--color-border);
  }

  .card-body {
    padding: 18px 20px 20px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    flex: 1;
    justify-content: space-between;
  }

  .camp-title {
    font-size: 17px;
    font-weight: 800;
    color: var(--text-main);
    line-height: 1.35;
    margin: 0;
  }

  .card-footer-meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 4px;
  }

  .status-indicator {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    font-weight: 700;
    color: var(--color-success);
    background: var(--color-success-wash);
    padding: 3px 10px;
    border-radius: var(--radius-full);
  }

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
  }

  /* Footer */
  .page-footer {
    margin-top: 56px;
    padding-top: 24px;
    border-top: 1px solid var(--color-border);
  }

  .footer-inner {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 13px;
    color: var(--text-dim);
    flex-wrap: wrap;
    gap: 12px;
  }

  .admin-link {
    color: var(--text-muted);
    font-weight: 600;
    transition: color 0.15s ease;
    text-decoration: none;
  }

  .admin-link:hover {
    color: var(--color-accent);
  }

  .chip-arrow {
    display: inline-block;
  }
</style>
