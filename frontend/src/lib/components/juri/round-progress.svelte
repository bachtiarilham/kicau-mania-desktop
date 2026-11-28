<script lang="ts">
  import { Check, Clock } from "@lucide/svelte";
  import { getRoundProgress, roundMilestones, roundStore } from "$lib/state/round.svelte";

  const stages = ["Mulai", "Shift 2", "Shift 3", "Shift 4", "Putaran akhir", "Selesai"];
  const progress = $derived(getRoundProgress(roundStore.elapsedSeconds, roundStore.state));
  const phase = $derived(roundStore.state === "IDLE" ? "Menunggu operator"
    : roundStore.state === "FINISHED" ? "Ronde selesai"
    : roundStore.roundRemainingSeconds === 0 ? "Waktu habis · menunggu operator"
    : roundStore.intervalLabel);

  function clock(seconds: number) {
    return `${String(Math.floor(seconds / 60)).padStart(2, "0")}:${String(seconds % 60).padStart(2, "0")}`;
  }
</script>

<section class="shrink-0 rounded-xl border border-border bg-card px-2.5 py-2.5 sm:px-4" aria-label="Waktu dan progres ronde">
  <div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-1 px-1 pb-2">
    <div class="flex items-center gap-2 text-xs">
      <span class="font-semibold text-foreground">{phase}</span>
      {#if roundStore.state === "RUNNING" && roundStore.roundRemainingSeconds > 0}
        <span class="rounded-md bg-muted px-2 py-0.5 font-mono tabular-nums text-muted-foreground" title="Sisa waktu zona">
          Zona {clock(roundStore.shiftRemainingSeconds)}
        </span>
      {/if}
    </div>
    <div class="flex items-center gap-2">
      <Clock class="size-3.5 text-primary" aria-hidden="true" />
      <span class="text-xs text-muted-foreground">Sisa ronde</span>
      <span class="font-mono text-xl font-bold leading-none tabular-nums text-foreground">{clock(roundStore.roundRemainingSeconds)}</span>
    </div>
  </div>

  <div class="relative">
    <div class="absolute inset-x-[8.333333%] top-2 h-0.5 -translate-y-1/2 rounded-full bg-border">
      <div
        role="progressbar"
        aria-label="Waktu ronde yang berjalan"
        aria-valuemin={0}
        aria-valuemax={roundStore.totalDurationSeconds}
        aria-valuenow={progress.elapsedSeconds}
        class="h-full rounded-full bg-primary transition-[width] duration-300 motion-reduce:transition-none"
        style:width={`${progress.percent}%`}
      ></div>
    </div>

    <ol aria-label="Tahap ronde" class="relative grid grid-cols-6">
      {#each roundMilestones as milestone, index (milestone)}
        {@const active = index === progress.activeStep}
        {@const done = index < progress.activeStep}
        <li aria-current={active ? "step" : undefined} class="flex min-w-0 flex-col items-center gap-1 text-center">
          <span class="grid size-4 place-items-center rounded-full border bg-card {active ? 'border-primary ring-2 ring-primary/20' : done ? 'border-primary/50 text-primary' : 'border-border'}">
            {#if done}
              <Check class="size-2.5" aria-label="Tahap terlewati" />
            {:else}
              <span aria-hidden="true" class="size-1.5 rounded-full {active ? 'bg-primary' : 'bg-muted-foreground/40'}"></span>
            {/if}
          </span>
          <span class="text-[9px] leading-tight sm:text-[11px] {active ? 'font-bold text-primary' : 'text-muted-foreground'}">
            {stages[index]}
          </span>
          <span class="font-mono text-[9px] leading-none tabular-nums text-muted-foreground">{clock(milestone)}</span>
        </li>
      {/each}
    </ol>
  </div>
</section>
