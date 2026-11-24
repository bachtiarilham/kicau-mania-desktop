<script lang="ts">
  import { Volume2, VolumeX, Maximize2 } from "@lucide/svelte";
  import NativePlayer from "$lib/components/cameras/native-player.svelte";
  import ZoomSlider from "$lib/components/cameras/zoom-slider.svelte";
  import AudioMeter from "$lib/components/metrics/audio-meter.svelte";
  import { roundStore } from "$lib/state/round.svelte";
  import { engine } from "$lib/state/engine.svelte";
  import type { GantanganStream } from "$lib/models";

  interface StreamItem extends GantanganStream {
    isLive: boolean;
    channelIndex: number;
  }

  interface Props {
    stream: StreamItem;
    isAudible: boolean;
    zoom: number;
    fitVertical?: boolean;
    onToggleAudio: (id: string) => void;
    onSpotlight: (id: string) => void;
    onZoomChange: (id: string, zoom: number) => void;
  }

  let {
    stream,
    isAudible,
    zoom = 100,
    fitVertical = false,
    onToggleAudio,
    onSpotlight,
    onZoomChange,
  }: Props = $props();

  let participant = $derived(roundStore.getParticipantByGantangan(stream.gantanganNumber));
  let isChannelAttached = $derived(
    engine.connected && (engine.channels.length === 0 || engine.channels.some((c) => c.id === stream.id))
  );
</script>

<article
  class="group relative flex {fitVertical ? 'h-full min-h-0' : ''} flex-col overflow-hidden rounded-xl border border-border bg-card shadow-sm transition hover:border-border"
>
  <header class="flex h-7 shrink-0 items-center justify-between border-b border-border bg-muted px-2 py-0.5">
    <div class="flex min-w-0 items-center gap-1.5">
      <span
        class="flex h-4 items-center justify-center rounded bg-gradient-to-br {stream.accentColor} px-1 font-mono text-[10px] font-black text-white"
      >
        {stream.code}
      </span>
      {#if participant}
        <span class="truncate text-[11px] font-semibold text-foreground" title="{participant.birdName} ({participant.ownerName})">
          {participant.birdName} <span class="text-muted-foreground font-normal">({participant.ownerName})</span>
        </span>
      {:else}
        <span class="truncate text-[11px] font-semibold text-foreground">
          {stream.category}
        </span>
      {/if}
    </div>

    <div class="flex shrink-0 items-center gap-1">
      <span class="font-mono text-[8px] text-muted-foreground">
        {#if engine.running}
          <span class="text-rose-400 font-bold animate-pulse">REC</span>
        {:else if engine.connected}
          <span class="text-emerald-400 font-bold">LIVE</span>
        {:else}
          <span class="text-muted-foreground">STANDBY</span>
        {/if}
      </span>

      <button
        type="button"
        onclick={() => onToggleAudio(stream.id)}
        class="flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:text-foreground"
        title={isAudible ? "Bisukan suara" : "Nyalakan suara"}
      >
        {#if !isAudible}
          <VolumeX class="h-3 w-3 text-muted-foreground" />
        {:else}
          <Volume2 class="h-3 w-3 text-sky-400" />
        {/if}
      </button>

      <button
        type="button"
        onclick={() => onSpotlight(stream.id)}
        class="flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:text-foreground"
        title="Perbesar sorotan kamera"
      >
        <Maximize2 class="h-3 w-3" />
      </button>
    </div>
  </header>

  <div class="relative mx-0.5 {fitVertical ? 'min-h-0 flex-1' : 'aspect-video'} overflow-hidden bg-muted">
    {#if isChannelAttached}
      <NativePlayer zoom={zoom} id={stream.id} visible={true} class={fitVertical ? 'h-full w-full' : ''} />
    {:else if engine.connected}
      <div class="flex h-full w-full items-center justify-center bg-muted text-muted-foreground text-xs">
        <span>Kamera tidak terpasang</span>
      </div>
    {:else}
      <div class="flex h-full w-full items-center justify-center bg-gradient-to-br {stream.previewGradient}">
        <span class="text-[11px] font-bold text-muted-foreground">
          {stream.code} · {stream.channelName}
        </span>
      </div>
    {/if}
  </div>

  <div class="shrink-0 bg-card p-0.5">
    {#if engine.connected && engine.channels.some((c) => c.id === stream.id)}
      <ZoomSlider value={zoom} onChange={(val) => onZoomChange(stream.id, val)} />
    {/if}
    <AudioMeter metric={engine.metrics[stream.id]} compact />
  </div>
</article>
