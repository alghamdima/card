<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { t, locale, formatTime, translateError } from '$lib/i18n';
  import { campaignsApi } from '$lib/api/campaigns';
  import type { CampaignAnalytics, Card, DashboardStats as Stats } from '$lib/types/campaign.types';
  import Pager from '$lib/components/ui/Pager.svelte';
  import DashboardStats from '$lib/components/admin/DashboardStats.svelte';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';
  import { showToast } from '$lib/components/ui/toast.store';

  const PAGE_SIZE = 50;
  const SEARCH_DEBOUNCE_MS = 300;

  let analyticsCampaigns = $state<CampaignAnalytics[]>([]);
  let stats = $state<Stats | null>(null);
  let activeTabSlug = $state('');
  let loading = $state(true);
  let errorMsg = $state('');
  let lastUpdated = $state<Date | null>(null);

  // Cards of the selected campaign: one server-side page at a time, filtered by the search box.
  let campaignCards = $state<Card[]>([]);
  let cardsTotal = $state(0);
  let offset = $state(0);
  let loadingCards = $state(false);
  let searchQuery = $state('');
  let exporting = $state(false);

  let activeCampaign = $derived(analyticsCampaigns.find((c) => c.slug === activeTabSlug) ?? analyticsCampaigns[0]);

  // Ignores the response of a superseded request (fast tab switching / typing).
  let cardsRequest = 0;
  let searchTimer: ReturnType<typeof setTimeout> | undefined;

  async function loadData() {
    loading = true;
    errorMsg = '';
    try {
      // The global counters are a nice-to-have: their failure must not hide the per-campaign analytics.
      const [overview, dashboard] = await Promise.all([
        campaignsApi.getAnalyticsOverview(),
        campaignsApi.getDashboardStats().catch(() => null)
      ]);
      analyticsCampaigns = overview.campaigns || [];
      stats = dashboard;
      lastUpdated = new Date();

      if (!analyticsCampaigns.some((c) => c.slug === activeTabSlug)) {
        activeTabSlug = analyticsCampaigns[0]?.slug ?? '';
      }
      if (activeTabSlug) await loadCampaignCards(0);
    } catch (e) {
      errorMsg = translateError(e);
    } finally {
      loading = false;
    }
  }

  async function loadCampaignCards(newOffset: number) {
    const slug = activeTabSlug;
    if (!slug) return;

    const request = ++cardsRequest;
    loadingCards = true;
    try {
      const res = await campaignsApi.getCampaignCards(slug, {
        limit: PAGE_SIZE,
        offset: newOffset,
        q: searchQuery.trim()
      });
      if (request !== cardsRequest) return;
      campaignCards = res.cards || [];
      cardsTotal = res.total;
      offset = newOffset;
    } catch (e) {
      if (request !== cardsRequest) return;
      showToast(translateError(e), 'error');
    } finally {
      if (request === cardsRequest) loadingCards = false;
    }
  }

  async function handleTabChange(slug: string) {
    if (slug === activeTabSlug) return;
    activeTabSlug = slug;
    campaignCards = [];
    searchQuery = '';
    clearTimeout(searchTimer);
    await loadCampaignCards(0);
  }

  function handleSearchInput() {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => loadCampaignCards(0), SEARCH_DEBOUNCE_MS);
  }

  async function handleExportCsv() {
    if (!activeTabSlug || exporting) return;
    exporting = true;
    try {
      await campaignsApi.downloadExport(activeTabSlug);
    } catch (e) {
      showToast(translateError(e), 'error');
    } finally {
      exporting = false;
    }
  }

  onMount(loadData);
  onDestroy(() => clearTimeout(searchTimer));
</script>

<svelte:head>
  <title>{$t('admin.analyticsTitle')} | {$t('app.title')}</title>
</svelte:head>

