<script lang="ts">
  import { onMount } from "svelte";
  import {
    Shield,
    Lock,
    Mail,
    Globe,
    AlertCircle,
    Zap,
  } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { auth, type UserRole } from "$lib/state/auth.svelte";
  import { roundStore } from "$lib/state/round.svelte";
  import { api, message } from "$lib/services/api";

  let { sessionError = "" }: { sessionError?: string } = $props();
  let dismissSessionError = $state(false);

  let email = $state("");
  let password = $state("");
  let error = $state("");
  let busy = $state(false);

  let showServerConfig = $state(false);
  let currentCloudUrl = $state("https://dev.megaallstararena.com/api/v1");
  let savingUrl = $state(false);
  let urlMessage = $state("");

  function handleAutoFill(role: UserRole) {
    if (role === "juri") {
      email = "juri.1@mock.kicau.test";
      password = "KicauMock2026!";
    } else if (role === "operator") {
      email = "operator.a@mock.kicau.test";
      password = "KicauMock2026!";
    } else {
      email = "admin@mock.kicau.test";
      password = "KicauMock2026!";
    }
    error = "";
  }

  onMount(() => {
    void api
      .config()
      .then((cfg) => {
        if (cfg?.cloud_url) currentCloudUrl = cfg.cloud_url;
      })
      .catch(() => {});
  });

  async function handleSaveCloudUrl() {
    if (!currentCloudUrl.trim()) {
      urlMessage = "Gagal: URL server tidak boleh kosong";
      return;
    }
    savingUrl = true;
    urlMessage = "";
    try {
      const existing = await api.config();
      await api.saveConfig({
        ...existing,
        cloud_url: currentCloudUrl.trim(),
      });
      urlMessage = "Berhasil: Alamat server cloud telah disimpan";
    } catch (err: any) {
      urlMessage = "Gagal: " + (message(err) || "Format URL cloud tidak valid");
    } finally {
      savingUrl = false;
    }
  }

  async function handleFormSubmit(e: Event) {
    e.preventDefault();
    dismissSessionError = true;
    if (!email.trim() || !password) {
      error = "Silakan masukkan email dan kata sandi.";
      return;
    }

    busy = true;
    error = "";
    try {
      const profile = await api.login(email.trim(), password);
      let role: UserRole = "operator";
      const roles: string[] = Array.isArray(profile?.roles)
        ? profile.roles.map((r: any) => String(r).toUpperCase())
        : [];
      const arenaRoles: string[] = Array.isArray(profile?.arena_roles)
        ? profile.arena_roles.map((r: any) => String(r?.role || "").toUpperCase())
        : [];
      const activeRole = String(profile?.active_scope?.role || "").toUpperCase();
      const allRoles = activeRole ? [activeRole] : [...roles, ...arenaRoles];

      if (allRoles.includes("ADMIN")) {
        role = "admin";
      } else if (allRoles.includes("OPERATOR")) {
        role = "operator";
      } else if (allRoles.includes("JUDGE") || allRoles.includes("JURI")) {
        role = "juri";
      }

      const userName = profile?.name || profile?.nickname || "";
      const arenaId = profile?.active_scope?.arena_id || profile?.arena_roles?.[0]?.arena_id || profile?.arena_id || "";
      roundStore.clearCloudRound();
      auth.login(role, profile?.email || email.trim(), userName, undefined, arenaId);
    } catch (err: any) {
      error = message(err) || "Gagal terhubung ke server atau kredensial salah.";
    } finally {
      busy = false;
    }
  }
</script>

