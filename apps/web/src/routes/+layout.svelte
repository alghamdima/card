<script lang="ts">
  import '../lib/styles/app.css';
  // Side-effect imports: these stores apply the saved theme and keep <html lang/dir> in sync as soon as the app boots.
  import '../lib/stores/theme.store';
  import '../lib/i18n';
  import Toast from '../lib/components/ui/Toast.svelte';
  import type { Snippet } from 'svelte';

  interface Props {
    children?: Snippet;
  }

  let { children }: Props = $props();
</script>

<div class="app-shell">
  {#if children}
    {@render children()}
  {/if}
  <Toast />
</div>

<style>
  .app-shell {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }
</style>
