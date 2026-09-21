import { goto } from '$app/navigation';
import { get, writable } from 'svelte/store';
import { authApi } from '../api/auth';
import { ApiError, setUnauthorizedHandler, TOKEN_KEY } from '../api/client';
import { getItem, removeItem, setItem } from '../utils/storage';

const initialToken = getItem(TOKEN_KEY);

export const adminToken = writable<string | null>(initialToken);
export const isAuthenticated = writable<boolean>(!!initialToken);

function clearSession() {
  removeItem(TOKEN_KEY);
  adminToken.set(null);
  isAuthenticated.set(false);
}

// An expired or revoked session mid-use: drop it and send the admin back to the login page.
setUnauthorizedHandler(() => {
  if (!get(isAuthenticated)) return;
  clearSession();
  goto('/login');
});

export const authStore = {
  login: async (password: string) => {
    const res = await authApi.login(password);
    if (res.token) {
      setItem(TOKEN_KEY, res.token);
      adminToken.set(res.token);
      isAuthenticated.set(true);
    }
    return res;
  },

  logout: async () => {
    try {
      await authApi.logout();
    } catch {
      // The local session is cleared regardless of the server response.
    }
    clearSession();
  },

  checkAuth: async () => {
    if (!getItem(TOKEN_KEY)) {
      isAuthenticated.set(false);
      return false;
    }
    try {
      await authApi.me();
      isAuthenticated.set(true);
      return true;
    } catch (e) {
      // Only a rejected token ends the session. A network error, timeout or aborted request (for example a page
      // navigation) says nothing about the token, and wiping it would log the admin out for no reason.
      if (e instanceof ApiError && e.status === 401) {
        clearSession();
        return false;
      }
      isAuthenticated.set(true);
      return true;
    }
  }
};
