<script lang="ts">
  import { onMount } from 'svelte';
  import { t } from '$lib/i18n';
  import { campaignsApi } from '$lib/api/campaigns';
  import type { CampaignAnalytics, Card } from '$lib/types/campaign.types';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';
  import { showToast } from '$lib/components/ui/toast.store';

  let analyticsCampaigns = $state<CampaignAnalytics[]>([]);
  let activeTabSlug = $state<string>('');
  let loading = $state(true);
  let errorMsg = $state('');
  let lastUpdatedTime = $state('');

  // Selected Campaign details & cards
  let campaignCards = $state<Card[]>([]);
  let loadingCards = $state(false);
  let searchQuery = $state('');

  let filteredCards = $derived.by(() => {
    if (!searchQuery.trim()) return campaignCards;
    const q = searchQuery.toLowerCase();
    return campaignCards.filter(
      (c) =>
        (c.to && c.to.toLowerCase().includes(q)) ||
        (c.from && c.from.toLowerCase().includes(q)) ||
        (c.message && c.message.toLowerCase().includes(q))
    );
  });

  let activeCampaign = $derived.by(() => {
    return analyticsCampaigns.find((c) => c.slug === activeTabSlug) || analyticsCampaigns[0];
  });

  function updateClock() {
    const now = new Date();
    lastUpdatedTime = now.toLocaleTimeString('en-US', { hour12: false });
  }

  async function loadData() {
    loading = true;
    errorMsg = '';
    updateClock();
    try {
      const res = await campaignsApi.getAnalyticsOverview();
      analyticsCampaigns = res.campaigns || [];
      if (analyticsCampaigns.length > 0 && !activeTabSlug) {
        activeTabSlug = analyticsCampaigns[0].slug;
      }
      if (activeTabSlug) {
        await loadCampaignCards(activeTabSlug);
      }
    } catch (e: any) {
      errorMsg = e.message || 'Failed to load analytics';
    } finally {
      loading = false;
    }
  }

  async function loadCampaignCards(slug: string) {
    loadingCards = true;
    try {
      const res = await campaignsApi.getCampaignCards(slug);
      campaignCards = res.cards || [];
    } catch (e: any) {
      showToast('Error loading cards: ' + e.message, 'error');
    } finally {
      loadingCards = false;
    }
  }

  async function handleTabChange(slug: string) {
    activeTabSlug = slug;
    searchQuery = '';
    await loadCampaignCards(slug);
  }

  function handleExportCsv() {
    if (!activeTabSlug) return;
    const url = campaignsApi.getExportUrl(activeTabSlug);
    window.open(url, '_blank');
  }

  onMount(() => {
    loadData();
    const interval = setInterval(updateClock, 1000);
    return () => clearInterval(interval);
  });
</script>

<svelte:head>
  <title>{$t('admin.analyticsTitle')} | {$t('app.title')}</title>
</svelte:head>

