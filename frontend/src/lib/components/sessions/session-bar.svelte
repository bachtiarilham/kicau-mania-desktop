<script lang="ts">
  import { onMount } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import * as Dialog from "$lib/components/ui/alert-dialog";
  import {
    Circle,
    Square,
    RefreshCw,
    HardDrive,
    Cloud,
    Pause,
    CheckCircle2,
    Video,
  } from "@lucide/svelte";
  import { api } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  import { roundStore } from "$lib/state/round.svelte";
  import RoundTimeUpDialog from "./round-timeup-dialog.svelte";

  let confirm = $state(false);
  let starting = $state(false);
  let sessionName = $state("");
  let selectedFormat = $state("mp4");
  let storageDir = $state("D:\\KicauRecordings");

  let finalizing = $derived(["FINALIZING", "STOPPING"].includes(engine.session?.state));
  let retryFinalization = $derived(finalizing && !!engine.session?.error);
  let now = $state(Date.now());

  onMount(() => {
    void api
      .config()
      .then((cfg) => {
        if (cfg?.recordings_dir) storageDir = cfg.recordings_dir;
      })
      .catch(() => {});
    const t = setInterval(() => (now = Date.now()), 1000);
    return () => clearInterval(t);
  });

  function generateSessionName(): string {
    const d = new Date();
    const pad = (n: number) => String(n).padStart(2, "0");
    const dateStr = `${d.getFullYear()}${pad(d.getMonth() + 1)}${pad(d.getDate())}`;
    const cleanEvent = roundStore.eventName ? roundStore.eventName.replace(/\s+/g, "_") : "Lomba";
    const cleanSession = roundStore.sessionName ? roundStore.sessionName.replace(/\s+/g, "_") : "Ronde";
    return `${cleanEvent}_${cleanSession}_${dateStr}`;
  }

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

  let formattedElapsed = $derived(
    `${String(Math.floor(elapsed / 60)).padStart(2, "0")}:${String(
      elapsed % 60,
    ).padStart(2, "0")}`,
  );

  let formattedRoundRemaining = $derived(
    `${String(Math.floor(roundStore.roundRemainingSeconds / 60)).padStart(2, "0")}:${String(
      roundStore.roundRemainingSeconds % 60,
    ).padStart(2, "0")}`,
  );

  $effect(() => {
    engine.overlay = confirm;
  });

  async function start() {
    if (starting || engine.busy) return;
    starting = true;
    sessionName = generateSessionName();
    try {
      if (roundStore.roundId) {
        await roundStore.startRound();
      } else {
        if (!engine.draftID) engine.draftID = await api.sessionID();
        await engine.start(engine.draftID);
        engine.draftID = "";
      }
    } catch {
    } finally {
      starting = false;
    }
  }

  async function handleStopClick() {
    void api.hidePlayers();
    if (roundStore.state === "RUNNING" && !(await roundStore.stopRound())) return;
    if (!sessionName) {
      sessionName = generateSessionName();
    }
    void api
      .config()
      .then((cfg) => {
        if (cfg?.recordings_dir) storageDir = cfg.recordings_dir;
      })
      .catch(() => {});
    confirm = true;
  }

  async function handleConfirmSave() {
    confirm = false;
    const currentSessionId = engine.session?.id;
    await engine.stop();
    engine.notice = `Menyimpan rekaman sesi '${sessionName}' (${selectedFormat.toUpperCase()}) ke ${storageDir}...`;

    if (currentSessionId) {
      void api
        .saveSession(currentSessionId, sessionName, storageDir, selectedFormat)
        .then((savedFiles) => {
          if (Array.isArray(savedFiles) && savedFiles.length > 0) {
            engine.notice = `Berhasil menyimpan ${savedFiles.length} file rekaman sesi '${sessionName}' ke: ${storageDir}`;
          }
        })
        .catch((err) => {
          console.warn("Gagal simpan rekaman sesi:", err);
          engine.error =
            "Gagal mengekspor rekaman ke folder tujuan: " +
            (err?.message || String(err));
        });
    }
  }
</script>

<div
  class="flex flex-wrap items-center justify-between gap-4 border-b border-border bg-card px-6 py-3"
