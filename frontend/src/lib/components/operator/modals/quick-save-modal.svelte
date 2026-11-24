<script lang="ts">
  import { HardDrive } from "@lucide/svelte";
  import * as Dialog from "$lib/components/ui/alert-dialog";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { api } from "$lib/services/api";

  interface Props {
    open: boolean;
    sessionName: string;
    selectedFormat: string;
    storageDir: string;
    onConfirm: () => void;
    onCancel: () => void;
  }

  let {
    open = $bindable(false),
    sessionName = $bindable(""),
    selectedFormat = $bindable("mp4"),
    storageDir = $bindable("D:\\KicauRecordings"),
    onConfirm,
    onCancel,
  }: Props = $props();

  async function selectFolder() {
    try {
      const dir = await api.selectFolder();
      if (dir) storageDir = dir;
    } catch (e) {
      console.error(e);
    }
  }

  async function openFolder() {
    try {
      await api.openFolder();
    } catch (e) {
      console.error(e);
    }
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-md bg-card border-border text-foreground">
    <Dialog.Header>
      <Dialog.Title class="text-base font-bold flex items-center gap-2">
        <HardDrive size={18} class="text-primary" />
        Simpan Rekaman Sesi Lomba
      </Dialog.Title>
      <Dialog.Description class="text-xs text-muted-foreground">
        Sesi ronde telah dihentikan. Anda dapat mengubah nama sesi dan format sebelum disimpan permanen.
      </Dialog.Description>
    </Dialog.Header>

    <div class="space-y-4 py-3">
      <div class="space-y-1.5">
        <Label for="op-session-name" class="text-xs font-semibold text-foreground">Nama Sesi Rekaman</Label>
        <Input
          id="op-session-name"
          bind:value={sessionName}
          placeholder="Contoh: Murai Batu A - Ronde 1"
          class="bg-card border-border text-sm"
        />
        <p class="text-[10px] text-muted-foreground">
          Bisa diubah sesuai kelas burung, atau biarkan default berformat event & tanggal.
        </p>
      </div>

      <div class="space-y-1.5">
        <Label class="text-xs font-semibold text-foreground">Format File Rekaman</Label>
        <div class="grid grid-cols-2 gap-2">
          <button
            type="button"
            class={`flex items-center justify-between px-3 py-2 rounded-lg border text-xs font-medium transition-colors ${
              selectedFormat === "mp4"
                ? "border-primary bg-primary/10 text-primary"
                : "border-border text-muted-foreground"
            }`}
            onclick={() => (selectedFormat = "mp4")}
          >
            <span>MP4 (Rekomendasi)</span>
            <span class="text-[10px] opacity-70">Universal</span>
          </button>

          <button
            type="button"
            class={`flex items-center justify-between px-3 py-2 rounded-lg border text-xs font-medium transition-colors ${
              selectedFormat === "mkv"
                ? "border-primary bg-primary/10 text-primary"
                : "border-border text-muted-foreground"
            }`}
            onclick={() => (selectedFormat = "mkv")}
          >
            <span>MKV</span>
            <span class="text-[10px] opacity-70">Tahan Crash</span>
          </button>
        </div>
      </div>

      <div class="space-y-1.5">
        <Label class="text-xs font-semibold text-foreground">Lokasi Harddisk</Label>
        <div class="flex gap-2">
          <Input
            bind:value={storageDir}
            placeholder="D:\KicauRecordings"
            class="bg-card border-border text-xs font-mono"
          />
          <Button
            type="button"
            variant="outline"
            size="sm"
            onclick={selectFolder}
          >
            Pilih Folder...
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onclick={openFolder}
            title="Buka folder di File Explorer"
          >
            📁 Buka
          </Button>
        </div>
        <p class="text-[10px] text-muted-foreground">
          Folder tujuan penyimpanan video rekaman ronde ini.
        </p>
      </div>
    </div>

    <Dialog.Footer class="gap-2 sm:gap-0">
      <Dialog.Cancel class="text-xs" onclick={onCancel}>
        Batal & Lanjutkan Sesi
      </Dialog.Cancel>
      <Dialog.Action
        class="bg-primary text-primary-foreground hover:bg-primary/90 text-xs font-semibold"
        onclick={onConfirm}
      >
        Simpan Rekaman Ronde
      </Dialog.Action>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
