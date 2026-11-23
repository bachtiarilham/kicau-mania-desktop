<script lang="ts">
  import {
    Trophy,
    Clock,
    Users,
    Timer as TimerIcon,
    AlertTriangle,
    ChevronLeft,
    ChevronRight,
    Sparkles,
  } from "@lucide/svelte";
  import type { Snippet } from "svelte";
  import { roundStore } from "$lib/state/round.svelte";

  interface Props {
    collapsed?: boolean;
    onToggleCollapse?: () => void;
    identity?: Snippet<[boolean]>;
  }

  let { collapsed = false, onToggleCollapse, identity }: Props = $props();

  // Progress percentage (0 to 1)
  const progress = $derived(
    roundStore.durationSeconds > 0
      ? Math.max(0, Math.min(1, roundStore.remainingSeconds / roundStore.durationSeconds))
      : 0
  );

  const isLowTime = $derived(roundStore.remainingSeconds <= 30 && roundStore.remainingSeconds > 0);
  const isTimeUp = $derived(roundStore.remainingSeconds === 0 || roundStore.rotationAlert);

  const formattedTime = $derived.by(() => {
    const mins = Math.floor(roundStore.remainingSeconds / 60);
    const secs = roundStore.remainingSeconds % 60;
    return `${String(mins).padStart(2, "0")}:${String(secs).padStart(2, "0")}`;
  });
</script>

