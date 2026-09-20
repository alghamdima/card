<script lang="ts">
  import { onMount } from 'svelte';
  import { CANVAS_W, CANVAS_H, renderCard, ensureFontsLoaded } from './canvas-renderer';
  import type { BoxesConfig, TextFieldConfig } from '../../types/campaign.types';

  interface Props {
    imageSrc: string;
    boxes?: BoxesConfig | null;
    dynamicFields?: TextFieldConfig[];
    fieldValues?: Record<string, any>;
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
  let bgImg: HTMLImageElement | null = null;
  let isReady = $state(false);

  function draw() {
    if (!canvasEl) return;
    renderCard(canvasEl, bgImg, boxes, { to, from, message, heading, fieldValues }, dynamicFields);
  }

  $effect(() => {
    // Redraw whenever text, fields or values change
    if (isReady && (to !== undefined || from !== undefined || message !== undefined || fieldValues !== undefined || dynamicFields !== undefined)) {
      draw();
    }
  });

  $effect(() => {
    // Reload image when imageSrc changes
    if (imageSrc && typeof window !== 'undefined') {
      ensureFontsLoaded().then(() => {
        const img = new Image();
        img.crossOrigin = 'anonymous';
        img.onload = () => {
          bgImg = img;
          isReady = true;
          draw();
          if (onready && canvasEl) onready(canvasEl);
        };
        img.src = imageSrc;
      });
    }
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
