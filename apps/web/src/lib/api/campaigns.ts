import { downloadFile, request } from './client';
import type {
  Campaign,
  CampaignAnalytics,
  CampaignSummary,
  CardsPage,
  DashboardStats,
  PublicCampaignSummary
} from '../types/campaign.types';

export interface SubmitCardInput {
  from?: string;
  to?: string;
  message?: string;
  heading?: string;
  lang?: string;
  fieldValues?: Record<string, string>;
  device?: string;
}

export interface CardsQuery {
  limit?: number;
  offset?: number;
  q?: string;
}

function toQuery(params: CardsQuery): string {
  const qs = new URLSearchParams();
  if (params.limit !== undefined) qs.set('limit', String(params.limit));
  if (params.offset !== undefined) qs.set('offset', String(params.offset));
  if (params.q) qs.set('q', params.q);
  const text = qs.toString();
  return text ? `?${text}` : '';
}

// Slugs come from the URL bar on public pages, so they are always encoded.
const seg = encodeURIComponent;

export const campaignsApi = {
  // Public endpoints
  listPublic: () => request<{ campaigns: PublicCampaignSummary[] }>('/campaigns'),

  getPublic: (slug: string) => request<Campaign>(`/campaigns/${seg(slug)}`),

  submitCard: (slug: string, data: SubmitCardInput) =>
    request<{ saved: boolean; id: number }>(`/campaigns/${seg(slug)}/cards`, {
      method: 'POST',
      body: JSON.stringify(data)
    }),

  // Admin endpoints
  listAdmin: () => request<{ campaigns: CampaignSummary[] }>('/admin/campaigns'),

  getAdmin: (slug: string) => request<Campaign>(`/admin/campaigns/${seg(slug)}`),

  // Uploads carry base64 artwork, so they get a longer timeout than regular calls.
  create: (data: Partial<Campaign>) =>
    request<Campaign>('/admin/campaigns', { method: 'POST', body: JSON.stringify(data), timeoutMs: 120_000 }),

  update: (slug: string, data: Partial<Campaign>) =>
    request<Campaign>(`/admin/campaigns/${seg(slug)}`, {
      method: 'PUT',
      body: JSON.stringify(data),
      timeoutMs: 120_000
    }),

  delete: (slug: string) => request<{ deleted: boolean }>(`/admin/campaigns/${seg(slug)}`, { method: 'DELETE' }),

  getCampaignCards: (slug: string, query: CardsQuery = {}) =>
    request<CardsPage>(`/admin/campaigns/${seg(slug)}/cards${toQuery(query)}`),

  getAllCards: (query: CardsQuery = {}) => request<CardsPage>(`/admin/cards${toQuery(query)}`),

  getDashboardStats: () => request<DashboardStats>('/admin/dashboard'),

  getAnalyticsOverview: () => request<{ campaigns: CampaignAnalytics[] }>('/admin/analytics/overview'),

  downloadExport: (slug: string) => downloadFile(`/admin/campaigns/${seg(slug)}/export`, `${slug}-cards.csv`)
};
