<script lang="ts">
  import {
    Maximize2,
    Minimize2,
    Volume2,
    VolumeX,
    Radio,
    Sparkles,
    Info,
    PanelLeftClose,
    PanelLeftOpen,
  } from "@lucide/svelte";
  import type { Snippet } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import NativePlayer from "./native-player.svelte";
  import ParticipantLabel from "./participant-label.svelte";
  import ZoomSlider from "./zoom-slider.svelte";
  import AudioMeter from "../metrics/audio-meter.svelte";
  import JuriRoundHud from "../juri/juri-round-hud.svelte";
  import RoundProgress from "../juri/round-progress.svelte";
  import { api } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  import { roundStore, getJudgeAssignedZone } from "$lib/state/round.svelte";
  import { auth } from "$lib/state/auth.svelte";
  import { cameraSlots } from "./camera-slots";

  let { identity, account }: { identity?: Snippet<[() => void]>; account?: Snippet<[boolean]> } = $props();

  const STORAGE_KEY = "kicau_active_zone_3x2";
  const ZONES = [
    { id: "A" as const, label: "Zona A", range: "G01 – G06", desc: "Arena Utama · Sisi Utara" },
    { id: "B" as const, label: "Zona B", range: "G07 – G12", desc: "Arena Timur · Sisi Timur" },
    { id: "C" as const, label: "Zona C", range: "G13 – G18", desc: "Arena Barat · Sisi Selatan" },
    { id: "D" as const, label: "Zona D", range: "G19 – G24", desc: "Arena Selatan · Sisi Barat" },
  ];

  let previewZone = $state<"A" | "B" | "C" | "D">(
    (typeof localStorage !== "undefined" &&
      (localStorage.getItem(STORAGE_KEY) as "A" | "B" | "C" | "D")) ||
      "A"
  );

  let focusId = $state<string | null>(null);
  let audibleId = $state<string | null>(null);
  let zoomed = $state<Record<string, number>>({});
  let hudCollapsed = $state(false);
  const judgeIndex = $derived(roundStore.judgeIndex);
  const selectedZone = $derived(auth.role === "juri"
    ? judgeIndex === null ? null : getJudgeAssignedZone(judgeIndex, roundStore.elapsedSeconds)
    : previewZone);

  $effect(() => {
    if (auth.role !== "juri" && typeof localStorage !== "undefined") {
      localStorage.setItem(STORAGE_KEY, previewZone);
    }
  });

  const zoneSlots = $derived(cameraSlots(engine.channels).filter((slot) => slot.zone === selectedZone));
  $effect(() => {
    if (focusId && !zoneSlots.some((slot) => slot.id === focusId)) focusId = null;
    if (audibleId && !zoneSlots.some((slot) => slot.id === audibleId)) audibleId = null;
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

<div class="flex h-full w-full flex-col overflow-hidden bg-background text-foreground select-none">
  <!-- Top Bar: Zone Selector & Control -->
  {#if auth.role !== "juri"}
    <div
      class="flex h-12 shrink-0 items-center justify-between border-b border-border bg-card px-3 sm:px-4"
    >
      <div class="flex items-center gap-2">
          <button
            type="button"
            onclick={() => (hudCollapsed = !hudCollapsed)}
            class="flex items-center gap-1.5 rounded-lg border border-border bg-muted px-2.5 py-1 text-xs font-semibold text-foreground transition hover:bg-accent hover:text-foreground"
            title={hudCollapsed ? "Tampilkan panel info juri" : "Perkecil panel info juri"}
          >
            {#if hudCollapsed}
              <PanelLeftOpen class="h-3.5 w-3.5 text-amber-400" />
              <span class="hidden sm:inline">Info Juri</span>
            {:else}
              <PanelLeftClose class="h-3.5 w-3.5" />
              <span class="hidden sm:inline">Ringkas Info</span>
            {/if}
          </button>

          <div class="flex items-center gap-1 rounded-lg bg-muted p-1 border border-border">
          {#each ZONES as z}
            <button
              type="button"
              onclick={() => {
                previewZone = z.id;
                focusId = null;
              }}
              class="flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs font-semibold transition {selectedZone === z.id ? 'bg-primary text-primary-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground hover:bg-accent'}"
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
            class="h-7 text-xs border-amber-500/40 text-amber-400 hover:bg-amber-500/10"
          >
            Kembali ke Grid 3x2
          </Button>
        {/if}
      </div>

      <!-- Right: Live Timer Status Badge -->
      <div class="flex items-center gap-2.5">
        <div class="flex items-center gap-2 rounded-lg border border-border bg-muted px-2.5 py-1">
          <span
            class="h-2 w-2 rounded-full {roundStore.state === 'RUNNING' ? 'bg-emerald-400 animate-pulse' : 'bg-zinc-500'}"
          ></span>
          <div class="flex flex-col text-right font-mono">
            <span class="text-xs font-black text-amber-400 leading-none">
              Area: {roundStore.shiftRemainingSeconds}s
            </span>
            <span class="text-[9px] text-muted-foreground leading-tight">
              {roundStore.intervalLabel} · {Math.floor(roundStore.roundRemainingSeconds / 60)}:{String(roundStore.roundRemainingSeconds % 60).padStart(2, "0")}
            </span>
          </div>
          <span class="rounded bg-accent px-1.5 py-0.5 text-[9px] text-foreground font-bold uppercase">
            {roundStore.sessionName}
          </span>
        </div>

        <div class="hidden items-center gap-2 text-[11px] text-muted-foreground md:flex">
          <span class="text-muted-foreground">|</span>
          <span class="font-mono text-[10px] text-muted-foreground">6 KAMERA 3x2</span>
        </div>
      </div>
    </div>
  {/if}

  {#if roundStore.syncError}
    <div class="border-b border-border bg-muted px-4 py-2 text-xs text-muted-foreground" role="status">
      {roundStore.syncError}
    </div>
  {/if}

  <!-- Main View: Juri HUD Sidebar + 3x2 Camera Grid -->
  <div class="flex min-h-0 flex-1 overflow-hidden">
      <JuriRoundHud
        {identity}
        {account}
        collapsed={hudCollapsed}
        onToggleCollapse={() => (hudCollapsed = !hudCollapsed)}
      />

    <!-- 3x2 Grid Filling Remainder of the Screen -->
    <div class="flex min-h-0 min-w-0 flex-1 flex-col gap-2 overflow-hidden p-2 md:p-3">
      {#if auth.role === "juri"}
        <RoundProgress />
      {/if}
      {#if !selectedZone}
        <p class="grid min-h-0 flex-1 place-items-center text-center text-sm text-muted-foreground" role="status">
          {roundStore.roundId ? "Akun ini belum cocok dengan nama juri pada ronde aktif." : "Menunggu operator memulai ronde."}
        </p>
      {:else}
      <div
        class="grid min-h-0 w-full flex-1 gap-2"
        style:grid-template-columns={focusId ? "1fr" : "repeat(3, minmax(0, 1fr))"}
        style:grid-template-rows={focusId ? "1fr" : "repeat(2, minmax(0, 1fr))"}
      >
        {#each zoneSlots as slot (slot.id)}
          <article
            class:hidden={!!focusId && focusId !== slot.id}
            class="group relative flex h-full min-h-0 flex-col overflow-hidden rounded-xl border border-border bg-card shadow-md transition-all duration-150 hover:border-border"
          >
            <!-- Cell Header: Compact (height ~32px) -->
            <header
              class="flex h-8 shrink-0 items-center justify-between border-b border-border bg-muted px-2.5 py-0.5"
            >
              <div class="flex min-w-0 items-center gap-2">
                <ParticipantLabel number={slot.number} class="text-xs font-semibold" />
              </div>

              <div class="flex shrink-0 items-center gap-1">
                <span class="font-mono text-[9px] text-muted-foreground">
                  {#if slot.attached && engine.running}
                    <span class="inline-flex items-center gap-1 text-rose-400 font-bold">
                      <span class="h-1.5 w-1.5 rounded-full bg-rose-400 animate-pulse"></span>
                      REC
                    </span>
                  {:else if slot.attached && engine.connected}
                    <span class="inline-flex items-center gap-1 text-emerald-400">
                      <span class="h-1.5 w-1.5 rounded-full bg-emerald-400"></span>
                      {engine.playerStatus[slot.id]?.fps !== undefined
                        ? `${Number(engine.playerStatus[slot.id].fps).toFixed(0)} FPS`
                        : "LIVE"}
                    </span>
                  {:else}
                    <span class="text-muted-foreground">STANDBY</span>
                  {/if}
                </span>

                <button
                  type="button"
                  onclick={() => toggleAudio(slot.id)}
                  disabled={!slot.attached || !engine.connected}
                  class="flex h-6 w-6 items-center justify-center rounded text-foreground transition hover:bg-accent"
                  aria-label={audibleId !== slot.id ? "Nyalakan suara" : "Bisukan suara"}
                  title={audibleId !== slot.id ? "Nyalakan suara" : "Bisukan"}
                >
                  {#if audibleId !== slot.id}
                    <VolumeX class="h-3.5 w-3.5 text-muted-foreground" />
                  {:else}
                    <Volume2 class="h-3.5 w-3.5 text-sky-400" />
                  {/if}
                </button>

                <button
                  type="button"
                  onclick={() => toggleFocus(slot.id)}
                  class="flex h-6 w-6 items-center justify-center rounded text-foreground transition hover:bg-accent"
                  aria-label={focusId === slot.id ? "Kembali ke 3x2" : "Fokus layar penuh"}
                  title={focusId === slot.id ? "Kembali ke 3x2" : "Fokus"}
                >
                  {#if focusId === slot.id}
                    <Minimize2 class="h-3.5 w-3.5 text-amber-400" />
                  {:else}
                    <Maximize2 class="h-3.5 w-3.5 text-muted-foreground group-hover:text-foreground" />
                  {/if}
                </button>
              </div>
            </header>

            <!-- Video Player Area: Fills remaining height -->
            <div class="relative mx-0.5 min-h-0 flex-1 overflow-hidden bg-muted">
              {#if engine.connected && slot.attached}
                <NativePlayer
                  id={slot.id}
                  zoom={zoomed[slot.id] ?? 100}
                  visible={!focusId || focusId === slot.id}
                  class="h-full w-full"
                />
              {:else if engine.connected}
                <div class="grid h-full place-items-center text-xs text-muted-foreground">Kamera tidak terpasang</div>
              {/if}
            </div>

            <!-- Bottom: Active Audio Meter with Live Waveform Graph -->
            <div class="shrink-0 bg-card pt-0.5">
              {#if engine.connected && slot.attached}
                <ZoomSlider
                  value={zoomed[slot.id] ?? 100}
                  onChange={(value) => (zoomed[slot.id] = value)}
                />
              {/if}
              <AudioMeter metric={engine.metrics[slot.id]} compact />
            </div>
          </article>
        {/each}
      </div>
      {/if}
    </div>
  </div>
</div>
