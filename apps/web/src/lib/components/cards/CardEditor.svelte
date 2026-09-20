<script lang="ts">
  import { t } from '../../i18n';
  import Input from '../ui/Input.svelte';

  interface Props {
    to: string;
    message: string;
    from: string;
    isAnonymous: boolean;
    ontochange: (val: string) => void;
    onmessagechange: (val: string) => void;
    onfromchange: (val: string) => void;
    onanonchange: (val: boolean) => void;
  }

  let {
    to,
    message,
    from,
    isAnonymous,
    ontochange,
    onmessagechange,
    onfromchange,
    onanonchange
  }: Props = $props();
</script>

<div class="editor-fields">
  <div class="field">
    <Input
      label={$t('card.to')}
      placeholder={$t('card.toPlaceholder')}
      value={to}
      maxlength={45}
      oninput={(e) => ontochange((e.target as HTMLInputElement).value)}
    />
  </div>

  <div class="field">
    <label class="label" for="msgText">{$t('card.message')}</label>
    <textarea
      id="msgText"
      class="textarea"
      placeholder={$t('card.messagePlaceholder')}
      maxlength={280}
      value={message}
      oninput={(e) => onmessagechange((e.target as HTMLTextAreaElement).value)}
    ></textarea>
  </div>

  <div class="field">
    <div class="from-header">
      <label class="label" for="fromText">
        <span>{$t('card.from')}</span>
        <span class="opt">{$t('card.optional')}</span>
      </label>

      <label class="anon-toggle">
        <input
          type="checkbox"
          checked={isAnonymous}
          onchange={(e) => onanonchange((e.target as HTMLInputElement).checked)}
        />
        <span>{$t('card.sendAnonymous')}</span>
      </label>
    </div>

    <Input
      placeholder={$t('card.fromPlaceholder')}
      value={isAnonymous ? '' : from}
      disabled={isAnonymous}
      maxlength={45}
      oninput={(e) => onfromchange((e.target as HTMLInputElement).value)}
    />
  </div>
</div>

<style>
  .editor-fields {
    display: flex;
    flex-direction: column;
    gap: 16px;
    width: 100%;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .label {
    font-size: 13.5px;
    font-weight: 700;
    color: var(--text-muted);
    display: flex;
    gap: 6px;
  }

  .opt {
    font-size: 12px;
    font-weight: 400;
    color: var(--text-dim);
  }

  .from-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .anon-toggle {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: var(--text-muted);
    cursor: pointer;
  }

  .anon-toggle input {
    cursor: pointer;
    accent-color: var(--color-accent);
  }

  .textarea {
    width: 100%;
    min-height: 90px;
    padding: 12px 14px;
    background: var(--surface-1);
    border: 1.5px solid var(--color-border);
    border-radius: var(--radius-md);
    color: var(--text-main);
    font-size: 15px;
    outline: none;
    resize: vertical;
    transition: all 0.15s ease;
  }

  .textarea:focus {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 3px var(--ring-focus);
  }
</style>
