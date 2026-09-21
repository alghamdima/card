<script lang="ts">
  import { base } from '$app/paths';
  import { page } from '$app/state';
  import { untrack } from 'svelte';
  import { t, locale, translateError } from '$lib/i18n';
  import { get } from 'svelte/store';
  import { campaignsApi } from '$lib/api/campaigns';
  import type { Campaign, TemplateVariant } from '$lib/types/campaign.types';
  import CardPreview from '$lib/components/cards/CardPreview.svelte';
  import CardEditor from '$lib/components/cards/CardEditor.svelte';
  import BrandLogo from '$lib/components/ui/BrandLogo.svelte';
  import Button from '$lib/components/ui/Button.svelte';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';
  import LanguageSwitcher from '$lib/components/ui/LanguageSwitcher.svelte';
  import ThemeSwitcher from '$lib/components/ui/ThemeSwitcher.svelte';
  import { showToast } from '$lib/components/ui/toast.store';

  type CardLang = 'ar' | 'en';

  let slug = $derived(page.params.slug);
  let campaign = $state<Campaign | null>(null);
  let loading = $state(true);
  let errorMsg = $state('');

  // Language of the card template the employee designs (independent from the UI language).
  let cardLang = $state<CardLang>('ar');

  // Values typed into the dynamic template fields, keyed by field id.
  let fieldValues = $state<Record<string, string>>({});

  // Legacy campaigns (fixed boxes instead of dynamic fields) use these fields.
  let toName = $state('');
  let messageText = $state('');
  let fromName = $state('');
  let isAnonymous = $state(false);

  let isGenerating = $state(false);
  let downloadedSuccess = $state(false);
  let cardPreviewComponent = $state<CardPreview | null>(null);

  const isIOS = typeof navigator !== 'undefined' && /iPhone|iPad|iPod/i.test(navigator.userAgent);

  let hasBothTemplates = $derived(!!campaign?.templateAR && !!campaign?.templateEN);

  // Template variant for the selected card language; a missing variant falls back to the other one.
  let activeVariant = $derived.by<TemplateVariant | null>(() => {
    if (!campaign) return null;
    const { templateAR, templateEN } = campaign;
    return (cardLang === 'en' ? (templateEN ?? templateAR) : (templateAR ?? templateEN)) ?? null;
  });

  // The server omits a variant image identical to the campaign image.
  let activeImage = $derived(activeVariant?.image || campaign?.image || '');
  let activeFields = $derived(activeVariant?.fields ?? []);

  let displayTitle = $derived.by(() => {
    if (!campaign) return '';
    return cardLang === 'en' ? campaign.titleEN || campaign.title : campaign.titleAR || campaign.title;
  });

  // Discards the response of a superseded load (fast navigation between two occasions).
  let loadToken = 0;

  async function loadCampaign() {
    const current = slug;
    if (!current) return;

    const token = ++loadToken;
    loading = true;
    errorMsg = '';
    campaign = null;
    downloadedSuccess = false;
    try {
      const loaded = await campaignsApi.getPublic(current);
      if (token !== loadToken) return;
      campaign = loaded;
      cardLang = initialCardLang(loaded);
      resetInputs();
    } catch (e) {
      if (token !== loadToken) return;
      errorMsg = translateError(e, 'card.notFound');
    } finally {
      if (token === loadToken) loading = false;
    }
  }

  // Both/no templates: follow the UI language; a single template dictates the card language.
  function initialCardLang(c: Campaign): CardLang {
    if (c.templateEN && !c.templateAR) return 'en';
    if (c.templateAR && !c.templateEN) return 'ar';
    return get(locale) === 'en' ? 'en' : 'ar';
  }

  // Clears typed text and makes sure every field of the active template has a (possibly empty) value.
  function resetInputs() {
    const cleared: Record<string, string> = {};
    for (const f of activeFields) cleared[f.id] = '';
    fieldValues = cleared;
    toName = '';
    messageText = '';
    fromName = '';
    isAnonymous = false;
  }

  $effect(() => {
    if (slug) untrack(loadCampaign);
  });

  function handleCardLangChange(newLang: CardLang) {
    if (newLang === cardLang) return;
    cardLang = newLang;
    // Keep what was typed for fields both templates share; add empty entries for the new ones.
    const next: Record<string, string> = {};
    for (const f of activeFields) next[f.id] = fieldValues[f.id] ?? '';
    fieldValues = next;
  }

  /** Labels of the fields the employee still has to fill; empty when the card can be exported. */
  function missingFields(): string[] {
    if (activeFields.length > 0) {
      const missing = activeFields.filter((f) => f.required && !fieldValues[f.id]?.trim());
      if (missing.length === 0 && activeFields.every((f) => !fieldValues[f.id]?.trim())) {
        missing.push(activeFields[0]);
      }
      return missing.map((f) => f.label || f.name);
    }
    return toName.trim() || messageText.trim() ? [] : [$t('card.to')];
  }

  async function toPngBlob(canvas: HTMLCanvasElement): Promise<Blob> {
    const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, 'image/png'));
    if (!blob) throw new Error('Canvas export failed');
    return blob;
  }

  /** Saves the PNG. Returns false when the user dismissed the iOS share sheet without saving. */
  async function savePng(blob: Blob, filename: string): Promise<boolean> {
    const file = new File([blob], filename, { type: 'image/png' });

    // iOS Safari ignores <a download>: the share sheet ("Save Image") is the reliable way to reach Photos.
    if (isIOS && navigator.canShare?.({ files: [file] })) {
      try {
        await navigator.share({ files: [file] });
        return true;
      } catch (e) {
        if (e instanceof DOMException && e.name === 'AbortError') return false;
        // Any other failure falls through to the regular download.
      }
    }

    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    setTimeout(() => URL.revokeObjectURL(url), 10_000);
    return true;
  }

  async function handleDownload() {
    if (!campaign || !cardPreviewComponent || isGenerating) return;

    const missing = missingFields();
    if (missing.length > 0) {
      showToast($t('card.requiredFields', { fields: missing.join($locale === 'ar' ? '، ' : ', ') }), 'error');
      return;
    }

    isGenerating = true;
    try {
      const canvas = cardPreviewComponent.getCanvas();
      const saved = await savePng(await toPngBlob(canvas), `${campaign.slug}-${cardLang}-card.png`);
      if (!saved) return;

      // Anonymous telemetry for the admin analytics; never blocks or fails the export.
      const device = isIOS ? 'iPhone' : /Android/i.test(navigator.userAgent) ? 'Android' : 'Desktop';

      // Resolve recipient (to), message, and sender (from) from dynamic fields if not fixed
      const sortedFields = [...activeFields].sort((a, b) => (a.order || 0) - (b.order || 0));
      const firstFieldVal = sortedFields[0] ? (fieldValues[sortedFields[0].id] || '') : '';
      const secondFieldVal = sortedFields[1] ? (fieldValues[sortedFields[1].id] || '') : '';
      const thirdFieldVal = sortedFields[2] ? (fieldValues[sortedFields[2].id] || '') : '';

      const resolvedTo = fieldValues['emp_name'] || fieldValues['name'] || toName || firstFieldVal;
      const resolvedMessage = fieldValues['job_title'] || fieldValues['title'] || messageText || secondFieldVal;
      const resolvedFrom = isAnonymous ? '' : (fromName.trim() || thirdFieldVal.trim());

      campaignsApi
        .submitCard(campaign.slug, {
          from: resolvedFrom,
          to: resolvedTo,
          message: resolvedMessage,
          lang: cardLang,
          fieldValues: { ...fieldValues },
          device
        })
        .catch(() => {});

      downloadedSuccess = true;
      showToast($t('card.successToast'), 'success');
    } catch {
      showToast($t('card.exportFailed'), 'error');
    } finally {
      isGenerating = false;
    }
  }

  function makeAnother() {
    resetInputs();
    downloadedSuccess = false;
  }
