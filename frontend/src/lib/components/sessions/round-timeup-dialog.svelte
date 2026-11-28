<script lang="ts">
  import { Clock } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import * as Dialog from "$lib/components/ui/alert-dialog";
  import { roundStore } from "$lib/state/round.svelte";

  let { onStopAndSave }: { onStopAndSave: () => void | Promise<void> } = $props();
</script>

<Dialog.Root open={roundStore.timeUpDialog}>
  <Dialog.Content class="max-w-md border-border bg-card text-foreground">
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-2 text-base font-bold">
        <Clock class="size-5 text-amber-600" />
        Waktu ronde selesai
      </Dialog.Title>
      <Dialog.Description>
        Timer dan rotasi berhenti pada menit ke-10. Tekan STOP untuk mengakhiri ronde dan menyimpan rekaman.
      </Dialog.Description>
    </Dialog.Header>
    {#if roundStore.commandError}
      <p class="text-xs text-destructive" role="alert">{roundStore.commandError}</p>
    {/if}
    <Dialog.Footer>
      <Button variant="destructive" disabled={roundStore.busy} onclick={onStopAndSave}>
        STOP &amp; Simpan Rekaman
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
