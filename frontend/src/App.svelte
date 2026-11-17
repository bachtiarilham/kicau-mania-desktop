<script lang="ts">
  import { onMount } from "svelte";
  import {
    Camera,
    Cctv,
    Settings,
    SlidersHorizontal,
    Cloud,
    HardDrive,
    Upload,
    CloudDownload,
    RefreshCw,
  } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import SettingsPage from "$lib/components/layout/settings-page.svelte";
  import Dashboard from "$lib/components/cameras/dashboard.svelte";
  import Setup from "$lib/components/setup/setup-page.svelte";
  import CloudPage from "$lib/components/cloud/cloud-page.svelte";
  import Recordings from "$lib/components/recordings/recordings-page.svelte";
  import SessionBar from "$lib/components/sessions/session-bar.svelte";
  import { api } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  let tab = $state("cameras");
  const nav = [
    { id: "cameras", label: "Kamera live", icon: Camera },
    { id: "setup", label: "Konfigurasi NVR", icon: SlidersHorizontal },
    { id: "cloud", label: "Ronde & peserta", icon: Cloud },
    { id: "recordings", label: "Rekaman lokal", icon: HardDrive },
    { id: "uploads", label: "Antrean upload", icon: Upload },
    { id: "archive", label: "Rekaman cloud", icon: CloudDownload },
    { id: "settings", label: "Pengaturan desktop", icon: Settings },
  ];
  onMount(() => {
    const off = engine.listen();
    void engine.connect();
    let pending = false;
    const timer = setInterval(async () => {
      if (pending) return;
      pending = true;
      try {
        await engine.refresh();
      } finally {
        pending = false;
      }
    }, 3000);
    return () => {
      off();
      clearInterval(timer);
    };
  });
  async function navigate(id: string) {
    await api.hidePlayers();
    tab = id;
  }
</script>

<div class="flex h-screen overflow-hidden">
  <aside
    class="flex w-56 shrink-0 flex-col border-r border-border bg-[#131c1d]"
  >
    <div class="flex items-center gap-3 px-5 py-7">
      <div class="rounded-xl bg-primary p-2 text-primary-foreground">
        <Cctv size={23} />
      </div>
      <div>
        <div class="text-sm font-semibold tracking-tight">KICAU MANIA</div>
        <div class="mt-1 text-[9px] tracking-[.2em] text-muted-foreground">
          DESKTOP CONTROL
        </div>
      </div>
    </div>
    <div class="px-5 pb-3 pt-4 text-[9px] tracking-[.2em] text-[#627975]">
      OPERASIONAL ARENA
    </div>
    <nav class="space-y-1 px-3">
      {#each nav as item}<button
          onclick={() => navigate(item.id)}
          class="flex w-full items-center gap-3 rounded-lg px-3 py-3 text-left text-xs transition-colors"
          class:bg-[#29362b]={tab === item.id}
          class:text-primary={tab === item.id}
          class:text-muted-foreground={tab !== item.id}
          ><item.icon size={16} />{item.label}</button
        >{/each}
    </nav>
    <div class="mt-auto border-t border-border p-5">
      <div class="mb-2 flex items-center gap-2 text-[11px]">
        <span
          class="h-1.5 w-1.5 rounded-full"
          class:bg-primary={engine.connected}
          class:bg-destructive={!engine.connected}
        ></span>{engine.connected ? "Engine terhubung" : "Engine terputus"}
      </div>
      <p class="text-[10px] leading-5 text-muted-foreground">
        Saving dan upload dikelola local-app. Menutup desktop tidak menghentikan
        engine.
      </p>
    </div>
  </aside>
  <div class="flex min-w-0 flex-1 flex-col">
    <header
      class="flex h-16 shrink-0 items-center justify-between border-b border-border px-6"
    >
      <div class="text-xs text-muted-foreground">
        Arena / <span class="text-foreground"
          >{nav.find((n) => n.id === tab)?.label}</span
        >
      </div>
      <div class="flex items-center gap-3">
        <span class="chip">WINDOWS · LOCAL ENGINE</span><Button
          variant="ghost"
          size="icon-sm"
          aria-label="Hubungkan ulang engine"
          onclick={() => engine.connect()}><RefreshCw size={14} /></Button
        >
      </div>
    </header>
    <SessionBar />
    <main data-viewport class="min-h-0 flex-1 overflow-y-auto p-6">
      {#if engine.error}<div class="error mb-5" role="alert">
          {engine.error}
          <button class="ml-2 underline" onclick={() => engine.connect()}
            >Hubungkan ulang</button
          >
        </div>{/if}{#if engine.notice}<div class="notice mb-5">
          {engine.notice}
        </div>{/if}{#if tab === "cameras"}<Dashboard
        />{:else if tab === "setup"}<Setup
        />{:else if tab === "cloud"}<CloudPage
        />{:else if tab === "recordings"}<Recordings
        />{:else if tab === "uploads"}<Recordings
          queue
        />{:else if tab === "settings"}<SettingsPage />{:else}<Recordings
          cloud
        />{/if}
    </main>
    <footer
      class="flex h-8 shrink-0 items-center justify-between border-t border-border px-6 text-[9px] tracking-wide text-muted-foreground"
    >
      <span>KICAU MANIA DESKTOP / CCTV WORKSPACE</span><span
        >Metrik {engine.metricConnection} · Video RTSP/TCP · Cloud HTTPS</span
      >
    </footer>
  </div>
</div>
