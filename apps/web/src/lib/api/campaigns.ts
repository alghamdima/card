import { request } from './client';
import type { Campaign, CampaignSummary, Card, DashboardStats } from '../types/campaign.types';

export const campaignsApi = {
  // Public endpoints
  listPublic: async (): Promise<{ campaigns: CampaignSummary[] }> => {
    return request<{ campaigns: CampaignSummary[] }>('/campaigns');
  },

  getPublic: async (slug: string): Promise<Campaign> => {
    return request<Campaign>(`/campaigns/${slug}`);
  },

  submitCard: async (slug: string, data: { from: string; to: string; message: string; heading?: string; device?: string }): Promise<{ saved: boolean; id: number }> => {
    return request<{ saved: boolean; id: number }>(`/campaigns/${slug}/cards`, {
      method: 'POST',
      body: JSON.stringify(data)
    });
  },

  // Admin endpoints
  listAdmin: async (): Promise<{ campaigns: CampaignSummary[] }> => {
    return request<{ campaigns: CampaignSummary[] }>('/admin/campaigns');
  },

  getAdmin: async (slug: string): Promise<Campaign> => {
    return request<Campaign>(`/admin/campaigns/${slug}`);
  },

  create: async (data: Partial<Campaign>): Promise<Campaign> => {
    return request<Campaign>('/admin/campaigns', {
      method: 'POST',
      body: JSON.stringify(data)
    });
  },

  update: async (slug: string, data: Partial<Campaign>): Promise<Campaign> => {
    return request<Campaign>(`/admin/campaigns/${slug}`, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
  },

  delete: async (slug: string): Promise<{ deleted: boolean }> => {
    return request<{ deleted: boolean }>(`/admin/campaigns/${slug}`, {
      method: 'DELETE'
    });
  },

  getCampaignCards: async (slug: string): Promise<{ slug: string; cards: Card[]; total: number }> => {
    return request<{ slug: string; cards: Card[]; total: number }>(`/admin/campaigns/${slug}/cards`);
  },

  getAllCards: async (limit = 50, offset = 0): Promise<{ cards: Card[]; total: number }> => {
    return request<{ cards: Card[]; total: number }>(`/admin/cards?limit=${limit}&offset=${offset}`);
  },

  getDashboardStats: async (): Promise<DashboardStats> => {
    return request<DashboardStats>('/admin/dashboard');
  }
};
