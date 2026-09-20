<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { t, locale } from '$lib/i18n';
  import { authStore, isAuthenticated } from '$lib/stores/auth.store';
  import Button from '$lib/components/ui/Button.svelte';
  import Input from '$lib/components/ui/Input.svelte';
  import LanguageSwitcher from '$lib/components/ui/LanguageSwitcher.svelte';
  import ThemeSwitcher from '$lib/components/ui/ThemeSwitcher.svelte';

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
    <div class="controls-wrap">
      <ThemeSwitcher />
      <LanguageSwitcher />
    </div>
  </div>

  <div class="login-card">
    <div class="card-head">
      <div class="brand-badge-login">
        <img
          src="/images/brand/aljuf-ar-tight.png"
          alt="Abdul Latif Jameel Finance"
          class="login-brand-logo light-only"
        />
        <img
          src="/images/brand/aljuf-ar-white-tight.png"
          alt="Abdul Latif Jameel Finance"
          class="login-brand-logo dark-only"
        />
      </div>
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

  .controls-wrap {
    display: flex;
    align-items: center;
    gap: 8px;
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
    max-width: 420px;
    background: var(--surface-card);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-xl);
    padding: 36px 30px;
    display: flex;
    flex-direction: column;
    gap: 24px;
    box-shadow: var(--shadow-card);
  }

  .card-head {
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
  }

  .brand-badge-login {
    margin-bottom: 8px;
  }

  .login-brand-logo {
    height: 54px;
    width: auto;
    object-fit: contain;
  }

  :global([data-theme="light"]) .dark-only {
    display: none !important;
  }

  :global([data-theme="light"]) .light-only {
    display: block !important;
  }

  :global([data-theme="dark"]) .light-only,
  :global(:root:not([data-theme="light"])) .light-only {
    display: none !important;
  }

  :global([data-theme="dark"]) .dark-only,
  :global(:root:not([data-theme="light"])) .dark-only {
    display: block !important;
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
