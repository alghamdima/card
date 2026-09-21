import { writable } from 'svelte/store';

export interface ToastMessage {
  id: number;
  text: string;
  type?: 'success' | 'error' | 'info';
}

export const toasts = writable<ToastMessage[]>([]);

let nextId = 0;

export function showToast(text: string, type: 'success' | 'error' | 'info' = 'success', duration = 3000) {
  // A counter, not Date.now(): two toasts in the same millisecond used to share a key.
  const id = ++nextId;
  toasts.update((all) => [...all, { id, text, type }]);

  setTimeout(() => {
    toasts.update((all) => all.filter((t) => t.id !== id));
  }, duration);
}
