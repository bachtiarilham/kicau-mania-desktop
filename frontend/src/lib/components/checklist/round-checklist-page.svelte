<script lang="ts">
  import { untrack } from "svelte";
  import {
    CheckSquare,
    Square,
    Search,
    Phone,
    Users,
    Clock,
    RefreshCw,
    Trophy,
    CheckCircle2,
    AlertCircle,
    ListChecks,
  } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import OperatorHeader from "../operator/operator-header.svelte";
  import QuickSaveModal from "../operator/modals/quick-save-modal.svelte";
  import RoundTimeUpDialog from "../sessions/round-timeup-dialog.svelte";
  import { gantanganChannels } from "../cameras/camera-slots";
  import CloudSessionSelector from "../sessions/cloud-session-selector.svelte";
  import {
    roundStore,
    getJudgeAssignedZone,
    getJudgeRotationSchedule,
    type ParticipantMapping,
    type ZoneId,
  } from "$lib/state/round.svelte";
  import { engine } from "$lib/state/engine.svelte";
  import { auth } from "$lib/state/auth.svelte";
  import { all, api, message, type Row } from "$lib/services/api";

  let showCloudModal = $state(false);
  let search = $state("");

  let confirmSave = $state(false);
  let sessionName = $state("");
  let selectedFormat = $state("mp4");
  let storageDir = $state("D:\\KicauRecordings");

  let cloudRounds = $state<Row[]>([]);
  let requestVersion = 0;
  let selectingRound = $state(false);
  let loadingRounds = $state(false);
  let roundsError = $state("");

  const ZONES: { id: ZoneId; start: number; end: number }[] = [
    { id: "A", start: 1, end: 6 },
    { id: "B", start: 7, end: 12 },
    { id: "C", start: 13, end: 18 },
    { id: "D", start: 19, end: 24 },
  ];

  function zoneOf(num: number): ZoneId {
    const found = ZONES.find((z) => num >= z.start && num <= z.end);
    return found ? found.id : "A";
  }

  function generateSessionName(): string {
    const d = new Date();
    const pad = (n: number) => String(n).padStart(2, "0");
    const dateStr = `${d.getFullYear()}${pad(d.getMonth() + 1)}${pad(d.getDate())}`;
    const cleanEvent = roundStore.eventName ? roundStore.eventName.replace(/\s+/g, "_") : "Lomba";
    const cleanSession = roundStore.sessionName ? roundStore.sessionName.replace(/\s+/g, "_") : "Ronde";
    return `${cleanEvent}_${cleanSession}_${dateStr}`;
  }

  async function handleStopClick() {
    if (!(await roundStore.stopRound())) return;
    if (!sessionName) {
      sessionName = generateSessionName();
    }
    void api
      .config()
      .then((cfg) => {
        if (cfg?.recordings_dir) storageDir = cfg.recordings_dir;
      })
      .catch(() => {});
    confirmSave = true;
  }

  async function handleConfirmSave() {
    confirmSave = false;
    const currentSessionId = engine.session?.id;
    await engine.stop();
    engine.notice = `Menyimpan rekaman sesi '${sessionName}' (${selectedFormat.toUpperCase()}) ke ${storageDir}...`;

    if (currentSessionId) {
      void api
        .saveSession(currentSessionId, sessionName, storageDir, selectedFormat)
        .then((savedFiles) => {
          if (Array.isArray(savedFiles) && savedFiles.length > 0) {
            engine.notice = `Berhasil menyimpan ${savedFiles.length} file rekaman sesi '${sessionName}' ke: ${storageDir}`;
          }
        })
        .catch((err) => {
          engine.error = "Gagal mengekspor rekaman: " + (err?.message || String(err));
        });
    }
  }

  async function fetchEventRounds(arenaId = auth.arenaId, eventId = roundStore.eventId) {
    const request = ++requestVersion;
    cloudRounds = [];
    roundsError = "";
    loadingRounds = !!arenaId && !!eventId;
    if (!loadingRounds) return;
    try {
      const rounds = await all(`/arenas/${encodeURIComponent(arenaId)}/rounds`);
      if (request !== requestVersion) return;
      cloudRounds = rounds.filter((row) => row.event_id === eventId)
        .sort((a, b) => (a.round_number ?? 0) - (b.round_number ?? 0));
    } catch (err) {
      if (request === requestVersion) roundsError = message(err);
    } finally {
      if (request === requestVersion) loadingRounds = false;
    }
  }

  $effect(() => {
    const arenaId = auth.arenaId;
    const eventId = roundStore.eventId;
    untrack(() => void fetchEventRounds(arenaId, eventId));
  });

  async function loadRoundIntoWorkspace(r: Row) {
    const { arenaId } = auth;
    const { eventId, eventName } = roundStore;
    const request = requestVersion;
    if (!arenaId || r.event_id !== eventId || engine.active || roundStore.busy || roundStore.state === "RUNNING" || selectingRound) return;
    selectingRound = true;
    roundsError = "";
    try {
      const path = `/arenas/${encodeURIComponent(arenaId)}/rounds/${encodeURIComponent(r.id)}`;
      const [parts, judges] = await Promise.all([
        api.cloud<Row[]>(path + "/participants"),
        api.cloud<Row[]>(path + "/judges"),
      ]);
      if (request !== requestVersion || roundStore.eventId !== eventId || auth.arenaId !== arenaId || engine.active) return;

      const channels = gantanganChannels(engine.channels);
      const mappedParticipants: ParticipantMapping[] = (parts ?? []).filter((p) => Number.isInteger(p.gantangan_number) && p.gantangan_number >= 1 && p.gantangan_number <= 24).map((p) => {
        const num = p.gantangan_number;
        return {
          registrationId: p.id,
          gantanganId: p.gantangan_id,
          gantanganNumber: num,
          birdName: p.bird_name || "",
          birdCode: p.bird_code,
          ownerName: p.owner_name || p.participant_name || p.registrant_name || "",
          ownerPhone: p.owner_phone || p.participant_phone || p.registrant_phone || "",
          channelId: channels[num - 1]?.id,
        };
      });

      const assigned = [r.field_judge_1, r.field_judge_2, r.field_judge_3, r.field_judge_4];
      const formattedJurors = assigned.some(Boolean)
        ? assigned.map((name) => name || "")
        : (judges ?? []).slice(0, 4).map((judge) => judge.judge_name || "");

      roundStore.setCloudRound({
        eventId,
        roundId: r.id,
        eventName,
        sessionName: r.title || `Sesi ${r.round_number}`,
        startTime: r.start_time || "",
        endTime: r.end_time || "",
        jurors: formattedJurors,
        jurorIds: formattedJurors.map((name, index) => {
          if (!assigned.some(Boolean)) return String(judges?.[index]?.judge_user_id ?? "");
          const normalized = name.trim().toLocaleLowerCase("id-ID").replace(/\s+/g, " ");
          const matches = (judges ?? []).filter((judge) =>
            normalized && String(judge.judge_name ?? "").trim().toLocaleLowerCase("id-ID").replace(/\s+/g, " ") === normalized,
          );
          return matches.length === 1 ? String(matches[0].judge_user_id ?? "") : "";
        }),
        participants: mappedParticipants,
      });
    } catch (err) {
      if (request === requestVersion) roundsError = message(err);
    } finally {
      selectingRound = false;
    }
  }

  const gantanganRows = $derived.by(() => {
    const q = search.trim().toLowerCase();
    const rows = Array.from({ length: 24 }, (_, idx) => {
      const num = idx + 1;
      const p = roundStore.getParticipantByGantangan(num);
      const checked = roundStore.isGantanganChecked(roundStore.roundId, num);
      return {
        num,
        zone: zoneOf(num),
        participant: p,
        checked,
      };
    });
    if (!q) return rows;
    return rows.filter((row) => {
      const text = [
        String(row.num),
        `g${String(row.num).padStart(2, "0")}`,
        `zona ${row.zone}`,
        row.participant?.birdName || "",
        row.participant?.birdCode || "",
        row.participant?.ownerName || "",
        row.participant?.ownerPhone || "",
      ]
        .join(" ")
        .toLowerCase();
      return text.includes(q);
    });
  });

  const activeNumbers = $derived(
    roundStore.participants.length > 0
      ? roundStore.participants.map((p) => p.gantanganNumber)
      : Array.from({ length: 24 }, (_, i) => i + 1)
  );

  const verifiedCount = $derived(
    activeNumbers.filter((num) => roundStore.isGantanganChecked(roundStore.roundId, num)).length
  );

  const jurorList = $derived(roundStore.jurors);