>
  <!-- Sisi Kiri: Status Operasional & Monitoring -->
  <div class="flex items-center gap-4">
    <Button
      variant="ghost"
      size="icon-sm"
      aria-label="Perbarui status sesi"
      disabled={engine.busy || starting}
      onclick={() => engine.refresh()}
    >
      <RefreshCw size={13} class={engine.busy ? "animate-spin" : ""} />
    </Button>

    <!-- Status Sesi (Recording / Standby) -->
    <div class="flex items-center gap-2">
      <Circle
        size={10}
        class={engine.running
          ? "fill-red-500 text-red-400 animate-pulse"
          : "text-muted-foreground"}
      />
      <div class="text-xs font-semibold tracking-wide">
        {engine.running ? "MEREKAM" : engine.session?.state || "STANDBY"}
      </div>
      {#if engine.running || roundStore.state === "RUNNING"}
        <span class="font-mono text-xs font-bold text-foreground bg-muted px-2 py-0.5 rounded border border-border" title="Sisa Waktu Ronde 10 Menit">
          Ronde: {formattedRoundRemaining}
        </span>
        <span class="font-mono text-[11px] font-semibold text-amber-400 bg-amber-500/10 px-2 py-0.5 rounded border border-amber-500/20" title="Sisa waktu area">
          {roundStore.intervalLabel} ({roundStore.shiftRemainingSeconds}s)
        </span>
      {:else if elapsed > 0}
        <span class="font-mono text-xs font-bold text-foreground bg-muted px-2 py-0.5 rounded border border-border">
          {formattedElapsed}
        </span>
      {/if}
    </div>

    <div class="h-4 w-px bg-border"></div>

    <!-- Info Kamera & Harddisk -->
    <div class="flex items-center gap-3 text-[11px] text-muted-foreground">
      <span class="flex items-center gap-1.5">
        <Video size={13} class="text-primary" />
        {engine.channels.length} Kamera (RTSP)
      </span>
      <span>·</span>
      <span class="flex items-center gap-1.5" title="Lokasi penyimpanan otomatis diatur dari .env backend">
        <HardDrive size={13} class="text-primary" />
        Disk: {storageDir}
      </span>
    </div>

    <div class="h-4 w-px bg-border"></div>

    <!-- Status Cloud Upload (Proteksi Bandwidth) -->
    <div class="flex items-center gap-1.5 text-[11px]">
      <Cloud size={13} class="text-muted-foreground" />
      {#if engine.running}
        <span class="flex items-center gap-1 text-amber-400 font-medium bg-amber-500/15 px-2 py-0.5 rounded border border-amber-500/30">
          <Pause size={10} /> Upload Paused (Lomba Aktif)
        </span>
      {:else}
        <span class="flex items-center gap-1 text-emerald-400 font-medium bg-emerald-500/15 px-2 py-0.5 rounded border border-emerald-500/30">
          <CheckCircle2 size={10} /> Upload Siap (Jeda Sesi)
        </span>
      {/if}
    </div>
  </div>

  <!-- Sisi Kanan: Single Action Button (Stopwatch Lap Style) -->
  <div class="flex items-center gap-3">
    {#if engine.active}
      <Button
        variant="destructive"
        size="sm"
        class="bg-red-600 hover:bg-red-700 text-white font-medium px-4 shadow-sm"
        disabled={engine.busy || (finalizing && !retryFinalization)}
        onclick={handleStopClick}
      >
        <Square size={13} class="fill-current mr-1.5" />
        {retryFinalization
          ? "Ulangi finalisasi"
          : finalizing
            ? "Menunggu finalisasi…"
            : `STOP & SIMPAN (${formattedElapsed})`}
      </Button>
    {:else}
      <Button
        size="sm"
        class="bg-emerald-600 hover:bg-emerald-700 text-white font-medium px-4 shadow-sm"
        disabled={engine.busy || starting || !engine.connected || !engine.channels.length}
        onclick={start}
      >
        {#if engine.busy || starting}
          <RefreshCw size={13} class="animate-spin mr-1.5" />
          Menyiapkan...
        {:else}
          <Circle size={13} class="fill-current mr-1.5" />
          MULAI REKAM (START)
        {/if}
      </Button>
    {/if}
  </div>
</div>

{#if engine.session?.error}
  <div class="error mx-6 my-2" role="alert">{engine.session.error}</div>
{/if}

<!-- Dialog Keputusan Waktu Habis: Lanjutkan atau Stop -->
<RoundTimeUpDialog onStopAndSave={handleStopClick} />

<!-- Quick Save Dialog (Muncul setelah klik STOP) -->
<Dialog.Root bind:open={confirm}>
  <Dialog.Content class="max-w-md bg-card border-border text-foreground">
    <Dialog.Header>
      <Dialog.Title class="text-base font-bold flex items-center gap-2">
        <HardDrive size={18} class="text-primary" />
        Simpan Rekaman Sesi Lomba
      </Dialog.Title>
      <Dialog.Description class="text-xs text-muted-foreground">
        Sesi lomba telah dihentikan (durasi: {formattedElapsed}). Anda memiliki jeda 10–15 menit untuk mengonfirmasi nama sesi sebelum ronde berikutnya.
      </Dialog.Description>
    </Dialog.Header>

    <div class="space-y-4 py-3">
      <!-- Input Nama Sesi -->
      <div class="space-y-1.5">
        <Label for="session-name" class="text-xs font-semibold text-foreground">Nama Sesi Rekaman</Label>
        <Input
          id="session-name"
          bind:value={sessionName}
          placeholder="Contoh: Murai Batu A - Tiket 500k"
          class="bg-card border-border text-sm"
        />
        <p class="text-[10px] text-muted-foreground">
          Bisa diubah sesuai kelas burung, atau biarkan default berformat tanggal & jam.
        </p>
      </div>

      <!-- Pilihan Format Video -->
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

      <!-- Lokasi Penyimpanan Harddisk -->
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
            onclick={async () => {
              try {
                const dir = await api.selectFolder();
                if (dir) storageDir = dir;
              } catch (e) {
                console.error(e);
              }
            }}
          >
            Pilih Folder...
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onclick={async () => {
              try {
                await api.openFolder();
              } catch (e) {
                console.error(e);
              }
            }}
            title="Buka folder di File Explorer"
          >
            📁 Buka
          </Button>
        </div>
        <p class="text-[10px] text-muted-foreground">
          Folder tujuan penyimpanan video rekaman sesi. Pilihan terakhir akan selalu diingat otomatis.
        </p>
      </div>
    </div>

    <Dialog.Footer class="gap-2 sm:gap-0">
      <Dialog.Cancel class="text-xs">
        Batal & Lanjutkan Sesi
      </Dialog.Cancel>
      <Dialog.Action
        class="bg-primary text-primary-foreground hover:bg-primary/90 text-xs font-semibold"
        onclick={handleConfirmSave}
      >
        Simpan & Antrekan ke Cloud
      </Dialog.Action>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
