<script lang="ts">
  import { CheckCircle2, Circle, PanelLeftClose, PanelLeftOpen } from "@lucide/svelte";
  import type { Snippet } from "svelte";
  import { auth } from "$lib/state/auth.svelte";
  import { getJudgeAssignedZone, getJudgeRotationSchedule, roundStore } from "$lib/state/round.svelte";

  let { collapsed = false, onToggleCollapse, identity, account }: {
    collapsed?: boolean;
    onToggleCollapse: () => void;
    identity?: Snippet<[() => void]>;
    account?: Snippet<[boolean]>;
  } = $props();

  const judgeIndex = $derived(roundStore.judgeIndex);
  const schedule = $derived(judgeIndex === null ? [] : getJudgeRotationSchedule(judgeIndex));
  const assignedZone = $derived(judgeIndex === null ? null : getJudgeAssignedZone(judgeIndex, roundStore.elapsedSeconds));
  const nextZone = $derived(judgeIndex === null || roundStore.nextZoneElapsed >= roundStore.totalDurationSeconds
    ? null
    : getJudgeAssignedZone(judgeIndex, roundStore.nextZoneElapsed));
  const judgeName = $derived((judgeIndex === null ? "" : roundStore.jurors[judgeIndex]) || auth.userName || "—");

  function scheduledTime(value: string) {
    const clock = value.trim().match(/(?:^|T)(\d{1,2}):(\d{2})/);
    return clock ? `${clock[1].padStart(2, "0")}:${clock[2]}` : value.trim() || "—";
  }
</script>

<aside aria-label="Info juri" class="relative z-20 flex h-full shrink-0 flex-col overflow-hidden border-r border-border bg-card {collapsed ? 'w-12' : 'w-56'}">
  {#if collapsed}
    <button type="button" onclick={onToggleCollapse} aria-label="Buka panel juri" title="Buka panel juri" class="mx-auto my-4 rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
      <PanelLeftOpen size={16} />
    </button>
    {#if assignedZone}
      <span class="mx-auto rounded-md bg-accent px-2 py-1 text-xs font-semibold text-primary" title={`Zona ${assignedZone}`}>{assignedZone}</span>
    {/if}
    <div class="flex-1"></div>
  {:else}
    {#if identity}
      {@render identity(onToggleCollapse)}
    {:else}
      <div class="flex items-center justify-between px-5 py-5">
        <span class="text-sm font-semibold">Info juri</span>
        <button type="button" onclick={onToggleCollapse} aria-label="Ringkas panel juri" class="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
          <PanelLeftClose size={16} />
        </button>
      </div>
    {/if}

    <div class="min-h-0 flex-1 space-y-5 overflow-y-auto">
      <section aria-label="Informasi ronde">
        <h2 class="px-5 pb-2 text-[9px] tracking-[.2em] text-muted-foreground uppercase">Info ronde</h2>
        <dl class="space-y-1 px-3 text-xs">
          <div class="rounded-lg bg-accent/65 px-3 py-2">
            <dt class="text-[10px] text-muted-foreground">Nama juri</dt>
            <dd class="break-words font-semibold">{judgeName}</dd>
          </div>
          <div class="rounded-lg px-3 py-2">
            <dt class="text-[10px] text-muted-foreground">Nama event</dt>
            <dd class="break-words font-semibold">{roundStore.eventName || "Menunggu ronde dari operator"}</dd>
            {#if roundStore.sessionName}<dd class="break-words text-muted-foreground">{roundStore.sessionName}</dd>{/if}
          </div>
          <div class="rounded-lg px-3 py-2">
            <dt class="text-[10px] text-muted-foreground">Jadwal ronde</dt>
            <dd class="font-mono font-semibold tabular-nums">{scheduledTime(roundStore.startTime)} – {scheduledTime(roundStore.endTime)}</dd>
          </div>
        </dl>
      </section>

      {#if roundStore.roundId && judgeIndex !== null}
        <section aria-label="Zona penilaian">
          <h2 class="px-5 pb-2 text-[9px] tracking-[.2em] text-muted-foreground uppercase">Zona penilaian</h2>
          <div class="space-y-1 px-3">
            <div class="flex flex-wrap items-center justify-between gap-x-2 gap-y-0.5 rounded-lg bg-accent px-3 py-2 text-xs">
              <strong class="text-primary">Zona {assignedZone}</strong>
              {#if nextZone && roundStore.state === "RUNNING"}
                <span class="text-[10px] text-muted-foreground">Lanjut ke {nextZone}</span>
              {/if}
            </div>
            {#each schedule as item (item.zone)}
              {@const checked = roundStore.isJudgeZoneChecked(roundStore.roundId, judgeIndex, item.zone)}
              <div class="flex items-center justify-between gap-1 rounded-lg px-3 py-1.5 text-xs">
                <span class="flex items-center gap-1.5">
                  {#if checked}
                    <CheckCircle2 class="size-3.5 text-emerald-500" aria-label="Selesai" />
                  {:else}
                    <Circle class="size-3.5 text-muted-foreground" aria-label="Belum selesai" />
                  {/if}
                  Zona {item.zone}
                </span>
                <span class="font-mono text-[10px] text-muted-foreground">{item.startClock}–{item.endClock}</span>
              </div>
            {/each}
          </div>
        </section>
      {/if}
    </div>
  {/if}
  {@render account?.(collapsed)}
</aside>
