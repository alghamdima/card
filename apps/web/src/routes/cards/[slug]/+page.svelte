<script lang="ts">
  import { page } from '$app/state';
  import { onMount } from 'svelte';
  import { t, locale } from '$lib/i18n';
  import { campaignsApi } from '$lib/api/campaigns';
  import type { Campaign, TextFieldConfig } from '$lib/types/campaign.types';
  import CardPreview from '$lib/components/cards/CardPreview.svelte';
  import Button from '$lib/components/ui/Button.svelte';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';
  import LanguageSwitcher from '$lib/components/ui/LanguageSwitcher.svelte';
  import { showToast } from '$lib/components/ui/toast.store';

  let slug = $derived(page.params.slug);
  let campaign = $state<Campaign | null>(null);
  let loading = $state(true);
  let errorMsg = $state('');

  // Selected language for the card: 'ar' or 'en'
  let cardLang = $state<'ar' | 'en'>('ar');

  // Dynamic field values
  let fieldValues = $state<Record<string, string>>({});

  // Legacy fallback fields
  let toName = $state('');
  let messageText = $state('');
  let fromName = $state('');
  let isAnonymous = $state(false);

  let isGenerating = $state(false);
  let downloadedSuccess = $state(false);
  let cardPreviewComponent = $state<CardPreview | null>(null);

  // Active template variant derived from cardLang
  let activeVariant = $derived.by(() => {
    if (!campaign) return null;
    if (cardLang === 'en' && campaign.templateEN?.image) {
      return campaign.templateEN;
    }
    if (campaign.templateAR?.image) {
      return campaign.templateAR;
    }
    return null;
  });

  let activeImage = $derived(activeVariant?.image || campaign?.image || '');
  let activeFields = $derived(activeVariant?.fields || []);

  async function loadCampaign() {
    if (!slug) return;
    loading = true;
    errorMsg = '';
    try {
      campaign = await campaignsApi.getPublic(slug);
      // Initialize language from current app language
      let currentAppLocale = 'ar';
      locale.subscribe((l) => (currentAppLocale = l))();
      cardLang = currentAppLocale === 'en' ? 'en' : 'ar';
      initFieldDefaults();
    } catch (e: any) {
      errorMsg = e.message || $t('card.notFound');
    } finally {
      loading = false;
    }
  }

  function initFieldDefaults() {
    fieldValues = {};
    if (activeFields && activeFields.length > 0) {
      activeFields.forEach((f) => {
        fieldValues[f.id] = '';
      });
    }
  }

  $effect(() => {
    if (slug) {
      loadCampaign();
    }
  });

  function handleCardLangChange(newLang: 'ar' | 'en') {
    cardLang = newLang;
    initFieldDefaults();
  }

  async function handleDownload() {
    if (!campaign || !cardPreviewComponent) return;

    isGenerating = true;
    try {
      const canvas = cardPreviewComponent.getCanvas();
      if (!canvas) throw new Error('Canvas not initialized');

      // 1. Trigger PNG download
      const dataUrl = canvas.toDataURL('image/png');
      const link = document.createElement('a');
      link.download = `${campaign.slug}-${cardLang}-card.png`;
      link.href = dataUrl;
      link.click();

      // 2. Prepare submission details
      const primaryName = fieldValues['emp_name'] || fieldValues['name'] || toName || '';
      const primaryMsg = fieldValues['job_title'] || fieldValues['title'] || messageText || '';

      const isMobile = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent);
      const device = /iPhone|iPad|iPod/i.test(navigator.userAgent)
        ? 'iPhone'
        : /Android/i.test(navigator.userAgent)
        ? 'Android'
        : 'Desktop';

      campaignsApi.submitCard(campaign.slug, {
        from: isAnonymous ? 'Anonymous' : fromName || 'Anonymous',
        to: primaryName,
        message: primaryMsg,
        lang: cardLang,
        fieldValues: { ...fieldValues },
        device
      }).catch(() => {
        // Non-blocking telemetry
      });

      downloadedSuccess = true;
      showToast($t('card.successToast'), 'success');
      initFieldDefaults();
      toName = '';
      messageText = '';
      fromName = '';
      isAnonymous = false;
    } catch (e: any) {
      showToast(e.message || 'Export error', 'error');
    } finally {
      isGenerating = false;
    }
  }

  function makeAnother() {
    downloadedSuccess = false;
  }
</script>

<svelte:head>
  <title>{campaign ? `${campaign.title} | ${$t('app.title')}` : $t('app.title')}</title>
</svelte:head>

