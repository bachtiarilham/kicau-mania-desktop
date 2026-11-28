<script lang="ts">
  import { onMount } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { api, message } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  import { arenaCameraID, gantanganCameraID, gantanganChannels } from "$lib/components/cameras/camera-slots";

  type Camera = { id: string; ip: string; source: string };
  type Settings = {
    server_ip: string;
    judge_ips: string[];
    nvr_ip: string;
    gateway: string;
    channels: Camera[];
    revision: string;
    recordings_dir: string;
    effective_recordings_dir: string;
  };

  const endpoint = "/api/v1/setup/arena";
  let settings = $state<Settings | null>(null);
  let cameras = $state<Camera[]>([]);
  let tests = $state<Record<string, { pending?: boolean; ok?: boolean; text: string }>>({});
  let loading = $state(false);
  let saving = $state(false);
  let error = $state("");
  let notice = $state("");

  const label = (index: number) => index === 24 ? "Kamera Umum Arena" : `G${String(index + 1).padStart(2, "0")}`;
  const blank = (id: string): Camera => ({ id, ip: "", source: "" });

  async function load(clearNotice = true) {
    loading = true;
    error = "";
    if (clearNotice) notice = "";
    try {
      await api.bootstrap();
      const next = await api.engine<Settings>("GET", endpoint);
      next.judge_ips = Array.from({ length: 4 }, (_, index) => next.judge_ips[index] ?? "");
      const gantangan = gantanganChannels(next.channels);
      cameras = gantangan.map((camera, index) => camera ?? blank(gantanganCameraID(index + 1)));
      cameras.push(next.channels.find((camera) => camera.id === arenaCameraID) ?? blank(arenaCameraID));
      settings = next;
    } catch (failure) {
      error = message(failure);
    } finally {
      loading = false;
    }
  }

  onMount(() => { void load(); });

  async function save(event: SubmitEvent) {
    event.preventDefault();
    if (!settings || saving) return;
    saving = true;
    error = "";
    notice = "";
    try {
      await api.engine("POST", endpoint, {
        ...$state.snapshot(settings),
        judge_ips: settings.judge_ips.map((ip) => ip.trim()).filter(Boolean),
        nvr_ip: settings.nvr_ip.trim(),
        channels: cameras.map(({ id, ip, source }) => ({ id, ip: ip.trim(), source: source.trim() })),
      });
      await load(false);
      notice = "Tersimpan. Mulai ulang Local App saat ronde dan rekaman sudah berhenti untuk menerapkan perubahan.";
    } catch (failure) {
      error = message(failure);
    } finally {
      saving = false;
    }
  }

  async function testCamera(camera: Camera) {
    if (tests[camera.id]?.pending || !camera.source.trim()) return;
    tests[camera.id] = { pending: true, text: "Memeriksa video…" };
    try {
      const result = await api.engine("POST", endpoint + "/test-camera", {
        id: camera.id, ip: camera.ip.trim(), source: camera.source.trim(),
      });
      const source = result?.via === "mediamtx" ? "MediaMTX" : result?.via === "direct" ? "Sumber RTSP" : "";
      tests[camera.id] = { ok: true, text: [result?.video_codec?.toUpperCase() || "Video tersedia", result?.width && result?.height ? `${result.width} × ${result.height}` : "", source].filter(Boolean).join(" · ") };
    } catch (failure) {
      tests[camera.id] = { ok: false, text: message(failure) };
    }
  }

  async function testNVR() {
    if (!settings?.nvr_ip.trim() || tests.nvr?.pending) return;
    tests.nvr = { pending: true, text: "Menghubungi NVR…" };
    try {
      await api.engine("POST", endpoint + "/test-nvr", { ip: settings.nvr_ip.trim() });
      tests.nvr = { ok: true, text: "Port RTSP 554 dapat dijangkau." };
    } catch (failure) {
      tests.nvr = { ok: false, text: message(failure) };
    }
  }

  function updateCameraIP(camera: Camera, ip: string) {
    if (camera.ip === ip) return;
    if (ip && camera.source) {
      try {
        const source = new URL(camera.source);
        if (source.hostname === camera.ip) {
          source.hostname = ip;
          source.username = "";
          source.password = "";
          camera.source = source.toString();
        }
      } catch {}
    }
    camera.ip = ip;
    delete tests[camera.id];
  }
</script>

