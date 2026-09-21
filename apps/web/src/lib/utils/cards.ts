import type { Card } from '../types/campaign.types';

/** The API stores anonymous senders as this literal (or an empty string). */
export function isAnonymousSender(from: string | undefined): boolean {
  return !from || from === 'Anonymous';
}

/**
 * Extracts recipient (to), message (details), and sender (from) from a card,
 * gracefully falling back to dynamic field_values if top-level fields are empty.
 */
export function getCardData(card: Partial<Card>): { to: string; message: string; from: string } {
  const f = card.fieldValues || {};

  // Recipient: check top-level card.to, then known IDs / name patterns
  let to = (card.to || '').trim();
  if (!to) {
    to = (f['emp_name'] || f['name'] || f['recipient'] || f['to'] ||
          f['field_muaw7h5a'] || f['field_muazzxxu'] || '').trim();
  }

  // Message / Details: check top-level card.message, then known IDs / message patterns
  let message = (card.message || '').trim();
  if (!message) {
    message = (f['job_title'] || f['title'] || f['message'] || f['details'] ||
               f['field_muaw8lag'] || f['field_muazzyhs'] || '').trim();
  }

  // Sender: check top-level card.from (ignoring 'Anonymous'), then known IDs / sender patterns
  let from = (!isAnonymousSender(card.from) ? (card.from || '').trim() : '');
  if (!from) {
    from = (f['sender'] || f['from'] ||
            f['field_muazy5dq'] || f['field_muazzyzg'] || '').trim();
  }

  // If still empty and there are field values, intelligently pick remaining fields
  const values = Object.entries(f).filter(([_, val]) => !!val?.trim());
  if (values.length > 0) {
    if (!to && values.length >= 1) to = values[0][1].trim();
    if (!message && values.length >= 2) message = values[1][1].trim();
    if (!from && values.length >= 3) from = values[2][1].trim();
  }

  return { to, message, from };
}

