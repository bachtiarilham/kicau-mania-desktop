<script lang="ts">
  import { onDestroy } from "svelte";
  import { Video } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import OperatorHeader from "./operator-header.svelte";
  import GantanganGrid24 from "./grid/gantangan-grid-24.svelte";
  import SpotlightView from "./spotlight/spotlight-view.svelte";
  import QuickSaveModal from "./modals/quick-save-modal.svelte";
  import RoundConfigModal from "./modals/round-config-modal.svelte";
  import RoundTimeUpDialog from "../sessions/round-timeup-dialog.svelte";
  import CloudSessionSelector from "../sessions/cloud-session-selector.svelte";
  import { roundStore } from "$lib/state/round.svelte";
  import { engine } from "$lib/state/engine.svelte";
  import { api } from "$lib/services/api";
  import { MOCK_GANTANGAN_STREAMS } from "$lib/services/livestream.mock";
  import type { ZoneId } from "$lib/models";

  let selectedFilter = $state<"all" | ZoneId>("all");
  let spotlightId = $state<string | null>(null);
  let showEditModal = $state(false);
  let showCloudModal = $state(false);
  let openingEditModal = $state(false);
  let disposed = false;
  let audibleId = $state<string | null>(null);
  let zoomed = $state<Record<string, number>>({});

  let confirmSave = $state(false);
  let sessionName = $state("");
  let selectedFormat = $state("mp4");
  let storageDir = $state("D:\\KicauRecordings");

  $effect(() => {
    engine.overlay = confirmSave;
  });

  function generateSessionName(): string {
    const d = new Date();
    const pad = (n: number) => String(n).padStart(2, "0");
    const dateStr = `${d.getFullYear()}${pad(d.getMonth() + 1)}${pad(d.getDate())}`;
    const cleanEvent = roundStore.eventName ? roundStore.eventName.replace(/\s+/g, "_") : "Lomba";
    const cleanSession = roundStore.sessionName ? roundStore.sessionName.replace(/\s+/g, "_") : "Ronde";
    return `${cleanEvent}_${cleanSession}_${dateStr}`;
  }

  function handleStopClick() {
    void api.hidePlayers();
    roundStore.stopRound();
    if (!sessionName) {
      sessionName = generateSessionName();
    }
    void api
      .config()
      .then((cfg) => {
        if (cfg?.recordings_dir) storageDir = cfg.recordings_dir;
      })
      .catch(() => {});
    confirmSave = true;
  }

  async function handleConfirmSave() {
    confirmSave = false;
    const currentSessionId = engine.session?.id;
    await engine.stop();
    engine.notice = `Menyimpan rekaman sesi '${sessionName}' (${selectedFormat.toUpperCase()}) ke ${storageDir}...`;

    if (currentSessionId) {
      void api
        .saveSession(currentSessionId, sessionName, storageDir, selectedFormat)
        .then((savedFiles) => {
          if (Array.isArray(savedFiles) && savedFiles.length > 0) {
            engine.notice = `Berhasil menyimpan ${savedFiles.length} file rekaman sesi '${sessionName}' ke: ${storageDir}`;
          }
        })
        .catch((err) => {
          console.warn("Gagal simpan rekaman sesi:", err);
          engine.error = "Gagal mengekspor rekaman: " + (err?.message || String(err));
        });
    }
  }

  async function openEditModal() {
    if (openingEditModal || showEditModal) return;
    openingEditModal = true;
    engine.overlay = true;
    try {
      await api.hidePlayers();
    } catch {
      // Player geometry hides native windows while overlay is active
    } finally {
      if (!disposed) showEditModal = true;
      openingEditModal = false;
    }
  }

  function closeEditModal() {
    showEditModal = false;
    engine.overlay = false;
  }

  async function openCloudModal() {
    if (openingEditModal || showEditModal || showCloudModal) return;
    engine.overlay = true;
    try {
      await api.hidePlayers();
    } catch {
      // Player geometry hides native windows while overlay is active
    } finally {
      if (!disposed) showCloudModal = true;
    }
  }

  function closeCloudModal() {
    showCloudModal = false;
    engine.overlay = false;
  }

  onDestroy(() => {
    disposed = true;
    if (showEditModal || openingEditModal || showCloudModal) engine.overlay = false;
  });

  let allStreams = $derived.by(() => {
    return MOCK_GANTANGAN_STREAMS.map((m, idx) => {
      const realChan = engine.channels[idx];
      return {
        ...m,
        id: realChan?.id ?? m.id,
        isLive: engine.connected || engine.running,
        channelIndex: idx,
      };
    });
  });

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
  <!-- Tactical Match Control Panel -->
  <OperatorHeader
    {openingEditModal}
    onOpenCloudModal={openCloudModal}
    onOpenEditModal={openEditModal}
    onStopRound={handleStopClick}
  />

  <!-- Filter & Spotlight Status Bar -->
  <section class="flex h-9 shrink-0 items-center justify-between border-b border-border bg-card px-3 sm:px-4">
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
      <span class="flex items-center gap-1.5 font-mono text-[11px]">
        <Video class="h-3.5 w-3.5 text-primary" />
        <span>{filteredStreams.length} Gantangan Dipantau</span>
      </span>
      <span class="text-muted-foreground">|</span>
      <span class="font-mono text-[11px] text-muted-foreground">
        RTSP / H.265 Hardware Decoding
      </span>
    </div>
  </section>

  <!-- Main Area: Live 24 Gantangan Matrix Overview -->
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

<!-- Konfigurasi lomba dari sesi cloud -->
<RoundConfigModal
  open={showEditModal}
  initialConfig={{
    roundId: roundStore.roundId,
    eventName: roundStore.eventName,
    sessionName: roundStore.sessionName,
    startTime: roundStore.startTime,
    endTime: roundStore.endTime,
    totalDurationSeconds: roundStore.totalDurationSeconds,
    shiftDurationSeconds: roundStore.shiftDurationSeconds,
    jurors: roundStore.jurors,
  }}
  onClose={closeEditModal}
/>

<!-- Dialog Keputusan Waktu Habis: Lanjutkan atau Stop -->
<RoundTimeUpDialog onStopAndSave={handleStopClick} />

<!-- Quick Save Dialog (Muncul setelah klik STOP) -->
<QuickSaveModal
  bind:open={confirmSave}
  bind:sessionName
  bind:selectedFormat
  bind:storageDir
  onConfirm={handleConfirmSave}
  onCancel={() => (confirmSave = false)}
/>

<CloudSessionSelector open={showCloudModal} onClose={closeCloudModal} />
