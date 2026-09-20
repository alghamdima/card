<script lang="ts">
  import { onMount } from 'svelte';
  import { t } from '$lib/i18n';
  import { campaignsApi } from '$lib/api/campaigns';
  import type { Campaign, CampaignSummary, Card, TextFieldConfig } from '$lib/types/campaign.types';
  import Button from '$lib/components/ui/Button.svelte';
  import Input from '$lib/components/ui/Input.svelte';
  import Modal from '$lib/components/ui/Modal.svelte';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';
  import { showToast } from '$lib/components/ui/toast.store';
  import TemplateCanvasEditor from '$lib/components/admin/TemplateCanvasEditor.svelte';

  let campaigns = $state<CampaignSummary[]>([]);
  let loading = $state(true);
  let errorMsg = $state('');

  // Create Campaign Modal
  let showCreateModal = $state(false);
  let newTitle = $state('');
  let newTextColor = $state('#FFFFFF');
  let newHeadColor = $state('#FFCD00');
  let newImageBase64 = $state('');
  let isSubmitting = $state(false);

  // Template & Position Builder Modal
  let showBuilderModal = $state(false);
  let editingCampaign = $state<Campaign | null>(null);
  let activeLangTab = $state<'ar' | 'en'>('ar');
  let activeFieldId = $state<string>('emp_name');

  // Fields and templates for the editor
  let arImage = $state('');
  let arFields = $state<TextFieldConfig[]>([]);
  let enImage = $state('');
  let enFields = $state<TextFieldConfig[]>([]);

  // Sample values for real-time preview
  let sampleValues = $state<Record<string, string>>({
    emp_name: 'علاء أبوراشد | Alaa Aburashed',
    job_title: 'مدير التواصل الداخلي | Internal Communication Manager'
  });

  // Stats Modal
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

  function handleCreateFileUpload(file: File) {
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
      const initialFields: TextFieldConfig[] = [
        {
          id: 'emp_name',
          name: 'emp_name',
          label: 'اسم الموظف | Full Name',
          placeholder: 'مثال: علاء أبو راشد',
          x: 230,
          y: 620,
          width: 620,
          height: 70,
          fontSize: 47,
          color: '#FFFFFF',
          weight: 'bold',
          align: 'center',
          order: 1
        },
        {
          id: 'job_title',
          name: 'job_title',
          label: 'المسمى الوظيفي | Job Title',
          placeholder: 'مثال: مدير التواصل الداخلي',
          x: 230,
          y: 705,
          width: 620,
          height: 60,
          fontSize: 34,
          color: '#B9B9C2',
          weight: 'regular',
          align: 'center',
          order: 2
        }
      ];

      const created = await campaignsApi.create({
        title: newTitle.trim(),
        textColor: newTextColor,
        headColor: newHeadColor,
        image: newImageBase64,
        thumb: newImageBase64,
        templateAR: {
          image: newImageBase64,
          fields: initialFields
        },
        templateEN: {
          image: newImageBase64,
          fields: initialFields
        }
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

  async function openBuilder(slug: string) {
    try {
      const camp = await campaignsApi.getAdmin(slug);
      editingCampaign = camp;
      activeLangTab = 'ar';

      // Load AR template
      arImage = camp.templateAR?.image || camp.image;
      arFields = camp.templateAR?.fields && camp.templateAR.fields.length > 0
        ? JSON.parse(JSON.stringify(camp.templateAR.fields))
        : [
            {
              id: 'emp_name',
              name: 'emp_name',
              label: 'اسم الموظف',
              x: 230,
              y: 620,
              width: 620,
              height: 70,
              fontSize: 47,
              color: '#FFFFFF',
              weight: 'bold',
              align: 'center',
              order: 1
            },
            {
              id: 'job_title',
              name: 'job_title',
              label: 'المسمى الوظيفي',
              x: 230,
              y: 705,
              width: 620,
              height: 60,
              fontSize: 34,
              color: '#B9B9C2',
              weight: 'regular',
              align: 'center',
              order: 2
            }
          ];

      // Load EN template
      enImage = camp.templateEN?.image || camp.image;
      enFields = camp.templateEN?.fields && camp.templateEN.fields.length > 0
        ? JSON.parse(JSON.stringify(camp.templateEN.fields))
        : JSON.parse(JSON.stringify(arFields));

      activeFieldId = arFields[0]?.id || 'emp_name';
      showBuilderModal = true;
    } catch (e: any) {
      showToast(e.message || 'Failed to load template', 'error');
    }
  }

  function handleImageTabUpload(file: File, lang: 'ar' | 'en') {
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (e) => {
      const b64 = (e.target?.result as string) || '';
      if (lang === 'ar') arImage = b64;
      else enImage = b64;
    };
    reader.readAsDataURL(file);
  }

  function addField(lang: 'ar' | 'en') {
    const list = lang === 'ar' ? arFields : enFields;
    const newId = `field_${Date.now()}`;
    const newField: TextFieldConfig = {
      id: newId,
      name: newId,
      label: `حقل نصي جديد ${list.length + 1}`,
      x: 250,
      y: 780 + list.length * 30,
      width: 580,
      height: 65,
      fontSize: 36,
      color: '#FFCD00',
      weight: 'bold',
      align: 'center',
      order: list.length + 1
    };

    if (lang === 'ar') {
      arFields = [...arFields, newField];
    } else {
      enFields = [...enFields, newField];
    }
    activeFieldId = newId;
    sampleValues[newId] = 'نص توضيحي جديد';
  }

  function deleteField(lang: 'ar' | 'en', id: string) {
    if (lang === 'ar') {
      arFields = arFields.filter((f) => f.id !== id);
      if (activeFieldId === id && arFields.length > 0) activeFieldId = arFields[0].id;
    } else {
      enFields = enFields.filter((f) => f.id !== id);
      if (activeFieldId === id && enFields.length > 0) activeFieldId = enFields[0].id;
    }
  }

  async function handleSaveTemplate() {
    if (!editingCampaign) return;
    isSubmitting = true;
    try {
      await campaignsApi.update(editingCampaign.slug, {
        templateAR: {
          image: arImage,
          fields: arFields
        },
        templateEN: {
          image: enImage,
          fields: enFields
        }
      });
      showToast($t('app.save'), 'success');
      showBuilderModal = false;
      await loadCampaigns();
    } catch (e: any) {
      showToast(e.message || 'Error saving template', 'error');
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
                <td><strong class="count-num">{camp.totalCards}</strong></td>
                <td>
                  <span class="status-badge" class:active={camp.active}>
                    {camp.active ? $t('app.active') : $t('app.inactive')}
                  </span>
                </td>
                <td>{new Date(camp.createdAt).toLocaleDateString()}</td>
                <td>
                  <div class="row-actions">
                    <button class="btn-builder" onclick={() => openBuilder(camp.slug)}>
                      🎨 {$t('admin.templateSettings')}
                    </button>
                    <button class="btn-text" onclick={() => openStats(camp.slug)}>
                      {$t('admin.viewCards')}
                    </button>
                    <button class="btn-text danger" onclick={() => handleDelete(camp.slug)}>
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
<Modal
  open={showCreateModal}
  title={$t('admin.newCampaign')}
  onclose={() => (showCreateModal = false)}
>
  <div class="modal-form">
    <Input
      label={$t('admin.campaignName')}
      placeholder="مثال: تهنئة عيد الفطر المبارك 2026"
      value={newTitle}
      oninput={(e) => (newTitle = (e.target as HTMLInputElement).value)}
    />

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
            if (files && files[0]) handleCreateFileUpload(files[0]);
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

<!-- Visual Template & Position Builder Modal -->
<Modal
  open={showBuilderModal}
  title={`${$t('admin.templateSettings')} - ${editingCampaign?.title || ''}`}
  maxWidth="1100px"
  onclose={() => (showBuilderModal = false)}
>
  <div class="builder-modal-content">
    <!-- Language Switcher Tabs -->
    <div class="lang-builder-tabs">
      <button
        class="lang-tab"
        class:active={activeLangTab === 'ar'}
        onclick={() => (activeLangTab = 'ar')}
      >
        🇸🇦 {$t('admin.arabicTemplate')}
      </button>
      <button
        class="lang-tab"
        class:active={activeLangTab === 'en'}
        onclick={() => (activeLangTab = 'en')}
      >
        🇬🇧 {$t('admin.englishTemplate')}
      </button>
    </div>

    <!-- Artwork Background Upload for Current Tab -->
    <div class="tab-artwork-bar">
      <span>🖼️ {activeLangTab === 'ar' ? 'صورة القالب العربي' : 'English Template Image'}:</span>
      <label class="btn-change-image">
        <span>تغيير الصورة</span>
        <input
          type="file"
          accept="image/*"
          class="hidden-input"
          onchange={(e) => {
            const files = (e.target as HTMLInputElement).files;
            if (files && files[0]) handleImageTabUpload(files[0], activeLangTab);
          }}
        />
      </label>
    </div>

    <!-- Interactive Dual Canvas Component -->
    <TemplateCanvasEditor
      imageSrc={activeLangTab === 'ar' ? arImage : enImage}
      fields={activeLangTab === 'ar' ? arFields : enFields}
      {activeFieldId}
      {sampleValues}
      onfieldchange={(f) => {
        if (activeLangTab === 'ar') arFields = f;
        else enFields = f;
      }}
      onselectfield={(id) => (activeFieldId = id)}
    />

    <!-- Dynamic Fields Control Strip -->
    <div class="fields-control-strip">
      <div class="fields-header">
        <h5>الحقول النصية ({activeLangTab === 'ar' ? arFields.length : enFields.length})</h5>
        <button class="btn-add-field" onclick={() => addField(activeLangTab)}>
          ➕ {$t('admin.addTextField')}
        </button>
      </div>

      <div class="fields-editor-list">
        {#each activeLangTab === 'ar' ? arFields : enFields as field (field.id)}
          <div class="field-edit-card" class:selected={field.id === activeFieldId}>
            <div class="card-top-row">
              <span class="field-pill-tag">#{field.id}</span>
              <button
                class="btn-delete-field"
                onclick={() => deleteField(activeLangTab, field.id)}
                title={$t('admin.deleteField')}
              >
                🗑️
              </button>
            </div>

            <div class="form-grid">
              <div class="input-item">
                <label for={`label_${field.id}`}>تسمية الحقل</label>
                <input
                  id={`label_${field.id}`}
                  type="text"
                  bind:value={field.label}
                  class="mini-input"
                />
              </div>

              <div class="input-item">
                <label for={`sample_${field.id}`}>القيمة التجريبية</label>
                <input
                  id={`sample_${field.id}`}
                  type="text"
                  bind:value={sampleValues[field.id]}
                  class="mini-input"
                />
              </div>

              <div class="input-item">
                <label for={`size_${field.id}`}>حجم الخط (px)</label>
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
                <label for={`color_${field.id}`}>لون النص</label>
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
                <label for={`align_${field.id}`}>المحاذاة</label>
                <select id={`align_${field.id}`} bind:value={field.align} class="mini-select">
                  <option value="center">وسط (Center)</option>
                  <option value="right">يمين (Right)</option>
                  <option value="left">يسار (Left)</option>
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
        disabled={isSubmitting}
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
            <th>{$t('admin.dateAndTime')}</th>
          </tr>
        </thead>
        <tbody>
          {#each campaignCards as card (card.id)}
            <tr>
              <td><strong>{card.to || '-'}</strong></td>
              <td class="msg-cell">{card.message || '-'}</td>
              <td>{card.from || '-'}</td>
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
</style>
