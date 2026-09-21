<script lang="ts">
  import { locale, t } from '../../i18n';
  import { base } from '$app/paths';
  import { effectiveTheme } from '../../stores/theme.store';

  interface Props {
    height?: number;
  }

  let { height = 48 }: Props = $props();

  // One image at a time: the logo variant follows the language (AR/EN wordmark) and the theme (dark art on light, white art on dark).
  let src = $derived(
    `${base}/images/brand/aljuf-${$locale === 'en' ? 'en' : 'ar'}${$effectiveTheme === 'light' ? '' : '-white'}-tight.png`
  );
</script>

<img {src} alt={$t('app.companyName')} class="brand-logo" style:height="{height}px" />

<style>
  .brand-logo {
    width: auto;
    object-fit: contain;
    display: block;
  }
</style>
