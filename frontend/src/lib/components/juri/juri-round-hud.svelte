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
  import { roundStore } from "$lib/state/round.svelte";

  interface Props {
    collapsed?: boolean;
    onToggleCollapse?: () => void;
  }

  let { collapsed = false, onToggleCollapse }: Props = $props();

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

{#if collapsed}
  <!-- Mini Collapsed Bar (Pill on Top-Left or Side) -->
  <aside
    class="flex h-full w-12 shrink-0 flex-col items-center justify-between border-r border-white/10 bg-[#0d1214]/95 py-4 backdrop-blur-md select-none"
  >
    <div class="flex flex-col items-center gap-3">
      <button
        type="button"
        onclick={onToggleCollapse}
        class="flex h-8 w-8 items-center justify-center rounded-lg bg-white/5 text-zinc-300 transition hover:bg-white/15"
        title="Buka panel juri"
      >
        <ChevronRight class="h-4 w-4" />
      </button>

      <span class="rounded bg-amber-500/20 p-1.5 text-amber-400">
        <Trophy class="h-4 w-4" />
      </span>
    </div>

    <!-- Vertical Timer Reading -->
    <div class="flex flex-col items-center gap-1 font-mono">
      <TimerIcon
        class="h-4 w-4 {isTimeUp ? 'animate-bounce text-red-500' : isLowTime ? 'animate-pulse text-amber-400' : 'text-emerald-400'}"
      />
      <span
        class="text-xs font-black tracking-wider {isTimeUp ? 'text-red-500 font-bold' : isLowTime ? 'text-amber-400' : 'text-white'}"
      >
        {roundStore.remainingSeconds}s
      </span>
    </div>

    <div class="h-4"></div>
  </aside>
{:else}
  <!-- Full Modern Broadcast-Grade HUD Sidebar -->
  <aside
    class="relative flex h-full w-72 shrink-0 flex-col justify-between overflow-y-auto border-r border-white/10 bg-[#0f1517]/95 p-4 text-zinc-100 backdrop-blur-md select-none"
  >
    <!-- Top: Event & Sesi Header -->
    <div class="space-y-4">
      <div class="flex items-center justify-between border-b border-white/10 pb-3">
        <div class="flex items-center gap-2">
          <div class="flex h-7 w-7 items-center justify-center rounded-lg bg-gradient-to-br from-amber-400 to-amber-600 text-black shadow-md shadow-amber-500/20">
            <Trophy class="h-4 w-4" />
          </div>
          <div>
            <div class="text-[10px] font-bold tracking-widest text-amber-400 uppercase">
              Nama Event
            </div>
            <h2 class="text-sm font-extrabold tracking-tight text-white line-clamp-1">
              {roundStore.eventName}
            </h2>
          </div>
        </div>

        {#if onToggleCollapse}
          <button
            type="button"
            onclick={onToggleCollapse}
            class="flex h-7 w-7 items-center justify-center rounded-md text-zinc-400 transition hover:bg-white/10 hover:text-white"
            title="Sembunyikan panel"
          >
            <ChevronLeft class="h-4 w-4" />
          </button>
        {/if}
      </div>

      <!-- Sesi & Waktu Section -->
      <div class="rounded-xl border border-white/10 bg-black/40 p-3 shadow-inner">
        <div class="flex items-center justify-between">
          <span class="text-xs font-black tracking-wider text-amber-300 uppercase">
            {roundStore.sessionName}
          </span>
          <span
            class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[9px] font-bold tracking-wider uppercase {roundStore.state === 'RUNNING' ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' : 'bg-white/10 text-zinc-400'}"
          >
            <span
              class="h-1.5 w-1.5 rounded-full {roundStore.state === 'RUNNING' ? 'bg-emerald-400 animate-pulse' : 'bg-zinc-500'}"
            ></span>
            {roundStore.state}
          </span>
        </div>

        <div class="mt-2.5 flex items-center justify-between text-xs text-zinc-300">
          <div class="flex items-center gap-1.5 text-zinc-400 font-mono text-[11px]">
            <Clock class="h-3.5 w-3.5 text-amber-400" />
            <span>Mulai: <strong class="text-zinc-200">{roundStore.startTime}</strong></span>
          </div>
          <span class="text-zinc-600">·</span>
          <div class="text-zinc-400 font-mono text-[11px]">
            <span>Selesai: <strong class="text-zinc-200">{roundStore.endTime}</strong></span>
          </div>
        </div>
      </div>

      <!-- Dewan Juri List -->
      <div class="space-y-1.5">
        <div class="flex items-center gap-1.5 px-1 text-[11px] font-bold tracking-wider text-zinc-400 uppercase">
          <Users class="h-3 w-3 text-sky-400" />
          <span>Dewan Juri :</span>
        </div>

        <div class="space-y-1">
          {#each roundStore.jurors as juri, idx}
            <div
              class="flex items-center gap-2 rounded-lg border border-white/5 bg-black/30 px-2.5 py-1.5 text-xs text-zinc-200"
            >
              <span
                class="flex h-5 w-5 shrink-0 items-center justify-center rounded bg-gradient-to-br from-sky-500/30 to-blue-600/30 font-mono text-[10px] font-bold text-sky-300 border border-sky-400/20"
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
    <div class="mt-4 space-y-3 pt-3 border-t border-white/10">
      <!-- Digital Countdown Display -->
      <div
        class="relative flex flex-col items-center justify-center rounded-2xl border p-4 text-center transition-all duration-300 {isTimeUp ? 'border-red-500/80 bg-red-950/40 shadow-lg shadow-red-900/30 animate-pulse' : isLowTime ? 'border-amber-500/50 bg-amber-950/30' : 'border-white/10 bg-black/50'}"
      >
        <span class="text-[11px] font-bold tracking-widest text-amber-400 uppercase">
          Timer Penilaian
        </span>

        <!-- Giant Number Display -->
        <div class="my-1 font-mono text-5xl font-black tracking-tighter sm:text-6xl {isTimeUp ? 'text-red-500 animate-pulse' : isLowTime ? 'text-amber-400' : 'text-white'}">
          {roundStore.remainingSeconds}
        </div>

        <!-- Progress Bar -->
        <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-white/10">
          <div
            class="h-full rounded-full transition-all duration-300 {isTimeUp ? 'bg-red-500' : isLowTime ? 'bg-amber-400' : 'bg-emerald-400'}"
            style:width={`${progress * 100}%`}
          ></div>
        </div>

        <div class="mt-2 flex w-full items-center justify-between font-mono text-[10px] text-zinc-400">
          <span>Format: {formattedTime}</span>
          <span>Jika Timer: 00</span>
        </div>
      </div>

      <!-- Rotation Alert: Prompt judges to rotate to next zone -->
      {#if isTimeUp}
        <div
          class="flex items-center gap-2 rounded-xl border border-red-500 bg-red-600/20 p-3 text-red-200 shadow-md shadow-red-600/20 animate-bounce"
        >
          <AlertTriangle class="h-5 w-5 shrink-0 text-red-400" />
          <div class="text-[11px] font-bold leading-tight">
            Silahkan Pindah ke Area Berikutnya!
          </div>
        </div>
      {:else}
        <div
          class="flex items-center justify-center gap-1.5 rounded-xl border border-white/5 bg-white/5 py-2 text-center text-[10px] font-medium text-zinc-400"
        >
          <Sparkles class="h-3 w-3 text-amber-400" />
          <span>Rotasi juri saat timer mencapai 00</span>
        </div>
      {/if}
    </div>
  </aside>
{/if}
