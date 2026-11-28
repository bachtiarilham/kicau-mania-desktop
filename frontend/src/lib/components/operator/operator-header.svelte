<script lang="ts">
  import {
    Play,
    Square,
    RotateCcw,
    Trophy,
    Cloud,
    CheckCircle2,
  } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { roundStore } from "$lib/state/round.svelte";

  interface Props {
    selectingRound?: boolean;
    onOpenCloudModal: () => void;
    onStopRound: () => void;
  }

  let {
    selectingRound = false,
    onOpenCloudModal,
    onStopRound,
  }: Props = $props();

</script>

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
            {roundStore.eventName || "Belum ada event"}
          </div>
        </div>
      </div>

      <Button
        size="sm"
        onclick={onOpenCloudModal}
        class="h-7 gap-1.5 bg-primary/20 text-primary border border-primary/40 text-xs font-bold hover:bg-primary/30"
        title="Tarik event & jadwal ronde dari database cloud"
      >
        <Cloud class="h-3 w-3" />
        <span>Pilih Event Cloud</span>
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

    </div>

    <!-- Center & Right: Tactical Big START / STOP Buttons & Giant Timer -->
    <div class="flex flex-wrap items-center gap-3">
      <!-- Giant Synchronized Countdown Timer (Dual Reading: 10m Round & 150s Shift) -->
      <div
        class="flex items-center gap-2.5 rounded-xl border px-3 py-1 shadow-inner transition-all {roundStore.state === 'RUNNING' ? 'border-emerald-500/50 bg-emerald-500/15' : roundStore.roundRemainingSeconds === 0 ? 'border-red-500/80 bg-red-500/15 animate-pulse' : 'border-border bg-muted'}"
      >
        <div class="flex flex-col items-center">
          <span class="text-[8px] font-bold tracking-widest text-muted-foreground uppercase">
            Sisa waktu ronde
          </span>
          <div class="flex items-baseline gap-1.5 font-mono">
            <span
              class="text-2xl font-black tracking-tight {roundStore.state === 'RUNNING' ? 'text-emerald-400' : roundStore.roundRemainingSeconds === 0 ? 'text-red-400' : 'text-foreground'}"
            >
              {Math.floor(roundStore.roundRemainingSeconds / 60)}:{String(roundStore.roundRemainingSeconds % 60).padStart(2, "0")}
            </span>
            <span class="text-xs font-bold text-amber-400">
              {roundStore.intervalLabel} · {roundStore.shiftRemainingSeconds}s
            </span>
          </div>
        </div>

        <button
          type="button"
          onclick={() => roundStore.resetRound()}
          disabled={roundStore.busy || !roundStore.roundId || roundStore.state === "RUNNING"}
          class="flex h-6 w-6 items-center justify-center rounded-lg bg-accent text-muted-foreground transition hover:bg-accent hover:text-foreground"
          title="Reset timer ke awal ronde"
        >
          <RotateCcw class="h-3 w-3" />
        </button>
      </div>

      <!-- Modern Tactile START & STOP Buttons -->
      <div class="flex items-center gap-2">
        <button
          type="button"
          disabled={selectingRound || roundStore.busy || !roundStore.roundId || roundStore.state === "RUNNING"}
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

        <button
          type="button"
          disabled={roundStore.busy || roundStore.state !== "RUNNING"}
          onclick={onStopRound}
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
  {#if roundStore.commandError}
    <p class="px-3 pb-2 text-xs text-destructive" role="alert">{roundStore.commandError}</p>
  {/if}
</section>
