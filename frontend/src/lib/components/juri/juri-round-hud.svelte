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
    CheckCircle2,
    ArrowRight,
  } from "@lucide/svelte";
  import type { Snippet } from "svelte";
  import { roundStore, getJudgeAssignedZone } from "$lib/state/round.svelte";
  import { auth } from "$lib/state/auth.svelte";

  interface Props {
    collapsed?: boolean;
    onToggleCollapse?: () => void;
    identity?: Snippet<[boolean]>;
  }

  let { collapsed = false, onToggleCollapse, identity }: Props = $props();

  const ZONE_INFO: Record<"A" | "B" | "C" | "D", { name: string; range: string }> = {
    A: { name: "Zona A", range: "G01 – G06" },
    B: { name: "Zona B", range: "G07 – G12" },
    C: { name: "Zona C", range: "G13 – G18" },
    D: { name: "Zona D", range: "G19 – G24" },
  };

  // Current judge's assigned zone
  const myAssignedZone = $derived(
    getJudgeAssignedZone(auth.judgeIndex, roundStore.currentShift)
  );

  const nextAssignedZone = $derived(
    getJudgeAssignedZone(auth.judgeIndex, Math.min(4, roundStore.currentShift + 1))
  );

  // Shift countdown progress (0 to 1)
  const shiftProgress = $derived(
    roundStore.shiftDurationSeconds > 0
      ? Math.max(0, Math.min(1, roundStore.shiftRemainingSeconds / roundStore.shiftDurationSeconds))
      : 0
  );

  // Round countdown progress (0 to 1)
  const roundProgress = $derived(
    roundStore.totalDurationSeconds > 0
      ? Math.max(0, Math.min(1, roundStore.roundRemainingSeconds / roundStore.totalDurationSeconds))
      : 0
  );

  const isLowShiftTime = $derived(
    roundStore.shiftRemainingSeconds <= 30 && roundStore.shiftRemainingSeconds > 0
  );
  const isShiftTimeUp = $derived(
    roundStore.shiftRemainingSeconds === 0 || roundStore.rotationAlert
  );
  const isRoundFinished = $derived(
    roundStore.state === "FINISHED" || roundStore.roundRemainingSeconds === 0
  );

  const formattedShiftTime = $derived.by(() => {
    const mins = Math.floor(roundStore.shiftRemainingSeconds / 60);
    const secs = roundStore.shiftRemainingSeconds % 60;
    return `${String(mins).padStart(2, "0")}:${String(secs).padStart(2, "0")}`;
  });

  const formattedRoundTime = $derived.by(() => {
    const mins = Math.floor(roundStore.roundRemainingSeconds / 60);
    const secs = roundStore.roundRemainingSeconds % 60;
    return `${String(mins).padStart(2, "0")}:${String(secs).padStart(2, "0")}`;
  });
</script>

