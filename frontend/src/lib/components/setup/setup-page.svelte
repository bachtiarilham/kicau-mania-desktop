<script lang="ts">
  import { onMount } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Plus, Trash2, Radio, CheckCircle2 } from "@lucide/svelte";
  import { api, job, message, type Row } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  let ip = $state(""),
    endpoint = $state(""),
    username = $state(""),
    password = $state(""),
    channels = $state<Row[]>([]),
    devices = $state<Row[]>([]),
    profiles = $state<Row[]>([]),
    busy = $state(false),
    current = $state<Row | null>(null),
    error = $state(""),
    notice = $state("");
  onMount(() => {
    ip = engine.configuration?.lan_ip || engine.networks[0]?.ip || "";
    channels = structuredClone(
      $state.snapshot(engine.configuration?.channels || []),
    );
  });
  async function run(input: Row) {
    busy = true;
    error = "";
    notice = "";
    try {
      return await job(input, (v) => (current = v));
    } catch (e) {
      error = message(e);
      throw e;
    } finally {
      busy = false;
      password = "";
    }
  }
  async function discover() {
    try {
      const v = await run({ action: "discover", ip });
      devices = Array.isArray(v) ? v : v.devices || [];
    } catch {}
  }
  async function readProfiles() {
    try {
      const v = await run({
        action: "profiles",
        ip,
        endpoint,
        username,
        password,
      });
      profiles = Array.isArray(v) ? v : v.profiles || [];
    } catch {}
  }
  async function probe(i: number) {
    try {
      const v = await run({ action: "probe", channel: channels[i] });
      channels[i] = { ...channels[i], probe: v };
      notice = "Sumber teruji. Simpan konfigurasi untuk menerapkannya.";
    } catch {}
  }
  async function apply() {
    try {
      await run({
        action: "apply",
        configuration: {
          version: 1,
          lan_ip: ip,
          lan_cidr:
            engine.networks.find((n) => n.ip === ip)?.cidr ||
            engine.configuration?.lan_cidr,
          channels,
        },
      });
      notice = "Konfigurasi berhasil diterapkan.";
      await engine.connect();
    } catch {}
  }
  async function cancel() {
    if (current?.id)
      try {
        await api.engine("DELETE", `/api/v1/setup/jobs/${current.id}`);
      } catch (e) {
        error = message(e);
      }
  }
  function add(profile?: Row) {
    if (channels.length >= 24) return;
    channels.push({
      id: `gantangan${String(channels.length + 1).padStart(2, "0")}`,
      source: profile?.source || "",
      credential_ref: profile?.credential_ref,

    });
  }
</script>

<div class="section-head">
  <div>
    <h1>Konfigurasi NVR</h1>
    <p class="subtitle">
      Temukan perangkat, uji sumber, lalu petakan kamera ke gantangan.
    </p>
  </div>
  <Button
    onclick={() => add()}
    disabled={busy || engine.active || channels.length >= 24}
    ><Plus size={14} /> Kamera manual</Button
  >
