<script lang="ts">
  import { onDestroy } from "svelte";
  import {
    Play,
    Square,
    RotateCcw,
    Trophy,
    Clock,
    Users,
    Video,
    Volume2,
    VolumeX,
    Maximize2,
    Minimize2,
    Settings,
    Cloud,
    Radio,
    ShieldAlert,
    CheckCircle2,
    Sparkles,
    HardDrive,
  } from "@lucide/svelte";
  import * as Dialog from "$lib/components/ui/alert-dialog";
  import RoundTimeUpDialog from "../sessions/round-timeup-dialog.svelte";
  import CloudSessionSelector from "../sessions/cloud-session-selector.svelte";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import NativePlayer from "../cameras/native-player.svelte";
  import ZoomSlider from "../cameras/zoom-slider.svelte";
  import AudioMeter from "../metrics/audio-meter.svelte";
  import { roundStore } from "$lib/state/round.svelte";
  import { engine } from "$lib/state/engine.svelte";
  import { api } from "$lib/services/api";
  import { MOCK_GANTANGAN_STREAMS } from "$lib/services/livestream.mock";
  import type { GantanganStream } from "$lib/models/livestream.model";

  let selectedFilter = $state<"all" | "A" | "B" | "C" | "D">("all");
  let spotlightId = $state<string | null>(null);
  let showEditModal = $state(false);
  let showCloudModal = $state(false);
  let openingEditModal = $state(false);
  let disposed = false;
  let audibleId = $state<string | null>(null);
  let zoomed = $state<Record<string, number>>({});

  // Quick Save Session Dialog state
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
          engine.error =
            "Gagal mengekspor rekaman: " + (err?.message || String(err));
        });
    }
  }

  // Local config form state
  let editEventName = $state(roundStore.eventName);
  let editSessionName = $state(roundStore.sessionName);
  let editStartTime = $state(roundStore.startTime);
  let editEndTime = $state(roundStore.endTime);
  let editTotalDuration = $state(roundStore.totalDurationSeconds);
  let editShiftDuration = $state(roundStore.shiftDurationSeconds);
  let editJurors = $state([...roundStore.jurors]);

  async function openEditModal() {
    if (openingEditModal || showEditModal) return;
    editEventName = roundStore.eventName;
    editSessionName = roundStore.sessionName;
    editStartTime = roundStore.startTime;
    editEndTime = roundStore.endTime;
    editTotalDuration = roundStore.totalDurationSeconds;
    editShiftDuration = roundStore.shiftDurationSeconds;
    editJurors = [...roundStore.jurors];
    openingEditModal = true;
    engine.overlay = true;
    try {
      await api.hidePlayers();
    } catch {
      // Player geometry also hides native windows while the overlay is active.
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

  function saveConfig() {
    const total = Number(editTotalDuration) || 600;
    const shift = Number(editShiftDuration) || 150;
    roundStore.updateConfig({
      eventName: editEventName,
      sessionName: editSessionName,
      startTime: editStartTime,
      endTime: editEndTime,
      totalDurationSeconds: total,
      shiftDurationSeconds: shift,
      jurors: editJurors,
    });
    closeEditModal();
  }

  // Quick preset duration (total / 4 = shift)
  function setDurationPreset(totalSec: number) {
    const shiftSec = Math.floor(totalSec / 4);
    roundStore.updateConfig({
      totalDurationSeconds: totalSec,
      shiftDurationSeconds: shiftSec,
    });
  }

  // All 24 Gantangan Streams
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

{#snippet cameraZoom(id: string)}
  {#if engine.connected && engine.channels.some((channel) => channel.id === id)}
    <ZoomSlider value={zoomed[id] ?? 100} onChange={(value) => (zoomed[id] = value)} />
  {/if}
{/snippet}

<div class="flex h-full w-full flex-col overflow-hidden bg-background text-foreground select-none">
  <!-- Top Section: Modern Tactical Match Control Panel (Upgraded Image 2) -->
  <section class="shrink-0 border-b border-border bg-card px-3 py-2 sm:px-4">
    <div class="flex flex-col gap-2.5 lg:flex-row lg:items-center lg:justify-between">
      <!-- Left: Event, Sesi, & Schedule Meta -->
      <div class="flex flex-wrap items-center gap-2.5">
        <div class="flex items-center gap-2 rounded-xl border border-border bg-muted px-2.5 py-1.5 shadow-sm">
          <div class="flex h-7 w-7 items-center justify-center rounded-lg bg-gradient-to-br from-amber-400 to-amber-600 text-black shadow-md shadow-amber-500/20">
            <Trophy class="h-4 w-4" />
          </div>
          <div>
            <div class="text-[9px] font-bold tracking-widest text-amber-400 uppercase">
              Nama Event
            </div>
            <div class="text-xs font-extrabold tracking-tight text-foreground">
              {roundStore.eventName}
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2 rounded-xl border border-border bg-muted px-2.5 py-1.5 shadow-sm">
          <div>
            <div class="text-[9px] font-bold tracking-widest text-muted-foreground uppercase">
              Babak Penilaian
            </div>
            <div class="text-xs font-black text-amber-400">
              {roundStore.sessionName}
            </div>
          </div>
          <span class="text-muted-foreground">·</span>
          <div class="font-mono text-xs text-foreground">
            <span class="text-muted-foreground">{roundStore.startTime}</span> – <span class="text-foreground">{roundStore.endTime}</span>
          </div>
        </div>

        <Button
          size="sm"
          onclick={openCloudModal}
          class="h-7 gap-1.5 bg-primary/20 text-primary border border-primary/40 text-xs font-bold hover:bg-primary/30"
          title="Tarik event & jadwal ronde dari database cloud"
        >
          <Cloud class="h-3 w-3" />
          <span>Pilih Sesi Cloud</span>
        </Button>

        {#if roundStore.roundId}
          <span
            class="hidden sm:inline-flex items-center gap-1 text-[9px] font-bold uppercase tracking-wider bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 px-2 py-0.5 rounded-full"
            title="Sesi aktif tersinkron dengan cloud database"
          >
            <CheckCircle2 class="h-2.5 w-2.5" />
            <span>Cloud Sync</span>
          </span>
        {/if}

        <Button
          variant="outline"
          size="sm"
          onclick={openEditModal}
          disabled={openingEditModal}
          class="h-7 border-border bg-accent text-xs text-foreground hover:bg-accent hover:text-foreground"
        >
          <Settings class="h-3 w-3" />
          <span>Atur Lomba & Juri</span>
        </Button>
      </div>

      <!-- Center & Right: Tactical Big START / STOP Buttons & Giant Timer -->
      <div class="flex flex-wrap items-center gap-3">
        <!-- Duration Presets -->
        <div class="hidden items-center gap-1 rounded-lg border border-border bg-muted p-0.5 xl:flex">
          {#each [{ sec: 600, label: "10m (Std 600s)" }, { sec: 300, label: "5m" }, { sec: 150, label: "2.5m" }] as p}
            <button
              type="button"
              onclick={() => setDurationPreset(p.sec)}
              class="rounded px-2 py-0.5 font-mono text-[10px] font-semibold transition {roundStore.totalDurationSeconds === p.sec ? 'bg-primary text-primary-foreground shadow-xs' : 'text-muted-foreground hover:text-foreground'}"
            >
              {p.label}
            </button>
          {/each}
        </div>

        <!-- Giant Synchronized Countdown Timer (Dual Reading: 10m Round & 150s Shift) -->
        <div
          class="flex items-center gap-2.5 rounded-xl border px-3 py-1 shadow-inner transition-all {roundStore.state === 'RUNNING' ? 'border-emerald-500/50 bg-emerald-500/15' : roundStore.roundRemainingSeconds === 0 ? 'border-red-500/80 bg-red-500/15 animate-pulse' : 'border-border bg-muted'}"
        >
          <div class="flex flex-col items-center">
            <span class="text-[8px] font-bold tracking-widest text-muted-foreground uppercase">
              Ronde: Shift {roundStore.currentShift}/4
            </span>
            <div class="flex items-baseline gap-1.5 font-mono">
              <span
                class="text-2xl font-black tracking-tight {roundStore.state === 'RUNNING' ? 'text-emerald-400' : roundStore.roundRemainingSeconds === 0 ? 'text-red-400' : 'text-foreground'}"
              >
                {Math.floor(roundStore.roundRemainingSeconds / 60)}:{String(roundStore.roundRemainingSeconds % 60).padStart(2, "0")}
              </span>
              <span class="text-xs font-bold text-amber-400">
                ({roundStore.shiftRemainingSeconds}s)
              </span>
            </div>
          </div>

          <button
            type="button"
            onclick={() => roundStore.resetRound()}
            class="flex h-6 w-6 items-center justify-center rounded-lg bg-accent text-muted-foreground transition hover:bg-accent hover:text-foreground"
            title="Reset timer ke awal (10 Menit / 600 Detik)"
          >
            <RotateCcw class="h-3 w-3" />
          </button>
        </div>

        <!-- Modern Tactile START & STOP Buttons (Upgraded from Image 2) -->
        <div class="flex items-center gap-2">
          <!-- START BUTTON: Emerald Gloss with Pulse -->
          <button
            type="button"
            disabled={roundStore.state === "RUNNING"}
            onclick={() => roundStore.startRound()}
            class="group relative flex h-10 items-center gap-2 rounded-full px-5 text-xs font-black tracking-wider text-white uppercase shadow-lg transition-all duration-200 disabled:opacity-40 disabled:cursor-not-allowed {roundStore.state === 'RUNNING' ? 'bg-emerald-600/60 shadow-emerald-600/20' : 'bg-gradient-to-r from-emerald-600 to-green-500 hover:from-emerald-500 hover:to-green-400 active:scale-95 shadow-emerald-500/30 hover:shadow-emerald-500/50'}"
          >
            {#if roundStore.state === "RUNNING"}
              <span class="absolute -inset-1 rounded-full bg-emerald-500/20 animate-ping"></span>
            {/if}
            <div class="flex h-5 w-5 items-center justify-center rounded-full bg-accent">
              <Play class="h-3 w-3 fill-current text-foreground ml-0.5" />
            </div>
            <span>{roundStore.state === "RUNNING" ? "BERJALAN" : "START"}</span>
          </button>

          <!-- STOP BUTTON: Crimson Gloss -->
          <button
            type="button"
            disabled={roundStore.state === "IDLE" || roundStore.state === "FINISHED"}
            onclick={handleStopClick}
            class="flex h-10 items-center gap-2 rounded-full px-5 text-xs font-black tracking-wider text-white uppercase shadow-lg transition-all duration-200 disabled:opacity-40 disabled:cursor-not-allowed bg-gradient-to-r from-red-600 to-rose-600 hover:from-red-500 hover:to-rose-500 active:scale-95 shadow-red-600/30 hover:shadow-red-600/50"
          >
            <div class="flex h-5 w-5 items-center justify-center rounded-full bg-accent">
              <Square class="h-3 w-3 fill-current text-foreground" />
            </div>
            <span>STOP</span>
          </button>
        </div>
      </div>
    </div>
  </section>

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
      <!-- Single Spotlight High-Res View -->
      <div class="flex h-full w-full flex-col gap-3 rounded-2xl border border-border bg-muted p-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="rounded bg-gradient-to-br {spotlightStream.accentColor} px-2.5 py-1 text-sm font-black text-white">
              {spotlightStream.code}
            </span>
            <span class="text-base font-bold text-foreground">
              {spotlightStream.channelName} • {spotlightStream.category}
            </span>
            <span class="text-xs text-muted-foreground">({spotlightStream.arenaName})</span>
          </div>
          <Button
            variant="outline"
            size="sm"
            onclick={() => (spotlightId = null)}
            class="h-8 text-xs"
          >
            <Minimize2 class="h-3.5 w-3.5" />
            <span>Kembali ke 24 Kamera</span>
          </Button>
        </div>

        <div class="relative min-h-0 flex-1 overflow-hidden rounded-xl bg-muted">
          {#if engine.connected}
            <NativePlayer zoom={zoomed[spotlightStream.id] ?? 100} id={spotlightStream.id} visible={true} class="h-full w-full" />
          {:else}
            <div class="flex h-full w-full items-center justify-center text-muted-foreground text-sm">
              Menghubungkan ke engine siaran kamera...
            </div>
          {/if}
        </div>

        {@render cameraZoom(spotlightStream.id)}
        <AudioMeter metric={engine.metrics[spotlightStream.id]} />
      </div>
    {:else if selectedFilter === "all"}
      <!-- 24 Gantangan High-Density Matrix (Scrollable overview) -->
      <div
        class="grid gap-2"
        style:grid-template-columns="repeat(auto-fill, minmax(240px, 1fr))"
      >
        {#each filteredStreams as stream (stream.id)}
          <article
            class="group relative flex flex-col overflow-hidden rounded-xl border border-border bg-card shadow-sm transition hover:border-border"
          >
            <!-- Card Header -->
            <header class="flex h-7 shrink-0 items-center justify-between border-b border-border bg-muted px-2 py-0.5">
              <div class="flex min-w-0 items-center gap-1.5">
                <span
                  class="flex h-4 items-center justify-center rounded bg-gradient-to-br {stream.accentColor} px-1 font-mono text-[10px] font-black text-white"
                >
                  {stream.code}
                </span>
                {#if roundStore.getParticipantByGantangan(stream.gantanganNumber)}
                  {@const part = roundStore.getParticipantByGantangan(stream.gantanganNumber)}
                  <span class="truncate text-[11px] font-semibold text-foreground" title="{part?.birdName} ({part?.ownerName})">
                    {part?.birdName} <span class="text-muted-foreground font-normal">({part?.ownerName})</span>
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
                  onclick={() => toggleAudio(stream.id)}
                  class="flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:text-foreground"
                  title={audibleId === stream.id ? "Bisukan suara" : "Nyalakan suara"}
                >
                  {#if audibleId !== stream.id}
                    <VolumeX class="h-3 w-3 text-muted-foreground" />
                  {:else}
                    <Volume2 class="h-3 w-3 text-sky-400" />
                  {/if}
                </button>

                <button
                  type="button"
                  onclick={() => (spotlightId = stream.id)}
                  class="flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:text-foreground"
                  title="Perbesar sorotan kamera"
                >
                  <Maximize2 class="h-3 w-3" />
                </button>
              </div>
            </header>

            <!-- Video Aspect Canvas -->
            <div class="relative mx-0.5 aspect-video overflow-hidden bg-muted">
              {#if engine.connected && (engine.channels.length === 0 || engine.channels.some((c) => c.id === stream.id))}
                <NativePlayer zoom={zoomed[stream.id] ?? 100} id={stream.id} visible={true} />
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

            <!-- Mini Audio Meter -->
            <div class="shrink-0 bg-card p-0.5">
              {@render cameraZoom(stream.id)}
              <AudioMeter metric={engine.metrics[stream.id]} compact />
            </div>
          </article>
        {/each}
      </div>
    {:else}
      <!-- Exactly 6 Cameras in 3x2 Grid (Contained, Fits 100% Height & Width without overflow) -->
      <div
        class="grid h-full w-full min-h-0 flex-1 gap-2"
        style:grid-template-columns="repeat(3, minmax(0, 1fr))"
        style:grid-template-rows="repeat(2, minmax(0, 1fr))"
      >
        {#each filteredStreams as stream (stream.id)}
          <article
            class="group relative flex h-full min-h-0 flex-col overflow-hidden rounded-xl border border-border bg-card shadow-sm transition hover:border-border"
          >
            <!-- Card Header -->
            <header class="flex h-7 shrink-0 items-center justify-between border-b border-border bg-muted px-2 py-0.5">
              <div class="flex min-w-0 items-center gap-1.5">
                <span
                  class="flex h-4 items-center justify-center rounded bg-gradient-to-br {stream.accentColor} px-1 font-mono text-[10px] font-black text-white"
                >
                  {stream.code}
                </span>
                {#if roundStore.getParticipantByGantangan(stream.gantanganNumber)}
                  {@const part = roundStore.getParticipantByGantangan(stream.gantanganNumber)}
                  <span class="truncate text-[11px] font-semibold text-foreground" title="{part?.birdName} ({part?.ownerName})">
                    {part?.birdName} <span class="text-muted-foreground font-normal">({part?.ownerName})</span>
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
                  onclick={() => toggleAudio(stream.id)}
                  class="flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:text-foreground"
                  title={audibleId === stream.id ? "Bisukan suara" : "Nyalakan suara"}
                >
                  {#if audibleId !== stream.id}
                    <VolumeX class="h-3 w-3 text-muted-foreground" />
                  {:else}
                    <Volume2 class="h-3 w-3 text-sky-400" />
                  {/if}
                </button>

                <button
                  type="button"
                  onclick={() => (spotlightId = stream.id)}
                  class="flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:text-foreground"
                  title="Perbesar sorotan kamera"
                >
                  <Maximize2 class="h-3 w-3" />
                </button>
              </div>
            </header>

            <!-- Video Player Area: Dynamically fills available vertical space -->
            <div class="relative mx-0.5 min-h-0 flex-1 overflow-hidden bg-muted">
              {#if engine.connected && (engine.channels.length === 0 || engine.channels.some((c) => c.id === stream.id))}
                <NativePlayer zoom={zoomed[stream.id] ?? 100} id={stream.id} visible={true} class="h-full w-full" />
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

            <!-- Mini Audio Meter -->
            <div class="shrink-0 bg-card p-0.5">
              {@render cameraZoom(stream.id)}
              <AudioMeter metric={engine.metrics[stream.id]} compact />
            </div>
          </article>
        {/each}
      </div>
    {/if}
  </main>
</div>

<!-- Modal Edit Konfigurasi Lomba & Juri -->
{#if showEditModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-xs p-4">
    <div class="w-full max-w-md rounded-2xl border border-border bg-card p-6 text-foreground shadow-2xl">
      <div class="flex items-center gap-2 border-b border-border pb-3">
        <Trophy class="h-5 w-5 text-amber-400" />
        <h3 class="text-base font-bold text-foreground">Konfigurasi Lomba & Dewan Juri</h3>
      </div>

      <div class="mt-4 space-y-3.5">
        <div>
          <Label class="text-xs text-muted-foreground">Nama Event</Label>
          <Input bind:value={editEventName} placeholder="Piala Umi Kasum" class="mt-1 bg-muted" />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <Label class="text-xs text-muted-foreground">Babak / Sesi</Label>
            <Input bind:value={editSessionName} placeholder="Sesi 1" class="mt-1 bg-muted" />
          </div>
          <div>
            <Label class="text-xs text-muted-foreground">Total Ronde (detik)</Label>
            <Input type="number" bind:value={editTotalDuration} placeholder="600" class="mt-1 bg-muted" />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <Label class="text-xs text-muted-foreground">Durasi Shift Area (detik)</Label>
            <Input type="number" bind:value={editShiftDuration} placeholder="150" class="mt-1 bg-muted" />
          </div>
          <div>
            <Label class="text-xs text-muted-foreground">Total Shift</Label>
            <div class="mt-1 flex h-9 items-center rounded-md border border-input bg-muted px-3 font-mono text-xs text-foreground">
              4 Shift Rotasi Area
            </div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <Label class="text-xs text-muted-foreground">Jam Mulai</Label>
            <Input bind:value={editStartTime} placeholder="10.00" class="mt-1 bg-muted" />
          </div>
          <div>
            <Label class="text-xs text-muted-foreground">Jam Selesai</Label>
            <Input bind:value={editEndTime} placeholder="10.10 WIB" class="mt-1 bg-muted" />
          </div>
        </div>

        <div>
          <Label class="text-xs text-muted-foreground">Dewan Juri (4 Juri)</Label>
          <div class="mt-1.5 space-y-1.5">
            {#each editJurors as _, idx}
              <Input bind:value={editJurors[idx]} placeholder={`Juri ${idx + 1}`} class="bg-muted text-xs" />
            {/each}
          </div>
        </div>
      </div>

      <div class="mt-6 flex items-center justify-end gap-2 border-t border-border pt-4">
        <Button variant="ghost" size="sm" onclick={closeEditModal}>
          Batal
        </Button>
        <Button size="sm" onclick={saveConfig} class="bg-primary text-primary-foreground">
          Simpan Konfigurasi
        </Button>
      </div>
    </div>
  </div>
{/if}

<!-- Dialog Keputusan Waktu Habis: Lanjutkan atau Stop -->
<RoundTimeUpDialog onStopAndSave={handleStopClick} />

<!-- Quick Save Dialog (Muncul setelah klik STOP) -->
<Dialog.Root bind:open={confirmSave}>
  <Dialog.Content class="max-w-md bg-card border-border text-foreground">
    <Dialog.Header>
      <Dialog.Title class="text-base font-bold flex items-center gap-2">
        <HardDrive size={18} class="text-primary" />
        Simpan Rekaman Sesi Lomba
      </Dialog.Title>
      <Dialog.Description class="text-xs text-muted-foreground">
        Sesi ronde telah dihentikan. Anda dapat mengubah nama sesi dan format sebelum disimpan permanen.
      </Dialog.Description>
    </Dialog.Header>

    <div class="space-y-4 py-3">
      <!-- Input Nama Sesi -->
      <div class="space-y-1.5">
        <Label for="op-session-name" class="text-xs font-semibold text-foreground">Nama Sesi Rekaman</Label>
        <Input
          id="op-session-name"
          bind:value={sessionName}
          placeholder="Contoh: Murai Batu A - Ronde 1"
          class="bg-card border-border text-sm"
        />
        <p class="text-[10px] text-muted-foreground">
          Bisa diubah sesuai kelas burung, atau biarkan default berformat event & tanggal.
        </p>
      </div>

      <!-- Pilihan Format Video -->
      <div class="space-y-1.5">
        <Label class="text-xs font-semibold text-foreground">Format File Rekaman</Label>
        <div class="grid grid-cols-2 gap-2">
          <button
            type="button"
            class={`flex items-center justify-between px-3 py-2 rounded-lg border text-xs font-medium transition-colors ${
              selectedFormat === "mp4"
                ? "border-primary bg-primary/10 text-primary"
                : "border-border text-muted-foreground"
            }`}
            onclick={() => (selectedFormat = "mp4")}
          >
            <span>MP4 (Rekomendasi)</span>
            <span class="text-[10px] opacity-70">Universal</span>
          </button>

          <button
            type="button"
            class={`flex items-center justify-between px-3 py-2 rounded-lg border text-xs font-medium transition-colors ${
              selectedFormat === "mkv"
                ? "border-primary bg-primary/10 text-primary"
                : "border-border text-muted-foreground"
            }`}
            onclick={() => (selectedFormat = "mkv")}
          >
            <span>MKV</span>
            <span class="text-[10px] opacity-70">Tahan Crash</span>
          </button>
        </div>
      </div>

      <!-- Lokasi Penyimpanan Harddisk -->
      <div class="space-y-1.5">
        <Label class="text-xs font-semibold text-foreground">Lokasi Harddisk</Label>
        <div class="flex gap-2">
          <Input
            bind:value={storageDir}
            placeholder="D:\KicauRecordings"
            class="bg-card border-border text-xs font-mono"
          />
          <Button
            type="button"
            variant="outline"
            size="sm"
            onclick={async () => {
              try {
                const dir = await api.selectFolder();
                if (dir) storageDir = dir;
              } catch (e) {
                console.error(e);
              }
            }}
          >
            Pilih Folder...
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onclick={async () => {
              try {
                await api.openFolder();
              } catch (e) {
                console.error(e);
              }
            }}
            title="Buka folder di File Explorer"
          >
            📁 Buka
          </Button>
        </div>
        <p class="text-[10px] text-muted-foreground">
          Folder tujuan penyimpanan video rekaman ronde ini.
        </p>
      </div>
    </div>

    <Dialog.Footer class="gap-2 sm:gap-0">
      <Dialog.Cancel class="text-xs" onclick={() => (confirmSave = false)}>
        Batal & Lanjutkan Sesi
      </Dialog.Cancel>
      <Dialog.Action
        class="bg-primary text-primary-foreground hover:bg-primary/90 text-xs font-semibold"
        onclick={handleConfirmSave}
      >
        Simpan Rekaman Ronde
      </Dialog.Action>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<CloudSessionSelector open={showCloudModal} onClose={closeCloudModal} />
