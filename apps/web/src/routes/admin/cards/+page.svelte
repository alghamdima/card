<script lang="ts">
  import { onMount } from 'svelte';
  import { t } from '$lib/i18n';
  import { campaignsApi } from '$lib/api/campaigns';
  import type { CampaignSummary, Card } from '$lib/types/campaign.types';
  import Button from '$lib/components/ui/Button.svelte';
  import Input from '$lib/components/ui/Input.svelte';
  import Modal from '$lib/components/ui/Modal.svelte';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';
  import { showToast } from '$lib/components/ui/toast.store';

  let campaigns = $state<CampaignSummary[]>([]);
  let loading = $state(true);
  let errorMsg = $state('');

  // Create Modal state
  let showCreateModal = $state(false);
  let newTitle = $state('');
  let newTextColor = $state('#FFFFFF');
  let newHeadColor = $state('#FFCD00');
  let newImageBase64 = $state('');
  let isSubmitting = $state(false);

  // Stats Modal state
  let showStatsModal = $state(false);
  let activeStatsCampaign = $state<string | null>(null);
  let campaignCards = $state<Card[]>([]);
  let loadingCards = $state(false);

  async function loadCampaigns() {
    loading = true;
    errorMsg = '';
    try {
      const res = await campaignsApi.listAdmin();
      campaigns = res.campaigns || [];
    } catch (e: any) {
      errorMsg = e.message || 'Failed to load campaigns';
    } finally {
      loading = false;
    }
  }

  onMount(loadCampaigns);

  function handleFileUpload(file: File) {
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (e) => {
      newImageBase64 = (e.target?.result as string) || '';
    };
    reader.readAsDataURL(file);
  }

  async function handleCreateCampaign() {
    if (!newTitle.trim()) {
      showToast($t('admin.campaignName') + ' ' + $t('errors.VALIDATION_ERROR'), 'error');
      return;
    }
    if (!newImageBase64) {
      showToast($t('admin.uploadArtwork') + ' ' + $t('errors.VALIDATION_ERROR'), 'error');
      return;
    }

    isSubmitting = true;
    try {
      // Notice: slug is intentionally left empty so backend generates random URL-safe 8-char slug!
      const created = await campaignsApi.create({
        title: newTitle.trim(),
        textColor: newTextColor,
        headColor: newHeadColor,
        image: newImageBase64,
        thumb: newImageBase64
      });

      showToast(`${$t('app.save')} (/cards/${created.slug})`, 'success');
      showCreateModal = false;
      newTitle = '';
      newImageBase64 = '';
      await loadCampaigns();
    } catch (e: any) {
      showToast(e.message || 'Error creating campaign', 'error');
    } finally {
      isSubmitting = false;
    }
  }

  async function handleDelete(slug: string) {
    if (!confirm($t('admin.confirmDelete'))) return;

    try {
      await campaignsApi.delete(slug);
      showToast($t('app.delete'), 'info');
      await loadCampaigns();
    } catch (e: any) {
      showToast(e.message || 'Failed to delete campaign', 'error');
    }
  }

  async function openStats(slug: string) {
    activeStatsCampaign = slug;
    showStatsModal = true;
    loadingCards = true;
    try {
      const res = await campaignsApi.getCampaignCards(slug);
      campaignCards = res.cards || [];
    } catch (e: any) {
      showToast(e.message || 'Failed to load cards', 'error');
    } finally {
      loadingCards = false;
    }
  }

  function copyLink(slug: string) {
    const fullUrl = `${window.location.origin}/cards/${slug}`;
    navigator.clipboard.writeText(fullUrl).then(() => {
      showToast($t('app.copied'), 'success');
    });
  }
</script>

<svelte:head>
  <title>{$t('nav.campaigns')} | {$t('app.title')}</title>
</svelte:head>

