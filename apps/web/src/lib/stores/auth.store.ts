import { writable } from 'svelte/store';
import { authApi } from '../api/auth';

const initialToken = typeof window !== 'undefined' ? localStorage.getItem('admin_token') : null;

export const adminToken = writable<string | null>(initialToken);
export const isAuthenticated = writable<boolean>(!!initialToken);

export const authStore = {
  login: async (password: string) => {
    const res = await authApi.login(password);
    if (res.token) {
      if (typeof window !== 'undefined') {
        localStorage.setItem('admin_token', res.token);
      }
      adminToken.set(res.token);
      isAuthenticated.set(true);
    }
    return res;
  },

  logout: async () => {
    try {
      await authApi.logout();
    } catch {
      // ignore
    }
    if (typeof window !== 'undefined') {
      localStorage.removeItem('admin_token');
    }
    adminToken.set(null);
    isAuthenticated.set(false);
  },

  checkAuth: async () => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('admin_token') : null;
    if (!token) {
      isAuthenticated.set(false);
      return false;
    }
    try {
      await authApi.me();
      isAuthenticated.set(true);
      return true;
    } catch {
      if (typeof window !== 'undefined') {
        localStorage.removeItem('admin_token');
      }
      adminToken.set(null);
      isAuthenticated.set(false);
      return false;
    }
  }
};
