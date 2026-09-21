<script lang="ts">
  import { t } from '../../i18n';

  interface Props {
    id?: string;
    label?: string;
    value?: string;
    placeholder?: string;
    type?: string;
    maxlength?: number;
    error?: string;
    optional?: boolean;
    disabled?: boolean;
    autocomplete?: AutoFill;
    oninput?: (e: Event) => void;
  }

  const fallbackId = `inp-${Math.random().toString(36).slice(2, 9)}`;

  let {
    id = fallbackId,
    label,
    value = $bindable(''),
    placeholder = '',
    type = 'text',
    maxlength,
    error,
    optional = false,
    disabled = false,
    autocomplete,
    oninput
  }: Props = $props();
</script>

<div class="input-wrapper">
  {#if label}
    <label class="label" for={id}>
      <span>{label}</span>
      {#if optional}
        <span class="opt">{$t('card.optional')}</span>
      {/if}
    </label>
  {/if}

  <input
    {id}
    {type}
    bind:value
    {placeholder}
    {maxlength}
    {disabled}
    {autocomplete}
    class="input"
    class:has-error={!!error}
    aria-invalid={error ? 'true' : undefined}
    aria-describedby={error ? `${id}-error` : undefined}
    {oninput}
  />

  {#if error}
    <span class="error-msg" id="{id}-error" role="alert">{error}</span>
  {/if}
</div>
<style>
  .input-wrapper {
    display: flex;
    flex-direction: column;
    gap: 6px;
    width: 100%;
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

  .input {
    width: 100%;
    padding: 12px 14px;
    background: var(--surface-1);
    border: 1.5px solid var(--color-border);
    border-radius: var(--radius-md);
    color: var(--text-main);
    font-size: 15px;
    outline: none;
    transition: all 0.15s ease;
  }

  .input:focus {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 3px var(--ring-focus);
  }

  .input.has-error {
    border-color: var(--color-danger);
  }

  .error-msg {
    font-size: 12.5px;
    color: var(--color-danger);
  }
</style>
