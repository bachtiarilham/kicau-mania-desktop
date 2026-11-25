<script lang="ts">
  import { Trophy } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import * as Dialog from "$lib/components/ui/dialog";
  import type { RoundConfig } from "$lib/models";

  let { open, initialConfig, onClose }: {
    open: boolean;
    initialConfig: RoundConfig;
    onClose: () => void;
  } = $props();

  const fields = $derived([
    ["Nama event", initialConfig.eventName],
    ["Babak / sesi", initialConfig.sessionName],
    ["Jam mulai", initialConfig.startTime],
    ["Jam selesai", initialConfig.endTime],
    ["Durasi ronde", initialConfig.totalDurationSeconds ? initialConfig.totalDurationSeconds + " detik" : ""],
    ["Durasi shift", initialConfig.shiftDurationSeconds ? initialConfig.shiftDurationSeconds + " detik" : ""],
  ]);
</script>

<Dialog.Root {open} onOpenChange={(value) => { if (!value) onClose(); }}>
  <Dialog.Content showCloseButton={false} class="max-h-[90vh] gap-0 overflow-y-auto">
      <header class="flex items-center gap-2 border-b border-border pb-3">
        <Trophy class="size-5 text-amber-400" />
        <Dialog.Title>Konfigurasi Lomba & Dewan Juri</Dialog.Title>
      </header>
      {#if initialConfig.roundId}
        <Dialog.Description class="mt-3 text-xs">Data sesi cloud. Perubahan dilakukan di cloud-app.</Dialog.Description>
        <dl class="mt-4 grid grid-cols-2 gap-3">
          {#each fields as [label, value]}
            <div class={label === "Nama event" ? "col-span-2" : ""}>
              <dt class="text-xs text-muted-foreground">{label}</dt>
              <dd class="mt-1 break-words rounded-lg bg-muted px-3 py-2 text-sm">{value || "—"}</dd>
            </div>
          {/each}
        </dl>
        <h4 class="mt-4 text-xs font-semibold">Dewan juri</h4>
        <ul class="mt-2 space-y-1 text-sm">
          {#each initialConfig.jurors as juror, index}
            <li class="rounded-lg bg-muted px-3 py-2"><span class="mr-2 text-xs text-muted-foreground">Juri {index + 1}</span>{juror || "—"}</li>
          {:else}
            <li class="text-xs text-muted-foreground">Belum ada juri yang ditetapkan di cloud.</li>
          {/each}
        </ul>
      {:else}
        <Dialog.Description class="py-5">Pilih sesi cloud untuk melihat konfigurasi lomba dan dewan juri.</Dialog.Description>
      {/if}
      <footer class="mt-5 flex justify-end border-t border-border pt-3">
        <Button size="sm" variant="outline" onclick={onClose}>Tutup</Button>
      </footer>
  </Dialog.Content>
</Dialog.Root>
