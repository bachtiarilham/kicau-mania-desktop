<script lang="ts">
  import { onMount } from "svelte";
  import { api, message } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  import { playerGeometry } from "./player-geometry";
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
    const rect = playerGeometry(element, visible && !engine.overlay);
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
  class="relative min-h-0 bg-muted text-center {className || 'aspect-video w-full'}"
>
  <div
    class="absolute inset-0 grid place-content-center gap-2 p-4 text-xs text-muted-foreground"
  >
    <span
      >{error ||
        engine.playerStatus[id]?.message ||
        engine.playerStatus[id]?.state ||
        "Menghubungkan RTSP…"}</span
    ><span class="font-mono text-[9px] text-muted-foreground"
      >NATIVE PLAYER · RTSP/TCP</span
    >
  </div>
</div>
