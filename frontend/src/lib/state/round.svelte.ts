import { engine } from "./engine.svelte";
import { api } from "$lib/services/api";

export interface RoundConfig {
  eventName: string;
  sessionName: string;
  startTime: string;
  endTime: string;
  jurors: string[];
  durationSeconds: number;
}

export type RoundState = "IDLE" | "RUNNING" | "PAUSED" | "FINISHED";

const STORAGE_KEY = "kicau_round_config_v1";
const SYNC_CHANNEL = "kicau_round_sync_channel";

class RoundStore {
  eventName = $state("Piala Umi Kasum");
  sessionName = $state("Sesi 1");
  startTime = $state("10.00");
  endTime = $state("10.10 WIB");
  jurors = $state<string[]>([
    "Juri 1: Axxx Bxxx",
    "Juri 2: Cxxxx Dxxx",
    "Juri 3: Zffgfg Hnnn",
    "Juri 4: Ijdjff Jndd",
  ]);
  durationSeconds = $state(150);
  remainingSeconds = $state(150);
  state = $state<RoundState>("IDLE");
  rotationAlert = $state(false);

  private timerInterval: ReturnType<typeof setInterval> | null = null;
  private channel: BroadcastChannel | null = null;
  private startedAt: number | null = null;

  constructor() {
    this.loadFromStorage();
    this.initBroadcast();
  }

  private initBroadcast() {
    if (typeof window === "undefined" || !("BroadcastChannel" in window)) return;
    this.channel = new BroadcastChannel(SYNC_CHANNEL);
    this.channel.onmessage = (event) => {
      const data = event.data;
      if (!data || typeof data !== "object") return;
      if (data.type === "SYNC_STATE") {
        this.applyExternalState(data.payload);
      }
    };
  }

  private broadcast() {
    if (!this.channel) return;
    this.channel.postMessage({
      type: "SYNC_STATE",
      payload: {
        eventName: this.eventName,
        sessionName: this.sessionName,
        startTime: this.startTime,
        endTime: this.endTime,
        jurors: [...this.jurors],
        durationSeconds: this.durationSeconds,
        remainingSeconds: this.remainingSeconds,
        state: this.state,
        rotationAlert: this.rotationAlert,
        startedAt: this.startedAt,
      },
    });
  }

  private applyExternalState(payload: any) {
    if (!payload) return;
    this.eventName = payload.eventName ?? this.eventName;
    this.sessionName = payload.sessionName ?? this.sessionName;
    this.startTime = payload.startTime ?? this.startTime;
    this.endTime = payload.endTime ?? this.endTime;
    this.jurors = Array.isArray(payload.jurors) ? payload.jurors : this.jurors;
    this.durationSeconds = payload.durationSeconds ?? this.durationSeconds;
    this.remainingSeconds = payload.remainingSeconds ?? this.remainingSeconds;
    this.state = payload.state ?? this.state;
    this.rotationAlert = payload.rotationAlert ?? this.rotationAlert;
    this.startedAt = payload.startedAt ?? this.startedAt;

    if (this.state === "RUNNING" && !this.timerInterval) {
      this.startLocalTicker();
    } else if (this.state !== "RUNNING" && this.timerInterval) {
      clearInterval(this.timerInterval);
      this.timerInterval = null;
    }
  }

  private saveToStorage() {
    if (typeof localStorage === "undefined") return;
    try {
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({
          eventName: this.eventName,
          sessionName: this.sessionName,
          startTime: this.startTime,
          endTime: this.endTime,
          jurors: this.jurors,
          durationSeconds: this.durationSeconds,
        })
      );
    } catch {
      // Ignore storage write error
    }
  }

  private loadFromStorage() {
    if (typeof localStorage === "undefined") return;
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (raw) {
        const parsed = JSON.parse(raw);
        if (parsed.eventName) this.eventName = parsed.eventName;
        if (parsed.sessionName) this.sessionName = parsed.sessionName;
        if (parsed.startTime) this.startTime = parsed.startTime;
        if (parsed.endTime) this.endTime = parsed.endTime;
        if (Array.isArray(parsed.jurors)) this.jurors = parsed.jurors;
        if (typeof parsed.durationSeconds === "number") {
          this.durationSeconds = parsed.durationSeconds;
          this.remainingSeconds = parsed.durationSeconds;
        }
      }
    } catch {
      // Ignore storage parse error
    }
  }

  updateConfig(config: Partial<RoundConfig>) {
    if (config.eventName !== undefined) this.eventName = config.eventName;
    if (config.sessionName !== undefined) this.sessionName = config.sessionName;
    if (config.startTime !== undefined) this.startTime = config.startTime;
    if (config.endTime !== undefined) this.endTime = config.endTime;
    if (config.jurors !== undefined) this.jurors = config.jurors;
    if (config.durationSeconds !== undefined) {
      this.durationSeconds = config.durationSeconds;
      if (this.state === "IDLE" || this.state === "FINISHED") {
        this.remainingSeconds = config.durationSeconds;
      }
    }
    this.saveToStorage();
    this.broadcast();
  }

  async startRound() {
    if (this.state === "RUNNING") return;

    this.state = "RUNNING";
    this.rotationAlert = false;
    this.startedAt = Date.now();

    if (this.remainingSeconds <= 0) {
      this.remainingSeconds = this.durationSeconds;
    }

    this.startLocalTicker();
    this.broadcast();

    // Trigger local engine recording if not running
    try {
      if (typeof window !== "undefined" && !engine.running && !engine.busy) {
        if (!engine.draftID) {
          engine.draftID = await api.sessionID();
        }
        await engine.start(engine.draftID);
        engine.draftID = "";
      }
    } catch {
      // Gracefully ignore if engine disconnected or in headless test environment
    }
  }

  stopRound() {
    if (this.timerInterval) {
      clearInterval(this.timerInterval);
      this.timerInterval = null;
    }
    this.state = "FINISHED";
    this.rotationAlert = true;
    this.broadcast();

    // Stop recording on engine
    try {
      if (engine.running && !engine.busy) {
        void engine.stop();
      }
    } catch (e) {
      console.warn("Failed to auto-stop engine session:", e);
    }
  }

  pauseRound() {
    if (this.timerInterval) {
      clearInterval(this.timerInterval);
      this.timerInterval = null;
    }
    this.state = "PAUSED";
    this.broadcast();
  }

  resetRound() {
    if (this.timerInterval) {
      clearInterval(this.timerInterval);
      this.timerInterval = null;
    }
    this.state = "IDLE";
    this.remainingSeconds = this.durationSeconds;
    this.rotationAlert = false;
    this.startedAt = null;
    this.broadcast();
  }

  private startLocalTicker() {
    if (this.timerInterval) clearInterval(this.timerInterval);

    this.timerInterval = setInterval(() => {
      if (this.state !== "RUNNING") {
        if (this.timerInterval) clearInterval(this.timerInterval);
        return;
      }

      if (this.remainingSeconds > 0) {
        this.remainingSeconds -= 1;
        if (this.remainingSeconds === 0) {
          this.state = "FINISHED";
          this.rotationAlert = true;
          if (this.timerInterval) clearInterval(this.timerInterval);
          this.broadcast();
        }
      }
    }, 1000);
  }
}

export const roundStore = new RoundStore();