<div class="analytics-page">
  <header class="analytics-header">
    <div class="header-left">
      <div class="title-row">
        <span class="bar-icon" aria-hidden="true">📊</span>
        <h2>{$t('admin.analyticsTitle')}</h2>
      </div>
      {#if lastUpdated}
        <span class="updated-time">{$t('admin.updatedAt', { time: formatTime(lastUpdated, $locale) })}</span>
      {/if}
    </div>

    <div class="header-actions">
      <button type="button" class="btn-round" onclick={loadData} disabled={loading} title={$t('admin.refresh')}>
        <span class="btn-icon" aria-hidden="true">🔄</span>
        <span>{$t('admin.refresh')}</span>
      </button>

      <button
        type="button"
        class="btn-round"
        onclick={handleExportCsv}
        disabled={exporting || !activeTabSlug}
        title={$t('admin.exportCsv')}
      >
        <span class="btn-icon" aria-hidden="true">📥</span>
        <span>{$t('admin.exportCsv')}</span>
      </button>
    </div>
  </header>

  {#if loading}
    <LoadingState />
  {:else if errorMsg}
    <ErrorState message={errorMsg} onretry={loadData} />
  {:else}
    {#if stats}
      <DashboardStats {stats} />
    {/if}

    <!-- Horizontal campaign tabs -->
    <div class="campaign-tabs-nav" role="tablist">
      {#each analyticsCampaigns as camp (camp.slug)}
        <button
          type="button"
          role="tab"
          aria-selected={activeTabSlug === camp.slug}
          class="nav-tab-item"
          class:active={activeTabSlug === camp.slug}
          onclick={() => handleTabChange(camp.slug)}
        >
          {camp.title}
        </button>
      {/each}
    </div>

    <div class="stats-section">
      <div class="section-badge-title">
        <span class="icon" aria-hidden="true">📊</span>
        <span>{$t('admin.allTimeStats')}</span>
      </div>

      <div class="stat-card-wide">
        <div class="stat-top-meta">
          <span class="stat-label">{$t('admin.totalCardsGenerated')}</span>
          <span class="card-icon-pill" aria-hidden="true">🎴</span>
        </div>
        <div class="stat-big-number">
          {activeCampaign?.totalCards ?? 0}
        </div>
      </div>
    </div>

    <div class="employee-list-section">
      <div class="section-badge-title">
        <span class="icon" aria-hidden="true">📋</span>
        <span>{$t('admin.employeeList')}</span>
      </div>

      <div class="employee-table-card">
        <div class="table-card-header">
          <div class="header-counter-box">
            <h4>{$t('admin.employeeCards')}</h4>
            <span class="total-badge">{$t('admin.totalCount', { count: cardsTotal })}</span>
          </div>

          <div class="search-box">
            <span class="search-icon" aria-hidden="true">🔍</span>
            <input
              type="search"
              placeholder={$t('admin.searchPlaceholder')}
              aria-label={$t('admin.searchPlaceholder')}
              bind:value={searchQuery}
              oninput={handleSearchInput}
              class="search-input"
            />
          </div>
        </div>

        {#if loadingCards && campaignCards.length === 0}
          <div class="cards-loading-wrap">
            <LoadingState />
          </div>
        {:else if campaignCards.length === 0}
          <div class="empty-cards-wrap">
            <p>{$t('admin.noCardsForCampaign')}</p>
          </div>
        {:else}
          <div class="table-scroll-container" class:refreshing={loadingCards}>
            <table class="analytics-data-table">
              <thead>
                <tr>
                  <th>{$t('admin.fullName')}</th>
                  <th>{$t('admin.jobTitleDetails')}</th>
                  <th>{$t('admin.language')}</th>
                  <th>{$t('admin.date')}</th>
                  <th>{$t('admin.time')}</th>
                </tr>
              </thead>
              <tbody>
                {#each campaignCards as card (card.id)}
                  <tr>
                    <td class="name-cell">
                      <strong>{card.to || '-'}</strong>
                    </td>
                    <td class="details-cell">
                      {card.message || card.fieldValues?.job_title || '-'}
                    </td>
                    <td>
                      <span class="lang-tag" class:en={card.lang === 'en'}>
                        {card.lang === 'en' ? 'English' : 'العربية'}
                      </span>
                    </td>
                    <td class="date-cell">{card.date}</td>
                    <td class="time-cell">{card.time}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>

          <Pager {offset} total={cardsTotal} pageSize={PAGE_SIZE} disabled={loadingCards} onchange={loadCampaignCards} />
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
    color: var(--text-main);
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

  .table-scroll-container.refreshing {
    opacity: 0.55;
    transition: opacity 0.15s ease;
  }

  .btn-round:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
</style>
