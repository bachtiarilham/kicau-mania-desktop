<script lang="ts">
  import { Minimize2 } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import NativePlayer from "$lib/components/cameras/native-player.svelte";
  import ZoomSlider from "$lib/components/cameras/zoom-slider.svelte";
  import AudioMeter from "$lib/components/metrics/audio-meter.svelte";
  import { engine } from "$lib/state/engine.svelte";
  import type { GantanganStream } from "$lib/models";

  interface StreamItem extends GantanganStream {
    isLive: boolean;
    channelIndex: number;
  }

  interface Props {
    stream: StreamItem;
    zoom: number;
    onClose: () => void;
    onZoomChange: (zoom: number) => void;
  }

  let { stream, zoom = 100, onClose, onZoomChange }: Props = $props();
</script>

<div class="flex h-full w-full flex-col gap-3 rounded-2xl border border-border bg-muted p-4">
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-2">
      <span class="rounded bg-gradient-to-br {stream.accentColor} px-2.5 py-1 text-sm font-black text-white">
        {stream.code}
      </span>
      <span class="text-base font-bold text-foreground">
        {stream.channelName} • {stream.category}
      </span>
      <span class="text-xs text-muted-foreground">({stream.arenaName})</span>
    </div>
    <Button
      variant="outline"
      size="sm"
      onclick={onClose}
      class="h-8 text-xs"
    >
      <Minimize2 class="h-3.5 w-3.5" />
      <span>Kembali ke 24 Kamera</span>
    </Button>
  </div>

  <div class="relative min-h-0 flex-1 overflow-hidden rounded-xl bg-muted">
    {#if engine.connected}
      <NativePlayer zoom={zoom} id={stream.id} visible={true} class="h-full w-full" />
    {:else}
      <div class="flex h-full w-full items-center justify-center text-muted-foreground text-sm">
        Menghubungkan ke engine siaran kamera...
      </div>
    {/if}
  </div>

  {#if engine.connected && engine.channels.some((c) => c.id === stream.id)}
    <ZoomSlider value={zoom} onChange={onZoomChange} />
  {/if}
  <AudioMeter metric={engine.metrics[stream.id]} />
</div>
