<script lang="ts">
  import { onMount } from 'svelte';
  import { t } from '$lib/i18n';
  import { campaignsApi } from '$lib/api/campaigns';
  import type { DashboardStats as StatsType, CampaignSummary } from '$lib/types/campaign.types';
  import DashboardStats from '$lib/components/admin/DashboardStats.svelte';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';

  let stats = $state<StatsType | null>(null);
  let campaigns = $state<CampaignSummary[]>([]);
  let loading = $state(true);
  let errorMsg = $state('');

  async function loadData() {
    loading = true;
    errorMsg = '';
    try {
      const [s, c] = await Promise.all([
        campaignsApi.getDashboardStats(),
        campaignsApi.listAdmin()
      ]);
      stats = s;
      campaigns = c.campaigns || [];
    } catch (e: any) {
      errorMsg = e.message || 'Failed to load dashboard data';
    } finally {
      loading = false;
    }
  }

  onMount(loadData);
</script>

<svelte:head>
  <title>{$t('nav.dashboard')} | {$t('app.title')}</title>
</svelte:head>

<div class="dashboard-page">
  <div class="dash-head">
    <div>
      <h2>{$t('admin.title')}</h2>
      <p>{$t('admin.subtitle')}</p>
    </div>
    <a href="/admin/cards" class="action-btn">
      <span>➕</span>
      <span>{$t('admin.newCampaign')}</span>
    </a>
  </div>

  {#if loading}
    <LoadingState />
  {:else if errorMsg}
    <ErrorState message={errorMsg} onretry={loadData} />
  {:else if stats}
    <DashboardStats {stats} />

    <div class="section-box">
      <div class="box-head">
        <h3>{$t('nav.campaigns')}</h3>
        <a href="/admin/cards" class="view-all">{$t('app.actions')} →</a>
      </div>

      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{$t('admin.campaignName')}</th>
              <th>{$t('admin.campaignSlug')}</th>
              <th>{$t('admin.cardsCount')}</th>
              <th>{$t('app.status')}</th>
              <th>{$t('admin.createdDate')}</th>
            </tr>
          </thead>
          <tbody>
            {#each campaigns.slice(0, 5) as camp (camp.slug)}
              <tr>
                <td class="name-col">
                  <strong>{camp.title}</strong>
                </td>
                <td>
                  <a href="/cards/{camp.slug}" target="_blank" class="slug-badge">
                    /cards/{camp.slug}
                  </a>
                </td>
                <td>{camp.totalCards}</td>
                <td>
                  <span class="status-badge" class:active={camp.active}>
                    {camp.active ? $t('app.active') : $t('app.inactive')}
                  </span>
                </td>
                <td class="date-col">{new Date(camp.createdAt).toLocaleDateString()}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</div>

<style>
  .dashboard-page {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .dash-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 16px;
  }

  .dash-head h2 {
    font-size: 24px;
    font-weight: 800;
    color: var(--text-main);
  }

  .dash-head p {
    font-size: 14px;
    color: var(--text-muted);
    margin-top: 4px;
  }

  .action-btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    background: var(--color-primary);
    color: #FFFFFF;
    padding: 10px 18px;
    border-radius: var(--radius-md);
    font-size: 14px;
    font-weight: 700;
    transition: background 0.15s;
  }

  .action-btn:hover {
    background: var(--color-primary-light);
  }

  .section-box {
    background: var(--surface-card);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    overflow: hidden;
  }

  .box-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid var(--color-border);
  }

  .box-head h3 {
    font-size: 16px;
    font-weight: 700;
    color: var(--text-main);
  }

  .view-all {
    font-size: 13px;
    font-weight: 600;
    color: var(--color-accent);
  }

  .table-wrap {
    overflow-x: auto;
  }

  .data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 14px;
  }

  th, td {
    padding: 12px 18px;
    text-align: start;
    border-bottom: 1px solid var(--color-border);
  }

  th {
    background: var(--surface-2);
    color: var(--text-muted);
    font-weight: 700;
    font-size: 12.5px;
  }

  .slug-badge {
    direction: ltr;
    display: inline-block;
    padding: 3px 8px;
    background: var(--surface-2);
    border-radius: var(--radius-sm);
    color: var(--color-accent);
    font-size: 12.5px;
    font-family: monospace;
  }

  .status-badge {
    padding: 3px 8px;
    border-radius: var(--radius-full);
    font-size: 11px;
    font-weight: 700;
    background: var(--color-danger-wash);
    color: var(--color-danger);
  }

  .status-badge.active {
    background: var(--color-success-wash);
    color: var(--color-success);
  }

  .date-col {
    color: var(--text-dim);
    font-size: 13px;
  }
</style>