<div class="analytics-page">
  <!-- Top Header matching screenshot -->
  <header class="analytics-header">
    <div class="header-left">
      <div class="title-row">
        <span class="bar-icon">📊</span>
        <h2>Card Generating Analytics</h2>
      </div>
      <span class="updated-time">Updated: {lastUpdatedTime}</span>
    </div>

    <div class="header-actions">
      <button class="btn-round" onclick={loadData} title={$t('admin.refresh')}>
        <span class="btn-icon">🔄</span>
        <span>Refresh</span>
      </button>

      <button class="btn-round" onclick={handleExportCsv} title={$t('admin.exportCsv')}>
        <span class="btn-icon">📥</span>
        <span>Export CSV</span>
      </button>

      <a href="/admin/cards" class="btn-round admin-pill">
        <span>Admin</span>
      </a>
    </div>
  </header>

  {#if loading}
    <LoadingState />
  {:else if errorMsg}
    <ErrorState message={errorMsg} onretry={loadData} />
  {:else}
    <!-- Horizontal Campaign Navigation Tabs matching screenshot -->
    <div class="campaign-tabs-nav">
      {#each analyticsCampaigns as camp (camp.slug)}
        <button
          class="nav-tab-item"
          class:active={activeTabSlug === camp.slug}
          onclick={() => handleTabChange(camp.slug)}
        >
          {camp.title}
        </button>
      {/each}
    </div>

    <!-- ALL TIME STATS Card matching screenshot -->
    <div class="stats-section">
      <div class="section-badge-title">
        <span class="icon">📊</span>
        <span>ALL TIME STATS</span>
      </div>

      <div class="stat-card-wide">
        <div class="stat-top-meta">
          <span class="stat-label">TOTAL CARDS GENERATED</span>
          <span class="card-icon-pill">🎴</span>
        </div>
        <div class="stat-big-number">
          {activeCampaign?.totalCards ?? 0}
        </div>
      </div>
    </div>

    <!-- EMPLOYEE LIST Section matching screenshot -->
    <div class="employee-list-section">
      <div class="section-badge-title">
        <span class="icon">📋</span>
        <span>EMPLOYEE LIST</span>
      </div>

      <div class="employee-table-card">
        <div class="table-card-header">
          <div class="header-counter-box">
            <h4>Employee Cards</h4>
            <span class="total-badge">{campaignCards.length} total</span>
          </div>

          <div class="search-box">
            <span class="search-icon">🔍</span>
            <input
              type="text"
              placeholder="البحث بالاسم أو المسمى..."
              bind:value={searchQuery}
              class="search-input"
            />
          </div>
        </div>

        {#if loadingCards}
          <div class="cards-loading-wrap">
            <LoadingState />
          </div>
        {:else if filteredCards.length === 0}
          <div class="empty-cards-wrap">
            <p>{$t('admin.noCardsForCampaign')}</p>
          </div>
        {:else}
          <div class="table-scroll-container">
            <table class="analytics-data-table">
              <thead>
                <tr>
                  <th>FULL NAME</th>
                  <th>JOB TITLE / DETAILS</th>
                  <th>LANGUAGE</th>
                  <th>DATE</th>
                  <th>TIME</th>
                </tr>
              </thead>
              <tbody>
                {#each filteredCards as card (card.id)}
                  <tr>
                    <td class="name-cell">
                      <strong>{card.to || '-'}</strong>
                    </td>
                    <td class="details-cell">
                      {card.message || (card.fieldValues && card.fieldValues['job_title']) || '-'}
                    </td>
                    <td>
                      <span class="lang-tag" class:en={card.lang === 'en'}>
                        {card.lang === 'en' ? 'English' : 'عربي'}
                      </span>
                    </td>
                    <td class="date-cell">{card.date}</td>
                    <td class="time-cell">{card.time}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .analytics-page {
    display: flex;
    flex-direction: column;
    gap: 28px;
    width: 100%;
  }

  /* Header */
  .analytics-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;
  }

  .header-left {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .title-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .bar-icon {
    font-size: 20px;
  }

  .title-row h2 {
    font-size: 20px;
    font-weight: 800;
    color: var(--color-accent);
  }

  .updated-time {
    font-size: 12px;
    color: var(--text-dim);
    font-family: monospace;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .btn-round {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 7px 16px;
    background: transparent;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-full);
    color: var(--text-main);
    font-size: 13px;
    font-weight: 700;
    cursor: pointer;
    text-decoration: none;
    transition: all 0.15s ease;
  }

  .btn-round:hover {
    background: var(--surface-2);
    border-color: var(--color-accent);
    color: var(--color-accent);
  }

  .admin-pill {
    background: var(--surface-3);
  }

  /* Horizontal Tabs */
  .campaign-tabs-nav {
    display: flex;
    gap: 32px;
    border-bottom: 2px solid var(--color-border);
    overflow-x: auto;
    padding-bottom: 2px;
  }

  .nav-tab-item {
    background: none;
    border: none;
    padding: 12px 6px;
    color: var(--text-muted);
    font-size: 15px;
    font-weight: 700;
    cursor: pointer;
    position: relative;
    white-space: nowrap;
    transition: color 0.15s;
  }

  .nav-tab-item:hover {
    color: var(--text-main);
  }

  .nav-tab-item.active {
    color: #FFFFFF;
  }

  .nav-tab-item.active::after {
    content: '';
    position: absolute;
    bottom: -2px;
    left: 0;
    right: 0;
    height: 3px;
    background: var(--color-accent);
    border-radius: 3px 3px 0 0;
  }

  /* Section Title Badge */
  .section-badge-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 11.5px;
    font-weight: 800;
    letter-spacing: 0.8px;
    color: var(--text-dim);
    margin-bottom: 12px;
    text-transform: uppercase;
  }

  /* ALL TIME STATS Card */
  .stats-section {
    display: flex;
    flex-direction: column;
  }

  .stat-card-wide {
    background: var(--surface-1);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    padding: 24px 28px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  }

  .stat-top-meta {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .stat-label {
    font-size: 12.5px;
    font-weight: 800;
    letter-spacing: 0.6px;
    color: var(--text-muted);
  }

  .card-icon-pill {
    font-size: 16px;
    background: var(--surface-2);
    padding: 4px 8px;
    border-radius: var(--radius-sm);
  }

  .stat-big-number {
    font-size: 48px;
    font-weight: 900;
    color: var(--color-accent);
    line-height: 1;
  }

  /* EMPLOYEE LIST */
  .employee-list-section {
    display: flex;
    flex-direction: column;
  }

  .employee-table-card {
    background: var(--surface-card);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    overflow: hidden;
    box-shadow: var(--shadow-sm);
  }

  .table-card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid var(--color-border);
    flex-wrap: wrap;
    gap: 16px;
  }

  .header-counter-box {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .header-counter-box h4 {
    font-size: 16px;
    font-weight: 800;
    color: var(--text-main);
  }

  .total-badge {
    padding: 3px 10px;
    background: var(--surface-3);
    border-radius: var(--radius-full);
    font-size: 12px;
    font-weight: 700;
    color: var(--color-accent);
  }

  .search-box {
    display: flex;
    align-items: center;
    gap: 8px;
    background: var(--surface-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    padding: 6px 14px;
  }

  .search-icon {
    font-size: 13px;
    color: var(--text-dim);
  }

  .search-input {
    background: transparent;
    border: none;
    color: var(--text-main);
    font-size: 13.5px;
    outline: none;
    width: 220px;
  }

  .table-scroll-container {
    overflow-x: auto;
  }

  .analytics-data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13.5px;
  }

  .analytics-data-table th {
    padding: 14px 24px;
    background: var(--surface-2);
    color: var(--text-dim);
    font-weight: 800;
    font-size: 11.5px;
    letter-spacing: 0.5px;
    text-align: inherit;
    border-bottom: 1px solid var(--color-border);
  }

  .analytics-data-table td {
    padding: 16px 24px;
    border-bottom: 1px solid var(--color-border);
    color: var(--text-main);
  }

  .name-cell strong {
    font-size: 14.5px;
    color: var(--text-main);
  }

  .details-cell {
    color: var(--text-muted);
    max-width: 300px;
  }

  .lang-tag {
    padding: 3px 8px;
    background: rgba(60, 16, 83, 0.4);
    border: 1px solid rgba(255, 205, 0, 0.3);
    border-radius: var(--radius-sm);
    color: var(--color-accent);
    font-size: 11px;
    font-weight: 700;
  }

  .lang-tag.en {
    background: rgba(71, 133, 159, 0.2);
    border-color: rgba(71, 133, 159, 0.4);
    color: #7EE8BE;
  }

  .date-cell,
  .time-cell {
    color: var(--text-dim);
    font-family: monospace;
    font-size: 12.5px;
  }

  .empty-cards-wrap,
  .cards-loading-wrap {
    padding: 48px;
    text-align: center;
    color: var(--text-muted);
  }
</style>
