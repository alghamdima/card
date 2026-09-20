<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { t } from '$lib/i18n';
  import { authStore, isAuthenticated } from '$lib/stores/auth.store';
  import AdminHeader from '$lib/components/admin/AdminHeader.svelte';
  import LoadingState from '$lib/components/ui/LoadingState.svelte';
  import type { Snippet } from 'svelte';

  interface Props {
    children?: Snippet;
  }

  let { children }: Props = $props();
  let checking = $state(true);

  onMount(async () => {
    const ok = await authStore.checkAuth();
    if (!ok) {
      goto('/login');
    }
    checking = false;
  });
</script>

{#if checking}
  <LoadingState />
{:else if $isAuthenticated}
  <div class="admin-layout">
    <AdminHeader />
    <main class="admin-main">
      {#if children}
        {@render children()}
      {/if}
    </main>
  </div>
{/if}

<style>
  .admin-layout {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    background: var(--bg-app);
  }

  .admin-main {
    flex: 1;
    max-width: 1080px;
    width: 100%;
    margin: 0 auto;
    padding: 24px 16px 48px;
  }
</style>