</script>

<div class="flex h-full w-full flex-col overflow-hidden bg-background text-foreground select-none">
  <OperatorHeader
    {selectingRound}
    onOpenCloudModal={() => (showCloudModal = true)}
    onStopRound={handleStopClick}
  />

  <div class="min-h-0 flex-1 overflow-y-auto p-4 space-y-4">
    <!-- Row 1: Ronde Hari Ini & Posisi Rotasi Juri -->
    <div class="grid gap-4 lg:grid-cols-12">
      <!-- Daftar Ronde Hari Ini -->
      <section class="lg:col-span-5 rounded-xl border border-border bg-card p-3.5 shadow-sm flex flex-col gap-3">
        <div class="flex items-center justify-between border-b border-border pb-2.5">
          <div class="flex items-center gap-2">
            <Trophy class="h-4 w-4 text-amber-400" />
            <div>
              <h2 class="text-xs font-extrabold uppercase tracking-wider text-foreground">
                Checklist Ronde Event
              </h2>
              <p class="text-[11px] text-muted-foreground">
                {roundStore.eventName || "Pilih event cloud untuk melihat ronde."}
              </p>
            </div>
          </div>
          <Button
            variant="outline"
            size="sm"
            disabled={loadingRounds || selectingRound || !roundStore.eventId}
            onclick={() => fetchEventRounds()}
            class="h-7 gap-1 text-[11px]"
          >
            <RefreshCw class="h-3 w-3 {loadingRounds ? 'animate-spin' : ''}" />
            <span>Refresh</span>
          </Button>
        </div>

        {#if roundsError}
          <div class="flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 p-2 text-xs text-red-400">
            <AlertCircle class="h-4 w-4 shrink-0" />
            <span>{roundsError}</span>
          </div>
        {/if}

        <div class="space-y-1.5 max-h-60 overflow-y-auto pr-1">
          {#if cloudRounds.length === 0}
            <div class="rounded-lg border border-dashed border-border p-4 text-center text-xs text-muted-foreground">
              {loadingRounds ? "Memuat daftar ronde…" : roundStore.eventId ? "Belum ada ronde untuk event terpilih." : "Pilih event cloud terlebih dahulu."}
            </div>
          {:else}
            {#each cloudRounds as r}
              {@const isActive = roundStore.roundId === r.id}
              <button
                type="button"
                disabled={selectingRound || engine.active || roundStore.busy || roundStore.state === "RUNNING"}
                onclick={() => loadRoundIntoWorkspace(r)}
                class="w-full flex items-center justify-between gap-2 rounded-lg border px-3 py-2 text-left transition {isActive
                  ? 'border-emerald-500/60 bg-emerald-500/10'
                  : 'border-border bg-muted/40 hover:bg-accent'} disabled:opacity-60"
              >
                <div class="min-w-0">
                  <div class="flex items-center gap-1.5">
                    <span class="text-xs font-bold text-foreground truncate">{r.title}</span>
                    {#if isActive}
                      <span class="rounded bg-emerald-500/20 px-1.5 py-0.5 text-[9px] font-bold text-emerald-400 uppercase">
                        Dipilih
                      </span>
                    {/if}
                  </div>
                  <div class="mt-0.5 flex flex-wrap items-center gap-2 text-[10px] text-muted-foreground">
                    <span>{r.event_title || "Event"}</span>
                    <span>·</span>
                    <span class="text-amber-400 font-semibold">{r.bird_species_name || "Burung"}</span>
                    {#if r.start_time}
                      <span>·</span>
                      <span class="font-mono">{r.start_time} – {r.end_time || "-"}</span>
                    {/if}
                  </div>
                </div>
              </button>
            {/each}
          {/if}
        </div>
      </section>

      <!-- Posisi & Jadwal Juri Real-Time -->
      <section class="lg:col-span-7 rounded-xl border border-border bg-card p-3.5 shadow-sm flex flex-col gap-3">
        <div class="flex items-center justify-between border-b border-border pb-2.5">
          <div class="flex items-center gap-2">
            <Users class="h-4 w-4 text-sky-400" />
            <div>
              <h2 class="text-xs font-extrabold uppercase tracking-wider text-foreground">
                Monitoring Posisi & Waktu Rotasi Juri
              </h2>
              <p class="text-[11px] text-muted-foreground">
                Pantau zona aktif dan checklist otomatis setiap juri.
              </p>
            </div>
          </div>
          <div class="flex items-center gap-1.5 rounded-lg border border-border bg-muted px-2.5 py-1 font-mono text-[11px] font-bold">
            <Clock class="h-3.5 w-3.5 text-amber-400" />
            <span>{roundStore.intervalLabel}</span>
          </div>
        </div>

        <div class="grid gap-2.5 sm:grid-cols-2">
          {#each jurorList as juri, idx}
            {@const currentZone = getJudgeAssignedZone(idx, roundStore.elapsedSeconds)}
            {@const schedule = getJudgeRotationSchedule(idx)}
            <div class="rounded-lg border border-border bg-muted/40 p-2.5 space-y-2">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2 min-w-0">
                  <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded bg-sky-500/20 font-mono text-[11px] font-black text-sky-400 border border-sky-500/30">
                    J{idx + 1}
                  </span>
                  <span class="text-xs font-bold text-foreground truncate">{juri || "Belum ditetapkan"}</span>
                </div>
                <span class="shrink-0 rounded-full bg-emerald-500/20 border border-emerald-500/40 px-2 py-0.5 font-mono text-[10px] font-extrabold text-emerald-400">
                  Aktif: Zona {currentZone}
                </span>
              </div>

              <div class="grid grid-cols-2 gap-1">
                {#each schedule as item}
                  {@const isCurrent = !roundStore.interval.final && item.shift === roundStore.interval.index + 1}
                  {@const checked = roundStore.isJudgeZoneChecked(roundStore.roundId, idx, item.zone)}
                  <div
                    class="flex items-center justify-between rounded px-2 py-1 text-[10px] font-mono border {isCurrent
                      ? 'border-primary bg-primary/15 text-foreground font-bold'
                      : checked
                      ? 'border-emerald-500/30 bg-emerald-500/10 text-foreground'
                      : 'border-border bg-card text-muted-foreground'}"
                  >
                    <span class="flex items-center gap-1">
                      {#if checked}<CheckSquare class="size-3 text-emerald-500" aria-label="Selesai" />{/if}
                      S{item.shift} · Z{item.zone}
                    </span>
                    <span>{item.startClock}-{item.endClock}</span>
                  </div>
                {/each}
              </div>
            </div>
          {:else}
            <p class="col-span-full text-xs text-muted-foreground">{roundStore.roundId ? "Cloud belum menetapkan juri untuk ronde ini." : "Pilih ronde untuk melihat daftar juri."}</p>
          {/each}
        </div>
      </section>
    </div>

    <!-- Row 2: Tabel Checklist Gantangan, Burung & WhatsApp -->
    <section class="rounded-xl border border-border bg-card p-4 shadow-sm space-y-3">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
        <div class="flex items-center gap-2.5">
          <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
            <ListChecks class="h-4 w-4" />
          </div>
          <div>
            <h2 class="text-sm font-extrabold text-foreground">
              Checklist Kesesuaian Burung, Gantangan & WhatsApp
            </h2>
            <p class="text-xs text-muted-foreground">
              Centang setiap gantangan setelah memastikan burung, pemilik, dan nomor WhatsApp sesuai sebelum/selama ronde berjalan.
            </p>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <div class="relative w-56">
            <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
            <Input
              bind:value={search}
              placeholder="Cari gantangan / burung / WA..."
              class="h-8 pl-8 text-xs bg-muted"
            />
          </div>

          <span class="rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-2.5 py-1 font-mono text-xs font-bold text-emerald-400">
            {verifiedCount} / {activeNumbers.length} Terverifikasi
          </span>

          <Button
            variant="outline"
            size="sm"
            disabled={!roundStore.roundId}
            onclick={() => roundStore.setAllGantanganChecks(roundStore.roundId, activeNumbers, true)}
            class="h-8 text-xs"
          >
            Check Semua
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={!roundStore.roundId}
            onclick={() => roundStore.setAllGantanganChecks(roundStore.roundId, activeNumbers, false)}
            class="h-8 text-xs"
          >
            Uncheck Semua
          </Button>
        </div>
      </div>

      {#if !roundStore.roundId}
        <div class="rounded-xl border border-dashed border-border bg-muted/30 p-6 text-center text-xs text-muted-foreground">
          {#if roundStore.eventId}
            Pilih ronde dari daftar <strong class="text-foreground">Checklist Ronde Event</strong>.
          {:else}
            Pilih tanggal dan event lewat <strong class="text-foreground">Pilih Event Cloud</strong>.
          {/if}
        </div>
      {:else}
        <div class="overflow-x-auto rounded-lg border border-border">
          <table class="w-full text-left text-xs">
            <thead class="bg-muted/80 text-[11px] font-bold uppercase tracking-wider text-muted-foreground border-b border-border">
              <tr>
                <th class="px-3 py-2.5 w-14 text-center">Cek</th>
                <th class="px-3 py-2.5">Gantangan</th>
                <th class="px-3 py-2.5">Zona</th>
                <th class="px-3 py-2.5">Nama Burung & Kode</th>
                <th class="px-3 py-2.5">Pemilik / Peserta</th>
                <th class="px-3 py-2.5">No. WhatsApp</th>
                <th class="px-3 py-2.5 text-right">Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border">
              {#each gantanganRows as row}
                <tr
                  onclick={() => roundStore.toggleGantanganCheck(roundStore.roundId, row.num)}
                  class="cursor-pointer transition {row.checked
                    ? 'bg-emerald-500/5 hover:bg-emerald-500/10'
                    : 'hover:bg-muted/50'}"
                >
                  <td class="px-3 py-2 text-center">
                    <button
                      type="button"
                      aria-label={`Checklist Gantangan ${row.num}`}
                      onclick={(e) => {
                        e.stopPropagation();
                        roundStore.toggleGantanganCheck(roundStore.roundId, row.num);
                      }}
                      class="inline-flex items-center justify-center rounded p-0.5 text-foreground hover:bg-accent"
                    >
                      {#if row.checked}
                        <CheckSquare class="h-4 w-4 text-emerald-400" />
                      {:else}
                        <Square class="h-4 w-4 text-muted-foreground" />
                      {/if}
                    </button>
                  </td>
                  <td class="px-3 py-2 font-mono font-bold text-foreground">
                    #{String(row.num).padStart(2, "0")}
                  </td>
                  <td class="px-3 py-2">
                    <span class="rounded bg-accent px-2 py-0.5 font-mono text-[10px] font-bold text-foreground">
                      Zona {row.zone}
                    </span>
                  </td>
                  <td class="px-3 py-2">
                    {#if row.participant}
                      <div class="font-bold text-foreground">{row.participant.birdName}</div>
                      {#if row.participant.birdCode}
                        <div class="font-mono text-[10px] text-muted-foreground">{row.participant.birdCode}</div>
                      {/if}
                    {:else}
                      <span class="text-muted-foreground italic">Kosong / Belum terdaftar</span>
                    {/if}
                  </td>
                  <td class="px-3 py-2 text-foreground">
                    {row.participant?.ownerName || "—"}
                  </td>
                  <td class="px-3 py-2 font-mono">
                    {#if row.participant?.ownerPhone}
                      <span class="inline-flex items-center gap-1.5 rounded bg-emerald-500/10 px-2 py-0.5 text-[11px] font-semibold text-emerald-400 border border-emerald-500/20">
                        <Phone class="h-3 w-3" />
                        {row.participant.ownerPhone}
                      </span>
                    {:else}
                      <span class="text-muted-foreground">—</span>
                    {/if}
                  </td>
                  <td class="px-3 py-2 text-right">
                    {#if row.checked}
                      <span class="inline-flex items-center gap-1 rounded-full bg-emerald-500/15 px-2 py-0.5 text-[10px] font-bold text-emerald-400">
                        <CheckCircle2 class="h-3 w-3" />
                        Sesuai
                      </span>
                    {:else}
                      <span class="inline-flex items-center gap-1 rounded-full bg-accent px-2 py-0.5 text-[10px] font-semibold text-muted-foreground">
                        Belum Dicek
                      </span>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </section>
  </div>
</div>

<CloudSessionSelector open={showCloudModal} onClose={() => (showCloudModal = false)} />

<QuickSaveModal
  bind:open={confirmSave}
  bind:sessionName
  bind:selectedFormat
  bind:storageDir
  onConfirm={handleConfirmSave}
  onCancel={() => (confirmSave = false)}
/>

<RoundTimeUpDialog onStopAndSave={handleStopClick} />
