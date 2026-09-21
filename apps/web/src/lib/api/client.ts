import type { ApiResponse } from '../types/campaign.types';
import { getItem } from '../utils/storage';

export class ApiError extends Error {
  code: string;
  status: number;

  constructor(code: string, message: string, status = 0) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;
  }
}

import { base } from '$app/paths';
const API_BASE = `${base}/api/v1`;
const DEFAULT_TIMEOUT_MS = 30_000;
export const TOKEN_KEY = 'admin_token';

let unauthorizedHandler: (() => void) | null = null;

/** Called when an authenticated request is rejected with 401 (expired or invalid session). */
export function setUnauthorizedHandler(handler: (() => void) | null) {
  unauthorizedHandler = handler;
}

interface RequestOptions extends RequestInit {
  timeoutMs?: number;
}

async function send(endpoint: string, options: RequestOptions, accept: string): Promise<Response> {
  const { timeoutMs = DEFAULT_TIMEOUT_MS, headers: extraHeaders, ...init } = options;

  const headers = new Headers(extraHeaders);
  headers.set('Accept', accept);
  if (init.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json');

  const token = getItem(TOKEN_KEY);
  if (token) headers.set('Authorization', `Bearer ${token}`);

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);

  try {
    return await fetch(`${API_BASE}${endpoint}`, { ...init, headers, signal: controller.signal });
  } catch (err) {
    const timedOut = err instanceof DOMException && err.name === 'AbortError';
    throw new ApiError(timedOut ? 'TIMEOUT' : 'NETWORK_ERROR', err instanceof Error ? err.message : 'Network error');
  } finally {
    clearTimeout(timer);
  }
}

async function toApiError(response: Response, endpoint: string): Promise<ApiError> {
  let code = `HTTP_${response.status}`;
  let message = response.statusText;
  try {
    const body = (await response.json()) as ApiResponse<unknown>;
    if (body.error) {
      code = body.error.code || code;
      message = body.error.message || message;
    }
  } catch {
    // Non-JSON error body (for example a proxy error page): keep the HTTP status code.
  }

  // Login failures also return 401; only an expired session should log the admin out.
  if (response.status === 401 && endpoint !== '/auth/login') {
    unauthorizedHandler?.();
  }
  return new ApiError(code, message, response.status);
}

export async function request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const response = await send(endpoint, options, 'application/json');

  if (!response.ok) throw await toApiError(response, endpoint);

  let body: ApiResponse<T>;
  try {
    body = await response.json();
  } catch {
    throw new ApiError('INVALID_RESPONSE', 'Invalid response from server', response.status);
  }
  if (!body.success) {
    throw new ApiError(body.error?.code ?? 'UNKNOWN', body.error?.message ?? '', response.status);
  }
  return body.data as T;
}

/** Download an authenticated file (the bearer token cannot be attached to a plain link). */
export async function downloadFile(endpoint: string, fallbackName: string): Promise<void> {
  const response = await send(endpoint, { timeoutMs: 120_000 }, '*/*');
  if (!response.ok) throw await toApiError(response, endpoint);

  const disposition = response.headers.get('Content-Disposition') ?? '';
  const filename = /filename="?([^";]+)"?/i.exec(disposition)?.[1] ?? fallbackName;

  const url = URL.createObjectURL(await response.blob());
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  setTimeout(() => URL.revokeObjectURL(url), 10_000);
}
