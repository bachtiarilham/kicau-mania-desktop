<script lang="ts">
  import { onMount } from "svelte";
  import { Browser } from "@wailsio/runtime";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { api, all, message, type Row } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  import { auth } from "$lib/state/auth.svelte";
  import NativePlayer from "../cameras/native-player.svelte";
  import {
    HardDrive,
    Upload,
    CloudDownload,
    Video,
    FolderOpen,
    RefreshCw,
    Download,
    Play,
    Copy,
    Check
  } from "@lucide/svelte";

  let { cloud = false, queue = false }: { cloud?: boolean; queue?: boolean } =
    $props();

  let activeTab = $state<"local" | "queue" | "cloud">("local");

  $effect(() => {
    activeTab = cloud ? "cloud" : queue ? "queue" : "local";
  });
  let isCloud = $derived(activeTab === "cloud");
  let isQueue = $derived(activeTab === "queue");

  let rows = $state<Row[]>([]),
    cursor = $state(""),
    next = $state(""),
    hasMore = $state(false),
    history = $state<string[]>([]),
    busy = $state(false),
    error = $state(""),
    notice = $state(""),
    session = $state(""),
    channel = $state(""),
    round = $state(""),
    arena = $state(""),
    playing = $state("");
  let arenas = $state<Row[]>([]),
    rounds = $state<Row[]>([]);
  let recordingMuted = $state(true);
  let progress = $state<Record<string, Row>>({});
  let copiedId = $state("");

  function formatChannelName(chId: string): string {
    if (!chId) return "Kamera";
    const match = chId.match(/gantangan(\d+)/i);
    if (match) return `Gantangan ${match[1]}`;
    return chId.toUpperCase();
  }

  function formatFriendlyTitle(row: Row): string {
    if (row.snapshot?.bird_name) {
      return `${row.snapshot.bird_name} (${formatChannelName(row.channel_id)})`;
    }
    return formatChannelName(row.channel_id) || "Rekaman Kamera";
  }

  function formatDateTime(isoString: string): string {
    if (!isoString) return "-";
    try {
      const d = new Date(isoString);
      return (
        d.toLocaleString("id-ID", {
          day: "2-digit",
          month: "short",
          year: "numeric",
          hour: "2-digit",
          minute: "2-digit",
          second: "2-digit",
        }) + " WIB"
      );
    } catch {
      return isoString;
    }
  }

  function formatExportFileName(row: Row): string {
    const ch = formatChannelName(row.channel_id).replace(/\s+/g, "-");
    const d = new Date(row.started_at || Date.now());
    const yyyy = d.getFullYear();
    const mm = String(d.getMonth() + 1).padStart(2, "0");
    const dd = String(d.getDate()).padStart(2, "0");
    const hh = String(d.getHours()).padStart(2, "0");
    const min = String(d.getMinutes()).padStart(2, "0");
    const ss = String(d.getSeconds()).padStart(2, "0");
    const bird = row.snapshot?.bird_name
      ? `_${row.snapshot.bird_name.replace(/[^\w-]/g, "")}`
      : "";
    return `${ch}${bird}_${yyyy}-${mm}-${dd}_${hh}-${min}-${ss}.mkv`;
  }

  async function copyToClipboard(text: string) {
    try {
      await navigator.clipboard.writeText(text);
      copiedId = text;
      setTimeout(() => (copiedId = ""), 2000);
    } catch {}
  }

  async function load() {
    if (busy) return;
    busy = true;
    error = "";
    try {
      const q = new URLSearchParams({
        limit: "25",
        cursor,
        session_id: session,
        channel_id: channel,
      });
      if (isCloud) {
        q.set("arena_id", arena || engine.pairing.arena_id || "");
        q.set("round_id", round);
      } else {
        q.set("state", isQueue ? "READY" : "");
      }
      const page = isCloud
        ? await api.cloud<Row>("/media/recordings?" + q)
        : await api.engine<Row>("GET", "/api/v1/recordings?" + q);
      rows = page.items || [];
      next = page.next_cursor || "";
      hasMore = Boolean(page.has_more);
      if (!isCloud) {
        const states = await Promise.all(
          rows.map(async (r) => [
            r.id,
            await api.engine("GET", `/api/v1/recordings/${r.id}/upload`),
          ]),
        );
        progress = Object.fromEntries(states);
      }
    } catch (e) {
      error = message(e);
    } finally {
      busy = false;
    }
  }

  async function switchTab(tab: "local" | "queue" | "cloud") {
    activeTab = tab;
    cursor = "";
    history = [];
    rows = [];
    if (tab === "cloud") {
      if (arenas.length === 0) {
        try {
          const v = await all("/arenas");
          arenas = v;
          if (!arena) arena = v[0]?.id || "";
          await arenaChanged();
        } catch (e) {
          error = message(e);
        }
      } else {
        await load();
      }
    } else {
      await load();
    }
  }

  onMount(() => {
    arena = engine.pairing.arena_id || "";
    if (isCloud) {
      void all("/arenas")
        .then(async (v) => {
          arenas = v;
          if (!arena) arena = v[0]?.id || "";
          await arenaChanged();
        })
        .catch((e) => (error = message(e)));
    } else {
      void load();
    }
    const timer = setInterval(() => {
      if (!playing) void load();
    }, 5000);
    return () => clearInterval(timer);
  });

  async function arenaChanged() {
    round = "";
    rounds = arena ? await all(`/arenas/${arena}/rounds`) : [];
    filter();
  }

  function filter() {
    cursor = "";
    history = [];
    void load();
  }

  async function retry(id: string) {
    try {
      await api.engine("POST", `/api/v1/recordings/${id}/retry`, {});
      await load();
    } catch (e) {
      error = message(e);
    }
  }

  async function exportFile(row: Row) {
    engine.overlay = true;
    try {
      const filename = formatExportFileName(row);
      const path = await api.export(row.id, filename);
      if (path) notice = "Salinan video berhasil disimpan ke: " + path;
    } catch (e) {
      error = message(e);
    } finally {
      engine.overlay = false;
    }
  }

  async function download(id: string) {
    busy = true;
    try {
      const result = await api.cloud<Row>(`/media/recordings/${id}/download`);
      await Browser.OpenURL(result.url);
    } catch (e) {
      error = message(e);
    } finally {
      busy = false;
    }
  }
