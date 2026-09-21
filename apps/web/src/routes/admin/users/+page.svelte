<script lang="ts">
  import { onMount } from 'svelte';
  import { isAnonymousSender, getCardData } from '$lib/utils/cards';
  import { t, translateError } from '$lib/i18n';
  import { campaignsApi } from '$lib/api/campaigns';
  import type { Card } from '$lib/types/campaign.types';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';
  import Pager from '$lib/components/ui/Pager.svelte';

  const PAGE_SIZE = 50;

  let cards = $state<Card[]>([]);
  let total = $state(0);
  let offset = $state(0);
  let loading = $state(true);
  let errorMsg = $state('');

  // Ignores the response of a superseded request (fast paging).
  let request = 0;

  async function loadCards(newOffset = 0) {
    const current = ++request;
    loading = true;
    errorMsg = '';
    try {
      const res = await campaignsApi.getAllCards({ limit: PAGE_SIZE, offset: newOffset });
      if (current !== request) return;
      cards = res.cards || [];
      total = res.total || 0;
      offset = newOffset;
    } catch (e) {
      if (current !== request) return;
      errorMsg = translateError(e);
    } finally {
      if (current === request) loading = false;
    }
  }


  onMount(() => loadCards());
</script>

<svelte:head>
  <title>{$t('nav.cards')} | {$t('app.title')}</title>
</svelte:head>

<div class="users-page">
  <div class="page-head">
    <div>
      <h2>{$t('nav.cards')}</h2>
      <p>{$t('admin.totalCards')}: {total}</p>
    </div>
  </div>

  {#if loading && cards.length === 0}
    <LoadingState />
  {:else if errorMsg}
    <ErrorState message={errorMsg} onretry={() => loadCards(offset)} />
  {:else if cards.length === 0}
    <div class="empty-box">
      <p>{$t('admin.noCardsYet')}</p>
    </div>
  {:else}
    <div class="table-card" class:refreshing={loading}>
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>#</th>
              <th>{$t('nav.campaigns')}</th>
              <th>{$t('admin.receiver')}</th>
              <th>{$t('card.message')}</th>
              <th>{$t('admin.sender')}</th>
              <th>{$t('admin.dateAndTime')}</th>
            </tr>
          </thead>
          <tbody>
            {#each cards as card, idx (card.id)}
              {@const cardData = getCardData(card)}
              <tr>
                <td class="num-col">{total - offset - idx}</td>
                <td>
                  <span class="slug-tag">{card.campaignSlug}</span>
                </td>
                <td><strong>{cardData.to || '-'}</strong></td>
                <td class="msg-col">{cardData.message || '-'}</td>
                <td>{cardData.from ? cardData.from : $t('app.anonymous')}</td>
                <td class="date-col">{card.date} {card.time}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <Pager {offset} {total} pageSize={PAGE_SIZE} disabled={loading} onchange={loadCards} />
    </div>
  {/if}
</div>

<style>
  .users-page {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .page-head h2 {
    font-size: 24px;
    font-weight: 800;
    color: var(--text-main);
  }

  .page-head p {
    font-size: 14px;
    color: var(--text-muted);
    margin-top: 4px;
  }

  .table-card {
    background: var(--surface-card);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    overflow: hidden;
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

  .num-col {
    color: var(--text-dim);
    font-size: 12px;
  }

  .slug-tag {
    padding: 2px 6px;
    background: var(--surface-2);
    border-radius: var(--radius-sm);
    color: var(--color-accent);
    font-family: monospace;
    font-size: 12px;
  }

  .msg-col {
    max-width: 300px;
    font-size: 13.5px;
    color: var(--text-muted);
  }

  .date-col {
    color: var(--text-dim);
    font-size: 13px;
    white-space: nowrap;
  }

  .empty-box {
    padding: 48px;
    text-align: center;
    background: var(--surface-card);
    border-radius: var(--radius-xl);
    border: 1px solid var(--color-border);
    color: var(--text-muted);
  }

  .table-card.refreshing {
    opacity: 0.6;
    transition: opacity 0.15s ease;
  }
</style>
