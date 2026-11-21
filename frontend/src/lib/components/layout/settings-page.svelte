<script lang="ts">
  import { onMount } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { api, message, type Row } from "$lib/services/api";
  let config = $state<Row | null>(null),
    error = $state(""),
    notice = $state(""),
    busy = $state(false);
  onMount(() => {
    void api
      .config()
      .then((c) => (config = c))
      .catch((e) => (error = message(e)));
  });
  async function save() {
    if (!config) return;
    busy = true;
    error = "";
    try {
      await api.saveConfig(config);
      notice =
        "Pengaturan tersimpan. Tutup dan jalankan kembali desktop untuk menerapkan koneksi baru. Recording dan upload local-app tetap berjalan.";
    } catch (e) {
      error = message(e);
    } finally {
      busy = false;
    }
  }
</script>

<div class="section-head">
  <div>
    <h1>Pengaturan desktop</h1>
    <p class="subtitle">Koneksi berlaku setelah desktop dijalankan ulang.</p>
  </div>
</div>
{#if error}<div class="error mb-4">{error}</div>{/if}{#if notice}<div
    class="notice mb-4"
  >
    {notice}
  </div>{/if}{#if config}<form
    class="panel max-w-2xl space-y-5"
    onsubmit={(e) => {
      e.preventDefault();
      void save();
    }}
  >
    <label class="field"
      >API local-app<Input bind:value={config.local_url} required /></label
    ><label class="field"
      >RTSP MediaMTX<Input bind:value={config.rtsp_base} required /></label
    ><label class="field"
      >Cloud API<Input bind:value={config.cloud_url} required /></label
    ><label class="field"
      >Lokasi libvlc.dll<Input bind:value={config.vlc} required /></label
    ><label class="field"
      >Folder rekaman video
      <div class="flex gap-2">
        <Input bind:value={config.recordings_dir} placeholder="C:\Users\...\Videos\KicauMania" />
        <Button
          type="button"
          variant="outline"
          onclick={async () => {
            try {
              const dir = await api.selectFolder();
              if (dir && config) config.recordings_dir = dir;
            } catch (e) {
              error = message(e);
            }
          }}>Pilih Folder...</Button
        >
        <Button
          type="button"
          variant="outline"
          onclick={async () => {
            try {
              await api.openFolder();
            } catch (e) {
              error = message(e);
            }
          }}>Buka di Explorer</Button
        >
      </div>
      <span class="text-xs text-muted-foreground"
        >Lokasi penyimpanan video rekaman dan ekspor. Pilihan terakhir akan selalu diingat otomatis.</span
      ></label
    ><label class="field"
      >Buffer live (ms)<Input type="number" min="0" max="3000" step="1" bind:value={config.live_cache_ms} required />
      <span class="text-xs text-muted-foreground">Default 50 ms. Jika gambar tersendat, coba 100 ms. Nilai 0 lebih rentan tersendat. Berlaku setelah desktop dijalankan ulang.</span></label
    ><Button type="submit" disabled={busy}>Simpan pengaturan</Button>
  </form>{/if}