</script>

<!-- Sub-Tab Navigation Bar -->
<div class="flex items-center gap-1.5 border-b border-border pb-3 mb-5">
  <button
    type="button"
    class="flex items-center gap-2 rounded-lg px-3.5 py-2 text-xs font-semibold transition-colors"
    class:bg-primary={activeTab === "local"}
    class:text-primary-foreground={activeTab === "local"}
    class:text-muted-foreground={activeTab !== "local"}
    class:hover:bg-accent={activeTab !== "local"}
    class:hover:text-foreground={activeTab !== "local"}
    onclick={() => switchTab("local")}
  >
    <HardDrive class="h-4 w-4" />
    <span>Rekaman Lokal</span>
  </button>

  <button
    type="button"
    class="flex items-center gap-2 rounded-lg px-3.5 py-2 text-xs font-semibold transition-colors"
    class:bg-primary={activeTab === "queue"}
    class:text-primary-foreground={activeTab === "queue"}
    class:text-muted-foreground={activeTab !== "queue"}
    class:hover:bg-accent={activeTab !== "queue"}
    class:hover:text-foreground={activeTab !== "queue"}
    onclick={() => switchTab("queue")}
  >
    <Upload class="h-4 w-4" />
    <span>Antrean Upload</span>
  </button>

  {#if auth.role === "admin"}
    <button
      type="button"
      class="flex items-center gap-2 rounded-lg px-3.5 py-2 text-xs font-semibold transition-colors"
      class:bg-primary={activeTab === "cloud"}
      class:text-primary-foreground={activeTab === "cloud"}
      class:text-muted-foreground={activeTab !== "cloud"}
      class:hover:bg-accent={activeTab !== "cloud"}
      class:hover:text-foreground={activeTab !== "cloud"}
      onclick={() => switchTab("cloud")}
    >
      <CloudDownload class="h-4 w-4" />
      <span>Arsip Cloud</span>
    </button>
  {/if}
</div>

<div class="section-head mb-4">
  <div>
    <h1 class="text-xl font-bold tracking-tight">
      {isCloud ? "Arsip Rekaman Cloud" : isQueue ? "Antrean Upload Cloud" : "Daftar Rekaman Lokal"}
    </h1>
    <p class="subtitle text-xs text-muted-foreground mt-0.5">
      {isCloud
        ? "Arsip privat cloud yang sudah dikaitkan ke tiket peserta lomba."
        : isQueue
          ? "File rekaman lokal yang sedang atau menunggu sinkronisasi ke cloud server."
          : "File rekaman tersimpan di engine lokal arena, aman diakses dan diputar offline."}
    </p>
  </div>
  <div class="flex items-center gap-2">
    {#if !isCloud}
      <Button
        variant="outline"
        size="sm"
        onclick={async () => {
          try {
            await api.openFolder();
          } catch (e) {
            error = message(e);
          }
        }}
        title="Buka folder rekaman di Windows File Explorer"
      >
        <FolderOpen class="mr-1.5 h-3.5 w-3.5" /> Buka di Explorer
      </Button>
    {/if}
    <Button variant="outline" size="sm" onclick={load} disabled={busy}>
      <RefreshCw class="mr-1.5 h-3.5 w-3.5 {busy ? 'animate-spin' : ''}" /> Muat Ulang
    </Button>
  </div>
</div>

{#if error}
  <div class="error mb-4" role="alert">
    {error} <button class="underline ml-2" onclick={load}>Coba lagi</button>
  </div>
{/if}

{#if notice}
  <div class="notice mb-4">
    {notice}
  </div>
{/if}

{#if playing}
  <div class="panel mb-5">
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-sm font-semibold">Putar Rekaman Video</h2>
      <div class="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onclick={async () => {
            recordingMuted = !recordingMuted;
            await api.mute("recording-" + playing, recordingMuted);
          }}
        >
          {recordingMuted ? "Aktifkan audio" : "Matikan audio"}
        </Button>
        <Button variant="outline" size="sm" onclick={() => (playing = "")}>
          Tutup Player
        </Button>
      </div>
    </div>
    {#key playing}
      <NativePlayer id={playing} recording />
    {/key}
  </div>
{/if}

<form
  class="mb-4 flex gap-3"
  onsubmit={(e) => {
    e.preventDefault();
    filter();
  }}
>
  {#if isCloud}
    <select
      aria-label="Arena"
      bind:value={arena}
      disabled={busy}
      class="h-9 rounded-md border border-input bg-transparent px-3 py-1 text-xs shadow-xs"
      onchange={() => arenaChanged().catch((e) => (error = message(e)))}
    >
      <option value="">Pilih arena</option>
      {#each arenas as a}
        <option value={a.id}>{a.name}</option>
      {/each}
    </select>
    <select
      aria-label="Ronde"
      bind:value={round}
      disabled={busy}
      class="h-9 rounded-md border border-input bg-transparent px-3 py-1 text-xs shadow-xs"
    >
      <option value="">Semua ronde</option>
      {#each rounds as r}
        <option value={r.id}>{r.title}</option>
      {/each}
    </select>
  {/if}
  <Input
    aria-label="Sesi"
    placeholder="Filter ID sesi..."
    bind:value={session}
    class="max-w-44 text-xs"
  />
  <Input
    aria-label="Kamera"
    placeholder="Filter ID kamera..."
    bind:value={channel}
    class="max-w-44 text-xs"
  />
  <Button type="submit" variant="outline" size="sm" disabled={busy}>
    Filter
  </Button>
</form>

<div class="overflow-x-auto rounded-xl border border-border">
  <table>
    <thead>
      <tr>
        <th>Rekaman Video</th>
        <th>Gantangan / Sesi</th>
        <th>Status Lokal</th>
        <th>Status Cloud</th>
        <th>Aksi</th>
      </tr>
    </thead>
    <tbody>
      {#each rows as r (r.id)}
        <tr>
          <!-- Kolom Rekaman: Human Readable -->
          <td>
            <div class="flex items-center gap-2.5">
              <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <Video class="h-4 w-4" />
              </span>
              <div class="min-w-0">
                <div class="font-bold text-sm text-foreground truncate">
                  {formatFriendlyTitle(r)}
                </div>
                <div class="mt-0.5 text-xs text-muted-foreground flex items-center gap-1.5">
                  <span>📅 {formatDateTime(r.started_at)}</span>
                  {#if r.snapshot?.registrant_name}
                    <span>· 👤 {r.snapshot.registrant_name}</span>
                  {/if}
                </div>
                <div class="mt-1 flex items-center gap-1.5 font-mono text-[10px] text-muted-foreground/60">
                  <span>ID: {r.id.slice(0, 8)}...{r.id.slice(-6)}</span>
                  <button
                    type="button"
                    onclick={() => copyToClipboard(r.id)}
                    class="rounded px-1 hover:bg-accent hover:text-foreground transition-colors"
                    title="Salin ID lengkap"
                  >
                    {#if copiedId === r.id}
                      <Check class="inline h-2.5 w-2.5 text-emerald-400" />
                    {:else}
                      <Copy class="inline h-2.5 w-2.5" />
                    {/if}
                  </button>
                </div>
              </div>
            </div>
          </td>

          <!-- Kolom Gantangan / Sesi -->
          <td>
            <div class="flex flex-col gap-1">
              <span class="inline-flex w-fit items-center rounded-md bg-secondary px-2 py-0.5 font-mono text-xs font-bold text-foreground">
                {formatChannelName(r.channel_id)}
              </span>
              <span class="text-[11px] text-muted-foreground truncate max-w-44" title={r.local_session_id || r.round_id}>
                {r.snapshot?.session_name || (r.started_at ? `Sesi ${new Date(r.started_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}` : r.round_id || '-')}
              </span>
            </div>
          </td>

          <!-- Status Lokal -->
          <td>
            <span class="chip">{isCloud ? "—" : r.state}</span>
            {#if r.error}
              <p class="mt-1 text-xs text-destructive">{r.error}</p>
            {/if}
          </td>

          <!-- Status Cloud -->
          <td>
            <span class="chip">
              {isCloud
                ? r.state
                : r.state === "UPLOADED"
                  ? "UPLOADED"
                  : progress[r.id]?.upload_state || "WAITING_MAPPING"}
            </span>
            {#if progress[r.id]?.size_bytes}
              <div class="mt-2 w-36">
                <progress
                  value={progress[r.id].uploaded_bytes}
                  max={progress[r.id].size_bytes}
                ></progress>
                <div class="text-[10px] text-muted-foreground">
                  {(progress[r.id].uploaded_bytes / 1048576).toFixed(1)} / {(
                    progress[r.id].size_bytes / 1048576
                  ).toFixed(1)} MiB
                </div>
              </div>
            {/if}
          </td>

          <!-- Aksi -->
          <td>
            <div class="flex gap-1.5">
              {#if isCloud}
                <Button
                  variant="outline"
                  size="sm"
                  disabled={busy || r.state !== "UPLOADED"}
                  onclick={() => download(r.id)}
                >
                  <Download class="mr-1 h-3.5 w-3.5" /> Unduh
                </Button>
              {:else}
                <Button
                  variant="outline"
                  size="sm"
                  disabled={!["READY", "UPLOADED"].includes(r.state)}
                  onclick={() => {
                    recordingMuted = true;
                    playing = r.id;
                  }}
                >
                  <Play class="mr-1 h-3.5 w-3.5" /> Putar
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={!["READY", "UPLOADED"].includes(r.state)}
                  onclick={() => exportFile(r)}
                  title="Simpan file video dengan nama ramah dibaca"
                >
                  <FolderOpen class="mr-1 h-3.5 w-3.5" /> Ekspor
                </Button>
                {#if r.state === "READY"}
                  <Button
                    variant="ghost"
                    size="sm"
                    onclick={() => retry(r.id)}
                  >
                    Coba upload
                  </Button>
                {/if}
              {/if}
            </div>
          </td>
        </tr>
      {:else}
        <tr>
          <td colspan="5" class="py-12 text-center text-muted-foreground">
            {busy
              ? "Memuat…"
              : error
                ? "Data belum dapat dimuat."
                : "Belum ada rekaman untuk pilihan ini."}
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<div class="mt-4 flex justify-end gap-2">
  <Button
    variant="outline"
    size="sm"
    disabled={busy || !history.length}
    onclick={() => {
      cursor = history.pop() || "";
      void load();
    }}
  >
    Sebelumnya
  </Button>
  <Button
    variant="outline"
    size="sm"
    disabled={busy || !hasMore}
    onclick={() => {
      history.push(cursor);
      cursor = next;
      void load();
    }}
  >
    Berikutnya
  </Button>
</div>
