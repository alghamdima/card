<script lang="ts">
  import { t } from '../../i18n';
  import { CANVAS_W, CANVAS_H, renderCard, ensureFontsLoaded } from '../cards/canvas-renderer';
  import type { TextFieldConfig } from '../../types/campaign.types';

  interface Props {
    imageSrc: string;
    fields: TextFieldConfig[];
    activeFieldId: string;
    sampleValues: Record<string, string>;
    onfieldchange: (fields: TextFieldConfig[]) => void;
    onselectfield: (id: string) => void;
  }

  let { imageSrc, fields, activeFieldId, sampleValues, onfieldchange, onselectfield }: Props = $props();

  const HANDLE_HIT_RADIUS = 30;
  const MIN_WIDTH = 100;
  const MIN_HEIGHT = 40;

  type Corner = 'tl' | 'tr' | 'bl' | 'br';

  let interactiveCanvasEl: HTMLCanvasElement;
  let previewCanvasEl: HTMLCanvasElement;
  let bgImg = $state<HTMLImageElement | null>(null);

  // Drag / resize interaction state (plain variables: they never drive rendering).
  let mode: 'idle' | 'drag' | 'resize' = 'idle';
  let resizeCorner: Corner | null = null;
  let targetId = ''; // the field being dragged or resized
  let dragStart = { x: 0, y: 0 };
  let initial = { x: 0, y: 0, w: 0, h: 0 };

  // A stale activeFieldId (for example after switching language tabs) falls back to the first field.
  let selectedId = $derived(fields.some((f) => f.id === activeFieldId) ? activeFieldId : (fields[0]?.id ?? ''));

  function redrawAll() {
    if (!interactiveCanvasEl || !previewCanvasEl || !bgImg) return;

    // 1. Live preview: exactly what an employee will get
    renderCard(previewCanvasEl, bgImg, null, { fieldValues: sampleValues }, fields);

    // 2. Interactive canvas: artwork + field boxes + resize handles
    const ctx = interactiveCanvasEl.getContext('2d');
    if (!ctx) return;

    ctx.clearRect(0, 0, CANVAS_W, CANVAS_H);
    ctx.drawImage(bgImg, 0, 0, CANVAS_W, CANVAS_H);

    for (const field of fields) {
      const isActive = field.id === selectedId;
      ctx.save();
      ctx.strokeStyle = isActive ? '#FFCD00' : 'rgba(255, 205, 0, 0.4)';
      ctx.lineWidth = isActive ? 5 : 2;
      ctx.fillStyle = isActive ? 'rgba(255, 205, 0, 0.12)' : 'rgba(0, 0, 0, 0.2)';

      ctx.beginPath();
      ctx.roundRect(field.x, field.y, field.width, field.height, Math.min(20, field.height / 2));
      ctx.fill();
      ctx.stroke();

      // Field name above the box
      ctx.font = '22px sans-serif';
      ctx.fillStyle = '#FFFFFF';
      ctx.textAlign = 'start';
      ctx.fillText(field.label || field.name, field.x + 12, Math.max(28, field.y - 8));

      // Sample text inside the box
      ctx.font = `${field.fontSize || 38}px sans-serif`;
      ctx.fillStyle = field.color || '#FFFFFF';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(sampleValues[field.id] || field.label, field.x + field.width / 2, field.y + field.height / 2);

      if (isActive) {
        ctx.fillStyle = '#FFCD00';
        ctx.strokeStyle = '#2B0A3D';
        ctx.lineWidth = 4;
        for (const corner of cornersOf(field)) {
          ctx.beginPath();
          ctx.arc(corner.x, corner.y, 12, 0, Math.PI * 2);
          ctx.fill();
          ctx.stroke();
        }
      }
      ctx.restore();
    }
  }

  function cornersOf(f: TextFieldConfig): { name: Corner; x: number; y: number }[] {
    return [
      { name: 'tl', x: f.x, y: f.y },
      { name: 'tr', x: f.x + f.width, y: f.y },
      { name: 'bl', x: f.x, y: f.y + f.height },
      { name: 'br', x: f.x + f.width, y: f.y + f.height }
    ];
  }

  function contains(f: TextFieldConfig, p: { x: number; y: number }): boolean {
    return p.x >= f.x && p.x <= f.x + f.width && p.y >= f.y && p.y <= f.y + f.height;
  }

  // Redraw when fields, samples, selection or the background change.
  $effect(() => {
    if (bgImg) redrawAll();
  });

  // Load the background; `cancelled` drops a load that finishes after the image was replaced.
  $effect(() => {
    const src = imageSrc;
    if (!src) return;

    let cancelled = false;
    ensureFontsLoaded().then(() => {
      if (cancelled) return;
      const img = new Image();
      img.onload = () => {
        if (!cancelled) bgImg = img;
      };
      img.src = src;
    });
    return () => {
      cancelled = true;
    };
  });

  function toCanvasCoords(e: PointerEvent): { x: number; y: number } {
    const rect = interactiveCanvasEl.getBoundingClientRect();
    return {
      x: ((e.clientX - rect.left) * CANVAS_W) / rect.width,
      y: ((e.clientY - rect.top) * CANVAS_H) / rect.height
    };
  }

  function beginDrag(field: TextFieldConfig, p: { x: number; y: number }) {
    mode = 'drag';
    targetId = field.id;
    dragStart = p;
    initial = { x: field.x, y: field.y, w: field.width, h: field.height };
  }

  function handlePointerDown(e: PointerEvent) {
    e.preventDefault();
    const p = toCanvasCoords(e);
    const active = fields.find((f) => f.id === selectedId);

    if (active) {
      const corner = cornersOf(active).find((c) => Math.hypot(p.x - c.x, p.y - c.y) <= HANDLE_HIT_RADIUS);
      if (corner) {
        mode = 'resize';
        targetId = active.id;
        resizeCorner = corner.name;
        dragStart = p;
        initial = { x: active.x, y: active.y, w: active.width, h: active.height };
        interactiveCanvasEl.setPointerCapture(e.pointerId);
        return;
      }
      if (contains(active, p)) {
        beginDrag(active, p);
        interactiveCanvasEl.setPointerCapture(e.pointerId);
        return;
      }
    }

    // Clicking another field selects it and starts dragging it.
    const hit = fields.find((f) => contains(f, p));
    if (hit) {
      onselectfield(hit.id);
      beginDrag(hit, p);
      interactiveCanvasEl.setPointerCapture(e.pointerId);
    }
  }

  function clamp(value: number, min: number, max: number): number {
    return Math.max(min, Math.min(max, value));
  }

  function handlePointerMove(e: PointerEvent) {
    if (mode === 'idle') return;
    const p = toCanvasCoords(e);
    const dx = p.x - dragStart.x;
    const dy = p.y - dragStart.y;

    const patch = mode === 'drag' ? dragPatch(dx, dy) : resizePatch(dx, dy);
    onfieldchange(fields.map((f) => (f.id === targetId ? { ...f, ...patch } : f)));
  }

  function dragPatch(dx: number, dy: number) {
    return {
      x: Math.round(clamp(initial.x + dx, 0, CANVAS_W - initial.w)),
      y: Math.round(clamp(initial.y + dy, 0, CANVAS_H - initial.h))
    };
  }

  // Resizing keeps the opposite corner anchored and never lets the box leave the card or collapse.
  function resizePatch(dx: number, dy: number) {
    const right = initial.x + initial.w;
    const bottom = initial.y + initial.h;
    const movesLeft = resizeCorner === 'tl' || resizeCorner === 'bl';
    const movesTop = resizeCorner === 'tl' || resizeCorner === 'tr';

    const x = movesLeft ? clamp(initial.x + dx, 0, right - MIN_WIDTH) : initial.x;
    const y = movesTop ? clamp(initial.y + dy, 0, bottom - MIN_HEIGHT) : initial.y;
    const width = movesLeft ? right - x : clamp(initial.w + dx, MIN_WIDTH, CANVAS_W - initial.x);
    const height = movesTop ? bottom - y : clamp(initial.h + dy, MIN_HEIGHT, CANVAS_H - initial.y);

    return { x: Math.round(x), y: Math.round(y), width: Math.round(width), height: Math.round(height) };
  }

  function handlePointerUp() {
    mode = 'idle';
    resizeCorner = null;
  }
