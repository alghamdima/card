<script lang="ts">
  import { t } from '../../i18n';
  import type { DashboardStats } from '../../types/campaign.types';

  interface Props {
    stats: DashboardStats;
  }

  let { stats }: Props = $props();
</script>

<div class="stats-grid">
  <div class="stat-card">
    <div class="stat-head">
      <span class="label">{$t('admin.totalCampaigns')}</span>
      <span class="icon">🎯</span>
    </div>
    <span class="value">{stats.totalCampaigns}</span>
  </div>

  <div class="stat-card">
    <div class="stat-head">
      <span class="label">{$t('admin.activeCampaigns')}</span>
      <span class="icon">🟢</span>
    </div>
    <span class="value highlight">{stats.activeCampaigns}</span>
  </div>

  <div class="stat-card">
    <div class="stat-head">
      <span class="label">{$t('admin.totalCards')}</span>
      <span class="icon">💌</span>
    </div>
    <span class="value">{stats.totalCards}</span>
  </div>

  <div class="stat-card">
    <div class="stat-head">
      <span class="label">{$t('admin.cardsToday')}</span>
      <span class="icon">⚡</span>
    </div>
    <span class="value">{stats.cardsToday}</span>
  </div>
</div>

<style>
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 16px;
    width: 100%;
    margin-bottom: 24px;
  }

  .stat-card {
    background: var(--surface-card);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .stat-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .label {
    font-size: 13px;
    font-weight: 700;
    color: var(--text-muted);
  }

  .value {
    font-size: 28px;
    font-weight: 900;
    color: var(--text-main);
  }

  .value.highlight {
    color: var(--color-accent);
  }
</style>