<div class="campaigns-page">
  <div class="page-head">
    <div>
      <h2>{$t('nav.campaigns')}</h2>
      <p>{$t('admin.subtitle')}</p>
    </div>

    <Button variant="primary" onclick={() => (showCreateModal = true)}>
      <span>➕</span>
      <span>{$t('admin.newCampaign')}</span>
    </Button>
  </div>

  {#if loading}
    <LoadingState />
  {:else if errorMsg}
    <ErrorState message={errorMsg} onretry={loadCampaigns} />
  {:else}
    <div class="table-card">
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>{$t('admin.campaignName')}</th>
              <th>{$t('admin.campaignSlug')}</th>
              <th>{$t('admin.cardsCount')}</th>
              <th>{$t('app.status')}</th>
              <th>{$t('admin.createdDate')}</th>
              <th>{$t('app.actions')}</th>
            </tr>
          </thead>
          <tbody>
            {#each campaigns as camp (camp.slug)}
              <tr>
                <td class="camp-title-cell">
                  <strong>{camp.title}</strong>
                </td>
                <td>
                  <div class="link-actions">
                    <a href="/cards/{camp.slug}" target="_blank" class="slug-badge">
                      /cards/{camp.slug}
                    </a>
                    <button class="copy-btn" onclick={() => copyLink(camp.slug)} title={$t('app.copyLink')}>
                      📋
                    </button>
                  </div>
                </td>
                <td>{camp.totalCards}</td>
                <td>
                  <span class="status-badge" class:active={camp.active}>
                    {camp.active ? $t('app.active') : $t('app.inactive')}
                  </span>
                </td>
                <td>{new Date(camp.createdAt).toLocaleDateString()}</td>
                <td>
                  <div class="row-actions">
                    <button class="btn-text" onclick={() => openStats(camp.slug)}>
                      📊 {$t('admin.viewCards')}
                    </button>
                    <button class="btn-text danger" onclick={() => handleDelete(camp.slug)}>
                      🗑️ {$t('app.delete')}
                    </button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</div>

<!-- Create Campaign Modal -->
<Modal
  open={showCreateModal}
  title={$t('admin.newCampaign')}
  onclose={() => (showCreateModal = false)}
>
  <div class="modal-form">
    <Input
      label={$t('admin.campaignName')}
      placeholder="مثال: عيد الفطر 2026"
      value={newTitle}
      oninput={(e) => (newTitle = (e.target as HTMLInputElement).value)}
    />

    <div class="note-box">
      <span>🎲</span>
      <p>{$t('admin.autoGeneratedNote')}</p>
    </div>

    <div class="color-row">
      <div class="color-picker">
        <label for="colorText">{$t('admin.textColor')}</label>
        <div class="picker-inner">
          <input id="colorText" type="color" bind:value={newTextColor} />
          <span>{newTextColor}</span>
        </div>
      </div>

      <div class="color-picker">
        <label for="colorHead">{$t('admin.headColor')}</label>
        <div class="picker-inner">
          <input id="colorHead" type="color" bind:value={newHeadColor} />
          <span>{newHeadColor}</span>
        </div>
      </div>
    </div>

    <div class="upload-section">
      <label class="upload-label" for="fileInputUpload">{$t('admin.uploadArtwork')}</label>
      <label class="drop-zone" class:has-file={!!newImageBase64}>
        <span class="drop-icon">🖼️</span>
        <p>{newImageBase64 ? $t('admin.imageSelected') : $t('admin.dropImage')}</p>
        <input
          id="fileInputUpload"
          type="file"
          accept="image/*"
          class="hidden-input"
          onchange={(e) => {
            const files = (e.target as HTMLInputElement).files;
            if (files && files[0]) handleFileUpload(files[0]);
          }}
        />
      </label>

      {#if newImageBase64}
        <div class="preview-thumb-box">
          <img src={newImageBase64} alt="Preview" class="preview-thumb" />
        </div>
      {/if}
    </div>

    <div class="modal-actions">
      <Button
        variant="primary"
        loading={isSubmitting}
        disabled={isSubmitting}
        onclick={handleCreateCampaign}
      >
        {$t('admin.saveAndPublish')}
      </Button>
    </div>
  </div>
</Modal>

<!-- View Cards Stats Modal -->
<Modal
  open={showStatsModal}
  title={`${$t('admin.stats')} (${activeStatsCampaign})`}
  onclose={() => (showStatsModal = false)}
>
  {#if loadingCards}
    <LoadingState />
  {:else if campaignCards.length === 0}
    <p class="empty-text">{$t('admin.noCardsYet')}</p>
  {:else}
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>{$t('admin.receiver')}</th>
            <th>{$t('card.message')}</th>
            <th>{$t('admin.sender')}</th>
            <th>{$t('admin.device')}</th>
            <th>{$t('admin.dateAndTime')}</th>
          </tr>
        </thead>
        <tbody>
          {#each campaignCards as card (card.id)}
            <tr>
              <td><strong>{card.to || '-'}</strong></td>
              <td class="msg-cell">{card.message || '-'}</td>
              <td>{card.from || '-'}</td>
              <td><span class="device-tag">{card.device}</span></td>
              <td>{card.date} {card.time}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</Modal>

<style>
  .campaigns-page {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .page-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 16px;
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

  .link-actions {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .copy-btn {
    background: transparent;
    border: none;
    cursor: pointer;
    font-size: 14px;
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

  .row-actions {
    display: flex;
    gap: 10px;
  }

  .btn-text {
    background: transparent;
    border: none;
    color: var(--text-muted);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
  }

  .btn-text:hover {
    color: var(--color-accent);
  }

  .btn-text.danger:hover {
    color: var(--color-danger);
  }

  /* Modal Form */
  .modal-form {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .note-box {
    display: flex;
    gap: 8px;
    align-items: center;
    padding: 10px 14px;
    background: var(--surface-2);
    border-radius: var(--radius-md);
    font-size: 13px;
    color: var(--color-accent);
  }

  .color-row {
    display: flex;
    gap: 16px;
  }

  .color-picker {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .color-picker label {
    font-size: 13px;
    font-weight: 700;
    color: var(--text-muted);
  }

  .picker-inner {
    display: flex;
    align-items: center;
    gap: 8px;
    background: var(--surface-1);
    padding: 6px 12px;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
  }

  .picker-inner input[type="color"] {
    width: 32px;
    height: 32px;
    border: none;
    border-radius: var(--radius-sm);
    cursor: pointer;
    background: none;
  }

  .upload-section {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .upload-label {
    font-size: 13px;
    font-weight: 700;
    color: var(--text-muted);
  }

  .drop-zone {
    border: 2px dashed var(--color-border);
    border-radius: var(--radius-lg);
    padding: 24px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    cursor: pointer;
    background: var(--surface-1);
    text-align: center;
  }

  .drop-zone.has-file {
    border-color: var(--color-accent);
    background: var(--surface-2);
  }

  .drop-icon {
    font-size: 32px;
  }

  .drop-zone p {
    font-size: 13.5px;
    color: var(--text-muted);
  }

  .hidden-input {
    display: none;
  }

  .preview-thumb-box {
    margin-top: 8px;
    border-radius: var(--radius-md);
    overflow: hidden;
    max-height: 140px;
    border: 1px solid var(--color-border);
  }

  .preview-thumb {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .modal-actions {
    margin-top: 8px;
  }

  .modal-actions :global(button) {
    width: 100%;
  }

  .msg-cell {
    max-width: 250px;
    font-size: 13px;
    color: var(--text-muted);
  }

  .device-tag {
    font-size: 11px;
    padding: 2px 6px;
    background: var(--surface-2);
    border-radius: var(--radius-sm);
    color: var(--text-dim);
  }

  .empty-text {
    padding: 32px;
    text-align: center;
    color: var(--text-muted);
  }
</style>
