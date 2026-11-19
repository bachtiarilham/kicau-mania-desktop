<script lang="ts">
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
    Radio,
    ShieldAlert,
    CheckCircle2,
    Sparkles,
  } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import NativePlayer from "../cameras/native-player.svelte";
  import AudioMeter from "../metrics/audio-meter.svelte";
  import { roundStore } from "$lib/state/round.svelte";
  import { engine } from "$lib/state/engine.svelte";
  import { api } from "$lib/services/api";
  import { MOCK_GANTANGAN_STREAMS } from "$lib/services/livestream.mock";
  import type { GantanganStream } from "$lib/models/livestream.model";

  let selectedFilter = $state<"all" | "A" | "B" | "C" | "D">("all");
  let spotlightId = $state<string | null>(null);
  let showEditModal = $state(false);
  let audioMuted = $state<Record<string, boolean>>({});

  // Local config form state
  let editEventName = $state(roundStore.eventName);
  let editSessionName = $state(roundStore.sessionName);
  let editStartTime = $state(roundStore.startTime);
  let editEndTime = $state(roundStore.endTime);
  let editDuration = $state(roundStore.durationSeconds);
  let editJurors = $state([...roundStore.jurors]);

  function openEditModal() {
    editEventName = roundStore.eventName;
    editSessionName = roundStore.sessionName;
    editStartTime = roundStore.startTime;
    editEndTime = roundStore.endTime;
    editDuration = roundStore.durationSeconds;
    editJurors = [...roundStore.jurors];
    showEditModal = true;
  }

  function saveConfig() {
    roundStore.updateConfig({
      eventName: editEventName,
      sessionName: editSessionName,
      startTime: editStartTime,
      endTime: editEndTime,
      durationSeconds: Number(editDuration) || 150,
      jurors: editJurors,
    });
    showEditModal = false;
  }

  // Quick preset duration
  function setDurationPreset(sec: number) {
    roundStore.updateConfig({ durationSeconds: sec });
  }

  // All 24 Gantangan Streams
  let allStreams = $derived.by(() => {
    return MOCK_GANTANGAN_STREAMS.map((m, idx) => {
      const realChan = engine.channels[idx];
      return {
        ...m,
        id: realChan?.id ?? m.id,
        isLive: engine.running,
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
    const nextMuted = !audioMuted[id];
    audioMuted[id] = nextMuted;
    try {
      await api.mute(id, nextMuted);
    } catch {
      // Ignore audio mute error
    }
  }
</script>

<div class="flex h-full w-full flex-col overflow-hidden bg-[#0c1011] text-zinc-100 select-none">
  <!-- Top Section: Modern Tactical Match Control Panel (Upgraded Image 2) -->
  <section class="shrink-0 border-b border-white/10 bg-[#12181a] p-3 sm:p-4">
    <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
      <!-- Left: Event, Sesi, & Schedule Meta -->
      <div class="flex flex-wrap items-center gap-3">
        <div class="flex items-center gap-2.5 rounded-xl border border-white/10 bg-black/40 p-2.5 shadow-sm">
          <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-gradient-to-br from-amber-400 to-amber-600 text-black shadow-md shadow-amber-500/20">
            <Trophy class="h-5 w-5" />
          </div>
          <div>
            <div class="text-[10px] font-bold tracking-widest text-amber-400 uppercase">
              Nama Event
            </div>
            <div class="text-sm font-extrabold tracking-tight text-white">
              {roundStore.eventName}
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2.5 rounded-xl border border-white/10 bg-black/40 px-3 py-2 shadow-sm">
          <div>
            <div class="text-[10px] font-bold tracking-widest text-zinc-400 uppercase">
              Babak Penilaian
            </div>
            <div class="text-sm font-black text-amber-300">
              {roundStore.sessionName}
            </div>
          </div>
          <span class="text-zinc-600">·</span>
          <div class="font-mono text-xs text-zinc-300">
            <span class="text-zinc-400">{roundStore.startTime}</span> – <span class="text-zinc-200">{roundStore.endTime}</span>
          </div>
        </div>

        <Button
          variant="outline"
          size="sm"
          onclick={openEditModal}
          class="border-white/10 bg-white/5 text-xs text-zinc-300 hover:bg-white/10 hover:text-white"
        >
          <Settings class="h-3.5 w-3.5" />
          <span>Atur Lomba & Juri</span>
        </Button>
      </div>

      <!-- Center & Right: Tactical Big START / STOP Buttons & Giant Timer -->
      <div class="flex flex-wrap items-center gap-4">
        <!-- Duration Presets -->
        <div class="hidden items-center gap-1 rounded-lg border border-white/10 bg-black/40 p-1 xl:flex">
          {#each [120, 150, 180, 300] as sec}
            <button
              type="button"
              onclick={() => setDurationPreset(sec)}
              class="rounded px-2 py-1 font-mono text-[11px] font-semibold transition {roundStore.durationSeconds === sec ? 'bg-white/15 text-white' : 'text-zinc-400 hover:text-zinc-200'}"
            >
              {sec}s
            </button>
          {/each}
        </div>

        <!-- Giant Synchronized Countdown Timer -->
        <div
          class="flex items-center gap-3 rounded-2xl border px-4 py-1.5 shadow-inner transition-all {roundStore.state === 'RUNNING' ? 'border-emerald-500/50 bg-emerald-950/30' : roundStore.remainingSeconds === 0 ? 'border-red-500/80 bg-red-950/40 animate-pulse' : 'border-white/10 bg-black/60'}"
        >
          <div class="flex flex-col items-center">
            <span class="text-[9px] font-bold tracking-widest text-zinc-400 uppercase">
              Timer Sisa
            </span>
            <span
              class="font-mono text-3xl font-black tracking-tight {roundStore.state === 'RUNNING' ? 'text-emerald-400' : roundStore.remainingSeconds === 0 ? 'text-red-500' : 'text-white'}"
            >
              {roundStore.remainingSeconds}s
            </span>
          </div>

          <button
            type="button"
            onclick={() => roundStore.resetRound()}
            class="flex h-7 w-7 items-center justify-center rounded-lg bg-white/5 text-zinc-400 transition hover:bg-white/15 hover:text-white"
            title="Reset timer ke awal"
          >
            <RotateCcw class="h-3.5 w-3.5" />
          </button>
        </div>

        <!-- Modern Tactile START & STOP Buttons (Upgraded from Image 2) -->
        <div class="flex items-center gap-2.5">
          <!-- START BUTTON: Emerald Gloss with Pulse -->
          <button
            type="button"
            disabled={roundStore.state === "RUNNING"}
            onclick={() => roundStore.startRound()}
            class="group relative flex h-12 items-center gap-2.5 rounded-full px-6 text-sm font-black tracking-wider text-white uppercase shadow-lg transition-all duration-200 disabled:opacity-40 disabled:cursor-not-allowed {roundStore.state === 'RUNNING' ? 'bg-emerald-600/60 shadow-emerald-600/20' : 'bg-gradient-to-r from-emerald-600 to-green-500 hover:from-emerald-500 hover:to-green-400 active:scale-95 shadow-emerald-500/30 hover:shadow-emerald-500/50'}"
          >
            {#if roundStore.state === "RUNNING"}
              <span class="absolute -inset-1 rounded-full bg-emerald-500/20 animate-ping"></span>
            {/if}
            <div class="flex h-6 w-6 items-center justify-center rounded-full bg-white/20">
              <Play class="h-3.5 w-3.5 fill-current text-white ml-0.5" />
            </div>
            <span>{roundStore.state === "RUNNING" ? "SEDANG BERJALAN" : "START"}</span>
          </button>

          <!-- STOP BUTTON: Crimson Gloss -->
          <button
            type="button"
            disabled={roundStore.state === "IDLE" || roundStore.state === "FINISHED"}
            onclick={() => roundStore.stopRound()}
            class="flex h-12 items-center gap-2.5 rounded-full px-6 text-sm font-black tracking-wider text-white uppercase shadow-lg transition-all duration-200 disabled:opacity-40 disabled:cursor-not-allowed bg-gradient-to-r from-red-600 to-rose-600 hover:from-red-500 hover:to-rose-500 active:scale-95 shadow-red-600/30 hover:shadow-red-600/50"
          >
            <div class="flex h-6 w-6 items-center justify-center rounded-full bg-white/20">
              <Square class="h-3.5 w-3.5 fill-current text-white" />
            </div>
            <span>STOP</span>
          </button>
        </div>
      </div>
    </div>
  </section>

  <!-- Filter & Spotlight Status Bar -->
  <section class="flex h-11 shrink-0 items-center justify-between border-b border-white/10 bg-[#0e1416] px-4">
    <div class="flex items-center gap-2">
      <div class="flex items-center gap-1 rounded-lg border border-white/10 bg-black/40 p-0.5">
        <button
          type="button"
          onclick={() => {
            selectedFilter = "all";
            spotlightId = null;
          }}
          class="rounded px-2.5 py-1 text-xs font-semibold transition {selectedFilter === 'all' ? 'bg-primary text-primary-foreground' : 'text-zinc-400 hover:text-white'}"
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
            class="rounded px-2.5 py-1 text-xs font-semibold transition {selectedFilter === z ? 'bg-primary text-primary-foreground' : 'text-zinc-400 hover:text-white'}"
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
          class="h-7 text-xs border-amber-500/40 text-amber-300 hover:bg-amber-500/10"
        >
          Tutup Sorotan (Kembali ke Matrix)
        </Button>
      {/if}
    </div>

    <div class="flex items-center gap-3 text-xs text-zinc-400">
      <span class="flex items-center gap-1.5 font-mono text-[11px]">
        <Video class="h-3.5 w-3.5 text-primary" />
        <span>{filteredStreams.length} Gantangan Dipantau</span>
      </span>
      <span class="text-zinc-600">|</span>
      <span class="font-mono text-[11px] text-zinc-500">
        RTSP / H.265 Hardware Decoding
      </span>
    </div>
  </section>

  <!-- Main Area: Live 24 Gantangan Matrix Overview -->
  <main class="min-h-0 flex-1 overflow-y-auto p-3">
    {#if spotlightStream}
      <!-- Single Spotlight High-Res View -->
      <div class="flex h-full w-full flex-col gap-3 rounded-2xl border border-white/10 bg-black/60 p-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="rounded bg-gradient-to-br {spotlightStream.accentColor} px-2.5 py-1 text-sm font-black text-white">
              {spotlightStream.code}
            </span>
            <span class="text-base font-bold text-white">
              {spotlightStream.channelName} • {spotlightStream.category}
            </span>
            <span class="text-xs text-zinc-400">({spotlightStream.arenaName})</span>
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

        <div class="relative min-h-0 flex-1 overflow-hidden rounded-xl bg-black">
          {#if engine.running}
            <NativePlayer id={spotlightStream.id} visible={true} />
          {:else}
            <div class="flex h-full w-full items-center justify-center text-zinc-400 text-sm">
              Mulai sesi lomba untuk melihat siaran RTSP waktu nyata
            </div>
          {/if}
        </div>

        <AudioMeter metric={engine.metrics[spotlightStream.id]} />
      </div>
    {:else}
      <!-- 24 Gantangan High-Density Matrix (6 columns x 4 rows on 1080p) -->
      <div
        class="grid gap-2.5"
        style:grid-template-columns={selectedFilter === "all" ? "repeat(auto-fill, minmax(240px, 1fr))" : "repeat(3, minmax(0, 1fr))"}
      >
        {#each filteredStreams as stream (stream.id)}
          <article
            class="group relative flex flex-col overflow-hidden rounded-xl border border-white/10 bg-[#131a1c] shadow-sm transition hover:border-white/25"
          >
            <!-- Card Header -->
            <header class="flex h-7 shrink-0 items-center justify-between border-b border-white/5 bg-black/40 px-2 py-0.5">
              <div class="flex min-w-0 items-center gap-1.5">
                <span
                  class="flex h-4 items-center justify-center rounded bg-gradient-to-br {stream.accentColor} px-1 font-mono text-[10px] font-black text-white"
                >
                  {stream.code}
                </span>
                <span class="truncate text-[11px] font-semibold text-zinc-200">
                  {stream.category}
                </span>
              </div>

              <div class="flex shrink-0 items-center gap-1">
                <span class="font-mono text-[8px] text-zinc-400">
                  {#if engine.running}
                    <span class="text-emerald-400 font-bold">LIVE</span>
                  {:else}
                    <span class="text-zinc-500">STANDBY</span>
                  {/if}
                </span>

                <button
                  type="button"
                  onclick={() => toggleAudio(stream.id)}
                  class="flex h-5 w-5 items-center justify-center rounded text-zinc-400 hover:text-white"
                  title="Audio"
                >
                  {#if audioMuted[stream.id]}
                    <VolumeX class="h-3 w-3 text-zinc-500" />
                  {:else}
                    <Volume2 class="h-3 w-3 text-sky-400" />
                  {/if}
                </button>

                <button
                  type="button"
                  onclick={() => (spotlightId = stream.id)}
                  class="flex h-5 w-5 items-center justify-center rounded text-zinc-400 hover:text-white"
                  title="Perbesar sorotan kamera"
                >
                  <Maximize2 class="h-3 w-3" />
                </button>
              </div>
            </header>

            <!-- Video Aspect Canvas -->
            <div class="relative aspect-video w-full overflow-hidden bg-black/60">
              {#if engine.running}
                <NativePlayer id={stream.id} visible={true} />
              {:else}
                <div class="flex h-full w-full items-center justify-center bg-gradient-to-br {stream.previewGradient}">
                  <span class="text-[11px] font-bold text-zinc-400">
                    {stream.code} · {stream.channelName}
                  </span>
                </div>
              {/if}
            </div>

            <!-- Mini Audio Meter -->
            <div class="shrink-0 bg-[#0e1416]/90 p-0.5">
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
    <div class="w-full max-w-md rounded-2xl border border-white/10 bg-[#161e20] p-6 text-zinc-100 shadow-2xl">
      <div class="flex items-center gap-2 border-b border-white/10 pb-3">
        <Trophy class="h-5 w-5 text-amber-400" />
        <h3 class="text-base font-bold text-white">Konfigurasi Lomba & Dewan Juri</h3>
      </div>

      <div class="mt-4 space-y-3.5">
        <div>
          <Label class="text-xs text-zinc-400">Nama Event</Label>
          <Input bind:value={editEventName} placeholder="Piala Umi Kasum" class="mt-1 bg-black/40" />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <Label class="text-xs text-zinc-400">Babak / Sesi</Label>
            <Input bind:value={editSessionName} placeholder="Sesi 1" class="mt-1 bg-black/40" />
          </div>
          <div>
            <Label class="text-xs text-zinc-400">Durasi Timer (detik)</Label>
            <Input type="number" bind:value={editDuration} placeholder="150" class="mt-1 bg-black/40" />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <Label class="text-xs text-zinc-400">Jam Mulai</Label>
            <Input bind:value={editStartTime} placeholder="10.00" class="mt-1 bg-black/40" />
          </div>
          <div>
            <Label class="text-xs text-zinc-400">Jam Selesai</Label>
            <Input bind:value={editEndTime} placeholder="10.10 WIB" class="mt-1 bg-black/40" />
          </div>
        </div>

        <div>
          <Label class="text-xs text-zinc-400">Dewan Juri (4 Juri)</Label>
          <div class="mt-1.5 space-y-1.5">
            {#each editJurors as _, idx}
              <Input bind:value={editJurors[idx]} placeholder={`Juri ${idx + 1}`} class="bg-black/40 text-xs" />
            {/each}
          </div>
        </div>
      </div>

      <div class="mt-6 flex items-center justify-end gap-2 border-t border-white/10 pt-4">
        <Button variant="ghost" size="sm" onclick={() => (showEditModal = false)}>
          Batal
        </Button>
        <Button size="sm" onclick={saveConfig} class="bg-primary text-primary-foreground">
          Simpan Konfigurasi
        </Button>
      </div>
    </div>
  </div>
{/if}
