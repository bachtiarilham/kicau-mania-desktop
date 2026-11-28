import { api, message, type Row } from "$lib/services/api";
import { isActiveSession, readSession } from "./session-state";
import { arenaCameraID } from "$lib/components/cameras/camera-slots";
class Engine {
  overlay = $state(false);
  draftID = $state("");
  session = $state<Row | null>(null);
  channels = $state<Row[]>([]);
  networks = $state<Row[]>([]);
  configuration = $state<Row | null>(null);
  pairing = $state<Row>({ paired: false });
  error = $state("");
  notice = $state("");
  connected = $state(false);
  busy = $state(false);
  metrics = $state<Record<string, Row>>({});
  playerStatus = $state<Record<string, Row>>({});
  metricConnection = $state("disconnected");
  private mutationVersion = 0;
  private refreshVersion = 0;
  get running() { return !!this.session?.id && ["RUNNING", "DEGRADED"].includes(this.session.state); }
  get active() { return isActiveSession(this.session); }
  async connect(readOnly = false) {
    this.error = "";
    if (readOnly) {
      await this.refresh(true);
      return;
    }
    try {
      const b = await api.bootstrap();
      this.networks = b.networks || [];
      this.connected = true;
      await this.refresh();
      try {
        this.configuration = await api.engine(
          "GET",
          "/api/v1/setup/configuration",
        );
      } catch (e) {
        this.error = message(e);
      }
    } catch (e) {
      this.connected = false;
      this.error = message(e);
    }
  }
  async refresh(readOnly = false) {
    if (this.busy) return;
    const request = ++this.refreshVersion;
    const mutation = this.mutationVersion;
    const stale = () => request !== this.refreshVersion || mutation !== this.mutationVersion;
    try {
      if (!this.connected && !readOnly) {
        const b = await api.bootstrap();
        this.networks = b.networks || [];
      }
      const status = await api.engine<Row>("GET", "/api/v1/local/status");
      if (stale()) return;
      if (typeof status.configured !== "boolean")
        throw Error("Status konfigurasi engine tidak valid");
      this.notice = status.advice || status.capacity_notice || "";
      if (readOnly) {
        const channels = status.configured ? await api.engine<Row[]>("GET", "/api/v1/channels") : [];
        if (stale()) return;
        if (!Array.isArray(channels)) throw Error("Daftar channel engine tidak valid");
        this.channels = channels;
        this.session = readSession(status.session);
        this.connected = true;
        this.error = "";
        return;
      }
      if (!status.configured) {
        const pairing = await api.engine<Row>("GET", "/api/v1/desktop/cloud");
        if (stale()) return;
        this.pairing = pairing;
        this.channels = [];
        this.session = null;
        this.connected = true;
        this.error = "";
        if (!this.notice)
          this.notice =
            "Engine terhubung. Konfigurasikan NVR sebelum memulai sesi.";
        return;
      }
      // Session belongs to this status snapshot. A separate session request can
      // race Start/Stop or a backend restart and restore an obsolete UI state.
      const session = readSession(status.session);
      const [channels, pairing] = await Promise.all([
        api.engine<Row[]>("GET", "/api/v1/channels"),
        api.engine<Row>("GET", "/api/v1/desktop/cloud"),
      ]);
      if (stale()) return;
      if (!Array.isArray(channels)) throw Error("Daftar channel engine tidak valid");
      this.channels = channels;
      this.session = session;
      this.pairing = pairing;
      this.connected = true;
      this.error = "";
    } catch (e) {
      if (stale()) return;
      this.connected = false;
      this.error = message(e);
    }
  }
  async start(id: string) {
    if (this.busy) throw Error("Operasi sesi masih berlangsung");
    ++this.mutationVersion;
    this.busy = true;
    try {
      const session = readSession(await api.engine("POST", "/api/v1/media/sessions", {
        id,
        channels: this.channels.filter((c) => c.id !== arenaCameraID).map((c) => c.id),
        restreams: [],
      }));
      if (!session) throw Error("Engine tidak mengembalikan sesi yang dimulai");
      this.session = session;
      this.error = "";
    } catch (e) {
      this.error = message(e);
      throw e;
    } finally {
      ++this.mutationVersion;
      this.busy = false;
    }
  }
  async stop() {
    const id = this.session?.id;
    if (!id || !this.active || this.busy) return;
    ++this.mutationVersion;
    this.busy = true;
    try {
      const session = readSession(await api.engine(
        "POST",
        `/api/v1/media/sessions/${id}/stop`,
        {},
      ));
      if (!session) throw Error("Engine tidak mengembalikan status Stop");
      this.session = session;
      this.error = "";
    } catch (e) {
      this.error = message(e);
    } finally {
      ++this.mutationVersion;
      this.busy = false;
    }
  }
  listen() {
    const offMetric = api.on("engine:metric", (v) => {
      if (v.connection) {
        this.metricConnection = v.connection;
        return;
      }
      if (v.channel_id) {
        const old = this.metrics[v.channel_id]?.history || [];
        this.metrics[v.channel_id] = {
          ...v,
          history: [...old.slice(-49), v.peak_dbfs],
        };
      }
    });
    const offPlayer = api.on("player:status", (v) => {
      this.playerStatus[v.channel_id] = v;
    });
    return () => {
      offMetric();
      offPlayer();
    };
  }
}
export const engine = new Engine();
