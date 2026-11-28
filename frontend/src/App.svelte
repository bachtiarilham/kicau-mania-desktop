<script lang="ts">
  import { onMount } from "svelte";
  import {
    Cctv,
    Settings,
    HardDrive,
    CloudDownload,
    RefreshCw,
    LogOut,
    LayoutDashboard,
    ClipboardCheck,
    PanelLeftClose,
    Menu,
  } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import SettingsPage from "$lib/components/layout/settings-page.svelte";
  import ArenaSettings from "$lib/components/layout/arena-settings.svelte";
  import Recordings from "$lib/components/recordings/recordings-page.svelte";
  import SessionBar from "$lib/components/sessions/session-bar.svelte";
  import LoginPage from "$lib/components/auth/login-page.svelte";
  import OperatorDashboard from "$lib/components/operator/operator-dashboard.svelte";
  import RoundChecklistPage from "$lib/components/checklist/round-checklist-page.svelte";
  import ZoneGrid3x2 from "$lib/components/cameras/zone-grid-3x2.svelte";
  import { api, message } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  import { auth } from "$lib/state/auth.svelte";
  import { roundStore } from "$lib/state/round.svelte";

  let tab = $state("operator");
  let sidebarOpen = $state(true);
  let noticeDismissed = $state(false);
  let checkingCloudSession = $state(true);
  let cloudSessionError = $state("");

  $effect(() => {
    noticeDismissed = false;
    if (checkingCloudSession || !auth.isLoggedIn || auth.role === "juri" || !engine.notice) return;
    const timer = setTimeout(() => (noticeDismissed = true), 15_000);
    return () => clearTimeout(timer);
  });

  function toggleSidebar() {
    sidebarOpen = !sidebarOpen;
    setTimeout(() => {
      window.dispatchEvent(new Event("resize"));
    }, 50);
  }

  $effect(() => {
    if (checkingCloudSession || !auth.arenaId) return;
    if (auth.role === "juri") return roundStore.followArena(auth.arenaId);
    if (auth.role === "operator" || auth.role === "admin") {
      void roundStore.recoverOperatorRound(auth.arenaId);
    }
  });

  const allNav = [
    { id: "checklist", label: "Checklist & Kontrol Ronde", icon: ClipboardCheck, roles: ["operator", "admin"] },
    { id: "operator", label: "Livestream (Full Video 24)", icon: LayoutDashboard, roles: ["operator", "admin"] },
    { id: "cameras", label: "Monitor Juri (Grid 3x2)", icon: Cctv, roles: ["admin"] },
    { id: "recordings", label: "Rekaman & Arsip", icon: HardDrive, roles: ["operator", "admin"] },
    { id: "arena-settings", label: "Pengaturan Arena", icon: Settings, roles: ["operator", "admin"] },
    { id: "settings", label: "Pengaturan desktop", icon: Settings, roles: ["admin"] },
  ];

  let visibleNav = $derived(
    allNav.filter((item) => auth.role && item.roles.includes(auth.role))
  );

  $effect(() => {
    if (auth.role === "operator" && !visibleNav.some((item) => item.id === tab)) {
      tab = "checklist";
    }
  });

  onMount(() => {
    const off = engine.listen();
    void engine.connect(auth.role === "juri");
    void verifyCloudSession();
    let pending = false;
    const timer = setInterval(async () => {
      if (pending) return;
      pending = true;
      try {
        await engine.refresh(auth.role === "juri");
      } finally {
        pending = false;
      }
    }, 3000);
    return () => {
      off();
      clearInterval(timer);
    };
  });

  async function verifyCloudSession() {
    try {
      if (!auth.isLoggedIn) return;
      const profile = await api.cloud("/auth/me");
      if (!profile?.id || (auth.userEmail && String(profile.email || "").toLowerCase() !== auth.userEmail.toLowerCase())) {
        throw Error("Identitas akun cloud berubah. Silakan masuk kembali.");
      }
      auth.setUserId(String(profile.id));
    } catch (error) {
      cloudSessionError = message(error);
      roundStore.clearCloudRound();
      auth.logout();
    } finally {
      checkingCloudSession = false;
    }
  }

  async function navigate(id: string) {
    if (!visibleNav.some((item) => item.id === id)) return;
    if ((tab === "cameras" || tab === "operator") && id !== tab) {
      await api.hidePlayers();
    }
    tab = id;
  }

  async function handleLogout() {
    await Promise.allSettled([api.hidePlayers(), api.logout()]);
    cloudSessionError = "";
    roundStore.clearCloudRound();
    auth.logout();
  }
