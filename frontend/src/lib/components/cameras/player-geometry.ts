type Edges = Pick<DOMRect, "left" | "top" | "right" | "bottom">;

export function pixelRect(bounds: Edges, clip: Edges, scale: number, visible: boolean) {
  const x = Math.round(bounds.left * scale);
  const y = Math.round(bounds.top * scale);
  const right = Math.round(bounds.right * scale);
  const bottom = Math.round(bounds.bottom * scale);
  // Round inward so the native region never covers a pixel outside the DOM clip.
  const clipLeft = Math.max(x, Math.ceil(clip.left * scale));
  const clipTop = Math.max(y, Math.ceil(clip.top * scale));
  const clipRight = Math.min(right, Math.floor(clip.right * scale));
  const clipBottom = Math.min(bottom, Math.floor(clip.bottom * scale));
  const clipWidth = Math.max(0, clipRight - clipLeft);
  const clipHeight = Math.max(0, clipBottom - clipTop);
  return {
    x, y, width: right - x, height: bottom - y,
    clip_x: clipLeft - x, clip_y: clipTop - y,
    clip_width: clipWidth, clip_height: clipHeight,
    visible: visible && clipWidth > 0 && clipHeight > 0,
    device_pixels: true,
  };
}

export function playerGeometry(element: HTMLElement, visible: boolean) {
  const bounds = element.getBoundingClientRect();
  const clip = {
    left: Math.max(0, bounds.left), top: Math.max(0, bounds.top),
    right: Math.min(innerWidth, bounds.right),
    bottom: Math.min(innerHeight, bounds.bottom),
  };
  // VLC paints in its own HWND, outside WebView's CSS clipping hierarchy.
  for (let parent = element.parentElement; parent; parent = parent.parentElement) {
    const { overflowX, overflowY } = getComputedStyle(parent);
    if (overflowX === "visible" && overflowY === "visible") continue;
    const rect = parent.getBoundingClientRect();
    const left = rect.left + parent.clientLeft;
    const top = rect.top + parent.clientTop;
    if (overflowX !== "visible") {
      clip.left = Math.max(clip.left, left);
      clip.right = Math.min(clip.right, left + parent.clientWidth);
    }
    if (overflowY !== "visible") {
      clip.top = Math.max(clip.top, top);
      clip.bottom = Math.min(clip.bottom, top + parent.clientHeight);
    }
  }
  return pixelRect(bounds, clip, window.devicePixelRatio || 1, visible);
}
