<script lang="ts">
  import '../lib/styles/app.css';
  import Toast from '../lib/components/ui/Toast.svelte';
  import { onMount } from 'svelte';
  import { applyTheme } from '../lib/stores/theme.store';
  import type { Snippet } from 'svelte';

  interface Props {
    children?: Snippet;
  }

  let { children }: Props = $props();

  onMount(() => {
    // ensure theme initialized on client
    const saved = localStorage.getItem('user_theme') || 'system';
    applyTheme(saved as any);
  });
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
