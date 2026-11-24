<script lang="ts">
  import { Trophy } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import type { RoundConfig } from "$lib/models";

  interface Props {
    open: boolean;
    initialConfig: RoundConfig;
    onClose: () => void;
    onSave: (config: RoundConfig) => void;
  }

  let { open, initialConfig, onClose, onSave }: Props = $props();

  let editEventName = $state("");
  let editSessionName = $state("");
  let editStartTime = $state("");
  let editEndTime = $state("");
  let editTotalDuration = $state(600);
  let editShiftDuration = $state(150);
  let editJurors = $state<string[]>([]);

  $effect(() => {
    if (open) {
      editEventName = initialConfig.eventName;
      editSessionName = initialConfig.sessionName;
      editStartTime = initialConfig.startTime;
      editEndTime = initialConfig.endTime;
      editTotalDuration = initialConfig.totalDurationSeconds ?? 600;
      editShiftDuration = initialConfig.shiftDurationSeconds ?? 150;
      editJurors = [...initialConfig.jurors];
    }
  });

  function handleSave() {
    onSave({
      eventName: editEventName,
      sessionName: editSessionName,
      startTime: editStartTime,
      endTime: editEndTime,
      totalDurationSeconds: Number(editTotalDuration) || 600,
      shiftDurationSeconds: Number(editShiftDuration) || 150,
      jurors: editJurors,
    });
  }
</script>

{#if open}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-xs p-4">
    <div class="w-full max-w-md rounded-2xl border border-border bg-card p-6 text-foreground shadow-2xl">
      <div class="flex items-center gap-2 border-b border-border pb-3">
        <Trophy class="h-5 w-5 text-amber-400" />
        <h3 class="text-base font-bold text-foreground">Konfigurasi Lomba & Dewan Juri</h3>
      </div>

      <div class="mt-4 space-y-3.5">
        <div>
          <Label class="text-xs text-muted-foreground">Nama Event</Label>
          <Input bind:value={editEventName} placeholder="Piala Umi Kasum" class="mt-1 bg-muted" />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <Label class="text-xs text-muted-foreground">Babak / Sesi</Label>
            <Input bind:value={editSessionName} placeholder="Sesi 1" class="mt-1 bg-muted" />
          </div>
          <div>
            <Label class="text-xs text-muted-foreground">Total Ronde (detik)</Label>
            <Input type="number" bind:value={editTotalDuration} placeholder="600" class="mt-1 bg-muted" />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <Label class="text-xs text-muted-foreground">Durasi Shift Area (detik)</Label>
            <Input type="number" bind:value={editShiftDuration} placeholder="150" class="mt-1 bg-muted" />
          </div>
          <div>
            <Label class="text-xs text-muted-foreground">Total Shift</Label>
            <div class="mt-1 flex h-9 items-center rounded-md border border-input bg-muted px-3 font-mono text-xs text-foreground">
              4 Shift Rotasi Area
            </div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <Label class="text-xs text-muted-foreground">Jam Mulai</Label>
            <Input bind:value={editStartTime} placeholder="10.00" class="mt-1 bg-muted" />
          </div>
          <div>
            <Label class="text-xs text-muted-foreground">Jam Selesai</Label>
            <Input bind:value={editEndTime} placeholder="10.10 WIB" class="mt-1 bg-muted" />
          </div>
        </div>

        <div>
          <Label class="text-xs text-muted-foreground">Dewan Juri (4 Juri)</Label>
          <div class="mt-1.5 space-y-1.5">
            {#each editJurors as _, idx}
              <Input bind:value={editJurors[idx]} placeholder={`Juri ${idx + 1}`} class="bg-muted text-xs" />
            {/each}
          </div>
        </div>
      </div>

      <div class="mt-6 flex items-center justify-end gap-2 border-t border-border pt-4">
        <Button variant="ghost" size="sm" onclick={onClose}>
          Batal
        </Button>
        <Button size="sm" onclick={handleSave} class="bg-primary text-primary-foreground">
          Simpan Konfigurasi
        </Button>
      </div>
    </div>
  </div>
{/if}
