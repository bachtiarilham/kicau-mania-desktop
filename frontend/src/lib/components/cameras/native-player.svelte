<script lang="ts">
  import { onMount } from "svelte";
  import { api, message } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  let {
    id,
    recording = false,
    zoom = 100,
    visible = true,
    class: className = "",
  }: {
    id: string;
    recording?: boolean;
    zoom?: number;
    visible?: boolean;
    class?: string;
  } = $props();
  let element: HTMLDivElement;
  let error = $state("");
  let busy = false;
  let pending = false;
  let disposed = false;
  let requestFrame = 0;
  let sync: () => void = () => {};
  function update() {
    if (!element || disposed) return;
    if (busy) {
      pending = true;
      return;
    }
    const r = element.getBoundingClientRect();
    const scrollParent =
      element.closest(".overflow-y-auto") ||
      document.querySelector("[data-viewport]");
    const viewport =
      scrollParent?.getBoundingClientRect() ||
      new DOMRect(0, 0, innerWidth, innerHeight);
    const left = Math.max(0, r.left, viewport.left),
      top = Math.max(0, r.top, viewport.top),
      right = Math.min(innerWidth, r.right, viewport.right),
      bottom = Math.min(innerHeight, r.bottom, viewport.bottom);
    const rect = {
      x: Math.round(r.left),
      y: Math.round(r.top),
      width: Math.round(r.width),
      height: Math.round(r.height),
      clip_x: Math.max(0, Math.round(left - r.left)),
      clip_y: Math.max(0, Math.round(top - r.top)),
      clip_width: Math.max(0, Math.round(right - left)),
      clip_height: Math.max(0, Math.round(bottom - top)),
      visible: visible && !engine.overlay && right > left && bottom > top,
    };
    busy = true;
    (recording ? api.playRecording(id, rect) : api.player(id, rect, zoom))
      .then(() => (error = ""))
      .catch((e) => (error = message(e)))
      .finally(() => {
        busy = false;
        if (disposed) {
          void api.stopPlayer(recording ? "recording-" + id : id);
        } else if (pending) {
          pending = false;
          sync();
        }
      });
  }
  onMount(() => {
    disposed = false;
    sync = () => {
      cancelAnimationFrame(requestFrame);
      requestFrame = requestAnimationFrame(update);
    };
    const resize = new ResizeObserver(sync);
    resize.observe(element);
    window.addEventListener("resize", sync);
    window.addEventListener("scroll", sync, true);
    const timer = setInterval(sync, 3000);
    sync();
    return () => {
      disposed = true;
      clearInterval(timer);
      cancelAnimationFrame(requestFrame);
      resize.disconnect();
      window.removeEventListener("resize", sync);
      window.removeEventListener("scroll", sync, true);
      void api.stopPlayer(recording ? "recording-" + id : id);
    };
  });
  $effect(() => {
    visible;
    zoom;
    engine.overlay;
    sync();
  });
</script>

<div
  bind:this={element}
  class="relative min-h-0 bg-[#090f10] text-center {className || 'aspect-video w-full'}"
>
  <div
    class="absolute inset-0 grid place-content-center gap-2 p-4 text-xs text-muted-foreground"
  >
    <span
      >{error ||
        engine.playerStatus[id]?.message ||
        engine.playerStatus[id]?.state ||
        "Menghubungkan RTSP…"}</span
    ><span class="font-mono text-[9px] text-[#51635d]"
      >NATIVE PLAYER · RTSP/TCP</span
    >
  </div>
</div>
