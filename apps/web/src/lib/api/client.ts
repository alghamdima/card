import type { ApiResponse } from '../types/campaign.types';

export class ApiError extends Error {
  code: string;

  constructor(code: string, message: string) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
  }
}

const API_BASE = '/api/v1';

export async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE}${endpoint}`;

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    Accept: 'application/json',
    ...(options.headers as Record<string, string>)
  };

  const token = typeof window !== 'undefined' ? localStorage.getItem('admin_token') : null;
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  let response: Response;
  try {
    response = await fetch(url, {
      ...options,
      headers
    });
  } catch (err: any) {
    throw new ApiError('NETWORK_ERROR', err?.message || 'Network connection failed');
  }

  let data: ApiResponse<T>;
  try {
    data = await response.json();
  } catch {
    throw new ApiError('INVALID_RESPONSE', 'Invalid response from server');
  }

  if (!response.ok || !data.success) {
    const errCode = data.error?.code || `HTTP_${response.status}`;
    const errMsg = data.error?.message || response.statusText;
    throw new ApiError(errCode, errMsg);
  }

  return data.data as T;
}
