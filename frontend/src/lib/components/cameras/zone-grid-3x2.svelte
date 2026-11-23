<script lang="ts">
  import {
    Maximize2,
    Minimize2,
    Volume2,
    VolumeX,
    Radio,
    Sparkles,
    RefreshCw,
    Info,
    PanelLeftClose,
    PanelLeftOpen,
  } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import NativePlayer from "./native-player.svelte";
  import ZoomSlider from "./zoom-slider.svelte";
  import AudioMeter from "../metrics/audio-meter.svelte";
  import JuriRoundHud from "../juri/juri-round-hud.svelte";
  import { api } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  import { roundStore } from "$lib/state/round.svelte";
  import { MOCK_GANTANGAN_STREAMS } from "$lib/services/livestream.mock";
  import type { GantanganStream } from "$lib/models/livestream.model";

  interface Props {
    showZoneBar?: boolean;
    showJuriHud?: boolean;
  }

  let { showZoneBar = true, showJuriHud = true }: Props = $props();

  const STORAGE_KEY = "kicau_active_zone_3x2";
  const ZONES = [
    { id: "A" as const, label: "Zona A", range: "G01 – G06", desc: "Arena Utama · Sisi Utara" },
    { id: "B" as const, label: "Zona B", range: "G07 – G12", desc: "Arena Timur · Sisi Timur" },
    { id: "C" as const, label: "Zona C", range: "G13 – G18", desc: "Arena Barat · Sisi Selatan" },
    { id: "D" as const, label: "Zona D", range: "G19 – G24", desc: "Arena Selatan · Sisi Barat" },
  ];

  let selectedZone = $state<"A" | "B" | "C" | "D">(
    (typeof localStorage !== "undefined" &&
      (localStorage.getItem(STORAGE_KEY) as "A" | "B" | "C" | "D")) ||
      "A"
  );

  let focusId = $state<string | null>(null);
  let audibleId = $state<string | null>(null);
  let zoomed = $state<Record<string, number>>({});
  let hudCollapsed = $state(false);

  $effect(() => {
    if (typeof localStorage !== "undefined") {
      localStorage.setItem(STORAGE_KEY, selectedZone);
    }
  });

  // Exactly 6 slots per zone
  interface ZoneSlot {
    index: number;
    stream: GantanganStream;
    channelId: string;
    channelIndex: number;
  }

  let zoneSlots = $derived.by<ZoneSlot[]>(() => {
    const startIndex =
      selectedZone === "A"
        ? 0
        : selectedZone === "B"
        ? 6
        : selectedZone === "C"
        ? 12
        : 18;

    const slots: ZoneSlot[] = [];
    for (let i = 0; i < 6; i++) {
      const idx = startIndex + i;
      const stream = MOCK_GANTANGAN_STREAMS[idx] ?? {
        id: `g${String(idx + 1).padStart(2, "0")}`,
        code: `G${String(idx + 1).padStart(2, "0")}`,
        channelName: `Gantangan ${String(idx + 1).padStart(2, "0")}`,
        category: "Murai Batu",
        arenaName: `Arena ${selectedZone}`,
        roundLabel: "Ronde 01",
        viewerCount: 0,
        isLive: engine.connected || engine.running,
        quality: "1080P60",
        zone: selectedZone,
        previewGradient: "from-zinc-800 to-zinc-950",
        accentColor: "from-primary to-primary/80",
      };

      const realChan = engine.channels[idx];
      const channelId = realChan?.id ?? stream.id;

      slots.push({
        index: i,
        stream,
        channelId,
        channelIndex: idx,
      });
    }
    return slots;
  });

  async function toggleAudio(id: string) {
    const nextMuted = audibleId === id;
    try {
      await api.mute(id, nextMuted);
      audibleId = nextMuted ? null : id;
    } catch {
      // Ignore audio mute error if engine in mock
    }
  }

  function toggleFocus(id: string) {
    focusId = focusId === id ? null : id;
  }
</script>

