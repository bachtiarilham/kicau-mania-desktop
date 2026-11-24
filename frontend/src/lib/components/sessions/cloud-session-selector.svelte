<script lang="ts">
  import { onMount } from "svelte";
  import {
    Cloud,
    Calendar,
    Trophy,
    Users,
    CheckCircle2,
    RefreshCw,
    X,
    Phone,
    Shield,
    Clock,
    AlertCircle,
  } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { roundStore, type ParticipantMapping } from "$lib/state/round.svelte";
  import { auth } from "$lib/state/auth.svelte";
  import { api, message } from "$lib/services/api";

  interface Props {
    open: boolean;
    onClose: () => void;
  }

  let { open = false, onClose }: Props = $props();

  let loading = $state(false);
  let error = $state("");
  let success = $state("");

  let arenas = $state<any[]>([]);
  let selectedArenaId = $state(auth.arenaId || "");

  let events = $state<any[]>([]);
  let selectedEventId = $state("");

  let rounds = $state<any[]>([]);
  let selectedRoundId = $state("");

  let participants = $state<any[]>([]);
  let judges = $state<any[]>([]);

  let binding = $state(false);

  // Computed selected objects
  let selectedEvent = $derived(events.find((e) => e.id === selectedEventId));
  let selectedRound = $derived(rounds.find((r) => r.id === selectedRoundId));

  $effect(() => {
    if (open) {
      error = "";
      success = "";
      initData();
    }
  });

  async function initData() {
    loading = true;
    error = "";
    try {
      // 1. If arenaId not set, fetch arenas
      if (!selectedArenaId) {
        try {
          const arenaRes = await api.cloud<any>("/arenas?limit=50");
          arenas = Array.isArray(arenaRes?.items) ? arenaRes.items : [];
          if (arenas.length > 0) {
            selectedArenaId = arenas[0].id;
            auth.setArenaId(selectedArenaId);
          }
        } catch {
          // If /arenas fails, proceed if we have a fallback
        }
      }

      if (selectedArenaId) {
        await loadEvents(selectedArenaId);
      } else {
        error = "Arena belum ditentukan. Pastikan akun memiliki akses ke Arena lomba.";
      }
    } catch (err: any) {
      error = message(err) || "Gagal memuat data dari cloud server.";
    } finally {
      loading = false;
    }
  }

  async function loadEvents(arenaId: string) {
    loading = true;
    error = "";
    try {
      const res = await api.cloud<any>(`/arenas/${arenaId}/events?limit=50`);
      events = Array.isArray(res?.items) ? res.items : [];
      if (events.length > 0) {
        selectedEventId = events[0].id;
        await loadRounds(arenaId, selectedEventId);
      } else {
        rounds = [];
        selectedRoundId = "";
        participants = [];
        judges = [];
      }
    } catch (err: any) {
      error = "Gagal memuat jadwal event: " + message(err);
    } finally {
      loading = false;
    }
  }

  async function loadRounds(arenaId: string, eventId: string) {
    loading = true;
    error = "";
    try {
      const res = await api.cloud<any>(`/arenas/${arenaId}/rounds?limit=50`);
      const allRounds = Array.isArray(res?.items) ? res.items : [];
      // Filter rounds matching event if eventId is set
      rounds = allRounds.filter((r: any) => !r.event_id || r.event_id === eventId);
      if (rounds.length > 0) {
        selectedRoundId = rounds[0].id;
        await loadRoundDetails(arenaId, selectedRoundId);
      } else {
        selectedRoundId = "";
        participants = [];
        judges = [];
      }
    } catch (err: any) {
      error = "Gagal memuat daftar ronde/sesi: " + message(err);
    } finally {
      loading = false;
    }
  }

  async function loadRoundDetails(arenaId: string, roundId: string) {
    loading = true;
    error = "";
    try {
      const [partsRes, judgesRes] = await Promise.allSettled([
        api.cloud<any[]>(`/arenas/${arenaId}/rounds/${roundId}/participants`),
        api.cloud<any[]>(`/arenas/${arenaId}/rounds/${roundId}/judges`),
      ]);

      if (partsRes.status === "fulfilled" && Array.isArray(partsRes.value)) {
        participants = partsRes.value.sort(
          (a, b) => (a.gantangan_number || 0) - (b.gantangan_number || 0)
        );
      } else {
        participants = [];
      }

      if (judgesRes.status === "fulfilled" && Array.isArray(judgesRes.value)) {
        judges = judgesRes.value;
      } else {
        judges = [];
      }
    } catch (err: any) {
      error = "Gagal memuat detail peserta ronde: " + message(err);
    } finally {
      loading = false;
    }
  }

  async function handleApplyToLivestream() {
    if (!selectedRound) {
      error = "Silakan pilih sesi lomba terlebih dahulu.";
      return;
    }

    binding = true;
    error = "";
    try {
      // 1. Prepare Juror names (from round fields or judges list)
      const formattedJurors: string[] = [
        selectedRound.field_judge_1 || judges[0]?.judge_name || "Juri 1",
        selectedRound.field_judge_2 || judges[1]?.judge_name || "Juri 2",
        selectedRound.field_judge_3 || judges[2]?.judge_name || "Juri 3",
        selectedRound.field_judge_4 || judges[3]?.judge_name || "Juri 4",
      ];

      // 2. Prepare participant mappings
      const mappedParticipants: ParticipantMapping[] = participants.map((p) => {
        const num = p.gantangan_number || 1;
        const channelId = `cam-${String(num).padStart(2, "0")}`;
        return {
          registrationId: p.id,
          gantanganId: p.gantangan_id,
          gantanganNumber: num,
          birdName: p.bird_name || `Gantangan ${num}`,
          birdCode: p.bird_code,
          ownerName: p.owner_name || p.participant_name || p.registrant_name || "Pemilik",
          ownerPhone: p.owner_phone || p.participant_phone || p.registrant_phone || "",
          channelId,
        };
      });

      // 3. Update RoundStore (drives Livestream UI, HUD, and timers)
      roundStore.setCloudRound({
        eventId: selectedEvent?.id || selectedRound.event_id || "",
        roundId: selectedRound.id,
        eventName: selectedEvent?.title || selectedRound.event_title || roundStore.eventName,
        sessionName: selectedRound.title || `Sesi ${selectedRound.round_number}`,
        startTime: selectedRound.start_time || roundStore.startTime,
        endTime: selectedRound.end_time || roundStore.endTime,
        jurors: formattedJurors,
        participants: mappedParticipants,
      });

      // 4. Ingest Mappings to Desktop CCTV Engine
      if (mappedParticipants.length > 0) {
        try {
          const mappings = mappedParticipants.map((m) => ({
            channel_id: m.channelId,
            registration_id: m.registrationId,
            gantangan_id: m.gantanganId,
          }));

          const localSessionId = roundStore.roundId || api.sessionID();
          await api.bind({
            session_id: localSessionId,
            round_id: selectedRound.id,
            mappings,
          });
        } catch (engineErr) {
          console.warn("Desktop engine ingestion mapping notice:", engineErr);
          // Non-fatal for UI preview
        }
      }

      success = "Sesi berhasil disinkronkan ke livestream & rekaman!";
      setTimeout(() => {
        onClose();
      }, 700);
    } catch (err: any) {
      error = "Gagal menerapkan sesi: " + message(err);
    } finally {
      binding = false;
    }
  }
