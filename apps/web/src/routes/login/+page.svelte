<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { t, locale, translateError } from '$lib/i18n';
  import { authStore } from '$lib/stores/auth.store';
  import BrandLogo from '$lib/components/ui/BrandLogo.svelte';
  import Button from '$lib/components/ui/Button.svelte';
  import Input from '$lib/components/ui/Input.svelte';
  import LanguageSwitcher from '$lib/components/ui/LanguageSwitcher.svelte';
  import ThemeSwitcher from '$lib/components/ui/ThemeSwitcher.svelte';

  let password = $state('');
  let loading = $state(false);
  let errorMsg = $state('');

  onMount(async () => {
    if (await authStore.checkAuth()) {
      goto('/admin/dashboard');
    }
  });

  async function handleLogin() {
    if (loading) return;
    if (!password) {
      errorMsg = $t('errors.VALIDATION_ERROR');
      return;
    }

    loading = true;
    errorMsg = '';
    try {
      await authStore.login(password);
      goto('/admin/dashboard');
    } catch (e) {
      errorMsg = (e as { code?: string })?.code === 'INVALID_CREDENTIALS' ? $t('login.invalidPassword') : translateError(e);
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>{$t('login.title')} | {$t('app.title')}</title>
</svelte:head>

<div class="login-page">
  <div class="top-bar">
    <a href="/" class="back-link">{$locale === 'ar' ? '→' : '←'} {$t('app.back')}</a>
    <div class="controls-wrap">
      <ThemeSwitcher />
      <LanguageSwitcher />
    </div>
  </div>

  <div class="login-card">
    <div class="card-head">
      <div class="brand-badge-login">
        <BrandLogo height={54} />
      </div>
      <h2>{$t('login.title')}</h2>
      <p>{$t('login.subtitle')}</p>
    </div>

    <form
      class="login-form"
      onsubmit={(e) => {
        e.preventDefault();
        handleLogin();
      }}
    >
      <Input
        type="password"
        autocomplete="current-password"
        label={$t('login.password')}
        placeholder={$t('login.passwordPlaceholder')}
        value={password}
        error={errorMsg}
        oninput={(e) => (password = (e.target as HTMLInputElement).value)}
      />

      <Button type="submit" variant="primary" {loading} disabled={loading}>
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

  .brand-badge-login {
    margin-bottom: 8px;
  }
</style>