<div class="flex h-full w-full flex-col overflow-hidden bg-[#0c1011] text-zinc-100 select-none">
  <!-- Top Bar: Zone Selector & Control -->
  {#if showZoneBar}
    <div
      class="flex h-12 shrink-0 items-center justify-between border-b border-white/10 bg-[#12181a] px-3 sm:px-4"
    >
      <div class="flex items-center gap-2">
        {#if showJuriHud}
          <button
            type="button"
            onclick={() => (hudCollapsed = !hudCollapsed)}
            class="flex items-center gap-1.5 rounded-lg border border-white/10 bg-black/40 px-2.5 py-1 text-xs font-semibold text-zinc-300 transition hover:bg-white/10 hover:text-white"
            title={hudCollapsed ? "Tampilkan panel info juri" : "Perkecil panel info juri"}
          >
            {#if hudCollapsed}
              <PanelLeftOpen class="h-3.5 w-3.5 text-amber-400" />
              <span class="hidden sm:inline">Info Juri</span>
            {:else}
              <PanelLeftClose class="h-3.5 w-3.5" />
              <span class="hidden sm:inline">Sembunyikan Info</span>
            {/if}
          </button>
        {/if}

        <div class="flex items-center gap-1 rounded-lg bg-black/40 p-1 border border-white/10">
          {#each ZONES as z}
            <button
              type="button"
              onclick={() => {
                selectedZone = z.id;
                focusId = null;
              }}
              class="flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs font-semibold transition {selectedZone === z.id ? 'bg-primary text-primary-foreground shadow-sm' : 'text-zinc-400 hover:text-zinc-200 hover:bg-white/5'}"
            >
              <span>{z.label}</span>
              <span class="text-[10px] opacity-75 font-mono hidden sm:inline">({z.range})</span>
            </button>
          {/each}
        </div>

        {#if focusId}
          <Button
            variant="outline"
            size="sm"
            onclick={() => (focusId = null)}
            class="h-7 text-xs border-amber-500/40 text-amber-300 hover:bg-amber-500/10"
          >
            Kembali ke Grid 3x2
          </Button>
        {/if}
      </div>

      <!-- Right: Live Timer Status Badge -->
      <div class="flex items-center gap-2.5">
        <div class="flex items-center gap-2 rounded-lg border border-white/10 bg-black/50 px-2.5 py-1">
          <span
            class="h-2 w-2 rounded-full {roundStore.state === 'RUNNING' ? 'bg-emerald-400 animate-pulse' : 'bg-zinc-500'}"
          ></span>
          <span class="font-mono text-xs font-black text-amber-300">
            {roundStore.remainingSeconds}s
          </span>
          <span class="text-[10px] text-zinc-400 font-semibold uppercase">
            {roundStore.sessionName}
          </span>
        </div>

        <div class="hidden items-center gap-2 text-[11px] text-zinc-400 md:flex">
          <span class="text-zinc-600">|</span>
          <span class="font-mono text-[10px] text-zinc-500">6 KAMERA 3x2</span>
        </div>
      </div>
    </div>
  {/if}

  <!-- Main View: Juri HUD Sidebar + 3x2 Camera Grid -->
  <div class="flex min-h-0 flex-1 overflow-hidden">
    {#if showJuriHud}
      <JuriRoundHud
        collapsed={hudCollapsed}
        onToggleCollapse={() => (hudCollapsed = !hudCollapsed)}
      />
    {/if}

    <!-- 3x2 Grid Filling Remainder of the Screen -->
    <div class="min-h-0 flex-1 p-2 md:p-3 overflow-hidden">
      <div
        class="grid h-full w-full gap-2"
        style:grid-template-columns={focusId ? "1fr" : "repeat(3, minmax(0, 1fr))"}
        style:grid-template-rows={focusId ? "1fr" : "repeat(2, minmax(0, 1fr))"}
      >
        {#each zoneSlots as slot (slot.channelId)}
          <article
            class:hidden={!!focusId && focusId !== slot.channelId}
            class="group relative flex h-full min-h-0 flex-col overflow-hidden rounded-xl border border-white/10 bg-[#141b1d]/90 shadow-md transition-all duration-150 hover:border-white/20"
          >
            <!-- Cell Header: Compact (height ~32px) -->
            <header
              class="flex h-8 shrink-0 items-center justify-between border-b border-white/5 bg-black/40 px-2.5 py-0.5"
            >
              <div class="flex min-w-0 items-center gap-2">
                <span
                  class="flex h-5 items-center justify-center rounded bg-gradient-to-br {slot.stream.accentColor} px-1.5 text-[11px] font-extrabold tracking-tight text-white shadow-xs"
                >
                  {slot.stream.code}
                </span>
                <span class="truncate text-xs font-semibold text-zinc-200">
                  {slot.stream.category}
                </span>
                <span class="hidden text-[10px] text-zinc-400 sm:inline truncate">
                  · {slot.stream.arenaName}
                </span>
              </div>

              <div class="flex shrink-0 items-center gap-1">
                <span class="font-mono text-[9px] text-zinc-400">
                  {#if engine.running}
                    <span class="inline-flex items-center gap-1 text-rose-400 font-bold">
                      <span class="h-1.5 w-1.5 rounded-full bg-rose-400 animate-pulse"></span>
                      REC
                    </span>
                  {:else if engine.connected}
                    <span class="inline-flex items-center gap-1 text-emerald-400">
                      <span class="h-1.5 w-1.5 rounded-full bg-emerald-400"></span>
                      {engine.playerStatus[slot.channelId]?.fps !== undefined
                        ? `${Number(engine.playerStatus[slot.channelId].fps).toFixed(0)} FPS`
                        : "LIVE"}
                    </span>
                  {:else}
                    <span class="text-zinc-500">STANDBY</span>
                  {/if}
                </span>

                <button
                  type="button"
                  onclick={() => toggleAudio(slot.channelId)}
                  class="flex h-6 w-6 items-center justify-center rounded text-zinc-300 transition hover:bg-white/15"
                  aria-label={audibleId !== slot.channelId ? "Nyalakan suara" : "Bisukan suara"}
                  title={audibleId !== slot.channelId ? "Nyalakan suara" : "Bisukan"}
                >
                  {#if audibleId !== slot.channelId}
                    <VolumeX class="h-3.5 w-3.5 text-zinc-500" />
                  {:else}
                    <Volume2 class="h-3.5 w-3.5 text-sky-400" />
                  {/if}
                </button>

                <button
                  type="button"
                  onclick={() => toggleFocus(slot.channelId)}
                  class="flex h-6 w-6 items-center justify-center rounded text-zinc-300 transition hover:bg-white/15"
                  aria-label={focusId === slot.channelId ? "Kembali ke 3x2" : "Fokus layar penuh"}
                  title={focusId === slot.channelId ? "Kembali ke 3x2" : "Fokus"}
                >
                  {#if focusId === slot.channelId}
                    <Minimize2 class="h-3.5 w-3.5 text-amber-400" />
                  {:else}
                    <Maximize2 class="h-3.5 w-3.5 text-zinc-400 group-hover:text-white" />
                  {/if}
                </button>
              </div>
            </header>

            <!-- Video Player Area: Fills remaining height -->
            <div class="relative min-h-0 flex-1 overflow-hidden bg-black/60">
              {#if engine.connected}
                <NativePlayer
                  id={slot.channelId}
                  zoom={zoomed[slot.channelId] ?? 100}
                  visible={!focusId || focusId === slot.channelId}
                  class="h-full w-full"
                />
              {:else}
                <!-- Aesthetic Standby / Preview Canvas -->
                <div
                  class="relative flex h-full w-full items-center justify-center bg-gradient-to-br {slot.stream.previewGradient}"
                >
                  <div
                    class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_50%_35%,rgba(255,255,255,0.1),transparent_65%)]"
                  ></div>
                  <div
                    class="pointer-events-none absolute inset-0 opacity-15"
                    style="background-image: linear-gradient(rgba(255,255,255,0.06) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.06) 1px, transparent 1px); background-size: 24px 24px;"
                  ></div>

                  <div class="relative z-10 flex flex-col items-center gap-1.5 text-center">
                    <div
                      class="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br {slot.stream.accentColor} text-sm font-black text-white shadow-md ring-2 ring-white/10"
                    >
                      {slot.stream.code}
                    </div>
                    <span class="text-[11px] font-medium text-zinc-300">
                      {slot.stream.channelName}
                    </span>
                    <span class="rounded bg-black/50 px-2 py-0.5 text-[9px] font-semibold text-zinc-400 backdrop-blur-xs">
                      Mulai sesi untuk stream live
                    </span>
                  </div>
                </div>
              {/if}
            </div>

            <!-- Bottom: Active Audio Meter with Live Waveform Graph -->
            <div class="shrink-0 bg-[#0e1416]/95 pt-0.5">
              {#if engine.connected && engine.channels.some((channel) => channel.id === slot.channelId)}
                <ZoomSlider
                  value={zoomed[slot.channelId] ?? 100}
                  onChange={(value) => (zoomed[slot.channelId] = value)}
                />
              {/if}
              <AudioMeter metric={engine.metrics[slot.channelId]} compact />
            </div>
          </article>
        {/each}
      </div>
    </div>
  </div>
</div>
