<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { t } from '../../i18n';
  import { authStore } from '../../stores/auth.store';
  import BrandLogo from '../ui/BrandLogo.svelte';
  import LanguageSwitcher from '../ui/LanguageSwitcher.svelte';
  import ThemeSwitcher from '../ui/ThemeSwitcher.svelte';

  const links = [
    { href: '/admin/dashboard', label: 'admin.analytics' },
    { href: '/admin/cards', label: 'nav.campaigns' },
    { href: '/admin/users', label: 'nav.cards' }
  ];

  async function handleLogout() {
    await authStore.logout();
    goto('/login');
  }
</script>

<header class="admin-header">
  <div class="brand">
    <a href="/admin/dashboard" class="logo" title={$t('app.companyName')}>
      <BrandLogo />
    </a>
  </div>

  <nav class="nav-links">
    {#each links as link (link.href)}
      <a
        href={link.href}
        class="nav-link"
        class:active={page.url.pathname.startsWith(link.href)}
        aria-current={page.url.pathname.startsWith(link.href) ? 'page' : undefined}
      >
        {$t(link.label)}
      </a>
    {/each}
  </nav>

  <div class="actions">
    <ThemeSwitcher />
    <LanguageSwitcher />
    <button type="button" class="logout-btn" onclick={handleLogout}>
      {$t('nav.logout')}
    </button>
  </div>
</header>

<style>
  .admin-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 28px;
    background: var(--surface-1);
    border-bottom: 1px solid var(--color-border);
    gap: 16px;
    box-shadow: var(--shadow-sm);
  }

  .brand .logo {
    display: flex;
    align-items: center;
    gap: 14px;
    font-size: 16px;
    font-weight: 800;
    color: var(--text-main);
    text-decoration: none;
  }

  .nav-links {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .nav-link {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-muted);
    transition: color 0.15s;
  }

  .nav-link:hover {
    color: var(--color-accent);
  }

  .actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .logout-btn {
    padding: 6px 14px;
    background: transparent;
    border: 1px solid var(--color-border);
    color: var(--color-danger);
    border-radius: var(--radius-full);
    font-size: 13.5px;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.15s;
  }

  .logout-btn:hover {
    background: var(--color-danger-wash);
  }

  @media (max-width: 640px) {
    .admin-header {
      flex-direction: column;
      align-items: flex-start;
    }
  }

  .nav-link.active {
    color: var(--color-accent);
  }
</style>
