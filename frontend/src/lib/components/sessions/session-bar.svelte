<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import * as Dialog from "$lib/components/ui/alert-dialog";
  import { Circle, Square, RefreshCw } from "@lucide/svelte";
  import { api } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  let confirm = $state(false);
  let starting = $state(false);
  let finalizing = $derived(["FINALIZING", "STOPPING"].includes(engine.session?.state));
  let retryFinalization = $derived(finalizing && !!engine.session?.error);
  let now = $state(Date.now());
  import { onMount } from "svelte";
  onMount(() => {
    const t = setInterval(() => (now = Date.now()), 1000);
    return () => clearInterval(t);
  });
  let elapsed = $derived(
    engine.session?.started_at
      ? Math.max(
          0,
          Math.floor(
            ((engine.session.stopped_at
              ? new Date(engine.session.stopped_at).getTime()
              : now) -
              new Date(engine.session.started_at).getTime()) /
              1000,
          ),
        )
      : 0,
  );
  $effect(() => {
    engine.overlay = confirm;
  });
  async function start() {
    if (starting || engine.busy) return;
    starting = true;
    try {
      if (!engine.draftID) engine.draftID = await api.sessionID();
      await engine.start(engine.draftID);
      engine.draftID = "";
    } catch {} finally { starting = false; }
  }
</script>

<div
  class="flex flex-wrap items-center justify-between gap-4 border-b border-border bg-[#182120] px-6 py-3"
>
  <div class="flex items-center gap-3">
    <Button variant="ghost" size="icon-sm" aria-label="Perbarui status sesi"
      disabled={engine.busy || starting} onclick={() => engine.refresh()}><RefreshCw size={13} /></Button>
    <Circle
      size={10}
      class={engine.running
        ? "fill-primary text-primary"
        : "text-muted-foreground"}
    />
    <div class="text-xs font-medium">{engine.session?.state || "STANDBY"}</div>
    <span class="font-mono text-xs text-muted-foreground"
      >{String(Math.floor(elapsed / 60)).padStart(2, "0")}:{String(
        elapsed % 60,
      ).padStart(2, "0")}</span
    ><span class="text-[10px] text-muted-foreground"
      >{engine.channels.length} kamera · RTSP/TCP</span
    >
  </div>
  <div class="flex items-center gap-3">
    <span class="text-[10px] text-muted-foreground"
      >{engine.pairing.paired
        ? "Cloud paired"
        : "Saving lokal · cloud belum paired"}</span
    >{#if engine.active}<Button
        variant="outline"
        size="sm"
        disabled={engine.busy ||
          (finalizing && !retryFinalization)}
        onclick={() => {
          void api.hidePlayers();
          confirm = true;
        }}><Square size={12} /> {retryFinalization ? "Ulangi finalisasi" : finalizing ? "Menunggu finalisasi…" : "Stop & finalisasi"}</Button
      >{:else}<Button
        size="sm"
        disabled={engine.busy || starting || !engine.connected || !engine.channels.length}
        onclick={start}
        >{#if engine.busy || starting}<RefreshCw
            size={13}
            class="animate-spin"
          />{:else}<Circle size={12} />{/if} Mulai sesi</Button
      >{/if}
  </div>
</div>
{#if engine.session?.error}
  <div class="error mx-6 my-2" role="alert">{engine.session.error}</div>
{/if}
<Dialog.Root bind:open={confirm}
  ><Dialog.Content
    ><Dialog.Header
      ><Dialog.Title>Hentikan sesi dan simpan rekaman?</Dialog.Title
      ><Dialog.Description
        >Local-app menghentikan media lalu memfinalisasi MKV. Upload otomatis
        dilanjutkan setelah file valid.</Dialog.Description
      ></Dialog.Header
    ><Dialog.Footer
      ><Dialog.Cancel>Lanjutkan sesi</Dialog.Cancel><Dialog.Action
        onclick={() => engine.stop()}>Stop & simpan</Dialog.Action
      ></Dialog.Footer
    ></Dialog.Content
  ></Dialog.Root
>
