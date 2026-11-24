<script lang="ts">
  import GantanganCard from "./gantangan-card.svelte";
  import type { GantanganStream } from "$lib/models";

  interface StreamItem extends GantanganStream {
    isLive: boolean;
    channelIndex: number;
  }

  interface Props {
    streams: StreamItem[];
    selectedFilter: "all" | "A" | "B" | "C" | "D";
    audibleId: string | null;
    zoomed: Record<string, number>;
    onToggleAudio: (id: string) => void;
    onSpotlight: (id: string) => void;
    onZoomChange: (id: string, zoom: number) => void;
  }

  let {
    streams,
    selectedFilter,
    audibleId,
    zoomed,
    onToggleAudio,
    onSpotlight,
    onZoomChange,
  }: Props = $props();
</script>

{#if selectedFilter === "all"}
  <div
    class="grid gap-2"
    style:grid-template-columns="repeat(auto-fill, minmax(240px, 1fr))"
  >
    {#each streams as stream (stream.id)}
      <GantanganCard
        {stream}
        isAudible={audibleId === stream.id}
        zoom={zoomed[stream.id] ?? 100}
        fitVertical={false}
        {onToggleAudio}
        {onSpotlight}
        {onZoomChange}
      />
    {/each}
  </div>
{:else}
  <div
    class="grid h-full w-full min-h-0 flex-1 gap-2"
    style:grid-template-columns="repeat(3, minmax(0, 1fr))"
    style:grid-template-rows="repeat(2, minmax(0, 1fr))"
  >
    {#each streams as stream (stream.id)}
      <GantanganCard
        {stream}
        isAudible={audibleId === stream.id}
        zoom={zoomed[stream.id] ?? 100}
        fitVertical={true}
        {onToggleAudio}
        {onSpotlight}
        {onZoomChange}
      />
    {/each}
  </div>
{/if}