<aside class="flex h-full shrink-0 flex-col overflow-hidden border-r border-border bg-card {collapsed ? 'w-12' : 'w-72'}">
  {@render identity?.(collapsed)}
{#if collapsed}
  <!-- Mini Collapsed Bar (Pill on Top-Left or Side) -->
  <div
    class="flex min-h-0 flex-1 flex-col items-center justify-between overflow-y-auto py-4"
  >
    <div class="flex flex-col items-center gap-3">
      <button
        type="button"
        onclick={onToggleCollapse}
        class="flex h-8 w-8 items-center justify-center rounded-lg bg-accent text-foreground transition hover:bg-accent"
        title="Buka panel juri"
      >
        <ChevronRight class="h-4 w-4" />
      </button>

      <span class="rounded bg-amber-500/20 p-1.5 text-amber-700">
        <Trophy class="h-4 w-4" />
      </span>
    </div>

    <!-- Vertical Timer Reading -->
    <div class="flex flex-col items-center gap-1 font-mono">
      <TimerIcon
        class="h-4 w-4 {isTimeUp ? 'animate-bounce text-red-700' : isLowTime ? 'animate-pulse text-amber-700' : 'text-emerald-700'}"
      />
      <span
        class="text-xs font-black tracking-wider {isTimeUp ? 'text-red-700 font-bold' : isLowTime ? 'text-amber-700' : 'text-foreground'}"
      >
        {roundStore.remainingSeconds}s
      </span>
    </div>

    <div class="h-4"></div>
  </div>
{:else}
  <!-- Full Modern Broadcast-Grade HUD Sidebar -->
  <div
    class="relative flex min-h-0 flex-1 flex-col justify-between overflow-y-auto p-4 text-foreground"
  >
    <!-- Top: Event & Sesi Header -->
    <div class="space-y-4">
      <div class="flex items-center justify-between border-b border-border pb-3">
        <div class="flex items-center gap-2">
          <div class="flex h-7 w-7 items-center justify-center rounded-lg bg-gradient-to-br from-amber-400 to-amber-600 text-black shadow-md shadow-amber-500/20">
            <Trophy class="h-4 w-4" />
          </div>
          <div>
            <div class="text-[10px] font-bold tracking-widest text-amber-700 uppercase">
              Nama Event
            </div>
            <h2 class="text-sm font-extrabold tracking-tight text-foreground line-clamp-1">
              {roundStore.eventName}
            </h2>
          </div>
        </div>

        {#if onToggleCollapse}
          <button
            type="button"
            onclick={onToggleCollapse}
            class="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition hover:bg-accent hover:text-foreground"
            title="Sembunyikan panel"
          >
            <ChevronLeft class="h-4 w-4" />
          </button>
        {/if}
      </div>

      <!-- Sesi & Waktu Section -->
      <div class="rounded-xl border border-border bg-muted p-3 shadow-inner">
        <div class="flex items-center justify-between">
          <span class="text-xs font-black tracking-wider text-amber-700 uppercase">
            {roundStore.sessionName}
          </span>
          <span
            class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[9px] font-bold tracking-wider uppercase {roundStore.state === 'RUNNING' ? 'bg-emerald-500/20 text-emerald-700 border border-emerald-500/30' : 'bg-accent text-muted-foreground'}"
          >
            <span
              class="h-1.5 w-1.5 rounded-full {roundStore.state === 'RUNNING' ? 'bg-emerald-400 animate-pulse' : 'bg-zinc-500'}"
            ></span>
            {roundStore.state}
          </span>
        </div>

        <div class="mt-2.5 flex items-center justify-between text-xs text-foreground">
          <div class="flex items-center gap-1.5 text-muted-foreground font-mono text-[11px]">
            <Clock class="h-3.5 w-3.5 text-amber-700" />
            <span>Mulai: <strong class="text-foreground">{roundStore.startTime}</strong></span>
          </div>
          <span class="text-muted-foreground">·</span>
          <div class="text-muted-foreground font-mono text-[11px]">
            <span>Selesai: <strong class="text-foreground">{roundStore.endTime}</strong></span>
          </div>
        </div>
      </div>

      <!-- Dewan Juri List -->
      <div class="space-y-1.5">
        <div class="flex items-center gap-1.5 px-1 text-[11px] font-bold tracking-wider text-muted-foreground uppercase">
          <Users class="h-3 w-3 text-sky-700" />
          <span>Dewan Juri :</span>
        </div>

        <div class="space-y-1">
          {#each roundStore.jurors as juri, idx}
            <div
              class="flex items-center gap-2 rounded-lg border border-border bg-muted px-2.5 py-1.5 text-xs text-foreground"
            >
              <span
                class="flex h-5 w-5 shrink-0 items-center justify-center rounded bg-gradient-to-br from-sky-500/30 to-blue-600/30 font-mono text-[10px] font-bold text-sky-700 border border-sky-400/20"
              >
                J{idx + 1}
              </span>
              <span class="truncate font-medium">{juri}</span>
            </div>
          {/each}
        </div>
      </div>
    </div>

    <!-- Bottom: Hero Timer & Rotation Banner -->
    <div class="mt-4 space-y-3 pt-3 border-t border-border">
      <!-- Digital Countdown Display -->
      <div
        class="relative flex flex-col items-center justify-center rounded-2xl border p-4 text-center transition-all duration-300 {isTimeUp ? 'border-red-500/80 bg-red-50 shadow-lg shadow-red-900/30 animate-pulse' : isLowTime ? 'border-amber-500/50 bg-amber-50' : 'border-border bg-muted'}"
      >
        <span class="text-[11px] font-bold tracking-widest text-amber-700 uppercase">
          Timer Penilaian
        </span>

        <!-- Giant Number Display -->
        <div class="my-1 font-mono text-5xl font-black tracking-tighter sm:text-6xl {isTimeUp ? 'text-red-700 animate-pulse' : isLowTime ? 'text-amber-700' : 'text-foreground'}">
          {roundStore.remainingSeconds}
        </div>

        <!-- Progress Bar -->
        <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-accent">
          <div
            class="h-full rounded-full transition-all duration-300 {isTimeUp ? 'bg-red-500' : isLowTime ? 'bg-amber-400' : 'bg-emerald-400'}"
            style:width={`${progress * 100}%`}
          ></div>
        </div>

        <div class="mt-2 flex w-full items-center justify-between font-mono text-[10px] text-muted-foreground">
          <span>Format: {formattedTime}</span>
          <span>Jika Timer: 00</span>
        </div>
      </div>

      <!-- Rotation Alert: Prompt judges to rotate to next zone -->
      {#if isTimeUp}
        <div
          class="flex items-center gap-2 rounded-xl border border-red-500 bg-red-600/20 p-3 text-red-700 shadow-md shadow-red-600/20 animate-bounce"
        >
          <AlertTriangle class="h-5 w-5 shrink-0 text-red-700" />
          <div class="text-[11px] font-bold leading-tight">
            Silahkan Pindah ke Area Berikutnya!
          </div>
        </div>
      {:else}
        <div
          class="flex items-center justify-center gap-1.5 rounded-xl border border-border bg-accent py-2 text-center text-[10px] font-medium text-muted-foreground"
        >
          <Sparkles class="h-3 w-3 text-amber-700" />
          <span>Rotasi juri saat timer mencapai 00</span>
        </div>
      {/if}
    </div>
  </div>
{/if}
</aside>