</script>

<svelte:window onpointermove={handlePointerMove} onpointerup={handlePointerUp} onpointercancel={handlePointerUp} />

<div class="dual-canvas-container">
  <div class="canvas-panel">
    <div class="panel-header">
      <span class="indicator interactive"></span>
      <h4>{$t('admin.positionEditor')}</h4>
    </div>
    <div class="canvas-box">
      <canvas
        bind:this={interactiveCanvasEl}
        width={CANVAS_W}
        height={CANVAS_H}
        class="editor-canvas"
        onpointerdown={handlePointerDown}
      ></canvas>
    </div>
  </div>

  <div class="canvas-panel">
    <div class="panel-header">
      <span class="indicator live"></span>
      <h4>{$t('admin.livePreviewLabel')}</h4>
    </div>
    <div class="canvas-box">
      <canvas
        bind:this={previewCanvasEl}
        width={CANVAS_W}
        height={CANVAS_H}
        class="preview-canvas"
      ></canvas>
    </div>
  </div>
</div>

<style>
  .dual-canvas-container {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 24px;
    width: 100%;
  }

  @media (max-width: 900px) {
    .dual-canvas-container {
      grid-template-columns: 1fr;
    }
  }

  .canvas-panel {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .panel-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .panel-header h4 {
    font-size: 13px;
    font-weight: 800;
    letter-spacing: 0.5px;
    color: var(--text-muted);
    text-transform: uppercase;
  }

  .indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .indicator.interactive {
    background: var(--color-accent);
    box-shadow: 0 0 10px rgba(255, 205, 0, 0.6);
  }

  .indicator.live {
    background: var(--color-success);
    box-shadow: 0 0 10px rgba(126, 232, 190, 0.6);
  }

  .canvas-box {
    width: 100%;
    aspect-ratio: 1080 / 1350;
    border-radius: var(--radius-lg);
    overflow: hidden;
    background: #000000;
    box-shadow: var(--shadow-card);
    border: 1px solid var(--color-border);
  }

  .editor-canvas,
  .preview-canvas {
    width: 100%;
    height: 100%;
    display: block;
    cursor: crosshair;
    touch-action: none; /* pointer events drive drag/resize instead of scrolling the page */
  }

  .preview-canvas {
    cursor: default;
  }
</style>
