<script lang="ts">
  import { onMount } from "svelte";
  import { Browser } from "@wailsio/runtime";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { api, all, message, type Row } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  import NativePlayer from "../cameras/native-player.svelte";
  let { cloud = false, queue = false }: { cloud?: boolean; queue?: boolean } =
    $props();
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
      if (cloud) {
        q.set("arena_id", arena || engine.pairing.arena_id || "");
        q.set("round_id", round);
      } else {
        q.set("state", queue ? "READY" : "");
      }
      const page = cloud
        ? await api.cloud<Row>("/media/recordings?" + q)
        : await api.engine<Row>("GET", "/api/v1/recordings?" + q);
      rows = page.items;
      next = page.next_cursor || "";
      hasMore = page.has_more;
      if (!cloud) {
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
  onMount(() => {
    arena = engine.pairing.arena_id || "";
    if (cloud) {
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
  async function exportFile(id: string) {
    engine.overlay = true;
    try {
      const path = await api.export(id);
      if (path) notice = "Salinan disimpan ke " + path;
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

<div class="section-head">
  <div>
    <h1>
      {cloud ? "Rekaman cloud" : queue ? "Antrean upload" : "Rekaman lokal"}
    </h1>
    <p class="subtitle">
      {cloud
        ? "Arsip privat yang sudah dikaitkan ke tiket peserta."
        : "File tetap tersimpan di engine, termasuk saat internet terputus."}
    </p>
  </div>
  <Button variant="outline" onclick={load} disabled={busy}>Muat ulang</Button>
</div>
{#if error}<div class="error mb-4" role="alert">
    {error} <button class="underline" onclick={load}>Coba lagi</button>
  </div>{/if}{#if notice}<div class="notice mb-4">{notice}</div>{/if}
{#if playing}<div class="panel mb-5">
    <div class="mb-4 flex items-center justify-between">
      <h2>Putar rekaman lokal</h2>
      <Button
        variant="outline"
        onclick={async () => {
          recordingMuted = !recordingMuted;
          await api.mute("recording-" + playing, recordingMuted);
        }}>{recordingMuted ? "Aktifkan audio" : "Matikan audio"}</Button
      >
      <Button variant="outline" onclick={() => (playing = "")}
        >Tutup player</Button
      >
    </div>
    {#key playing}<NativePlayer id={playing} recording />{/key}
  </div>{/if}
<form
  class="mb-4 flex gap-3"
  onsubmit={(e) => {
    e.preventDefault();
    filter();
  }}
>
  {#if cloud}<select
      aria-label="Arena"
      bind:value={arena}
      disabled={busy}
      onchange={() => arenaChanged().catch((e) => (error = message(e)))}
      ><option value="">Pilih arena</option>{#each arenas as a}<option
          value={a.id}>{a.name}</option
        >{/each}</select
    ><select aria-label="Ronde" bind:value={round} disabled={busy}
      ><option value="">Semua ronde</option>{#each rounds as r}<option
          value={r.id}>{r.title}</option
        >{/each}</select
    >{/if}<Input
    aria-label="Sesi"
    placeholder="ID sesi"
    bind:value={session}
  /><Input
    aria-label="Kamera"
    placeholder="ID kamera"
    bind:value={channel}
  /><Button type="submit" variant="outline" disabled={busy}>Filter</Button>
</form>
<div class="overflow-x-auto rounded-xl border border-border">
  <table>
    <thead
      ><tr
        ><th>Rekaman</th><th>Kamera / sesi</th><th>Saving lokal</th><th
          >Upload cloud</th
        ><th>Aksi</th></tr
      ></thead
    ><tbody
      >{#each rows as r}<tr
          ><td
            ><span class="font-medium">{r.snapshot?.bird_name || r.id}</span>
            <div class="mt-1 text-[10px] text-muted-foreground">
              {r.snapshot?.registrant_name ||
                new Date(r.started_at).toLocaleString("id-ID")}
            </div></td
          ><td
            >{r.channel_id}
            <div
              class="mt-1 max-w-48 truncate text-[10px] text-muted-foreground"
            >
              {r.local_session_id || r.round_id}
            </div></td
          ><td
            ><span class="chip">{cloud ? "—" : r.state}</span>{#if r.error}<p
                class="mt-2 text-xs text-destructive"
              >
                {r.error}
              </p>{/if}</td
          ><td
            ><span class="chip"
              >{cloud
                ? r.state
                : r.state === "UPLOADED"
                  ? "UPLOADED"
                  : progress[r.id]?.upload_state || "WAITING_MAPPING"}</span
            >{#if progress[r.id]?.size_bytes}<div class="mt-2 w-36">
                <progress
                  value={progress[r.id].uploaded_bytes}
                  max={progress[r.id].size_bytes}
                ></progress>
                <div class="text-[10px] text-muted-foreground">
                  {(progress[r.id].uploaded_bytes / 1048576).toFixed(1)} / {(
                    progress[r.id].size_bytes / 1048576
                  ).toFixed(1)} MiB
                </div>
              </div>{/if}</td
          ><td
            ><div class="flex gap-1">
              {#if cloud}<Button
                  variant="outline"
                  size="sm"
                  disabled={busy || r.state !== "UPLOADED"}
                  onclick={() => download(r.id)}>Unduh</Button
                >{:else}<Button
                  variant="outline"
                  size="sm"
                  disabled={!["READY", "UPLOADED"].includes(r.state)}
                  onclick={() => {
                    recordingMuted = true;
                    playing = r.id;
                  }}>Putar</Button
                ><Button
                  variant="outline"
                  size="sm"
                  disabled={!["READY", "UPLOADED"].includes(r.state)}
                  onclick={() => exportFile(r.id)}>Ekspor</Button
                >{#if r.state === "READY"}<Button
                    variant="ghost"
                    size="sm"
                    onclick={() => retry(r.id)}>Coba upload</Button
                  >{/if}{/if}
            </div></td
          ></tr
        >{:else}<tr
          ><td colspan="5" class="py-12 text-center text-muted-foreground"
            >{busy
              ? "Memuat…"
              : error
                ? "Data belum dapat dimuat."
                : "Belum ada rekaman untuk pilihan ini."}</td
          ></tr
        >{/each}</tbody
    >
  </table>
</div>
<div class="mt-4 flex justify-end gap-2">
  <Button
    variant="outline"
    disabled={busy || !history.length}
    onclick={() => {
      cursor = history.pop() || "";
      void load();
    }}>Sebelumnya</Button
  ><Button
    variant="outline"
    disabled={busy || !hasMore}
    onclick={() => {
      history.push(cursor);
      cursor = next;
      void load();
    }}>Berikutnya</Button
  >
</div>
