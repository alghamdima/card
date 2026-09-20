import { request } from './client';

export const authApi = {
  login: async (password: string): Promise<{ token: string; user: { role: string } }> => {
    return request<{ token: string; user: { role: string } }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ password })
    });
  },

  logout: async (): Promise<{ loggedOut: boolean }> => {
    return request<{ loggedOut: boolean }>('/auth/logout', {
      method: 'POST'
    });
  },

  me: async (): Promise<{ role: string; authenticated: boolean }> => {
    return request<{ role: string; authenticated: boolean }>('/auth/me');
  }
};
