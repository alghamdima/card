<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { isAnonymousSender, getCardData } from '$lib/utils/cards';
  import { t, locale, formatDate, translateError } from '$lib/i18n';
  import { campaignsApi } from '$lib/api/campaigns';
  import type { Campaign, CampaignSummary, Card, TextFieldConfig } from '$lib/types/campaign.types';
  import Button from '$lib/components/ui/Button.svelte';
  import Input from '$lib/components/ui/Input.svelte';
  import Modal from '$lib/components/ui/Modal.svelte';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';
  import EmptyState from '$lib/components/ui/EmptyState.svelte';
  import { showToast } from '$lib/components/ui/toast.store';
  import TemplateCanvasEditor from '$lib/components/admin/TemplateCanvasEditor.svelte';
  import { copyText } from '$lib/utils/clipboard';
  import { isSupportedImage, prepareArtwork } from '$lib/utils/image';
  import { defaultFields, defaultSamples, newField } from '$lib/utils/template-defaults';

  type Lang = 'ar' | 'en';

  let campaigns = $state<CampaignSummary[]>([]);
  let loading = $state(true);
  let errorMsg = $state('');

  // Create Campaign Modal
  let showCreateModal = $state(false);
  let newTitleAR = $state('');
  let newTitleEN = $state('');
  let newImage = $state('');
  let newThumb = $state('');
  let isProcessingImage = $state(false);
  let isSubmitting = $state(false);

  // Template & Position Builder Modal
  let showBuilderModal = $state(false);
  let editingCampaign = $state<Campaign | null>(null);
  let editingTitleAR = $state('');
  let editingTitleEN = $state('');
  let activeLangTab = $state<Lang>('ar');
  let activeFieldId = $state('emp_name');

  // Per-language template state. An image is only sent back to the server when the admin replaced it.
  let images = $state<Record<Lang, string>>({ ar: '', en: '' });
  let imageDirty = $state<Record<Lang, boolean>>({ ar: false, en: false });
  let fields = $state<Record<Lang, TextFieldConfig[]>>({ ar: [], en: [] });
  let samples = $state<Record<Lang, Record<string, string>>>({ ar: defaultSamples('ar'), en: defaultSamples('en') });
  let originalTemplateImages: Record<Lang, string> = { ar: '', en: '' };
  let newArThumb = '';

  let activeFields = $derived(fields[activeLangTab]);

  // Stats Modal
  let showStatsModal = $state(false);
  let activeStatsCampaign = $state<string | null>(null);
  let campaignCards = $state<Card[]>([]);
  let statsTotal = $state(0);
  let loadingCards = $state(false);

  async function loadCampaigns() {
    loading = true;
    errorMsg = '';
    try {
      const res = await campaignsApi.listAdmin();
      campaigns = res.campaigns || [];
    } catch (e) {
      errorMsg = translateError(e);
    } finally {
      loading = false;
    }
  }

  onMount(loadCampaigns);

  /** Validate and optimize an uploaded artwork file; returns null (after a toast) when it cannot be used. */
  async function processUpload(file: File | undefined) {
    if (!file) return null;
    if (!isSupportedImage(file)) {
      showToast($t('admin.imageInvalid'), 'error');
      return null;
    }
    isProcessingImage = true;
    try {
      return await prepareArtwork(file);
    } catch {
      showToast($t('admin.imageInvalid'), 'error');
      return null;
    } finally {
      isProcessingImage = false;
    }
  }

  async function handleCreateFileUpload(file: File | undefined) {
    const prepared = await processUpload(file);
    if (prepared) {
      newImage = prepared.image;
      newThumb = prepared.thumb;
    }
  }

  function closeCreateModal() {
    showCreateModal = false;
    newTitleAR = '';
    newTitleEN = '';
    newImage = '';
    newThumb = '';
  }

  async function handleCreateCampaign() {
    if (!newTitleAR.trim() && !newTitleEN.trim()) {
      showToast(`${$t('admin.campaignName')}: ${$t('errors.VALIDATION_ERROR')}`, 'error');
      return;
    }
    if (!newImage) {
      showToast(`${$t('admin.uploadArtwork')}: ${$t('errors.VALIDATION_ERROR')}`, 'error');
      return;
    }

    isSubmitting = true;
    try {
      const mainTitle = newTitleAR.trim() || newTitleEN.trim();

      const created = await campaignsApi.create({
        title: mainTitle,
        titleAR: newTitleAR.trim() || mainTitle,
        titleEN: newTitleEN.trim() || mainTitle,
        image: newImage,
        thumb: newThumb,
        templateAR: { image: newImage, fields: defaultFields('ar') },
        templateEN: { image: newImage, fields: defaultFields('en') }
      });

      showToast($t('admin.created', { link: `/cards/${created.slug}` }), 'success');
      closeCreateModal();
      await loadCampaigns();
    } catch (e) {
      showToast(translateError(e), 'error');
    } finally {
      isSubmitting = false;
    }
  }

  async function openBuilder(slug: string) {
    try {
      const camp = await campaignsApi.getAdmin(slug);
      editingCampaign = camp;
      editingTitleAR = camp.titleAR || camp.title || '';
      editingTitleEN = camp.titleEN || camp.title || '';
      activeLangTab = 'ar';

      // The server omits a variant image identical to the campaign image, so fall back to it for display.
      originalTemplateImages = { ar: camp.templateAR?.image ?? '', en: camp.templateEN?.image ?? '' };
      images = {
        ar: camp.templateAR?.image || camp.image,
        en: camp.templateEN?.image || camp.image
      };
      imageDirty = { ar: false, en: false };
      newArThumb = '';

      const arFields = camp.templateAR?.fields?.length ? structuredClone(camp.templateAR.fields) : defaultFields('ar');
      fields = {
        ar: arFields,
        en: camp.templateEN?.fields?.length ? structuredClone(camp.templateEN.fields) : defaultFields('en')
      };
      samples = { ar: defaultSamples('ar'), en: defaultSamples('en') };
      activeFieldId = arFields[0]?.id || 'emp_name';
      showBuilderModal = true;
    } catch (e) {
      showToast(translateError(e), 'error');
    }
  }

  function switchTab(lang: Lang) {
    activeLangTab = lang;
    // Field ids differ between the two templates: keep the selection valid.
    activeFieldId = fields[lang][0]?.id ?? '';
  }

  async function handleImageTabUpload(file: File | undefined, lang: Lang) {
    const prepared = await processUpload(file);
    if (!prepared) return;
    images[lang] = prepared.image;
    imageDirty[lang] = true;
    if (lang === 'ar') newArThumb = prepared.thumb;
  }

  function addField(lang: Lang) {
    const field = newField(lang, fields[lang].length);
    fields[lang] = [...fields[lang], field];
    samples[lang][field.id] = $t('admin.newFieldSample');
    activeFieldId = field.id;
  }

  function deleteField(lang: Lang, id: string) {
    fields[lang] = fields[lang].filter((f) => f.id !== id);
    if (activeFieldId === id) activeFieldId = fields[lang][0]?.id ?? '';
  }

  async function handleSaveTemplate() {
    if (!editingCampaign) return;
    isSubmitting = true;
    try {
      const payload: Partial<Campaign> = {
        titleAR: editingTitleAR.trim() || editingCampaign.title,
        titleEN: editingTitleEN.trim() || editingCampaign.title,
        templateAR: {
          image: imageDirty.ar ? images.ar : originalTemplateImages.ar,
          fields: $state.snapshot(fields.ar)
        },
        templateEN: {
          image: imageDirty.en ? images.en : originalTemplateImages.en,
          fields: $state.snapshot(fields.en)
        }
      };
      // The Arabic artwork is also the listing thumbnail.
      if (imageDirty.ar && newArThumb) payload.thumb = newArThumb;

      await campaignsApi.update(editingCampaign.slug, payload);
      showToast($t('app.saved'), 'success');
      showBuilderModal = false;
      await loadCampaigns();
    } catch (e) {
      showToast(translateError(e), 'error');
    } finally {
      isSubmitting = false;
    }
  }

  async function handleDelete(slug: string) {
    if (!confirm($t('admin.confirmDelete'))) return;
    try {
      await campaignsApi.delete(slug);
      showToast($t('app.deleted'), 'info');
      await loadCampaigns();
    } catch (e) {
      showToast(translateError(e), 'error');
    }
  }

  async function openStats(slug: string) {
    activeStatsCampaign = slug;
    showStatsModal = true;
    loadingCards = true;
    campaignCards = [];
    statsTotal = 0;
    try {
      const res = await campaignsApi.getCampaignCards(slug, { limit: 100 });
      campaignCards = res.cards || [];
      statsTotal = res.total;
    } catch (e) {
      showToast(translateError(e), 'error');
    } finally {
      loadingCards = false;
    }
  }

  async function copyLink(slug: string) {
    const ok = await copyText(`${window.location.origin}${base}/cards/${slug}`);
    showToast(ok ? $t('app.copied') : $t('app.copyFailed'), ok ? 'success' : 'error');
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
      <span aria-hidden="true">➕</span>
      <span>{$t('admin.newCampaign')}</span>
    </Button>
  </div>

  {#if loading}
    <LoadingState />
  {:else if errorMsg}
    <ErrorState message={errorMsg} onretry={loadCampaigns} />
  {:else if campaigns.length === 0}
    <EmptyState icon="🎨" message={$t('admin.noCampaignsYet')} />
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
                  <strong>{$locale === 'en' ? camp.titleEN || camp.title : camp.titleAR || camp.title}</strong>
                </td>
                <td>
                  <div class="link-actions">
                    <a href="{base}/cards/{camp.slug}" target="_blank" rel="noopener" class="slug-badge">
                      /cards/{camp.slug}
                    </a>
                    <button
                      type="button"
                      class="copy-btn"
                      onclick={() => copyLink(camp.slug)}
                      title={$t('app.copyLink')}
                      aria-label={$t('app.copyLink')}
                    >
                      📋
                    </button>
                  </div>
                </td>
                <td><strong class="count-num">{camp.totalCards}</strong></td>
                <td>
                  <span class="status-badge" class:active={camp.active}>
                    {camp.active ? $t('app.active') : $t('app.inactive')}
                  </span>
                </td>
                <td>{formatDate(camp.createdAt, $locale)}</td>
                <td>
                  <div class="row-actions">
                    <button type="button" class="btn-builder" onclick={() => openBuilder(camp.slug)}>
                      🎨 {$t('admin.templateSettings')}
                    </button>
                    <button type="button" class="btn-text" onclick={() => openStats(camp.slug)}>
                      {$t('admin.viewCards')}
                    </button>
                    <button type="button" class="btn-text danger" onclick={() => handleDelete(camp.slug)}>
                      {$t('app.delete')}
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
<Modal open={showCreateModal} title={$t('admin.newCampaign')} onclose={closeCreateModal}>
  <div class="modal-form">
    <Input
      label={$t('admin.titleArLabel')}
      placeholder={$t('admin.titleArPlaceholder')}
      value={newTitleAR}
      maxlength={120}
      oninput={(e) => (newTitleAR = (e.target as HTMLInputElement).value)}
    />

    <Input
      label={$t('admin.titleEnLabel')}
      placeholder={$t('admin.titleEnPlaceholder')}
      value={newTitleEN}
      maxlength={120}
      oninput={(e) => (newTitleEN = (e.target as HTMLInputElement).value)}
    />

    <div class="upload-section">
      <label class="upload-label" for="fileInputUpload">{$t('admin.uploadArtwork')}</label>
      <label class="drop-zone" class:has-file={!!newImage}>
        <span class="drop-icon" aria-hidden="true">🖼️</span>
        <p>
          {#if isProcessingImage}
            {$t('admin.imageProcessing')}
          {:else if newImage}
            {$t('admin.imageSelected')}
          {:else}
            {$t('admin.dropImage')}
          {/if}
        </p>
        <input
          id="fileInputUpload"
          type="file"
          accept="image/png,image/jpeg,image/webp"
          class="hidden-input"
          onchange={(e) => handleCreateFileUpload((e.target as HTMLInputElement).files?.[0])}
        />
      </label>

      {#if newThumb}
        <div class="preview-thumb-box">
          <img src={newThumb} alt="" class="preview-thumb" />
        </div>
      {/if}
    </div>

    <div class="modal-actions">
      <Button
        variant="primary"
        loading={isSubmitting}
        disabled={isSubmitting || isProcessingImage}
        onclick={handleCreateCampaign}
      >
        {$t('admin.saveAndPublish')}
      </Button>
    </div>
  </div>
</Modal>

<!-- Visual Template & Position Builder Modal -->
<Modal
  open={showBuilderModal}
  title={`${$t('admin.templateSettings')} - ${editingCampaign?.titleAR || editingCampaign?.title || ''}`}
  maxWidth="1100px"
  onclose={() => (showBuilderModal = false)}
>
  <div class="builder-modal-content">
    <!-- Campaign Titles in AR & EN -->
    <div class="titles-bilingual-bar">
      <div class="title-input-item">
        <label for="edit_title_ar">{$t('admin.titleArLabel')}</label>
        <input
          id="edit_title_ar"
          type="text"
          bind:value={editingTitleAR}
          maxlength="120"
          class="mini-input"
          placeholder={$t('admin.titleArPlaceholder')}
        />
      </div>
      <div class="title-input-item">
        <label for="edit_title_en">{$t('admin.titleEnLabel')}</label>
        <input
          id="edit_title_en"
          type="text"
          bind:value={editingTitleEN}
          maxlength="120"
          class="mini-input"
          placeholder={$t('admin.titleEnPlaceholder')}
        />
      </div>
    </div>

    <!-- Language Switcher Tabs -->
    <div class="lang-builder-tabs" role="tablist">
      <button
        type="button"
        role="tab"
        aria-selected={activeLangTab === 'ar'}
        class="lang-tab"
        class:active={activeLangTab === 'ar'}
        onclick={() => switchTab('ar')}
      >
        🇸🇦 {$t('admin.arabicTemplate')}
      </button>
      <button
        type="button"
        role="tab"
        aria-selected={activeLangTab === 'en'}
        class="lang-tab"
        class:active={activeLangTab === 'en'}
        onclick={() => switchTab('en')}
      >
        🇬🇧 {$t('admin.englishTemplate')}
      </button>
    </div>

    <!-- Artwork Background Upload for Current Tab -->
    <div class="tab-artwork-bar">
      <span>🖼️ {activeLangTab === 'ar' ? $t('admin.templateImageAr') : $t('admin.templateImageEn')}:</span>
      <label class="btn-change-image">
        <span>{isProcessingImage ? $t('admin.imageProcessing') : $t('admin.changeImage')}</span>
        <input
          type="file"
          accept="image/png,image/jpeg,image/webp"
          class="hidden-input"
          onchange={(e) => handleImageTabUpload((e.target as HTMLInputElement).files?.[0], activeLangTab)}
        />
      </label>
    </div>

    <!-- Interactive Dual Canvas Component -->
    <TemplateCanvasEditor
      imageSrc={images[activeLangTab]}
      fields={activeFields}
      {activeFieldId}
      sampleValues={samples[activeLangTab]}
      onfieldchange={(f) => (fields[activeLangTab] = f)}
      onselectfield={(id) => (activeFieldId = id)}
    />

    <!-- Dynamic Fields Control Strip -->
    <div class="fields-control-strip">
      <div class="fields-header">
        <h5>{$t('admin.textFields', { count: activeFields.length })}</h5>
        <button type="button" class="btn-add-field" onclick={() => addField(activeLangTab)}>
          ➕ {$t('admin.addTextField')}
        </button>
      </div>

      <div class="fields-editor-list">
        {#each activeFields as field (field.id)}
          <div class="field-edit-card" class:selected={field.id === activeFieldId}>
            <div class="card-top-row">
              <span class="field-pill-tag">#{field.id}</span>
              <button
                type="button"
                class="btn-delete-field"
                onclick={() => deleteField(activeLangTab, field.id)}
                title={$t('admin.deleteField')}
                aria-label={$t('admin.deleteField')}
              >
                🗑️
              </button>
            </div>

            <div class="form-grid">
              <div class="input-item">
                <label for={`label_${field.id}`}>{$t('admin.fieldLabel')}</label>
                <input
                  id={`label_${field.id}`}
                  type="text"
                  bind:value={field.label}
                  maxlength="100"
                  class="mini-input"
                />
              </div>

              <div class="input-item">
                <label for={`sample_${field.id}`}>{$t('admin.sampleValue')}</label>
                <input
                  id={`sample_${field.id}`}
                  type="text"
                  bind:value={samples[activeLangTab][field.id]}
                  class="mini-input"
                />
              </div>

              <div class="input-item">
                <label for={`size_${field.id}`}>{$t('admin.fontSize')}</label>
                <input
                  id={`size_${field.id}`}
                  type="number"
                  bind:value={field.fontSize}
                  class="mini-input"
                  min="12"
                  max="100"
                />
              </div>

              <div class="input-item">
                <label for={`color_${field.id}`}>{$t('admin.fontColor')}</label>
                <div class="mini-color-wrap">
                  <input
                    id={`color_${field.id}`}
                    type="color"
                    bind:value={field.color}
                    class="mini-color-input"
                  />
                  <span>{field.color}</span>
                </div>
              </div>

              <div class="input-item">
                <label for={`align_${field.id}`}>{$t('admin.textAlign')}</label>
                <select id={`align_${field.id}`} bind:value={field.align} class="mini-select">
                  <option value="center">{$t('admin.alignCenter')}</option>
                  <option value="right">{$t('admin.alignRight')}</option>
                  <option value="left">{$t('admin.alignLeft')}</option>
                </select>
              </div>
            </div>
          </div>
        {/each}
      </div>
    </div>

    <!-- Modal Footer Actions -->
    <div class="builder-actions">
      <Button
        variant="primary"
        loading={isSubmitting}
        disabled={isSubmitting || isProcessingImage}
        onclick={handleSaveTemplate}
      >
        💾 {$t('admin.saveTemplateAndPosition')}
      </Button>
    </div>
  </div>
</Modal>

<!-- View Cards Stats Modal -->
<Modal
  open={showStatsModal}
  title={`${$t('admin.stats')} (${activeStatsCampaign ?? ''})`}
  onclose={() => (showStatsModal = false)}
>
  {#if loadingCards}
    <LoadingState />
  {:else if campaignCards.length === 0}
    <p class="empty-text">{$t('admin.noCardsYet')}</p>
  {:else}
    <p class="stats-note">{$t('admin.totalCount', { count: statsTotal })}</p>
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>{$t('admin.receiver')}</th>
            <th>{$t('card.message')}</th>
            <th>{$t('admin.sender')}</th>
            <th>{$t('admin.dateAndTime')}</th>
          </tr>
        </thead>
        <tbody>
          {#each campaignCards as card (card.id)}
            {@const cardData = getCardData(card)}
            <tr>
              <td><strong>{cardData.to || '-'}</strong></td>
              <td class="msg-cell">{cardData.message || '-'}</td>
              <td>{cardData.from ? cardData.from : $t('app.anonymous')}</td>
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

  .count-num {
    color: var(--color-accent);
    font-size: 16px;
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

  .data-table th {
    padding: 14px 18px;
    background: var(--surface-2);
    color: var(--text-muted);
    font-weight: 700;
    text-align: inherit;
    border-bottom: 1px solid var(--color-border);
  }

  .data-table td {
    padding: 14px 18px;
    border-bottom: 1px solid var(--color-border);
    color: var(--text-main);
  }

  .link-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .slug-badge {
    color: var(--color-accent);
    font-family: monospace;
    font-size: 13px;
    text-decoration: none;
  }

  .copy-btn {
    background: transparent;
    border: none;
    cursor: pointer;
    font-size: 14px;
  }

  .status-badge {
    padding: 4px 10px;
    border-radius: var(--radius-full);
    font-size: 12px;
    font-weight: 700;
    background: rgba(255, 138, 122, 0.15);
    color: var(--color-danger);
  }

  .status-badge.active {
    background: rgba(126, 232, 190, 0.15);
    color: var(--color-success);
  }

  .row-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .btn-builder {
    background: var(--surface-3);
    border: 1px solid var(--color-accent);
    color: var(--color-accent);
    font-size: 13px;
    font-weight: 700;
    padding: 6px 12px;
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .btn-builder:hover {
    background: var(--color-accent);
    color: #2B0A3D;
  }

  .btn-text {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 13.5px;
    cursor: pointer;
    transition: color 0.15s;
  }

  .btn-text:hover {
    color: var(--color-accent);
  }

  .btn-text.danger:hover {
    color: var(--color-danger);
  }

  /* Builder Modal Styles */
  .builder-modal-content {
    display: flex;
    flex-direction: column;
    gap: 20px;
    max-height: 80vh;
    overflow-y: auto;
    padding: 8px 4px;
  }

  .titles-bilingual-bar {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
    background: var(--surface-1);
    padding: 14px;
    border-radius: var(--radius-md);
    border: 1px solid var(--color-border);
  }

  .title-input-item {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .title-input-item label {
    font-size: 12.5px;
    font-weight: 700;
    color: var(--text-muted);
  }

  .lang-builder-tabs {
    display: flex;
    gap: 12px;
    border-bottom: 2px solid var(--color-border);
    padding-bottom: 8px;
  }

  .lang-tab {
    padding: 8px 18px;
    background: var(--surface-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    color: var(--text-muted);
    font-size: 14.5px;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .lang-tab.active {
    background: var(--color-primary);
    border-color: var(--color-accent);
    color: var(--color-accent);
  }

  .tab-artwork-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    background: var(--surface-1);
    border-radius: var(--radius-md);
    border: 1px solid var(--color-border);
    font-size: 13.5px;
    color: var(--text-muted);
  }

  .btn-change-image {
    padding: 6px 14px;
    background: var(--surface-3);
    border: 1px solid var(--color-border);
    color: var(--text-main);
    border-radius: var(--radius-sm);
    cursor: pointer;
    font-size: 12.5px;
    font-weight: 700;
  }

  .fields-control-strip {
    display: flex;
    flex-direction: column;
    gap: 14px;
    background: var(--surface-1);
    padding: 16px;
    border-radius: var(--radius-lg);
    border: 1px solid var(--color-border);
  }

  .fields-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .fields-header h5 {
    font-size: 15px;
    font-weight: 800;
    color: var(--text-main);
  }

  .btn-add-field {
    padding: 6px 12px;
    background: var(--color-accent);
    color: #2B0A3D;
    font-weight: 800;
    font-size: 12.5px;
    border: none;
    border-radius: var(--radius-sm);
    cursor: pointer;
  }

  .fields-editor-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .field-edit-card {
    background: var(--surface-2);
    border: 1.5px solid var(--color-border);
    border-radius: var(--radius-md);
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .field-edit-card.selected {
    border-color: var(--color-accent);
  }

  .card-top-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .field-pill-tag {
    font-size: 12px;
    font-family: monospace;
    color: var(--color-accent);
  }

  .btn-delete-field {
    background: transparent;
    border: none;
    cursor: pointer;
    font-size: 14px;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
    gap: 10px;
  }

  .input-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .input-item label {
    font-size: 11.5px;
    font-weight: 700;
    color: var(--text-dim);
  }

  .mini-input,
  .mini-select {
    padding: 6px 8px;
    background: var(--surface-1);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    color: var(--text-main);
    font-size: 13px;
    outline: none;
  }

  .mini-color-wrap {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
  }

  .mini-color-input {
    width: 28px;
    height: 28px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    background: transparent;
  }

  .builder-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: 10px;
  }

  .upload-section {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-top: 12px;
  }

  .upload-label {
    font-size: 13.5px;
    font-weight: 700;
    color: var(--text-muted);
  }

  .drop-zone {
    border: 2px dashed var(--color-border-strong);
    padding: 24px 16px;
    border-radius: var(--radius-md);
    text-align: center;
    cursor: pointer;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  .drop-icon {
    font-size: 32px;
  }

  .hidden-input {
    display: none;
  }

  .preview-thumb-box {
    margin-top: 12px;
    text-align: center;
  }

  .preview-thumb {
    max-height: 140px;
    border-radius: var(--radius-md);
  }

  .modal-actions {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  .empty-text {
    text-align: center;
    color: var(--text-muted);
    padding: 24px 0;
    font-size: 14px;
  }

  .stats-note {
    font-size: 13px;
    font-weight: 700;
    color: var(--text-muted);
    margin-bottom: 12px;
  }

  .msg-cell {
    max-width: 260px;
    overflow-wrap: anywhere;
  }
</style>
