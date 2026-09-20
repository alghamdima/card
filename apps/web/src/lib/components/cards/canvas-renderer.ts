import type { BoxesConfig, BoxItem } from '../../types/campaign.types';

export const CANVAS_W = 1080;
export const CANVAS_H = 1350;
const SEP = ' \u2014 ';
const SECTION_GAP = 0.7;

let fontsLoadedPromise: Promise<void> | null = null;

export async function ensureFontsLoaded(): Promise<void> {
  if (typeof window === 'undefined') return;
  if (fontsLoadedPromise) return fontsLoadedPromise;

  fontsLoadedPromise = (async () => {
    try {
      const fonts = [
        new FontFace('LumaSemiBold', 'url(/fonts/Luma-SemiBold.ttf)'),
        new FontFace('LumaRegular', 'url(/fonts/Luma-Regular.ttf)'),
        new FontFace('KarbonBold', 'url(/fonts/Karbon-Bold.ttf)'),
        new FontFace('KarbonRegular', 'url(/fonts/Karbon-Regular.ttf)')
      ];

      const loaded = await Promise.all(fonts.map((f) => f.load()));
      loaded.forEach((f) => document.fonts.add(f));
    } catch (e) {
      console.warn('Custom font load notice, falling back to system fonts:', e);
    }
  })();

  return fontsLoadedPromise;
}

export function isArabic(text: string): boolean {
  return /[\u0600-\u06FF]/.test(text || '');
}

function fontFor(text: string, size: number, weight: 'bold' | 'regular'): string {
  const ar = isArabic(text);
  const fam =
    weight === 'bold'
      ? ar
        ? 'LumaSemiBold, Tajawal, sans-serif'
        : 'KarbonBold, Instrument Sans, sans-serif'
      : ar
      ? 'LumaRegular, Tajawal, sans-serif'
      : 'KarbonRegular, Instrument Sans, sans-serif';
  return `${size}px "${fam}"`;
}

function wrapWithPrefix(ctx: CanvasRenderingContext2D, text: string, maxW: number, firstLineIndent: number): string[] {
  const words = String(text).split(/\s+/).filter(Boolean);
  const lines: string[] = [];
  let cur = '';
  let limit = maxW - firstLineIndent;

  for (let i = 0; i < words.length; i++) {
    const w = words[i];
    const test = cur ? cur + ' ' + w : w;
    if (ctx.measureText(test).width <= limit) {
      cur = test;
      continue;
    }
    if (cur) {
      lines.push(cur);
      cur = '';
      limit = maxW;
    }
    if (ctx.measureText(w).width > limit) {
      let piece = '';
      for (let j = 0; j < w.length; j++) {
        const t2 = piece + w[j];
        if (ctx.measureText(t2).width > limit && piece) {
          lines.push(piece);
          piece = w[j];
          limit = maxW;
        } else {
          piece = t2;
        }
      }
      cur = piece;
    } else {
      cur = w;
    }
  }
  if (cur) lines.push(cur);
  return lines.length ? lines : [''];
}

function wrapPlain(ctx: CanvasRenderingContext2D, text: string, maxW: number): string[] {
  return wrapWithPrefix(ctx, text, maxW, 0);
}

function drawBox(ctx: CanvasRenderingContext2D, text: string, box: BoxItem) {
  if (!text || !box) return;
  const maxSize = Number(box.size) || 40;
  const minSize = Math.max(9, Math.round(maxSize * 0.35));
  const weight = box.weight === 'regular' ? 'regular' : 'bold';
  const boxH = box.bottom - box.top;
  let chosen: { lines: string[]; lh: number; total: number; size: number } | null = null;

  for (let size = maxSize; size >= minSize; size--) {
    ctx.font = fontFor(text, size, weight);
    const lines = wrapPlain(ctx, text, box.maxW);
    const lh = size * 1.3;
    const total = lines.length * lh;
    let widest = 0;
    lines.forEach((l) => {
      widest = Math.max(widest, ctx.measureText(l).width);
    });

    if ((total <= boxH && widest <= box.maxW) || size === minSize) {
      chosen = { lines, lh, total, size };
      break;
    }
  }

  if (!chosen) return;
  ctx.font = fontFor(text, chosen.size, weight);
  ctx.fillStyle = box.color || '#FFFFFF';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';

  let y = box.top + (boxH - chosen.total) / 2;
  chosen.lines.forEach((l) => {
    ctx.fillText(l, box.centerX, y + chosen.lh / 2);
    y += chosen.lh;
  });
}