<div class="card-page">
  <div class="page-wrap">
    <div class="top-nav">
      <a href="/" class="back-link">
        <span class="arrow">←</span>
        <span>{$t('app.back')}</span>
      </a>

      <!-- Brand Logo Header in Card Creation -->
      <div class="brand-badge">
        <img src="/images/brand/aljuf-ar.png" alt="ALJ Finance" class="aljuf-logo" />
      </div>

      <LanguageSwitcher />
    </div>

    {#if loading}
      <LoadingState />
    {:else if errorMsg}
      <ErrorState message={errorMsg} onretry={loadCampaign} />
    {:else if campaign}
      <header class="header">
        <h1>{campaign.title}</h1>
        <p>{$t('card.livePreview')}</p>

        <!-- Template Language Selection Tabs for Employee -->
        <div class="template-lang-pill-wrap">
          <button
            class="lang-pill"
            class:selected={cardLang === 'ar'}
            onclick={() => handleCardLangChange('ar')}
          >
            🇸🇦 العربية
          </button>
          <button
            class="lang-pill"
            class:selected={cardLang === 'en'}
            onclick={() => handleCardLangChange('en')}
          >
            🇬🇧 English
          </button>
        </div>
      </header>

      <div class="card-box">
        <CardPreview
          bind:this={cardPreviewComponent}
          imageSrc={activeImage}
          boxes={campaign.boxes}
          dynamicFields={activeFields}
          {fieldValues}
          to={fieldValues['emp_name'] || toName}
          from={isAnonymous ? '' : fromName}
          message={fieldValues['job_title'] || messageText}
        />
      </div>

      {#if !downloadedSuccess}
        <div class="form-card">
          {#if activeFields && activeFields.length > 0}
            <!-- Dynamic Form Fields Generated from Admin Config -->
            <div class="dynamic-inputs-wrap">
              {#each activeFields as field (field.id)}
                <div class="input-field">
                  <label for={`field_${field.id}`} class="field-label">
                    {field.label || field.name}
                  </label>
                  <input
                    id={`field_${field.id}`}
                    type="text"
                    class="card-text-input"
                    placeholder={field.placeholder || `أدخل ${field.label || field.name}...`}
                    maxlength={field.maxChars || 80}
                    bind:value={fieldValues[field.id]}
                  />
                </div>
              {/each}
            </div>
          {:else}
            <!-- Legacy Form Fallback -->
            <div class="dynamic-inputs-wrap">
              <div class="input-field">
                <label for="to_legacy" class="field-label">{$t('card.to')}</label>
                <input
                  id="to_legacy"
                  type="text"
                  class="card-text-input"
                  placeholder={$t('card.toPlaceholder')}
                  bind:value={toName}
                />
              </div>
              <div class="input-field">
                <label for="msg_legacy" class="field-label">{$t('card.message')}</label>
                <textarea
                  id="msg_legacy"
                  class="card-textarea"
                  placeholder={$t('card.messagePlaceholder')}
                  bind:value={messageText}
                ></textarea>
              </div>
            </div>
          {/if}

          <div class="submit-wrap">
            <Button
              variant="primary"
              loading={isGenerating}
              disabled={isGenerating}
              onclick={handleDownload}
            >
              <span>📥</span>
              <span>{isGenerating ? $t('card.generating') : $t('card.download')}</span>
            </Button>
          </div>
        </div>
      {:else}
        <div class="success-box">
          <div class="how-to-save">
            <h3>{$t('card.howToSave')}</h3>
            <p class="guide-item">📱 {$t('card.iosInstructions')}</p>
          </div>

          <Button variant="secondary" onclick={makeAnother}>
            <span>🔄</span>
            <span>{$t('card.makeAnother')}</span>
          </Button>
        </div>
      {/if}
    {/if}
  </div>
</div>

<style>
  .card-page {
    min-height: 100vh;
    display: flex;
    justify-content: center;
    padding: 20px 16px 48px;
    background: var(--bg-app);
  }

  .page-wrap {
    width: 100%;
    max-width: 500px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .top-nav {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .brand-badge .aljuf-logo {
    height: 38px;
    object-fit: contain;
  }

  .back-link {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--text-muted);
    font-size: 14px;
    font-weight: 600;
    transition: color 0.15s;
    text-decoration: none;
  }

  .back-link:hover {
    color: var(--color-accent);
  }

  .header {
    text-align: center;
    display: flex;
    flex-direction: column;
    gap: 8px;
    align-items: center;
  }

  .header h1 {
    font-size: 24px;
    font-weight: 800;
    color: var(--text-main);
  }

  .header p {
    font-size: 14px;
    color: var(--text-dim);
  }

  .template-lang-pill-wrap {
    display: inline-flex;
    background: var(--surface-1);
    padding: 4px;
    border-radius: var(--radius-full);
    border: 1px solid var(--color-border);
    gap: 4px;
    margin-top: 6px;
  }

  .lang-pill {
    padding: 6px 18px;
    border: none;
    background: transparent;
    border-radius: var(--radius-full);
    color: var(--text-muted);
    font-size: 13.5px;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .lang-pill.selected {
    background: var(--color-primary);
    color: var(--color-accent);
    box-shadow: 0 2px 8px rgba(60, 16, 83, 0.4);
  }

  .card-box {
    width: 100%;
  }

  .form-card {
    background: var(--surface-1);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    box-shadow: var(--shadow-sm);
  }

  .dynamic-inputs-wrap {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .input-field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .field-label {
    font-size: 13.5px;
    font-weight: 700;
    color: var(--text-muted);
  }

  .card-text-input,
  .card-textarea {
    width: 100%;
    padding: 12px 14px;
    background: var(--surface-2);
    border: 1.5px solid var(--color-border);
    border-radius: var(--radius-md);
    color: var(--text-main);
    font-size: 15px;
    outline: none;
    transition: all 0.15s ease;
  }

  .card-text-input:focus,
  .card-textarea:focus {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 3px var(--ring-focus);
  }

  .submit-wrap {
    margin-top: 6px;
  }

  .submit-wrap :global(button) {
    width: 100%;
  }

  .success-box {
    background: var(--surface-1);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    padding: 24px 20px;
    display: flex;
    flex-direction: column;
    gap: 20px;
    text-align: center;
  }

  .how-to-save {
    display: flex;
    flex-direction: column;
    gap: 10px;
    background: var(--surface-2);
    padding: 16px;
    border-radius: var(--radius-md);
  }

  .how-to-save h3 {
    font-size: 15px;
    font-weight: 800;
    color: var(--text-main);
  }

  .guide-item {
    font-size: 13px;
    color: var(--text-muted);
    line-height: 1.5;
  }
</style>