</script>

<svelte:head>
  <title>{displayTitle ? `${displayTitle} | ${$t('app.title')}` : $t('app.title')}</title>
</svelte:head>

<div class="card-page">
  <div class="page-wrap">
    <header class="card-header-bar">
      <!-- Start: Back to Home -->
      <div class="header-side start">
        <a href="{base}/" class="back-link" title={$t('app.back')}>
          <span class="arrow" aria-hidden="true">{$locale === 'ar' ? '→' : '←'}</span>
          <span class="back-text">{$t('app.back')}</span>
        </a>
      </div>

      <!-- Center: Clean Brand Identity -->
      <div class="header-center">
        <a href="{base}/" class="brand-link" title={$t('app.companyName')}>
          <BrandLogo />
        </a>
      </div>

      <!-- End: Minimal Controls -->
      <div class="header-side end">
        <div class="nav-controls">
          <ThemeSwitcher />
          <LanguageSwitcher />
        </div>
      </div>
    </header>

    {#if loading}
      <LoadingState />
    {:else if errorMsg}
      <ErrorState message={errorMsg} onretry={loadCampaign} />
    {:else if campaign}
      <section class="occasion-title-box">
        <h1 class="occasion-title">{displayTitle}</h1>

        <!-- Template language selection (only when the occasion has both variants) -->
        {#if hasBothTemplates}
          <div class="template-lang-pill-wrap" role="group" aria-label={$t('card.language')}>
            <button
              type="button"
              class="lang-pill"
              class:selected={cardLang === 'ar'}
              aria-pressed={cardLang === 'ar'}
              lang="ar"
              onclick={() => handleCardLangChange('ar')}
            >
              العربية
            </button>
            <button
              type="button"
              class="lang-pill"
              class:selected={cardLang === 'en'}
              aria-pressed={cardLang === 'en'}
              lang="en"
              onclick={() => handleCardLangChange('en')}
            >
              English
            </button>
          </div>
        {/if}
      </section>

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
        <form
          class="form-card"
          onsubmit={(e) => {
            e.preventDefault();
            handleDownload();
          }}
        >
          {#if activeFields.length > 0}
            <!-- Dynamic form fields generated from the admin template -->
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
                    placeholder={field.placeholder || $t('card.enterField', { label: field.label || field.name })}
                    maxlength={field.maxChars || 80}
                    aria-required={field.required ? 'true' : undefined}
                    bind:value={fieldValues[field.id]}
                  />
                </div>
              {/each}
            </div>
          {:else}
            <!-- Legacy campaigns: fixed recipient / message / sender boxes -->
            <CardEditor
              to={toName}
              message={messageText}
              from={fromName}
              {isAnonymous}
              ontochange={(v) => (toName = v)}
              onmessagechange={(v) => (messageText = v)}
              onfromchange={(v) => (fromName = v)}
              onanonchange={(v) => (isAnonymous = v)}
            />
          {/if}

          <div class="submit-wrap">
            <Button type="submit" variant="primary" loading={isGenerating} disabled={isGenerating}>
              <span aria-hidden="true">📥</span>
              <span>{isGenerating ? $t('card.generating') : $t('card.download')}</span>
            </Button>
          </div>
        </form>
      {:else}
        <div class="success-box">
          <div class="success-icon" aria-hidden="true">🎉</div>
          <h3>{$t('card.successToast')}</h3>
          <p class="success-hint">{$t('card.savedHint')}</p>
          {#if isIOS}
            <p class="success-hint">{$t('card.iosInstructions')}</p>
          {/if}

          <Button variant="secondary" onclick={makeAnother}>
            <span aria-hidden="true">🔄</span>
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
    padding: 24px 16px 56px;
    background: var(--bg-app);
  }

  .page-wrap {
    width: 100%;
    max-width: 520px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  /* Redesigned 3-column header bar */
  .card-header-bar {
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: center;
    width: 100%;
    padding: 4px 0 12px;
    border-bottom: 1px solid var(--color-border);
  }

  .header-side.start {
    display: flex;
    justify-content: flex-start;
  }

  .header-side.end {
    display: flex;
    justify-content: flex-end;
  }

  .header-center {
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .brand-link {
    display: flex;
    align-items: center;
    justify-content: center;
    text-decoration: none;
  }

  .nav-controls {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .back-link {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: var(--text-muted);
    font-size: 13.5px;
    font-weight: 700;
    padding: 6px 12px;
    border-radius: var(--radius-full);
    background: var(--surface-1);
    border: 1px solid var(--color-border);
    transition: all 0.2s ease;
    text-decoration: none;
  }

  .back-link:hover {
    color: var(--color-accent);
    border-color: var(--color-accent);
    background: var(--surface-2);
  }

  .arrow {
    font-size: 14px;
    transition: transform 0.2s ease;
  }

  .back-link:hover .arrow {
    transform: translateX(-2px);
  }

  :global([dir="rtl"]) .back-link:hover .arrow {
    transform: translateX(2px);
  }

  /* Occasion Title & Language Switcher Block */
  .occasion-title-box {
    text-align: center;
    display: flex;
    flex-direction: column;
    gap: 8px;
    align-items: center;
    padding: 8px 0;
  }

  .occasion-title {
    font-size: 26px;
    font-weight: 800;
    color: var(--text-main);
    letter-spacing: -0.4px;
    line-height: 1.3;
  }

  .template-lang-pill-wrap {
    display: inline-flex;
    background: var(--surface-1);
    padding: 4px;
    border-radius: var(--radius-full);
    border: 1px solid var(--color-border);
    gap: 4px;
    margin-top: 10px;
    box-shadow: var(--shadow-sm);
  }

  .lang-pill {
    padding: 6px 20px;
    border: none;
    background: transparent;
    border-radius: var(--radius-full);
    color: var(--text-muted);
    font-size: 13.5px;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .lang-pill.selected {
    background: var(--color-primary);
    color: var(--color-accent);
    box-shadow: 0 2px 10px rgba(60, 16, 83, 0.35);
  }

  .lang-pill:not(.selected):hover {
    color: var(--text-main);
  }

  .card-box {
    width: 100%;
  }

  .form-card {
    background: var(--surface-1);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    padding: 22px;
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

  .card-text-input {
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

  .card-text-input:focus {
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
    padding: 28px 20px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    align-items: center;
    text-align: center;
    box-shadow: var(--shadow-sm);
  }

  .success-icon {
    font-size: 40px;
  }

  .success-box h3 {
    font-size: 18px;
    font-weight: 800;
    color: var(--text-main);
  }

  .success-hint {
    font-size: 14px;
    color: var(--text-muted);
  }
</style>
