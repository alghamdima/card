<script lang="ts">
  import { CANVAS_W, CANVAS_H, renderCard, ensureFontsLoaded } from './canvas-renderer';
  import type { BoxesConfig, TextFieldConfig } from '../../types/campaign.types';

  interface Props {
    imageSrc: string;
    boxes?: BoxesConfig | null;
    dynamicFields?: TextFieldConfig[];
    fieldValues?: Record<string, string>;
    to?: string;
    from?: string;
    message?: string;
    heading?: string;
    onready?: (canvas: HTMLCanvasElement) => void;
  }

  let {
    imageSrc,
    boxes = null,
    dynamicFields = [],
    fieldValues = {},
    to = '',
    from = '',
    message = '',
    heading = '',
    onready
  }: Props = $props();

  let canvasEl: HTMLCanvasElement;
  let bgImg = $state<HTMLImageElement | null>(null);

  // Redraws whenever the background or any text input changes (all reads below are tracked).
  $effect(() => {
    if (!canvasEl || !bgImg) return;
    renderCard(canvasEl, bgImg, boxes, { to, from, message, heading, fieldValues }, dynamicFields);
  });

  // Load the background image. `cancelled` discards a load that finished after the source changed.
  $effect(() => {
    const src = imageSrc;
    if (!src) return;

    let cancelled = false;
    ensureFontsLoaded().then(() => {
      if (cancelled) return;
      const img = new Image();
      img.crossOrigin = 'anonymous';
      img.onload = () => {
        if (cancelled) return;
        bgImg = img;
        onready?.(canvasEl);
      };
      img.src = src;
    });

    return () => {
      cancelled = true;
    };
  });

  export function getCanvas(): HTMLCanvasElement {
    return canvasEl;
  }
</script>

<div class="canvas-wrapper">
  <canvas
    bind:this={canvasEl}
    width={CANVAS_W}
    height={CANVAS_H}
    class="card-canvas"
  ></canvas>
</div>

<style>
  .canvas-wrapper {
    width: 100%;
    aspect-ratio: 1080 / 1350;
    border-radius: var(--radius-lg);
    overflow: hidden;
    background: var(--surface-2);
    box-shadow: var(--shadow-card);
    border: 1px solid var(--color-border);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .card-canvas {
    width: 100%;
    height: 100%;
    object-fit: contain;
    display: block;
  }
</style>
