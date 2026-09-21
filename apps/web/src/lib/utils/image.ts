// Client-side artwork preparation. Card artwork is stored as a base64 data URL inside the database and is
// downloaded by every visitor, so raw 5-10 MB PNGs are re-encoded to a sensible size before upload.

const ACCEPTED_TYPES = ['image/png', 'image/jpeg', 'image/webp'];
const MAX_INPUT_BYTES = 20 * 1024 * 1024;
const MAX_WIDTH = 1080;
const MAX_HEIGHT = 1350;
const THUMB_WIDTH = 480;

export interface PreparedArtwork {
  /** Full-size artwork used to render cards. */
  image: string;
  /** Small preview used in the occasions listing. */
  thumb: string;
}

export function isSupportedImage(file: File): boolean {
  return ACCEPTED_TYPES.includes(file.type) && file.size <= MAX_INPUT_BYTES;
}

export async function prepareArtwork(file: File): Promise<PreparedArtwork> {
  const source = await createImageBitmap(file);
  try {
    const scale = Math.min(MAX_WIDTH / source.width, MAX_HEIGHT / source.height, 1);
    const full = draw(source, Math.round(source.width * scale), Math.round(source.height * scale), false);

    // Keep PNG only when the artwork really uses transparency; opaque art compresses far better as JPEG.
    const image = hasTransparency(source)
      ? await encode(full, 'image/png')
      : await encode(full, 'image/jpeg', 0.92);

    const thumbScale = Math.min(THUMB_WIDTH / source.width, 1);
    const thumbCanvas = draw(source, Math.round(source.width * thumbScale), Math.round(source.height * thumbScale), true);
    const thumb = await encode(thumbCanvas, 'image/jpeg', 0.8);

    return { image, thumb };
  } finally {
    source.close();
  }
}

function draw(source: ImageBitmap, width: number, height: number, opaqueBackground: boolean): HTMLCanvasElement {
  const canvas = document.createElement('canvas');
  canvas.width = Math.max(1, width);
  canvas.height = Math.max(1, height);
  const ctx = canvas.getContext('2d');
  if (!ctx) throw new Error('Canvas is not available');

  if (opaqueBackground) {
    // JPEG has no alpha channel: without this transparent areas turn black.
    ctx.fillStyle = '#FFFFFF';
    ctx.fillRect(0, 0, canvas.width, canvas.height);
  }
  ctx.imageSmoothingQuality = 'high';
  ctx.drawImage(source, 0, 0, canvas.width, canvas.height);
  return canvas;
}

function hasTransparency(source: ImageBitmap): boolean {
  const probe = document.createElement('canvas');
  probe.width = probe.height = 64;
  const ctx = probe.getContext('2d', { willReadFrequently: true });
  if (!ctx) return false;

  ctx.drawImage(source, 0, 0, 64, 64);
  const { data } = ctx.getImageData(0, 0, 64, 64);
  for (let i = 3; i < data.length; i += 4) {
    if (data[i] < 250) return true;
  }
  return false;
}

function encode(canvas: HTMLCanvasElement, type: 'image/png' | 'image/jpeg', quality?: number): Promise<string> {
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => {
        if (!blob) return reject(new Error('Image encoding failed'));
        const reader = new FileReader();
        reader.onload = () => resolve(reader.result as string);
        reader.onerror = () => reject(reader.error ?? new Error('Image read failed'));
        reader.readAsDataURL(blob);
      },
      type,
      quality
    );
  });
}
