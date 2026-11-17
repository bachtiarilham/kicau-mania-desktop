<script lang="ts">
  import { onMount } from "svelte";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { api, all, message, type Row } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  let user = $state<Row | null>(null),
    email = $state(""),
    password = $state(""),
    arenas = $state<Row[]>([]),
    rounds = $state<Row[]>([]),
    tickets = $state<Row[]>([]),
    arena = $state(""),
    round = $state(""),
    sessionID = $state(""),
    selected = $state<Record<string, string>>({}),
    busy = $state(false),
    error = $state(""),
    notice = $state("");
  async function load() {
    arenas = await all("/arenas");
    if (!arena) arena = engine.pairing.arena_id || arenas[0]?.id || "";
    await loadRounds();
  }
  onMount(() => {
    void api
      .cloud<Row>("/auth/me")
      .then(async (v) => {
        user = v;
        await load();
      })
      .catch(() => {});
  });
  async function login() {
    busy = true;
    error = "";
    try {
      user = await api.login(email, password);
      await load();
    } catch (e) {
      error = message(e);
    } finally {
      password = "";
      busy = false;
    }
  }
  async function logout() {
    try {
      await api.logout();
    } catch (e) {
      error = message(e);
    } finally {
      user = null;
      arenas = [];
      rounds = [];
      tickets = [];
    }
  }
  async function loadRounds() {
    round = "";
    tickets = [];
    rounds = arena ? await all(`/arenas/${arena}/rounds`) : [];
  }
  async function loadTickets() {
    tickets = round
      ? await all(`/registrations?arena_id=${arena}&round_id=${round}`)
      : [];
    selected = {};
    if (!sessionID) sessionID = await api.sessionID();
  }
  async function pair() {
    busy = true;
    error = "";
    try {
      await api.pair(arena);
      await engine.refresh();
      notice = "Engine tersambung ke arena. Upload berjalan melalui local-app.";
    } catch (e) {
      error = message(e);
    } finally {
      busy = false;
    }
  }
  async function bind() {
    busy = true;
    error = "";
    notice = "";
    try {
      const mappings = engine.channels
        .filter((c) => selected[c.id])
        .map((c) => {
          const t = tickets.find((v) => v.id === selected[c.id])!;
          return {
            channel_id: c.id,
            registration_id: t.id,
            gantangan_id: t.gantangan_id,
          };
        });
      if (!mappings.length) throw Error("Pilih setidaknya satu tiket peserta");
      const result = await api.bind({
        engine_id: engine.pairing.engine_id,
        session_id: sessionID,
        round_id: round,
        mappings,
      });
      if (!result.existing_session) engine.draftID = sessionID;
      notice =
        "Mapping tersimpan. Mulai sesi dari tombol di header. Sesi lama dengan ID ini juga dapat melanjutkan antrean upload.";
    } catch (e) {
      error = message(e);
    } finally {
      busy = false;
    }
  }
  async function revoke() {
    busy = true;
    try {
      await api.revoke(engine.pairing.engine_id);
      notice =
        "Credential engine dicabut. Upload berikutnya memerlukan pairing baru.";
    } catch (e) {
      error = message(e);
    } finally {
      busy = false;
    }
  }
</script>

<div class="section-head">
  <div>
    <h1>Hubungkan ke lomba</h1>
    <p class="subtitle">
      Tetapkan ronde dan tiket peserta agar rekaman sampai ke pemilik yang
      benar.
    </p>
  </div>
  {#if user}<Button variant="outline" onclick={logout}>Logout cloud</Button
    >{/if}
</div>
{#if error}<div class="error mb-4" role="alert">
    {error}
  </div>{/if}{#if notice}<div class="notice mb-4">{notice}</div>{/if}
{#if !user}<form
    class="panel mx-auto max-w-md space-y-5"
    onsubmit={(e) => {
      e.preventDefault();
      void login();
    }}
  >
    <div>
      <h2>Login petugas</h2>
      <p class="subtitle">CCTV lokal tetap tersedia tanpa login.</p>
    </div>
    <label class="field"
      >Email<Input
        type="email"
        bind:value={email}
        required
        autocomplete="username"
      /></label
    ><label class="field"
      >Password<Input
        type="password"
        bind:value={password}
        required
        autocomplete="current-password"
      /></label
    ><Button type="submit" class="w-full" disabled={busy}
      >{busy ? "Menghubungkan…" : "Masuk ke cloud"}</Button
    >
  </form>{:else}<div class="space-y-5">
    <div class="panel flex items-center justify-between">
      <div>
        <h2>{user.name}</h2>
        <p class="subtitle">{user.email}</p>
      </div>
      <span class="chip"
        >{engine.pairing.paired ? "ENGINE PAIRED" : "BELUM PAIRED"}</span
      >
    </div>
    <section class="panel">
      <h2 class="mb-5">1. Arena dan pairing engine</h2>
      <div class="flex items-end gap-3">
        <label class="field flex-1"
          >Arena<select
            bind:value={arena}
            onchange={() => loadRounds().catch((e) => (error = message(e)))}
            ><option value="">Pilih arena</option>{#each arenas as a}<option
                value={a.id}>{a.name}</option
              >{/each}</select
          ></label
        ><Button onclick={pair} disabled={busy || !arena}
          >Pair engine (ADMIN)</Button
        >{#if engine.pairing.paired}<Button
            variant="outline"
            onclick={revoke}
            disabled={busy}>Cabut akses</Button
          >{/if}
      </div>
    </section>
    <section class="panel">
      <h2 class="mb-5">2. Ronde dan mapping tiket</h2>
      <div class="grid grid-cols-2 gap-4">
        <label class="field"
          >Ronde<select
            bind:value={round}
            onchange={() => loadTickets().catch((e) => (error = message(e)))}
            ><option value="">Pilih ronde</option>{#each rounds as r}<option
                value={r.id}>{r.title} · {r.status}</option
              >{/each}</select
          ></label
        ><label class="field"
          >ID sesi lokal (baru atau rekaman sebelumnya)<Input
            bind:value={sessionID}
          /></label
        >
      </div>
      <div class="mt-5 overflow-x-auto">
        <table>
          <thead
            ><tr><th>Kamera</th><th>Tiket peserta / burung / gantangan</th></tr
            ></thead
          ><tbody
            >{#each engine.channels as c}<tr
                ><td>{c.id}</td><td
                  ><select
                    class="w-full"
                    aria-label={`Tiket ${c.id}`}
                    bind:value={selected[c.id]}
                    ><option value="">Simpan lokal saja</option
                    >{#each tickets.filter((t) => t.status === "CONFIRMED" && t.payment_status === "VERIFIED") as t}<option
                        value={t.id}
                        >{t.registrant_name} · {t.bird_name} · Gantangan {t.gantangan_number}</option
                      >{/each}</select
                  ></td
                ></tr
              >{/each}</tbody
          >
        </table>
      </div>
      <div class="mt-5 flex justify-end">
        <Button
          onclick={bind}
          disabled={busy || !round || !sessionID || !engine.pairing.paired}
          >Validasi & simpan mapping</Button
        >
      </div>
    </section>
  </div>{/if}
