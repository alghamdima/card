/** The API stores anonymous senders as this literal (or an empty string). */
export function isAnonymousSender(from: string | undefined): boolean {
  return !from || from === 'Anonymous';
}
