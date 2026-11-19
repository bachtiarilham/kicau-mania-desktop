<script lang="ts">
  import { onMount } from "svelte";
  import {
    Cctv,
    Settings,
    HardDrive,
    Upload,
    CloudDownload,
    RefreshCw,
    LogOut,
    LayoutDashboard,
  } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import SettingsPage from "$lib/components/layout/settings-page.svelte";
  import Recordings from "$lib/components/recordings/recordings-page.svelte";
  import SessionBar from "$lib/components/sessions/session-bar.svelte";
  import LoginPage from "$lib/components/auth/login-page.svelte";
  import OperatorDashboard from "$lib/components/operator/operator-dashboard.svelte";
  import ZoneGrid3x2 from "$lib/components/cameras/zone-grid-3x2.svelte";
  import { api } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  import { auth } from "$lib/state/auth.svelte";

  let tab = $state("operator");

  $effect(() => {
    if (auth.role === "juri") {
      tab = "cameras";
    } else if (auth.role === "operator" && !["operator", "cameras", "recordings", "uploads"].includes(tab)) {
      tab = "operator";
    }
  });

  const allNav = [
    { id: "operator", label: "Operator Arena (Live 24)", icon: LayoutDashboard, roles: ["operator", "admin"] },
    { id: "cameras", label: "Monitor Juri (Grid 3x2)", icon: Cctv, roles: ["juri", "operator", "admin"] },
    { id: "recordings", label: "Rekaman lokal", icon: HardDrive, roles: ["operator", "admin"] },
    { id: "uploads", label: "Antrean upload", icon: Upload, roles: ["operator", "admin"] },
    { id: "archive", label: "Rekaman cloud", icon: CloudDownload, roles: ["admin"] },
    { id: "settings", label: "Pengaturan desktop", icon: Settings, roles: ["admin"] },
  ];

  let visibleNav = $derived(
    allNav.filter((item) => auth.role && item.roles.includes(auth.role))
  );

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
    if ((tab === "cameras" || tab === "operator") && id !== tab) {
      await api.hidePlayers();
    }
    tab = id;
  }

  async function handleLogout() {
    await api.hidePlayers();
    auth.logout();
  }
</script>