</script>

{#snippet workspaceBrand(workspace: string, onCollapse: () => void)}
  <div class="flex items-center justify-between px-5 py-5">
    <div class="flex items-center gap-3">
      <div class="rounded-xl bg-primary p-2 text-primary-foreground">
        <Cctv size={20} />
      </div>
      <div>
        <div class="text-sm font-semibold tracking-tight">KICAU MANIA</div>
        <div class="mt-0.5 text-[9px] tracking-[.2em] text-muted-foreground uppercase">{workspace}</div>
      </div>
    </div>
    <button
      type="button"
      onclick={onCollapse}
      aria-label="Ringkas panel samping"
      title="Ringkas panel samping"
      class="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
    >
      <PanelLeftClose size={16} />
    </button>
  </div>
{/snippet}

{#if checkingCloudSession}
  <div class="grid h-screen place-items-center bg-background text-sm text-muted-foreground" role="status">Memeriksa sesi cloud…</div>
{:else if !auth.isLoggedIn}
  <!-- Fullscreen Login Page -->
  <LoginPage sessionError={cloudSessionError} />
{:else if auth.role === "juri"}
  <!-- Layar penilaian juri -->
  <div class="flex h-screen w-screen flex-col overflow-hidden bg-background text-foreground select-none">
    <main class="min-h-0 flex-1 overflow-hidden">
      <ZoneGrid3x2>
        {#snippet identity(onCollapse)}
          {@render workspaceBrand("Juri Arena", onCollapse)}
        {/snippet}
        {#snippet account(collapsed)}
          <div class="flex shrink-0 items-center border-t border-border {collapsed ? 'justify-center py-3' : 'justify-between gap-2 p-4'}">
            {#if !collapsed}<span class="truncate text-[10px] text-muted-foreground">{auth.userEmail}</span>{/if}
            <Button variant="ghost" size="icon-xs" onclick={handleLogout} aria-label="Ganti Akun" title="Ganti Akun" class="text-muted-foreground hover:text-foreground">
              <LogOut class="size-3.5" />
            </Button>
          </div>
        {/snippet}
      </ZoneGrid3x2>
    </main>
  </div>
{:else}
  <!-- Operator / Admin Workspace -->
  <div class="flex h-screen overflow-hidden bg-background">
    {#if sidebarOpen}
      <aside class="flex w-56 shrink-0 flex-col border-r border-border bg-card relative z-20">
        {@render workspaceBrand(auth.role === "admin" ? "Admin Workspace" : "Operator Workspace", toggleSidebar)}

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

      {#if !["checklist", "operator", "cameras", "recordings", "arena-settings"].includes(tab)}
        <SessionBar />
      {/if}

      <main
        data-viewport
        class="relative min-h-0 flex-1 flex flex-col {['checklist', 'cameras', 'operator'].includes(tab) ? 'overflow-hidden p-0' : 'overflow-y-auto p-6'}"
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

        {#if tab === "checklist"}
          <div class="flex-1 min-h-0 w-full overflow-hidden">
            <RoundChecklistPage />
          </div>
        {:else if tab === "operator"}
          <div class="flex-1 min-h-0 w-full overflow-hidden">
            <OperatorDashboard />
          </div>
        {:else if tab === "cameras" && auth.role === "admin"}
          <div class="flex-1 min-h-0 w-full overflow-hidden">
            <ZoneGrid3x2 />
          </div>
        {:else if tab === "recordings"}
          <Recordings />
        {:else if tab === "archive"}
          <Recordings cloud />
        {:else if tab === "arena-settings"}
          <ArenaSettings />
        {:else if tab === "settings"}
          <SettingsPage />
        {:else}
          <OperatorDashboard />
        {/if}
      </main>
    </div>
  </div>
{/if}