<aside class="flex h-full shrink-0 flex-col overflow-hidden border-r border-border bg-card {collapsed ? 'w-12' : 'w-80'}">
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

      <!-- Mini Zone Badge -->
      <span class="rounded bg-primary/20 px-1 py-0.5 text-[10px] font-black text-primary">
        {myAssignedZone}
      </span>
    </div>

    <!-- Vertical Timer Reading -->
    <div class="flex flex-col items-center gap-1 font-mono">
      <TimerIcon
        class="h-4 w-4 {isShiftTimeUp ? 'animate-bounce text-red-700' : isLowShiftTime ? 'animate-pulse text-amber-700' : 'text-emerald-700'}"
      />
      <span
        class="text-xs font-black tracking-wider {isShiftTimeUp ? 'text-red-700 font-bold' : isLowShiftTime ? 'text-amber-700' : 'text-foreground'}"
      >
        {roundStore.shiftRemainingSeconds}s
      </span>
      <span class="text-[9px] text-muted-foreground">S{roundStore.currentShift}/4</span>
    </div>

    <div class="h-4"></div>
  </div>
{:else}
  <!-- Full Modern Broadcast-Grade HUD Sidebar -->
  <div
    class="relative flex min-h-0 flex-1 flex-col justify-between overflow-y-auto p-4 text-foreground"
  >
    <!-- Top: Event & Sesi Header -->
    <div class="space-y-3.5">
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

      <!-- Sesi, Status, & Shift Section -->
      <div class="rounded-xl border border-border bg-muted p-3 shadow-inner space-y-2">
        <div class="flex items-center justify-between">
          <span class="text-xs font-black tracking-wider text-amber-700 uppercase">
            {roundStore.sessionName} (10 MENIT)
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

        <!-- Shift & Total Round Mini Bar -->
        <div class="flex items-center justify-between text-[11px] text-muted-foreground">
          <span class="font-semibold text-foreground">
            Shift {roundStore.currentShift} dari 4
          </span>
          <span class="font-mono text-[11px]">
            Total: <strong class="text-foreground">{formattedRoundTime}</strong> / 10:00
          </span>
        </div>
        <div class="h-1 w-full overflow-hidden rounded-full bg-accent">
          <div
            class="h-full rounded-full bg-primary transition-all duration-300"
            style:width={`${(1 - roundProgress) * 100}%`}
          ></div>
        </div>
      </div>

      <!-- Active Assigned Zone Banner for Current Judge -->
      <div class="rounded-xl border-2 border-primary/40 bg-primary/10 p-3 shadow-sm">
        <div class="flex items-center justify-between">
          <span class="text-[10px] font-bold tracking-widest text-primary uppercase">
            Area Penilaian Anda
          </span>
          <span class="rounded bg-primary px-1.5 py-0.5 font-mono text-[10px] font-extrabold text-primary-foreground">
            JURI {auth.judgeIndex + 1}
          </span>
        </div>
        <div class="mt-1 flex items-baseline justify-between">
          <div class="text-lg font-black tracking-tight text-foreground">
            {ZONE_INFO[myAssignedZone].name}
          </div>
          <div class="font-mono text-xs font-bold text-primary">
            {ZONE_INFO[myAssignedZone].range}
          </div>
        </div>
        <div class="mt-1 flex items-center justify-between text-[10px] text-muted-foreground">
          <span>Shift {roundStore.currentShift} · 150 Detik</span>
          {#if roundStore.currentShift < 4}
            <span class="flex items-center gap-1 text-[10px]">
              Berikutnya: <strong class="text-foreground">{nextAssignedZone}</strong>
            </span>
          {:else}
            <span class="text-emerald-700 font-semibold">Shift Terakhir</span>
          {/if}
        </div>
      </div>

      <!-- Dewan Juri & Rotasi Matrix -->
      <div class="space-y-1.5">
        <div class="flex items-center justify-between px-1 text-[11px] font-bold tracking-wider text-muted-foreground uppercase">
          <span class="flex items-center gap-1.5">
            <Users class="h-3 w-3 text-sky-700" />
            <span>4 Dewan Juri & Rotasi</span>
          </span>
          <span class="text-[9px] lowercase font-normal">klik untuk pilih</span>
        </div>

        <div class="space-y-1">
          {#each roundStore.jurors as juri, idx}
            {@const zone = getJudgeAssignedZone(idx, roundStore.currentShift)}
            {@const isMe = auth.judgeIndex === idx}
            <button
              type="button"
              onclick={() => auth.setJudgeIndex(idx)}
              class="w-full flex items-center justify-between rounded-lg border px-2.5 py-1.5 text-xs text-left transition {isMe ? 'border-primary bg-primary/15 font-semibold text-foreground ring-1 ring-primary/40' : 'border-border bg-muted text-muted-foreground hover:bg-accent hover:text-foreground'}"
              title="Pilih juri ini sebagai identitas Anda"
            >
              <div class="flex min-w-0 items-center gap-2">
                <span
                  class="flex h-5 w-5 shrink-0 items-center justify-center rounded font-mono text-[10px] font-bold border {isMe ? 'bg-primary text-primary-foreground border-primary' : 'bg-gradient-to-br from-sky-500/20 to-blue-600/20 text-sky-700 border-sky-400/20'}"
                >
                  J{idx + 1}
                </span>
                <span class="truncate font-medium">{juri}</span>
              </div>
              <span class="shrink-0 rounded px-1.5 py-0.5 font-mono text-[10px] font-bold {isMe ? 'bg-primary text-primary-foreground' : 'bg-accent text-foreground'}">
                Zona {zone}
              </span>
            </button>
          {/each}
        </div>
      </div>
    </div>

    <!-- Bottom: Hero Shift Timer & Prominent Rotation Alert -->
    <div class="mt-4 space-y-3 pt-3 border-t border-border">
      <!-- Digital Countdown Display for Area (150s) -->
      <div
        class="relative flex flex-col items-center justify-center rounded-2xl border p-4 text-center transition-all duration-300 {isRoundFinished ? 'border-emerald-500 bg-emerald-50' : isShiftTimeUp ? 'border-red-500/80 bg-red-50 shadow-lg shadow-red-900/30 animate-pulse' : isLowShiftTime ? 'border-amber-500/50 bg-amber-50' : 'border-border bg-muted'}"
      >
        <span class="text-[11px] font-bold tracking-widest text-amber-700 uppercase">
          {isRoundFinished ? 'Ronde Selesai' : `Timer Area · Shift ${roundStore.currentShift}/4`}
        </span>

        <!-- Giant Number Display (Shift Countdown) -->
        <div class="my-1 font-mono text-5xl font-black tracking-tighter sm:text-6xl {isRoundFinished ? 'text-emerald-700' : isShiftTimeUp ? 'text-red-700 animate-pulse' : isLowShiftTime ? 'text-amber-700' : 'text-foreground'}">
          {roundStore.shiftRemainingSeconds}
        </div>

        <!-- Shift Progress Bar -->
        <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-accent">
          <div
            class="h-full rounded-full transition-all duration-300 {isRoundFinished ? 'bg-emerald-500' : isShiftTimeUp ? 'bg-red-500' : isLowShiftTime ? 'bg-amber-400' : 'bg-emerald-400'}"
            style:width={`${shiftProgress * 100}%`}
          ></div>
        </div>

        <div class="mt-2 flex w-full items-center justify-between font-mono text-[10px] text-muted-foreground">
          <span>Sisa: {formattedShiftTime}</span>
          <span>Rotasi tiap 150 dtk</span>
        </div>
      </div>

      <!-- Rotation Alert: Prompt judges to rotate to next zone -->
      {#if isRoundFinished}
        <div
          class="flex items-center gap-2 rounded-xl border border-emerald-500 bg-emerald-500/20 p-3 text-emerald-800 shadow-md shadow-emerald-600/20"
        >
          <CheckCircle2 class="h-5 w-5 shrink-0 text-emerald-700" />
          <div class="text-[11px] font-bold leading-tight">
            Ronde 10 Menit Selesai! Rekaman otomatis disimpan.
          </div>
        </div>
      {:else if isShiftTimeUp}
        <div
          class="flex items-center gap-2 rounded-xl border border-red-500 bg-red-600/20 p-3 text-red-700 shadow-md shadow-red-600/20 animate-bounce"
        >
          <AlertTriangle class="h-5 w-5 shrink-0 text-red-700" />
          <div class="text-[11px] font-bold leading-tight">
            Waktu 150s Habis! Pindah ke {ZONE_INFO[nextAssignedZone].name} ({ZONE_INFO[nextAssignedZone].range})!
          </div>
        </div>
      {:else}
        <div
          class="flex items-center justify-center gap-1.5 rounded-xl border border-border bg-accent py-2 text-center text-[10px] font-medium text-muted-foreground"
        >
          <Sparkles class="h-3 w-3 text-amber-700" />
          <span>Rotasi ke area berikutnya saat timer 150s habis</span>
        </div>
      {/if}
    </div>
  </div>
{/if}
</aside>
