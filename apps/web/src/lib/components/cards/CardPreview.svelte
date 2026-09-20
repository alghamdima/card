<script lang="ts">
  import { onMount } from 'svelte';
  import { CANVAS_W, CANVAS_H, renderCard, ensureFontsLoaded } from './canvas-renderer';
  import type { BoxesConfig } from '../../types/campaign.types';

  interface Props {
    imageSrc: string;
    boxes: BoxesConfig;
    to?: string;
    from?: string;
    message?: string;
    heading?: string;
    onready?: (canvas: HTMLCanvasElement) => void;
  }

  let {
    imageSrc,
    boxes,
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
    renderCard(canvasEl, bgImg, boxes, { to, from, message, heading });
  }

  $effect(() => {
    // Redraw whenever text or boxes change
    if (isReady && (to !== undefined || from !== undefined || message !== undefined)) {
      draw();
    }
  });

  onMount(() => {
    ensureFontsLoaded().then(() => {
      bgImg = new Image();
      bgImg.crossOrigin = 'anonymous';
      bgImg.onload = () => {
        isReady = true;
        draw();
        if (onready) onready(canvasEl);
      };
      bgImg.src = imageSrc;
    });
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