</script>

{#if open}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-xs p-4">
    <div
      class="flex max-h-[90vh] w-full max-w-4xl flex-col rounded-2xl border border-border bg-card text-foreground shadow-2xl overflow-hidden"
    >
      <!-- Header Modal -->
      <div class="flex items-center justify-between border-b border-border px-6 py-4 bg-muted/40">
        <div class="flex items-center gap-3">
          <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/20 text-primary border border-primary/30">
            <Cloud class="h-5 w-5" />
          </div>
          <div>
            <h2 class="text-base font-extrabold tracking-tight">Pilih Sesi Lomba dari Cloud Database</h2>
            <p class="text-xs text-muted-foreground">
              Tarik jadwal, peserta gantangan, dan dewan juri resmi untuk livestreaming & rekaman
            </p>
          </div>
        </div>

        <button
          type="button"
          onclick={onClose}
          class="rounded-lg p-1.5 text-muted-foreground hover:bg-accent hover:text-foreground transition"
          aria-label="Tutup modal"
        >
          <X class="h-5 w-5" />
        </button>
      </div>

      <!-- Content Area -->
      <div class="flex-1 overflow-y-auto p-6 space-y-5">
        {#if error}
          <div class="flex items-center gap-2 rounded-xl border border-red-500/40 bg-red-500/10 p-3 text-xs text-red-400">
            <AlertCircle class="h-4 w-4 shrink-0" />
            <span>{error}</span>
          </div>
        {/if}

        {#if success}
          <div class="flex items-center gap-2 rounded-xl border border-emerald-500/40 bg-emerald-500/10 p-3 text-xs text-emerald-400">
            <CheckCircle2 class="h-4 w-4 shrink-0" />
            <span>{success}</span>
          </div>
        {/if}

        <!-- Step 1 & 2: Select Event & Round -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Event Selector -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-muted-foreground flex items-center gap-1.5 uppercase tracking-wider">
              <Trophy class="h-3.5 w-3.5 text-amber-400" />
              <span>Event Perlombaan</span>
            </label>
            <select
              bind:value={selectedEventId}
              onchange={() => {
                if (selectedArenaId && selectedEventId) {
                  loadRounds(selectedArenaId, selectedEventId);
                }
              }}
              disabled={loading || events.length === 0}
              class="w-full rounded-xl border border-border bg-muted px-3 py-2 text-xs font-semibold text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
            >
              {#if events.length === 0}
                <option value="">Belum ada event ditemukan</option>
              {:else}
                {#each events as ev}
                  <option value={ev.id}>
                    {ev.title} ({ev.status})
                  </option>
                {/each}
              {/if}
            </select>
            {#if selectedEvent}
              <div class="text-[11px] text-muted-foreground flex items-center gap-2 pt-0.5">
                <Calendar class="h-3 w-3" />
                <span>Tanggal: {selectedEvent.event_date ? new Date(selectedEvent.event_date).toLocaleDateString("id-ID") : "Minggu"}</span>
                <span>·</span>
                <span>{selectedEvent.round_count || 0} Sesi Terdaftar</span>
              </div>
            {/if}
          </div>

          <!-- Round Selector -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-muted-foreground flex items-center gap-1.5 uppercase tracking-wider">
              <Clock class="h-3.5 w-3.5 text-sky-400" />
              <span>Sesi / Ronde Lomba</span>
            </label>
            <select
              bind:value={selectedRoundId}
              onchange={() => {
                if (selectedArenaId && selectedRoundId) {
                  loadRoundDetails(selectedArenaId, selectedRoundId);
                }
              }}
              disabled={loading || rounds.length === 0}
              class="w-full rounded-xl border border-border bg-muted px-3 py-2 text-xs font-semibold text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
            >
              {#if rounds.length === 0}
                <option value="">Belum ada ronde di event ini</option>
              {:else}
                {#each rounds as r}
                  <option value={r.id}>
                    Ronde {r.round_number}: {r.title} ({r.category_name || r.bird_species_name || "Umum"})
                  </option>
                {/each}
              {/if}
            </select>
            {#if selectedRound}
              <div class="text-[11px] text-muted-foreground flex items-center gap-2 pt-0.5">
                <Clock class="h-3 w-3" />
                <span>Jadwal: {selectedRound.start_time || "10:00"} – {selectedRound.end_time || "10:10 WIB"}</span>
                <span>·</span>
                <span class="text-emerald-400 font-semibold">{participants.length} Gantangan Terisi</span>
              </div>
            {/if}
          </div>
        </div>

        <!-- Step 3: Round Overview Card -->
        {#if selectedRound}
          <div class="rounded-xl border border-border bg-muted/60 p-4 space-y-3">
            <div class="flex items-center justify-between border-b border-border/80 pb-2">
              <div class="flex items-center gap-2">
                <span class="rounded bg-primary/20 text-primary font-mono text-[10px] font-bold px-2 py-0.5 border border-primary/30">
                  RONDE {selectedRound.round_number}
                </span>
                <span class="text-sm font-extrabold text-foreground">{selectedRound.title}</span>
              </div>
              <span class="text-xs font-medium text-amber-400">
                Kategori: {selectedRound.category_name || selectedRound.bird_species_name || "Murai Batu"}
              </span>
            </div>

            <!-- 4 Judges Assigned -->
            <div>
              <div class="text-[10px] font-bold uppercase tracking-wider text-muted-foreground mb-1.5 flex items-center gap-1">
                <Shield class="h-3 w-3 text-primary" />
                <span>Dewan Juri yang Bertugas di Ronde Ini:</span>
              </div>
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
                <div class="rounded-lg border border-border bg-card p-2 text-xs">
                  <div class="text-[9px] text-muted-foreground font-bold">JURI 1 (BLOK A)</div>
                  <div class="font-bold text-foreground truncate">{selectedRound.field_judge_1 || judges[0]?.judge_name || "Juri 1"}</div>
                </div>
                <div class="rounded-lg border border-border bg-card p-2 text-xs">
                  <div class="text-[9px] text-muted-foreground font-bold">JURI 2 (BLOK B)</div>
                  <div class="font-bold text-foreground truncate">{selectedRound.field_judge_2 || judges[1]?.judge_name || "Juri 2"}</div>
                </div>
                <div class="rounded-lg border border-border bg-card p-2 text-xs">
                  <div class="text-[9px] text-muted-foreground font-bold">JURI 3 (BLOK C)</div>
                  <div class="font-bold text-foreground truncate">{selectedRound.field_judge_3 || judges[2]?.judge_name || "Juri 3"}</div>
                </div>
                <div class="rounded-lg border border-border bg-card p-2 text-xs">
                  <div class="text-[9px] text-muted-foreground font-bold">JURI 4 (BLOK D)</div>
                  <div class="font-bold text-foreground truncate">{selectedRound.field_judge_4 || judges[3]?.judge_name || "Juri 4"}</div>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 4: Participants Table / Gantangan Mapping (For WhatsApp & Video Mapping) -->
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold uppercase tracking-wider text-muted-foreground flex items-center gap-1.5">
                <Users class="h-3.5 w-3.5 text-primary" />
                <span>Pemetaan 24 Kamera Gantangan & No. WhatsApp Peserta ({participants.length} Terdaftar)</span>
              </span>
              <span class="text-[10px] text-muted-foreground">
                Rekaman video tiap kamera otomatis dikirim ke No. WhatsApp masing-masing
              </span>
            </div>

            {#if participants.length === 0}
              <div class="rounded-xl border border-dashed border-border p-6 text-center text-xs text-muted-foreground">
                Belum ada data peserta terdaftar untuk sesi ini di database cloud.
              </div>
            {:else}
              <div class="max-h-56 overflow-y-auto rounded-xl border border-border bg-card">
                <table class="w-full text-left text-xs border-collapse">
                  <thead class="sticky top-0 bg-muted border-b border-border font-bold text-muted-foreground text-[10px] uppercase">
                    <tr>
                      <th class="py-2 px-3">Kamera</th>
                      <th class="py-2 px-3">Gantangan</th>
                      <th class="py-2 px-3">Nama Burung</th>
                      <th class="py-2 px-3">Nama Pemilik</th>
                      <th class="py-2 px-3">No. WhatsApp</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-border font-medium">
                    {#each participants as p}
                      <tr class="hover:bg-accent/40 transition">
                        <td class="py-2 px-3 font-mono text-primary font-bold">
                          CAM-{String(p.gantangan_number || 1).padStart(2, "0")}
                        </td>
                        <td class="py-2 px-3 font-mono font-bold text-amber-400">
                          #{String(p.gantangan_number || 1).padStart(2, "0")}
                        </td>
                        <td class="py-2 px-3 text-foreground font-semibold">
                          {p.bird_name || "—"}
                        </td>
                        <td class="py-2 px-3 text-muted-foreground">
                          {p.owner_name || p.participant_name || p.registrant_name || "—"}
                        </td>
                        <td class="py-2 px-3 font-mono text-[11px] text-emerald-400 flex items-center gap-1">
                          <Phone class="h-3 w-3" />
                          <span>{p.owner_phone || p.participant_phone || p.registrant_phone || "—"}</span>
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            {/if}
          </div>
        {/if}
      </div>

      <!-- Footer Buttons -->
      <div class="flex items-center justify-between border-t border-border px-6 py-4 bg-muted/40">
        <Button
          variant="outline"
          size="sm"
          onclick={() => initData()}
          disabled={loading}
          class="gap-1.5 text-xs text-muted-foreground hover:text-foreground"
        >
          <RefreshCw class="h-3.5 w-3.5 {loading ? 'animate-spin' : ''}" />
          <span>Refresh Data Cloud</span>
        </Button>

        <div class="flex items-center gap-2">
          <Button variant="ghost" size="sm" onclick={onClose} class="text-xs">
            Batal
          </Button>

          <Button
            size="sm"
            onclick={handleApplyToLivestream}
            disabled={binding || !selectedRound}
            class="gap-2 text-xs font-bold bg-primary text-black hover:bg-primary/90"
          >
            {#if binding}
              <RefreshCw class="h-3.5 w-3.5 animate-spin" />
              <span>Menerapkan...</span>
            {:else}
              <CheckCircle2 class="h-3.5 w-3.5" />
              <span>Terapkan Sesi ke Arena & Livestream</span>
            {/if}
          </Button>
        </div>
      </div>
    </div>
  </div>
{/if}
