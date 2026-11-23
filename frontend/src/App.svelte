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
    PanelLeftClose,
    Menu,
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
  let sidebarOpen = $state(true);
  let noticeDismissed = $state(false);

  function toggleSidebar() {
    sidebarOpen = !sidebarOpen;
    setTimeout(() => {
      window.dispatchEvent(new Event("resize"));
    }, 50);
  }

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
  <!-- Layar penilaian juri -->
  <div class="flex h-screen w-screen flex-col overflow-hidden bg-background text-foreground select-none">
    <main class="min-h-0 flex-1 overflow-hidden">
      <ZoneGrid3x2>
        {#snippet identity(collapsed)}
          <header class="flex shrink-0 flex-col items-center gap-2 border-b border-border py-3 text-center {collapsed ? 'px-1' : 'px-4'}">
            <span class="grid size-7 place-items-center rounded-lg bg-amber-500/20 text-xs font-bold text-amber-700" aria-hidden="true">KM</span>
            {#if !collapsed}
              <h1 class="text-sm font-bold tracking-normal">Layar Penilaian Dewan Juri</h1>
              <span class="rounded bg-amber-500/20 px-2 py-0.5 text-[10px] font-semibold text-amber-700">JURI ARENA</span>
              <span class="max-w-full truncate text-xs font-medium text-muted-foreground">{auth.userName}</span>
            {/if}
            <Button variant="outline" size={collapsed ? "icon-sm" : "sm"} onclick={handleLogout} aria-label="Ganti Akun" title="Ganti Akun">
              <LogOut class="size-3.5" />
              {#if !collapsed}Ganti Akun{/if}
            </Button>
          </header>
        {/snippet}
      </ZoneGrid3x2>
    </main>
  </div>
{:else}
  <!-- Operator / Admin Workspace -->
  <div class="flex h-screen overflow-hidden bg-background">
    {#if sidebarOpen}
      <aside class="flex w-56 shrink-0 flex-col border-r border-border bg-card relative z-20">
        <div class="flex items-center justify-between px-5 py-5">
          <div class="flex items-center gap-3">
            <div class="rounded-xl bg-primary p-2 text-primary-foreground">
              <Cctv size={20} />
            </div>
            <div>
              <div class="text-sm font-semibold tracking-tight">KICAU MANIA</div>
              <div class="mt-0.5 text-[9px] tracking-[.2em] text-muted-foreground uppercase">
                {auth.role === "admin" ? "ADMIN WORKSPACE" : "OPERATOR WORKSPACE"}
              </div>
            </div>
          </div>
          <button
            onclick={toggleSidebar}
            title="Sembunyikan Sidebar"
            class="rounded-lg p-1.5 text-muted-foreground hover:bg-accent hover:text-foreground transition-colors"
          >
            <PanelLeftClose size={16} />
          </button>
        </div>

        <div class="px-5 pb-2 text-[9px] tracking-[.2em] text-muted-foreground">
          MENU OPERASIONAL
        </div>
        <nav class="space-y-1 px-3">
          {#each visibleNav as item}
            <button
              onclick={() => navigate(item.id)}
              class="flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left text-xs transition-colors"
              class:bg-accent={tab === item.id}
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
              <div class="text-xs font-semibold text-foreground truncate">{auth.userName}</div>
              <div class="text-[10px] text-muted-foreground truncate">{auth.userEmail}</div>
            </div>
            <Button
              variant="ghost"
              size="icon-xs"
              onclick={handleLogout}
              title="Keluar / Ganti Akun"
              class="text-muted-foreground hover:text-foreground"
            >
              <LogOut size={14} />
            </Button>
          </div>

          <div class="flex items-center justify-between text-[10px] text-muted-foreground">
            <div class="flex items-center gap-2">
              <span
                class="h-1.5 w-1.5 rounded-full"
                class:bg-primary={engine.connected}
                class:bg-destructive={!engine.connected}
              ></span>
              <span>{engine.connected ? "Engine terhubung" : "Engine terputus"}</span>
            </div>
            <button
              onclick={() => engine.connect()}
              title="Hubungkan ulang engine"
              class="rounded p-1 hover:bg-accent hover:text-foreground transition-colors"
            >
              <RefreshCw size={12} />
            </button>
          </div>
        </div>
      </aside>
    {/if}

    <div class="relative flex min-w-0 flex-1 flex-col h-full overflow-hidden">
      <!-- Floating sidebar open button when sidebar is hidden -->
      {#if !sidebarOpen}
        <button
          onclick={toggleSidebar}
          title="Tampilkan Sidebar"
          class="absolute top-3 left-3 z-30 flex h-8 w-8 items-center justify-center rounded-lg border border-border bg-card text-foreground shadow-md backdrop-blur hover:bg-accent hover:text-foreground transition-all"
        >
          <Menu size={16} />
        </button>
      {/if}

      {#if !["operator", "cameras"].includes(tab)}
        <SessionBar />
      {/if}

      <main
        data-viewport
        class="relative min-h-0 flex-1 flex flex-col {['cameras', 'operator'].includes(tab) ? 'overflow-hidden p-0' : 'overflow-y-auto p-6'}"
      >
        {#if engine.error}
          <div class="error m-3 mb-1 shrink-0 flex items-center justify-between" role="alert">
            <div>
              {engine.error}
              <button class="ml-2 underline" onclick={() => engine.connect()}>
                Hubungkan ulang
              </button>
            </div>
          </div>
        {/if}
        {#if engine.notice && !noticeDismissed}
          <div class="notice m-3 mb-1 shrink-0 flex items-center justify-between">
            <span class="text-xs">{engine.notice}</span>
            <button
              class="ml-3 rounded p-0.5 text-xs opacity-70 hover:opacity-100 transition-opacity"
              onclick={() => (noticeDismissed = true)}
              title="Tutup pesan"
            >
              ✕
            </button>
          </div>
        {/if}

        {#if tab === "operator"}
          <div class="flex-1 min-h-0 w-full overflow-hidden">
            <OperatorDashboard />
          </div>
        {:else if tab === "cameras"}
          <div class="flex-1 min-h-0 w-full overflow-hidden">
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
    </div>
  </div>
{/if}