function drawMessageSections(ctx: CanvasRenderingContext2D, sections: { heading?: string; text: string }[], box: BoxItem) {
  const maxSize = Number(box.size) || 40;
  const minSize = Math.max(9, Math.round(maxSize * 0.32));
  const boxH = box.bottom - box.top;
  let chosenPlan: {
    plan: { head: string; indent: number; lines: string[]; size: number }[];
    total: number;
    lh: number;
    gap: number;
    size: number;
  } | null = null;

  for (let size = maxSize; size >= minSize; size--) {
    const lh = size * 1.32;
    const gap = size * SECTION_GAP;
    const plan: { head: string; indent: number; lines: string[]; size: number }[] = [];
    let total = 0;
    let ok = true;

    for (let i = 0; i < sections.length; i++) {
      const head = (sections[i].heading || '').trim();
      const body = (sections[i].text || '').trim();
      if (!head && !body) continue;

      let indent = 0;
      if (head) {
        ctx.font = fontFor(head, size, 'bold');
        indent = ctx.measureText(head + SEP).width;
      }
      ctx.font = fontFor(body, size, 'regular');
      const lines = body ? wrapWithPrefix(ctx, body, box.maxW, indent) : [''];

      if (indent > box.maxW * 0.9) {
        ok = false;
        break;
      }

      plan.push({ head, indent, lines, size });
      total += lines.length * lh;
      if (plan.length > 1) total += gap;
    }

    if (!plan.length) return;
    if ((ok && total <= boxH) || size === minSize) {
      chosenPlan = { plan, total, lh, gap, size };
      break;
    }
  }

  if (!chosenPlan) return;
  const headColor = box.headColor || box.color || '#FFCD00';
  const bodyColor = box.bodyColor || box.color || '#FFFFFF';
  ctx.textBaseline = 'middle';
  let y = box.top + (box.bottom - box.top - chosenPlan.total) / 2;

  chosenPlan.plan.forEach((sec, si) => {
    sec.lines.forEach((line, li) => {
      const cy = y + chosenPlan!.lh / 2;
      if (li === 0 && sec.head) {
        ctx.font = fontFor(line, chosenPlan!.size, 'regular');
        const restW = ctx.measureText(line).width;
        const startX = box.centerX - (sec.indent + restW) / 2;

        ctx.textAlign = 'left';
        ctx.font = fontFor(sec.head, chosenPlan!.size, 'bold');
        ctx.fillStyle = headColor;
        ctx.fillText(sec.head + SEP, startX, cy);

        ctx.font = fontFor(line, chosenPlan!.size, 'regular');
        ctx.fillStyle = bodyColor;
        ctx.fillText(line, startX + sec.indent, cy);
        ctx.textAlign = 'center';
      } else {
        ctx.textAlign = 'center';
        ctx.font = fontFor(line, chosenPlan!.size, 'regular');
        ctx.fillStyle = bodyColor;
        ctx.fillText(line, box.centerX, cy);
      }
      y += chosenPlan!.lh;
    });
    if (si < chosenPlan!.plan.length - 1) y += chosenPlan!.gap;
  });
}

export function drawDynamicField(
  ctx: CanvasRenderingContext2D,
  text: string,
  field: import('../../types/campaign.types').TextFieldConfig
) {
  if (!text || !field) return;
  const rawText = String(text).trim();
  if (!rawText) return;

  const maxSize = Number(field.fontSize) || 40;
  const minSize = Math.max(10, Math.round(maxSize * 0.4));
  const weight = field.weight === 'regular' ? 'regular' : 'bold';
  const boxW = field.width || 400;
  const boxH = field.height || 80;
  const align = field.align || 'center';

  let chosen: { lines: string[]; lh: number; total: number; size: number } | null = null;

  for (let size = maxSize; size >= minSize; size--) {
    ctx.font = fontFor(rawText, size, weight);
    const lines = wrapPlain(ctx, rawText, boxW);
    const lh = size * 1.3;
    const total = lines.length * lh;
    let widest = 0;
    lines.forEach((l) => {
      widest = Math.max(widest, ctx.measureText(l).width);
    });

    if ((total <= boxH && widest <= boxW) || size === minSize) {
      chosen = { lines, lh, total, size };
      break;
    }
  }

  if (!chosen) return;
  ctx.font = fontFor(rawText, chosen.size, weight);
  ctx.fillStyle = field.color || '#FFFFFF';
  ctx.textBaseline = 'middle';

  let startX = field.x + boxW / 2;
  if (align === 'left') {
    ctx.textAlign = 'left';
    startX = field.x;
  } else if (align === 'right') {
    ctx.textAlign = 'right';
    startX = field.x + boxW;
  } else {
    ctx.textAlign = 'center';
  }

  let y = field.y + (boxH - chosen.total) / 2;
  chosen.lines.forEach((line) => {
    ctx.fillText(line, startX, y + chosen!.lh / 2);
    y += chosen!.lh;
  });
}

export function renderCard(
  canvas: HTMLCanvasElement,
  img: HTMLImageElement | null,
  boxes: BoxesConfig | null | undefined,
  data: { to?: string; from?: string; message?: string; heading?: string; fieldValues?: Record<string, any> },
  dynamicFields?: import('../../types/campaign.types').TextFieldConfig[]
) {
  const ctx = canvas.getContext('2d');
  if (!ctx) return;

  ctx.clearRect(0, 0, CANVAS_W, CANVAS_H);

  if (img && img.complete && img.naturalWidth) {
    ctx.drawImage(img, 0, 0, CANVAS_W, CANVAS_H);
  }

  // If dynamic fields are provided, render them
  if (dynamicFields && dynamicFields.length > 0) {
    const sorted = [...dynamicFields].sort((a, b) => (a.order || 0) - (b.order || 0));
    for (const f of sorted) {
      const val = data.fieldValues?.[f.id] ?? data.fieldValues?.[f.name] ?? (f.id === 'emp_name' || f.id === 'name' ? data.to : f.id === 'message' ? data.message : '');
      if (val) {
        drawDynamicField(ctx, String(val), f);
      }
    }
    return;
  }

  // Fallback to legacy boxes config
  if (!boxes) return;

  if (data.to && boxes.to) {
    drawBox(ctx, data.to.trim(), boxes.to);
  }

  if (data.message && boxes.message) {
    drawMessageSections(ctx, [{ heading: data.heading, text: data.message }], boxes.message);
  }

  if (data.from && boxes.from) {
    drawBox(ctx, data.from.trim(), boxes.from);
  }
}
