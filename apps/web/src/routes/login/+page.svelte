<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { t } from '$lib/i18n';
  import { authStore, isAuthenticated } from '$lib/stores/auth.store';
  import Button from '$lib/components/ui/Button.svelte';
  import Input from '$lib/components/ui/Input.svelte';
  import LanguageSwitcher from '$lib/components/ui/LanguageSwitcher.svelte';

  let password = $state('');
  let loading = $state(false);
  let errorMsg = $state('');

  onMount(async () => {
    const ok = await authStore.checkAuth();
    if (ok) {
      goto('/admin/dashboard');
    }
  });

  async function handleLogin() {
    if (!password) {
      errorMsg = $t('errors.VALIDATION_ERROR');
      return;
    }

    loading = true;
    errorMsg = '';
    try {
      await authStore.login(password);
      goto('/admin/dashboard');
    } catch (e: any) {
      errorMsg = e.code === 'INVALID_CREDENTIALS' 
        ? $t('login.invalidPassword')
        : (e.message || $t('app.error'));
    } finally {
      loading = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      handleLogin();
    }
  }
</script>

<svelte:head>
  <title>{$t('login.title')} | {$t('app.title')}</title>
</svelte:head>

<div class="login-page">
  <div class="top-bar">
    <a href="/" class="back-link">← {$t('app.back')}</a>
    <LanguageSwitcher />
  </div>

  <div class="login-card">
    <div class="card-head">
      <span class="icon">🔒</span>
      <h2>{$t('login.title')}</h2>
      <p>{$t('login.subtitle')}</p>
    </div>

    <form class="login-form" onsubmit={(e) => { e.preventDefault(); handleLogin(); }}>
      <Input
        type="password"
        label={$t('login.password')}
        placeholder={$t('login.passwordPlaceholder')}
        value={password}
        error={errorMsg}
        oninput={(e) => (password = (e.target as HTMLInputElement).value)}
      />

      <Button
        type="submit"
        variant="primary"
        {loading}
        disabled={loading}
      >
        {$t('login.submit')}
      </Button>
    </form>
  </div>
</div>

<style>
  .login-page {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 24px 16px;
    background: var(--bg-app);
  }

  .top-bar {
    position: absolute;
    top: 24px;
    left: 24px;
    right: 24px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .back-link {
    color: var(--text-muted);
    font-size: 14px;
    font-weight: 700;
  }

  .back-link:hover {
    color: var(--color-accent);
  }

  .login-card {
    width: 100%;
    max-width: 400px;
    background: var(--surface-card);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    padding: 32px 28px;
    display: flex;
    flex-direction: column;
    gap: 24px;
    box-shadow: var(--shadow-card);
  }

  .card-head {
    text-align: center;
  }

  .icon {
    font-size: 32px;
    display: inline-block;
    margin-bottom: 8px;
  }

  h2 {
    font-size: 20px;
    font-weight: 800;
    color: var(--text-main);
  }

  p {
    font-size: 13.5px;
    color: var(--text-muted);
    margin-top: 4px;
  }

  .login-form {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .login-form :global(button) {
    width: 100%;
    margin-top: 8px;
  }
</style>