<section aria-label="Pengaturan Arena" class="mx-auto w-full max-w-6xl space-y-4">
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if notice}<p class="notice" role="status">{notice}</p>{/if}
  {#if loading && !settings}<p role="status">Memuat pengaturan arena…</p>{/if}

  {#if settings}
    <form class="space-y-4" onsubmit={save}>
      <fieldset disabled={saving || loading} class="grid gap-4 xl:grid-cols-2">
        <section class="panel space-y-3">
          <h2 class="text-sm font-bold">Jaringan arena</h2>
          <label class="field">
            IP Operator / Server
            <select bind:value={settings.server_ip} required class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm">
              <option value="">Pilih alamat LAN</option>
              {#if settings.server_ip && !engine.networks.some((network) => network.ip === settings?.server_ip)}
                <option value={settings.server_ip}>{settings.server_ip}</option>
              {/if}
              {#each engine.networks as network (network.ip)}
                <option value={network.ip}>{network.ip} · {network.name}</option>
              {/each}
            </select>
          </label>
          <div class="space-y-1">
            <span class="text-xs font-semibold">IP Router / Gateway</span>
            <p class="text-sm font-mono text-muted-foreground">{settings.gateway || "Tidak terdeteksi"}</p>
            <p class="text-xs text-muted-foreground">Diambil dari pengaturan jaringan Windows untuk IP server yang tersimpan.</p>
          </div>
          <label class="field">
            IP NVR
            <div class="flex gap-2">
              <Input bind:value={settings.nvr_ip} placeholder="192.168.1.123" oninput={() => delete tests.nvr} />
              <Button type="button" variant="outline" onclick={testNVR} disabled={!settings.nvr_ip.trim() || tests.nvr?.pending}>Tes NVR</Button>
            </div>
          </label>
          {#if tests.nvr}<p role="status" class="text-xs {tests.nvr.ok === false ? 'text-destructive' : 'text-muted-foreground'}">{tests.nvr.text}</p>{/if}
        </section>

        <section class="panel space-y-3">
          <h2 class="text-sm font-bold">Akses PC Juri</h2>
          <div class="grid gap-3 sm:grid-cols-2">
            {#each settings.judge_ips as _, index (index)}
              <label class="field">IP PC Juri {index + 1}
                <Input bind:value={settings.judge_ips[index]} placeholder={index === 0 ? "192.168.1.58" : "Opsional"} />
              </label>
            {/each}
          </div>
          <p class="text-xs text-muted-foreground">Isi hanya PC juri yang akan dipakai. IP ini memberi akses ke video dan status ronde melalui LAN.</p>
        </section>

        <section class="panel space-y-3 xl:col-span-2">
          <div>
            <h2 class="text-sm font-bold">CCTV arena</h2>
            <p class="text-xs text-muted-foreground">Isi URL RTSP untuk memasang kamera. Kosongkan URL untuk melepas kamera dari slot.</p>
          </div>
          <div class="divide-y divide-border">
            <div class="hidden grid-cols-[7.5rem_10rem_minmax(0,1fr)_6.25rem] gap-x-3 pb-2 text-xs font-medium text-muted-foreground lg:grid" aria-hidden="true">
              <span>Kamera</span>
              <span>IP kamera</span>
              <span>URL RTSP</span>
            </div>
            {#each cameras as camera, index (camera.id)}
              <div class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-3 gap-y-2 py-2 lg:grid-cols-[7.5rem_10rem_minmax(0,1fr)_6.25rem]">
                <span class="col-span-2 text-xs font-semibold lg:col-span-1">{label(index)}</span>
                <Input
                  aria-label={`IP ${label(index)}`}
                  value={camera.ip}
                  placeholder="IP kamera"
                  class="col-span-2 font-mono text-sm lg:col-span-1"
                  oninput={(event) => updateCameraIP(camera, event.currentTarget.value)}
                />
                <Input
                  aria-label={`URL RTSP ${label(index)}`}
                  bind:value={camera.source}
                  placeholder="rtsp://alamat:554/path"
                  class="font-mono text-sm"
                  autocomplete="off"
                  spellcheck={false}
                  oninput={() => delete tests[camera.id]}
                />
                <Button type="button" variant="outline" class="h-9 w-full" disabled={!camera.source.trim() || tests[camera.id]?.pending} onclick={() => testCamera(camera)}>Tes Video</Button>
                {#if tests[camera.id]}
                  <p role="status" class="col-span-2 text-xs lg:col-start-3 {tests[camera.id].ok === false ? 'text-destructive' : 'text-muted-foreground'}">{tests[camera.id].text}</p>
                {/if}
              </div>
            {/each}
          </div>
        </section>
      </fieldset>
      <fieldset class="panel space-y-3" disabled={saving || loading}>
        <legend class="text-sm font-bold">Penyimpanan Rekaman Asli</legend>
        <label class="field">
          Folder rekaman Local App
          <Input bind:value={settings.recordings_dir} placeholder={settings.effective_recordings_dir} />
        </label>
        <p class="text-xs text-muted-foreground">Folder aktif {settings.effective_recordings_dir}. Perubahan berlaku setelah Local App dimulai ulang.</p>
        <Button type="button" variant="outline" onclick={async () => {
          if (!settings) return;
          try {
            const path = await api.selectOriginalFolder(settings.recordings_dir || settings.effective_recordings_dir);
            if (path && settings) settings.recordings_dir = path;
          } catch (failure) {
            error = message(failure);
          }
        }}>Pilih Folder</Button>
      </fieldset>
      <div class="flex flex-wrap gap-2">
        <Button type="submit" disabled={saving || loading}>{saving ? "Menyimpan…" : "Simpan pengaturan arena"}</Button>
        <Button type="button" variant="outline" disabled={loading || saving} onclick={() => load()}>Muat ulang</Button>
      </div>
    </form>
  {:else if !loading}
    <Button type="button" variant="outline" onclick={() => load()}>Muat ulang</Button>
  {/if}
</section>