</div>
{#if error}<div class="error mb-4" role="alert">
    {error}
  </div>{/if}{#if notice}<div class="notice mb-4">{notice}</div>{/if}
{#if engine.active}<div class="notice mb-4">
    Hentikan sesi dan tunggu finalisasi sebelum mengganti konfigurasi.
  </div>{/if}
<fieldset disabled={busy || engine.active} class="space-y-5">
  <section class="panel">
    <h2 class="mb-5">1. Koneksi dan discovery</h2>
    <div class="grid grid-cols-1 gap-4 md:grid-cols-2 2xl:grid-cols-4">
      <label class="field"
        >Interface arena<select bind:value={ip}
          ><option value="">Pilih interface</option
          >{#each engine.networks as n}<option value={n.ip}
              >{n.name} · {n.ip}</option
            >{/each}</select
        ></label
      ><label class="field"
        >Endpoint ONVIF<Input
          bind:value={endpoint}
          placeholder="http://IP-NVR/onvif/device_service"
        /></label
      ><label class="field"
        >Username<Input bind:value={username} autocomplete="off" /></label
      ><label class="field"
        >Password<Input
          type="password"
          bind:value={password}
          autocomplete="off"
        /></label
      >
    </div>
    <div class="mt-4 flex flex-wrap gap-2">
      <Button variant="outline" onclick={discover} disabled={!ip}
        ><Radio size={14} /> Temukan perangkat</Button
      ><Button
        variant="outline"
        onclick={readProfiles}
        disabled={!endpoint || !ip}>Baca profil NVR</Button
      >
    </div>
    {#if devices.length}<div class="mt-4 flex flex-wrap gap-2">
        {#each devices as d}<Button
            variant="outline"
            onclick={() => (endpoint = d.endpoint)}
            >{d.name || d.endpoint}</Button
          >{/each}
      </div>{/if}{#if profiles.length}<div
        class="mt-5 border-t border-border pt-4"
      >
        <h3 class="mb-3 text-sm">Pilih profil untuk ditambahkan</h3>
        <div class="grid grid-cols-2 gap-2">
          {#each profiles as p}<Button
              variant="outline"
              class="justify-between"
              onclick={() => add(p)}
              disabled={channels.length >= 24}
              >{p.name} · {p.video_codec || "Belum diuji"}
              <Plus size={14} /></Button
            >{/each}
        </div>
      </div>{/if}
  </section>
  <section class="panel">
    <h2 class="mb-5">
      2. Mapping dan pengujian kamera <span
        class="text-sm text-muted-foreground">({channels.length}/24)</span
      >
    </h2>
    {#if !channels.length}<div class="empty">
        Tambahkan kamera manual atau pilih profil NVR.
      </div>{/if}
    <div class="space-y-3">
      {#each channels as channel, i}<div
          class="rounded-xl border border-border p-4"
        >
          <div class="grid grid-cols-1 items-end gap-3 md:grid-cols-2 xl:grid-cols-[minmax(0,150px)_minmax(0,1fr)_minmax(0,145px)_auto]">
            <label class="field"
              >ID gantangan<Input bind:value={channel.id} /></label
            ><label class="field"
              >Sumber RTSP<Input
                bind:value={channel.source}
                placeholder="rtsp://IP-NVR/path"
              /></label
            ><div class="field"
              >Live RTSP<span class="rounded-md border border-border px-3 py-2 text-sm">Codec asli</span
              ></div
            >
            <div class="flex gap-1">
              <Button variant="outline" onclick={() => probe(i)}>Uji</Button
              ><Button
                variant="ghost"
                size="icon"
                aria-label="Hapus kamera"
                onclick={() => channels.splice(i, 1)}
                ><Trash2 size={14} /></Button
              >
            </div>
          </div>
          <div class="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2">
            <label class="field"
              >Username RTSP (opsional)<Input
                bind:value={channel.username}
                autocomplete="off"
              /></label
            ><label class="field"
              >Password RTSP (opsional)<Input
                type="password"
                bind:value={channel.password}
                autocomplete="off"
              /></label
            >
          </div>
          {#if channel.probe}<div
              class="mt-3 flex items-center gap-2 text-xs text-primary"
            >
              <CheckCircle2 size={13} />{channel.probe.video_codec} · {channel
                .probe.width}×{channel.probe.height} · {channel.probe.fps} FPS · {channel
                .probe.has_audio
                ? "Audio tersedia"
                : "Tanpa audio"}
            </div>{/if}
        </div>{/each}
    </div>
    <div class="mt-5 flex justify-end">
      <Button onclick={apply} disabled={!channels.length || !ip}
        >Simpan & terapkan konfigurasi</Button
      >
    </div>
  </section>
</fieldset>
{#if busy}<div class="panel mt-4 flex items-center justify-between">
    <span class="text-sm">{current?.progress || "Memproses…"}</span><Button
      variant="outline"
      onclick={cancel}>Batalkan pekerjaan</Button
    >
  </div>{/if}
