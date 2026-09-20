import { writable } from 'svelte/store';

export interface ToastMessage {
  id: number;
  text: string;
  type?: 'success' | 'error' | 'info';
}

export const toasts = writable<ToastMessage[]>([]);

export function showToast(text: string, type: 'success' | 'error' | 'info' = 'success', duration = 3000) {
  const id = Date.now();
  toasts.update((all) => [...all, { id, text, type }]);

  setTimeout(() => {
    toasts.update((all) => all.filter((t) => t.id !== id));
  }, duration);
}
