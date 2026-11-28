<script lang="ts">
  import { untrack } from "svelte";
  import { AlertCircle, Cloud, RefreshCw, Trophy, X } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import EventDatePicker from "./event-date-picker.svelte";
  import { roundStore } from "$lib/state/round.svelte";
  import { auth } from "$lib/state/auth.svelte";
  import { engine } from "$lib/state/engine.svelte";
  import { all, api, message, type Row } from "$lib/services/api";

  let { open, onClose }: { open: boolean; onClose: () => void } = $props();
  let loading = $state(false);
  let error = $state("");
  let signInRequired = $state(false);
  let arenaId = $state("");
  let arenaName = $state("");
  let events = $state<Row[]>([]);
  let date = $state("");
  let eventId = $state("");

  const visibleEvents = $derived(date ? events.filter((event) => event.event_date?.slice(0, 10) === date) : []);
  const selectedEvent = $derived(visibleEvents.find((event) => event.id === eventId));

  $effect(() => {
    if (open) untrack(() => void loadEvents());
  });

  async function loadEvents() {
    if (loading) return;
    loading = true;
    error = "";
    signInRequired = false;
    try {
      arenaId = auth.arenaId || (await all("/arenas"))[0]?.id || "";
      if (!arenaId) {
        events = [];
        error = "Akun ini belum memiliki arena.";
        return;
      }
      arenaName = (await api.cloud<Row>(`/arenas/${encodeURIComponent(arenaId)}`)).name;
      events = await all(`/arenas/${encodeURIComponent(arenaId)}/events`);
      const active = events.find((event) => event.id === roundStore.eventId);
      date = date || active?.event_date?.slice(0, 10) || "";
      eventId = visibleEvents.find((event) => event.id === eventId)?.id
        || visibleEvents.find((event) => event.id === active?.id)?.id
        || "";
    } catch (cause) {
      events = [];
      eventId = "";
      const detail = message(cause);
      signInRequired = /sesi cloud tidak aktif|login ke cloud|authentication required/i.test(detail);
      error = signInRequired ? "Sesi cloud tidak aktif. Masuk kembali dengan akun cloud." : "Gagal memuat event: " + detail;
    } finally {
      loading = false;
    }
  }

  function applyEvent() {
    if (loading || engine.active || roundStore.state === "RUNNING" || !selectedEvent) return;
    auth.setArenaId(arenaId);
    roundStore.selectCloudEvent(selectedEvent.id, selectedEvent.title, selectedEvent.event_date?.slice(0, 10) || "", arenaName);
    onClose();
  }
</script>

{#if open}
  <div class="fixed inset-0 z-50 grid place-items-center bg-black/75 p-4 backdrop-blur-xs">
    <div role="dialog" aria-modal="true" aria-labelledby="cloud-event-title" class="max-h-[90vh] w-full max-w-xl overflow-y-auto rounded-2xl border border-border bg-card text-foreground">
      <header class="flex items-center justify-between gap-3 border-b border-border px-5 py-4">
        <h2 id="cloud-event-title" class="flex items-center gap-2 text-base font-bold">
          <Cloud class="size-5 text-primary" /> Pilih Event Cloud
        </h2>
        <Button variant="ghost" size="icon-sm" onclick={onClose} aria-label="Tutup"><X class="size-4" /></Button>
      </header>
      <div class="space-y-4 p-5">
        {#if error}
          <div role="alert" class="flex flex-wrap items-center gap-2 rounded-lg border border-destructive/40 p-3 text-sm text-destructive">
            <AlertCircle class="size-4 shrink-0" />
            <span class="flex-1">{error}</span>
            {#if signInRequired}
              <Button size="sm" variant="outline" onclick={() => { onClose(); roundStore.clearCloudRound(); auth.logout(); }}>Masuk kembali</Button>
            {/if}
          </div>
        {/if}

        <div>
          <p class="mb-1.5 text-xs font-semibold">Tanggal event</p>
          <EventDatePicker value={date} onChange={(value) => { date = value; eventId = ""; }} />
        </div>

        <div>
          <label for="cloud-event-select" class="mb-1.5 flex items-center gap-1.5 text-xs font-semibold"><Trophy class="size-4 text-primary" /> Event perlombaan</label>
          <select id="cloud-event-select" bind:value={eventId} disabled={loading || !visibleEvents.length} class="w-full rounded-lg border border-border bg-muted px-3 py-2 text-sm">
            <option value="">{loading ? "Memuat event…" : !date ? "Pilih tanggal dulu" : !visibleEvents.length ? "Tidak ada event pada tanggal ini" : "Pilih event"}</option>
            {#each visibleEvents as event (event.id)}
              <option value={event.id}>{event.title}</option>
            {/each}
          </select>
        </div>

        {#if selectedEvent}
          <p class="text-xs text-muted-foreground">{selectedEvent.round_count ?? 0} ronde. Pilih ronde di Checklist &amp; Kontrol Ronde setelah event diterapkan.</p>
        {/if}
      </div>

      <footer class="flex flex-wrap justify-between gap-2 border-t border-border px-5 py-4">
        <Button variant="outline" size="sm" onclick={loadEvents} disabled={loading}>
          <RefreshCw class="size-3.5 {loading ? 'animate-spin' : ''}" /> Muat ulang
        </Button>
        <div class="flex items-center gap-2">
          <Button variant="ghost" size="sm" onclick={onClose}>Batal</Button>
          <Button size="sm" onclick={applyEvent} disabled={loading || !selectedEvent || engine.active || roundStore.state === "RUNNING"}>
            Terapkan ke Arena &amp; Livestream
          </Button>
        </div>
      </footer>
    </div>
  </div>
{/if}
