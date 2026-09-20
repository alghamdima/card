<script lang="ts">
  import { t, locale } from '../../i18n';
  import { authStore } from '../../stores/auth.store';
  import LanguageSwitcher from '../ui/LanguageSwitcher.svelte';
  import ThemeSwitcher from '../ui/ThemeSwitcher.svelte';

  function handleLogout() {
    authStore.logout().then(() => {
      window.location.href = '/login';
    });
  }
</script>

<header class="admin-header">
  <div class="brand">
    <a href="/admin/dashboard" class="logo">
      <img
        src={$locale === 'en' ? '/images/brand/aljuf-en.png' : '/images/brand/aljuf-ar.png'}
        alt="ALJ Finance"
        class="admin-brand-logo"
      />
      <div class="brand-text">
        <span class="product-name">Cards</span>
        <span class="name">{$t('admin.title')}</span>
      </div>
    </a>
  </div>

  <nav class="nav-links">
    <a href="/admin/dashboard" class="nav-link">{$t('admin.analytics')}</a>
    <a href="/admin/cards" class="nav-link">{$t('nav.campaigns')}</a>
    <a href="/admin/users" class="nav-link">{$t('nav.cards')}</a>
  </nav>

  <div class="actions">
    <ThemeSwitcher />
    <LanguageSwitcher />
    <button class="logout-btn" onclick={handleLogout}>
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

  .admin-brand-logo {
    height: 48px;
    width: auto;
    object-fit: contain;
  }

  .brand-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .product-name {
    font-size: 15px;
    font-weight: 800;
    color: var(--color-accent);
    letter-spacing: 0.5px;
  }

  .name {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-muted);
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
</style>