<div class="relative flex min-h-screen w-full items-center justify-center overflow-y-auto bg-background p-4 text-foreground select-none">
  <div class="pointer-events-none fixed inset-0 bg-[radial-gradient(ellipse_80%_60%_at_50%_-20%,rgba(16,185,129,0.12),transparent_70%)]"></div>
  <div class="pointer-events-none fixed inset-0 opacity-10" style="background-image: linear-gradient(rgba(40,75,61,0.24) 1px, transparent 1px), linear-gradient(90deg, rgba(40,75,61,0.24) 1px, transparent 1px); background-size: 36px 36px;"></div>

  <div class="relative z-10 w-full max-w-md py-6">
    <div class="mb-6 text-center">
      <div class="mx-auto mb-3 flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-700 text-white shadow-xl shadow-emerald-500/25 ring-4 ring-border">
        <Shield class="h-7 w-7" />
      </div>
      <h1 class="text-2xl font-black tracking-tight text-foreground">
        KICAU MANIA DESKTOP
      </h1>
      <p class="mt-1 text-xs tracking-widest text-emerald-400 font-semibold uppercase">
        Arena Surveillance & Competition Workspace
      </p>
    </div>

    <div class="rounded-2xl border border-border bg-card p-6 shadow-xl">
      <div class="mb-4 flex flex-wrap items-center justify-between gap-2 border-b border-border pb-3">
        <div>
          <h2 class="text-sm font-bold text-foreground">Masuk ke Server Cloud</h2>
          <p class="text-xs text-muted-foreground">Gunakan akun Admin, Operator, atau Juri yang terdaftar di server.</p>
        </div>
        <div class="flex items-center gap-1.5">
          <button
            type="button"
            onclick={() => handleAutoFill("juri")}
            class="rounded bg-accent px-2 py-1 text-[11px] font-semibold text-amber-400 hover:bg-accent/80 cursor-pointer"
          >
            Auto: Juri
          </button>
          <button
            type="button"
            onclick={() => handleAutoFill("operator")}
            class="rounded bg-accent px-2 py-1 text-[11px] font-semibold text-emerald-400 hover:bg-accent/80 cursor-pointer"
          >
            Auto: Operator
          </button>
          <button
            type="button"
            onclick={() => handleAutoFill("admin")}
            class="rounded bg-accent px-2 py-1 text-[11px] font-semibold text-sky-400 hover:bg-accent/80 cursor-pointer"
          >
            Auto: Admin
          </button>
        </div>
      </div>

      {#if error || (!dismissSessionError && sessionError)}
        <div class="mb-4 flex items-center gap-2 rounded-xl border border-red-500/40 bg-red-500/15 p-3 text-xs text-red-400">
          <AlertCircle class="h-4 w-4 shrink-0 text-red-400" />
          <span>{error || sessionError}</span>
        </div>
      {/if}

      <form onsubmit={handleFormSubmit} class="space-y-4">
        <div class="space-y-1.5">
          <Label for="login-email" class="text-xs text-foreground">Email Akun</Label>
          <div class="relative">
            <Input
              id="login-email"
              type="email"
              bind:value={email}
              placeholder="admin@mock.kicau.test"
              class="border-border bg-muted pl-9 text-xs text-foreground"
            />
            <Mail class="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
          </div>
        </div>

        <div class="space-y-1.5">
          <Label for="login-pwd" class="text-xs text-foreground">Kata Sandi</Label>
          <div class="relative">
            <Input
              id="login-pwd"
              type="password"
              bind:value={password}
              placeholder="••••••••"
              class="border-border bg-muted pl-9 text-xs text-foreground"
            />
            <Lock class="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
          </div>
        </div>

        <Button
          type="submit"
          disabled={busy}
          class="w-full bg-emerald-600 text-xs font-bold text-white hover:bg-emerald-500"
        >
          <Zap class="h-3.5 w-3.5 mr-1.5" />
          <span>{busy ? "Menghubungkan ke Server..." : "Masuk ke Sistem"}</span>
        </Button>
      </form>

      <div class="mt-6 border-t border-border pt-4">
        <div class="flex items-center justify-between gap-2">
          <button
            type="button"
            onclick={() => (showServerConfig = !showServerConfig)}
            class="flex items-center gap-1.5 text-[11px] font-semibold text-muted-foreground hover:text-foreground transition cursor-pointer min-w-0"
          >
            <Globe class="h-3.5 w-3.5 text-primary shrink-0" />
            <span class="shrink-0">Server:</span>
            <span class="font-mono text-[10px] text-primary/80 truncate">{currentCloudUrl}</span>
          </button>
          <button
            type="button"
            onclick={() => (showServerConfig = !showServerConfig)}
            class="text-[10px] font-bold text-primary hover:underline cursor-pointer shrink-0"
          >
            {showServerConfig ? "Tutup" : "Ubah"}
          </button>
        </div>

        {#if showServerConfig}
          <div class="mt-3 rounded-xl border border-border bg-muted/60 p-3 space-y-2.5">
            <div class="text-[10px] text-muted-foreground">
              Alamat HTTPS API Cloud Server (contoh: <code>https://dev.megaallstararena.com/api/v1</code>).
            </div>
            <div class="flex gap-2">
              <Input
                type="text"
                bind:value={currentCloudUrl}
                placeholder="https://api.domain.com/api/v1"
                class="bg-card text-xs font-mono"
              />
              <Button
                type="button"
                size="sm"
                disabled={savingUrl}
                onclick={handleSaveCloudUrl}
                class="text-xs font-bold bg-primary text-black hover:bg-primary/90 shrink-0"
              >
                {savingUrl ? "Menyimpan..." : "Simpan"}
              </Button>
            </div>
            {#if urlMessage}
              <div class="text-[11px] font-semibold {urlMessage.includes('Gagal') ? 'text-red-400' : 'text-emerald-400'}">
                {urlMessage}
              </div>
            {/if}
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>
