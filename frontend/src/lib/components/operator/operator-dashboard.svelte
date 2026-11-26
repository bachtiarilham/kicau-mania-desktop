<script lang="ts">
  import { Video } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import GantanganGrid24 from "./grid/gantangan-grid-24.svelte";
  import SpotlightView from "./spotlight/spotlight-view.svelte";
  import { roundStore } from "$lib/state/round.svelte";
  import { engine } from "$lib/state/engine.svelte";
  import { api } from "$lib/services/api";
  import { cameraSlots } from "$lib/components/cameras/camera-slots";
  import type { ZoneId } from "$lib/models";

  let selectedFilter = $state<"all" | ZoneId>("all");
  let spotlightId = $state<string | null>(null);
  let audibleId = $state<string | null>(null);
  let zoomed = $state<Record<string, number>>({});

  const allStreams = $derived(cameraSlots(engine.channels));

  let filteredStreams = $derived.by(() => {
    if (selectedFilter === "all") return allStreams;
    return allStreams.filter((s) => s.zone === selectedFilter);
  });

  let spotlightStream = $derived(
    spotlightId ? allStreams.find((s) => s.id === spotlightId) ?? null : null
  );

  async function toggleAudio(id: string) {
    const nextMuted = audibleId === id;
    try {
      await api.mute(id, nextMuted);
      audibleId = nextMuted ? null : id;
    } catch {
      // Ignore audio mute error
    }
  }
</script>

<div class="flex h-full w-full flex-col overflow-hidden bg-background text-foreground select-none">
  <section class="flex h-10 shrink-0 items-center justify-between border-b border-border bg-card px-3 sm:px-4">
    <div class="flex items-center gap-2">
      <div class="flex items-center gap-1 rounded-lg border border-border bg-muted p-0.5">
        <button
          type="button"
          onclick={() => {
            selectedFilter = "all";
            spotlightId = null;
          }}
          class="rounded px-2.5 py-1 text-xs font-semibold transition {selectedFilter === 'all' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground'}"
        >
          Semua (24 Kamera)
        </button>

        {#each ["A", "B", "C", "D"] as const as z}
          <button
            type="button"
            onclick={() => {
              selectedFilter = z;
              spotlightId = null;
            }}
            class="rounded px-2.5 py-1 text-xs font-semibold transition {selectedFilter === z ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground'}"
          >
            Zona {z} (6 Cam)
          </button>
        {/each}
      </div>

      {#if spotlightId}
        <Button
          variant="outline"
          size="sm"
          onclick={() => (spotlightId = null)}
          class="h-7 text-xs border-amber-500/40 text-amber-400 hover:bg-amber-500/10"
        >
          Tutup Sorotan (Kembali ke Matrix)
        </Button>
      {/if}
    </div>

    <div class="flex items-center gap-3 text-xs text-muted-foreground">
      {#if roundStore.sessionName}
        <span class="hidden md:inline-flex items-center gap-1.5 rounded-full border border-border bg-muted px-2.5 py-0.5 text-[11px] font-semibold text-foreground">
          <span class="h-1.5 w-1.5 rounded-full {roundStore.state === 'RUNNING' ? 'bg-emerald-400 animate-pulse' : 'bg-zinc-500'}"></span>
          <span>{roundStore.sessionName}</span>
          {#if roundStore.state === "RUNNING"}
            <span class="font-mono text-emerald-400">
              · {Math.floor(roundStore.roundRemainingSeconds / 60)}:{String(roundStore.roundRemainingSeconds % 60).padStart(2, "0")}
            </span>
          {/if}
        </span>
      {/if}
      <span class="flex items-center gap-1.5 font-mono text-[11px]">
        <Video class="h-3.5 w-3.5 text-primary" />
        <span>{filteredStreams.length} Kamera Live</span>
      </span>
    </div>
  </section>

  <main class="min-h-0 flex-1 {selectedFilter === 'all' ? 'overflow-y-auto p-2.5 sm:p-3' : 'overflow-hidden p-2 flex flex-col'}">
    {#if spotlightStream}
      <SpotlightView
        stream={spotlightStream}
        zoom={zoomed[spotlightStream.id] ?? 100}
        onClose={() => (spotlightId = null)}
        onZoomChange={(val) => (zoomed[spotlightStream.id] = val)}
      />
    {:else}
      <GantanganGrid24
        streams={filteredStreams}
        {selectedFilter}
        {audibleId}
        {zoomed}
        onToggleAudio={toggleAudio}
        onSpotlight={(id) => (spotlightId = id)}
        onZoomChange={(id, val) => (zoomed[id] = val)}
      />
    {/if}
  </main>
</div>