{#if !auth.isLoggedIn}
  <!-- Fullscreen Login Page -->
  <LoginPage />
{:else if auth.role === "juri"}
  <!-- Pure Juri Display (100% Fullscreen Grid 3x2 + Juri HUD) -->
  <div class="flex h-screen w-screen flex-col overflow-hidden bg-[#0c1011] text-zinc-100 select-none">
    <header class="flex h-12 shrink-0 items-center justify-between border-b border-white/10 bg-[#12181a] px-4">
      <div class="flex items-center gap-2.5">
        <span class="flex h-7 w-7 items-center justify-center rounded-lg bg-amber-500/20 text-amber-400 font-bold text-xs ring-1 ring-amber-500/30">
          KM
        </span>
        <div class="flex items-center gap-2">
          <span class="text-sm font-bold text-white">Layar Penilaian Dewan Juri</span>
          <span class="rounded bg-amber-500/20 px-2 py-0.5 text-[10px] font-semibold text-amber-300 uppercase">
            JURI ARENA
          </span>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <span class="text-xs text-zinc-400 font-medium hidden sm:inline">{auth.userName}</span>
        <Button
          variant="outline"
          size="sm"
          onclick={handleLogout}
          class="h-7 text-xs border-white/10 bg-white/5 text-zinc-300 hover:bg-white/10 hover:text-white"
        >
          <LogOut class="size-3.5" />
          <span>Ganti Akun</span>
        </Button>
      </div>
    </header>

    <main class="min-h-0 flex-1 overflow-hidden">
      <ZoneGrid3x2 showZoneBar={true} showJuriHud={true} />
    </main>
  </div>
{:else}
  <!-- Operator / Admin Workspace -->
  <div class="flex h-screen overflow-hidden bg-background">
    <aside class="flex w-56 shrink-0 flex-col border-r border-border bg-[#131c1d]">
      <div class="flex items-center gap-3 px-5 py-6">
        <div class="rounded-xl bg-primary p-2 text-primary-foreground">
          <Cctv size={22} />
        </div>
        <div>
          <div class="text-sm font-semibold tracking-tight">KICAU MANIA</div>
          <div class="mt-0.5 text-[9px] tracking-[.2em] text-muted-foreground uppercase">
            {auth.role === "admin" ? "ADMIN WORKSPACE" : "OPERATOR WORKSPACE"}
          </div>
        </div>
      </div>

      <div class="px-5 pb-2 text-[9px] tracking-[.2em] text-[#627975]">
        MENU OPERASIONAL
      </div>
      <nav class="space-y-1 px-3">
        {#each visibleNav as item}
          <button
            onclick={() => navigate(item.id)}
            class="flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left text-xs transition-colors"
            class:bg-[#29362b]={tab === item.id}
            class:text-primary={tab === item.id}
            class:text-muted-foreground={tab !== item.id}
          >
            <item.icon size={16} />{item.label}
          </button>
        {/each}
      </nav>

      <div class="mt-auto border-t border-border p-4">
        <div class="mb-3 flex items-center justify-between">
          <div class="min-w-0">
            <div class="text-xs font-semibold text-zinc-200 truncate">{auth.userName}</div>
            <div class="text-[10px] text-zinc-500 truncate">{auth.userEmail}</div>
          </div>
          <Button
            variant="ghost"
            size="icon-xs"
            onclick={handleLogout}
            title="Keluar / Ganti Akun"
            class="text-zinc-400 hover:text-white"
          >
            <LogOut size={14} />
          </Button>
        </div>

        <div class="flex items-center gap-2 text-[10px] text-muted-foreground">
          <span
            class="h-1.5 w-1.5 rounded-full"
            class:bg-primary={engine.connected}
            class:bg-destructive={!engine.connected}
          ></span>
          <span>{engine.connected ? "Engine terhubung" : "Engine terputus"}</span>
        </div>
      </div>
    </aside>

    <div class="flex min-w-0 flex-1 flex-col">
      <header
        class="flex h-13 shrink-0 items-center justify-between border-b border-border px-6"
      >
        <div class="text-xs text-muted-foreground">
          Arena / <span class="text-foreground">{visibleNav.find((n) => n.id === tab)?.label}</span>
        </div>
        <div class="flex items-center gap-3">
          <span class="chip">WINDOWS · LOCAL ENGINE</span>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Hubungkan ulang engine"
            onclick={() => engine.connect()}
          >
            <RefreshCw size={14} />
          </Button>
        </div>
      </header>

      {#if !["operator", "cameras"].includes(tab)}
        <SessionBar />
      {/if}

      <main
        data-viewport
        class="min-h-0 flex-1 {['cameras', 'operator'].includes(tab) ? 'overflow-hidden p-0' : 'overflow-y-auto p-6'}"
      >
        {#if engine.error}
          <div class="error m-4 mb-2" role="alert">
            {engine.error}
            <button class="ml-2 underline" onclick={() => engine.connect()}>
              Hubungkan ulang
            </button>
          </div>
        {/if}
        {#if engine.notice}
          <div class="notice m-4 mb-2">
            {engine.notice}
          </div>
        {/if}

        {#if tab === "operator"}
          <div class="h-full w-full overflow-hidden">
            <OperatorDashboard />
          </div>
        {:else if tab === "cameras"}
          <div class="h-full w-full overflow-hidden">
            <ZoneGrid3x2 />
          </div>
        {:else if tab === "recordings"}
          <Recordings />
        {:else if tab === "uploads"}
          <Recordings queue />
        {:else if tab === "archive"}
          <Recordings cloud />
        {:else if tab === "settings"}
          <SettingsPage />
        {:else}
          <OperatorDashboard />
        {/if}
      </main>

      <footer
        class="flex h-7 shrink-0 items-center justify-between border-t border-border px-6 text-[9px] tracking-wide text-muted-foreground"
      >
        <span>KICAU MANIA DESKTOP / SURVEILLANCE WORKSPACE</span>
        <span>Metrik {engine.metricConnection} · Video RTSP/TCP · Cloud HTTPS</span>
      </footer>
    </div>
  </div>
{/if}
