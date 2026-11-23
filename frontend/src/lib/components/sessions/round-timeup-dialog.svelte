<script lang="ts">
  import {
    AlertCircle,
    CheckCircle2,
    Clock,
    HardDrive,
    Play,
    Square,
    Video,
  } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import * as Dialog from "$lib/components/ui/alert-dialog";
  import { roundStore } from "$lib/state/round.svelte";
  import { engine } from "$lib/state/engine.svelte";

  interface Props {
    onStopAndSave: () => void;
  }

  let { onStopAndSave }: Props = $props();

  const formattedOvertime = $derived.by(() => {
    const mins = Math.floor(roundStore.overtimeSeconds / 60);
    const secs = roundStore.overtimeSeconds % 60;
    return `${String(mins).padStart(2, "0")}:${String(secs).padStart(2, "0")}`;
  });
</script>

<Dialog.Root open={roundStore.timeUpDialog}>
  <Dialog.Content class="max-w-md bg-card border-border text-foreground">
    <Dialog.Header>
      <Dialog.Title class="text-base font-bold flex items-center gap-2 text-foreground">
        <Clock class="h-5 w-5 text-amber-600" />
        Waktu Ronde (10 Menit) Telah Selesai!
      </Dialog.Title>
      <Dialog.Description class="text-xs text-muted-foreground">
        Waktu ronde resmi 10 menit (600 detik) telah tercapai. Rekaman 24 kamera saat ini <strong>tetap aktif merekam</strong> agar tidak ada momen lapangan yang terputus.
      </Dialog.Description>
    </Dialog.Header>

    <div class="space-y-3.5 py-2">
      <!-- Status Box -->
      <div class="rounded-xl border border-amber-500/30 bg-amber-500/10 p-3 text-xs">
        <div class="flex items-center justify-between">
          <span class="font-bold text-amber-700">Status Sesi Lapangan:</span>
          <span class="inline-flex items-center gap-1.5 font-mono text-[11px] font-bold text-rose-600">
            <span class="h-2 w-2 rounded-full bg-rose-500 animate-pulse"></span>
            MEREKAM ({engine.channels.length} Kamera)
          </span>
        </div>
        <div class="mt-2 flex items-center justify-between text-muted-foreground font-mono text-[11px]">
          <span>Overtime Waktu:</span>
          <span class="font-bold text-foreground">+{formattedOvertime}</span>
        </div>
      </div>

      <p class="text-xs text-foreground font-medium">
        Silakan tentukan keputusan operasional di lapangan:
      </p>

      <!-- Action Choice 1: Stop and Save (Red/Destructive) -->
      <button
        type="button"
        onclick={() => {
          roundStore.timeUpDialog = false;
          onStopAndSave();
        }}
        class="w-full flex items-center justify-between rounded-xl border border-red-500/40 bg-red-600/10 p-3 text-left transition hover:bg-red-600/20 active:scale-98"
      >
        <div class="flex items-center gap-3">
          <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-red-600 text-white shadow-sm">
            <Square class="h-4 w-4 fill-current" />
          </div>
          <div>
            <div class="text-xs font-bold text-red-700">STOP & SIMPAN REKAMAN</div>
            <div class="text-[10px] text-muted-foreground">Akhiri ronde sekarang & ekspor rekaman ke folder ronde</div>
          </div>
        </div>
      </button>

      <!-- Action Choice 2: Add Extension Time or Continue -->
      <div class="space-y-1.5 pt-1">
        <span class="text-[11px] font-semibold text-muted-foreground">Atau Lanjutkan Ronde:</span>
        <div class="grid grid-cols-3 gap-2">
          <Button
            variant="outline"
            size="sm"
            onclick={() => roundStore.continueRound(60)}
            class="text-xs h-8"
          >
            +1 Menit
          </Button>
          <Button
            variant="outline"
            size="sm"
            onclick={() => roundStore.continueRound(120)}
            class="text-xs h-8"
          >
            +2 Menit
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onclick={() => roundStore.continueRound(0)}
            class="text-xs h-8 text-muted-foreground hover:text-foreground"
          >
            Bebas
          </Button>
        </div>
      </div>
    </div>

    <Dialog.Footer class="pt-2">
      <Dialog.Cancel
        class="text-xs text-muted-foreground hover:text-foreground"
        onclick={() => roundStore.continueRound(0)}
      >
        Lanjutkan Ronde Saja (Tutup)
      </Dialog.Cancel>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
