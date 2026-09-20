<script lang="ts">
  import { onMount } from 'svelte';
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

  let {
    imageSrc,
    fields,
    activeFieldId,
    sampleValues,
    onfieldchange,
    onselectfield
  }: Props = $props();

  let interactiveCanvasEl: HTMLCanvasElement;
  let previewCanvasEl: HTMLCanvasElement;
  let bgImg: HTMLImageElement | null = null;
  let isReady = $state(false);

  // Dragging & Resizing interaction state
  let isDragging = $state(false);
  let isResizing = $state(false);
  let resizeHandle = $state<string | null>(null);
  let dragStartX = 0;
  let dragStartY = 0;
  let initialFieldX = 0;
  let initialFieldY = 0;
  let initialFieldW = 0;
  let initialFieldH = 0;

  function getActiveField(): TextFieldConfig | undefined {
    return fields.find((f) => f.id === activeFieldId) || fields[0];
  }

  function redrawAll() {
    if (!interactiveCanvasEl || !previewCanvasEl || !bgImg) return;

    // 1. Draw Live Preview
    renderCard(previewCanvasEl, bgImg, null, { fieldValues: sampleValues }, fields);

    // 2. Draw Interactive Canvas with bounding boxes & handles
    const ctx = interactiveCanvasEl.getContext('2d');
    if (!ctx) return;

    ctx.clearRect(0, 0, CANVAS_W, CANVAS_H);
    if (bgImg.complete && bgImg.naturalWidth) {
      ctx.drawImage(bgImg, 0, 0, CANVAS_W, CANVAS_H);
    }

    // Draw non-active boxes
    fields.forEach((field) => {
      const isActive = field.id === activeFieldId;
      ctx.save();
      ctx.strokeStyle = isActive ? '#FFCD00' : 'rgba(255, 205, 0, 0.4)';
      ctx.lineWidth = isActive ? 5 : 2;
      ctx.fillStyle = isActive ? 'rgba(255, 205, 0, 0.12)' : 'rgba(0, 0, 0, 0.2)';

      // Draw rounded rectangle for pill/box
      const r = Math.min(20, field.height / 2);
      ctx.beginPath();
      ctx.roundRect(field.x, field.y, field.width, field.height, r);
      ctx.fill();
      ctx.stroke();

      // Label on top-left of box
      ctx.font = '22px sans-serif';
      ctx.fillStyle = '#FFFFFF';
      ctx.fillText(field.label || field.name, field.x + 12, Math.max(28, field.y - 8));

      // Draw sample text inside interactive canvas
      const textVal = sampleValues[field.id] || field.label;
      ctx.font = `${field.fontSize || 38}px sans-serif`;
      ctx.fillStyle = field.color || '#FFFFFF';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(textVal, field.x + field.width / 2, field.y + field.height / 2);

      // If active, draw 4 yellow corner handles
      if (isActive) {
        const handleSize = 24;
        ctx.fillStyle = '#FFCD00';
        ctx.strokeStyle = '#2B0A3D';
        ctx.lineWidth = 4;

        const handles = [
          { x: field.x, y: field.y }, // top-left
          { x: field.x + field.width, y: field.y }, // top-right
          { x: field.x, y: field.y + field.height }, // bottom-left
          { x: field.x + field.width, y: field.y + field.height } // bottom-right
        ];

        handles.forEach((h) => {
          ctx.beginPath();
          ctx.arc(h.x, h.y, handleSize / 2, 0, Math.PI * 2);
          ctx.fill();
          ctx.stroke();
        });
      }
      ctx.restore();
    });
  }

  $effect(() => {
    if (isReady && fields && activeFieldId && sampleValues) {
      redrawAll();
    }
  });

  $effect(() => {
    if (imageSrc && typeof window !== 'undefined') {
      ensureFontsLoaded().then(() => {
        const img = new Image();
        img.crossOrigin = 'anonymous';
        img.onload = () => {
          bgImg = img;
          isReady = true;
          redrawAll();
        };
        img.src = imageSrc;
      });
    }
  });

  function getCanvasCoords(e: MouseEvent | Touch): { x: number; y: number } {
    const rect = interactiveCanvasEl.getBoundingClientRect();
    const scaleX = CANVAS_W / rect.width;
    const scaleY = CANVAS_H / rect.height;
    return {
      x: (e.clientX - rect.left) * scaleX,
      y: (e.clientY - rect.top) * scaleY
    };
  }

  function handlePointerDown(e: MouseEvent) {
    const coords = getCanvasCoords(e);
    const active = getActiveField();

    if (active) {
      // Check handles first
      const handleRadius = 30;
      const corners = [
        { name: 'tl', x: active.x, y: active.y },
        { name: 'tr', x: active.x + active.width, y: active.y },
        { name: 'bl', x: active.x, y: active.y + active.height },
        { name: 'br', x: active.x + active.width, y: active.y + active.height }
      ];

      for (const corner of corners) {
        const dist = Math.hypot(coords.x - corner.x, coords.y - corner.y);
        if (dist <= handleRadius) {
          isResizing = true;
          resizeHandle = corner.name;
          dragStartX = coords.x;
          dragStartY = coords.y;
          initialFieldX = active.x;
          initialFieldY = active.y;
          initialFieldW = active.width;
          initialFieldH = active.height;
          return;
        }
      }

      // Check if inside active box for moving
      if (
        coords.x >= active.x &&
        coords.x <= active.x + active.width &&
        coords.y >= active.y &&
        coords.y <= active.y + active.height
      ) {
        isDragging = true;
        dragStartX = coords.x;
        dragStartY = coords.y;
        initialFieldX = active.x;
        initialFieldY = active.y;
        return;
      }
    }

    // Check if clicked another field
    for (const f of fields) {
      if (
        coords.x >= f.x &&
        coords.x <= f.x + f.width &&
        coords.y >= f.y &&
        coords.y <= f.y + f.height
      ) {
        onselectfield(f.id);
        isDragging = true;
        dragStartX = coords.x;
        dragStartY = coords.y;
        initialFieldX = f.x;
        initialFieldY = f.y;
        return;
      }
    }
  }

  function handlePointerMove(e: MouseEvent) {
    if (!isDragging && !isResizing) return;
    const coords = getCanvasCoords(e);
    const dx = coords.x - dragStartX;
    const dy = coords.y - dragStartY;
    const active = getActiveField();
    if (!active) return;

    if (isDragging) {
      const updated = fields.map((f) => {
        if (f.id === active.id) {
          return {
            ...f,
            x: Math.round(Math.max(0, Math.min(CANVAS_W - f.width, initialFieldX + dx))),
            y: Math.round(Math.max(0, Math.min(CANVAS_H - f.height, initialFieldY + dy)))
          };
        }
        return f;
      });
      onfieldchange(updated);
    } else if (isResizing && resizeHandle) {
      let newX = initialFieldX;
      let newY = initialFieldY;
      let newW = initialFieldW;
      let newH = initialFieldH;

      if (resizeHandle === 'br') {
        newW = Math.max(100, Math.min(CANVAS_W - initialFieldX, initialFieldW + dx));
        newH = Math.max(40, Math.min(CANVAS_H - initialFieldY, initialFieldH + dy));
      } else if (resizeHandle === 'bl') {
        const proposedX = initialFieldX + dx;
        newW = Math.max(100, initialFieldW - dx);
        newX = Math.max(0, proposedX);
        newH = Math.max(40, initialFieldH + dy);
      } else if (resizeHandle === 'tr') {
        newW = Math.max(100, initialFieldW + dx);
        const proposedY = initialFieldY + dy;
        newH = Math.max(40, initialFieldH - dy);
        newY = Math.max(0, proposedY);
      } else if (resizeHandle === 'tl') {
        newX = Math.max(0, initialFieldX + dx);
        newY = Math.max(0, initialFieldY + dy);
        newW = Math.max(100, initialFieldW - dx);
        newH = Math.max(40, initialFieldH - dy);
      }

      const updated = fields.map((f) => {
        if (f.id === active.id) {
          return {
            ...f,
            x: Math.round(newX),
            y: Math.round(newY),
            width: Math.round(newW),
            height: Math.round(newH)
          };
        }
        return f;
      });
      onfieldchange(updated);
    }
  }

  function handlePointerUp() {
    isDragging = false;
    isResizing = false;
    resizeHandle = null;
  }
</script>

<svelte:window onmousemove={handlePointerMove} onmouseup={handlePointerUp} />

<div class="dual-canvas-container">
  <div class="canvas-panel">
    <div class="panel-header">
      <span class="indicator interactive"></span>
      <h4>{$t('admin.dragBoxOntoPill')}</h4>
    </div>
    <div class="canvas-box">
      <canvas
        bind:this={interactiveCanvasEl}
        width={CANVAS_W}
        height={CANVAS_H}
        class="editor-canvas"
        onmousedown={handlePointerDown}
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
  }

  .preview-canvas {
    cursor: default;
  }
</style>
