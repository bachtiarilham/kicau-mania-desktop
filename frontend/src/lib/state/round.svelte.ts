import { engine } from "./engine.svelte";
import { api } from "$lib/services/api";

export interface RoundConfig {
  eventName: string;
  sessionName: string;
  startTime: string;
  endTime: string;
  jurors: string[];
  totalDurationSeconds?: number;
  shiftDurationSeconds?: number;
  durationSeconds?: number;
}

export type RoundState = "IDLE" | "RUNNING" | "PAUSED" | "FINISHED";

export function getJudgeAssignedZone(
  judgeIndex: number,
  currentShift: number
): "A" | "B" | "C" | "D" {
  const zones: ("A" | "B" | "C" | "D")[] = ["A", "B", "C", "D"];
  const zoneIndex = (((judgeIndex % 4) + (Math.max(1, currentShift) - 1)) % 4 + 4) % 4;
  return zones[zoneIndex];
}

const STORAGE_KEY = "kicau_round_config_v2";
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

  // Round is 10 minutes (600 seconds) divided into 4 shifts of 150 seconds each
  totalDurationSeconds = $state(600);
  shiftDurationSeconds = $state(150);
  roundRemainingSeconds = $state(600);
  shiftRemainingSeconds = $state(150);
  currentShift = $state(1);

  state = $state<RoundState>("IDLE");
  rotationAlert = $state(false);

  // Backward-compatibility properties for existing components/tests
  get durationSeconds(): number {
    return this.shiftDurationSeconds;
  }
  set durationSeconds(val: number) {
    this.shiftDurationSeconds = val;
  }

  get remainingSeconds(): number {
    return this.shiftRemainingSeconds;
  }
  set remainingSeconds(val: number) {
    this.shiftRemainingSeconds = val;
  }

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
        jurors: this.jurors,
        totalDurationSeconds: this.totalDurationSeconds,
        shiftDurationSeconds: this.shiftDurationSeconds,
        roundRemainingSeconds: this.roundRemainingSeconds,
        shiftRemainingSeconds: this.shiftRemainingSeconds,
        currentShift: this.currentShift,
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
    this.totalDurationSeconds = payload.totalDurationSeconds ?? this.totalDurationSeconds;
    this.shiftDurationSeconds = payload.shiftDurationSeconds ?? this.shiftDurationSeconds;
    this.roundRemainingSeconds = payload.roundRemainingSeconds ?? this.roundRemainingSeconds;
    this.shiftRemainingSeconds = payload.shiftRemainingSeconds ?? this.shiftRemainingSeconds;
    this.currentShift = payload.currentShift ?? this.currentShift;
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
          totalDurationSeconds: this.totalDurationSeconds,
          shiftDurationSeconds: this.shiftDurationSeconds,
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
        if (typeof parsed.totalDurationSeconds === "number") {
          this.totalDurationSeconds = parsed.totalDurationSeconds;
          this.roundRemainingSeconds = parsed.totalDurationSeconds;
        }
        if (typeof parsed.shiftDurationSeconds === "number") {
          this.shiftDurationSeconds = parsed.shiftDurationSeconds;
          this.shiftRemainingSeconds = parsed.shiftDurationSeconds;
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
    if (config.totalDurationSeconds !== undefined) {
      this.totalDurationSeconds = config.totalDurationSeconds;
      if (this.state === "IDLE" || this.state === "FINISHED") {
        this.roundRemainingSeconds = config.totalDurationSeconds;
      }
    }
    if (config.shiftDurationSeconds !== undefined) {
      this.shiftDurationSeconds = config.shiftDurationSeconds;
      if (this.state === "IDLE" || this.state === "FINISHED") {
        this.shiftRemainingSeconds = config.shiftDurationSeconds;
      }
    }
    if (config.durationSeconds !== undefined && config.shiftDurationSeconds === undefined) {
      this.shiftDurationSeconds = config.durationSeconds;
      if (this.state === "IDLE" || this.state === "FINISHED") {
        this.shiftRemainingSeconds = config.durationSeconds;
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

    if (this.roundRemainingSeconds <= 0) {
      this.roundRemainingSeconds = this.totalDurationSeconds;
      this.shiftRemainingSeconds = this.shiftDurationSeconds;
      this.currentShift = 1;
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
      // Gracefully ignore if engine disconnected or in test environment
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
      console.warn("Failed to stop engine recording:", e);
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
    this.roundRemainingSeconds = this.totalDurationSeconds;
    this.shiftRemainingSeconds = this.shiftDurationSeconds;
    this.currentShift = 1;
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

      // Decrement total round seconds
      if (this.roundRemainingSeconds > 0) {
        this.roundRemainingSeconds -= 1;
      }

      // Decrement shift seconds
      if (this.shiftRemainingSeconds > 0) {
        this.shiftRemainingSeconds -= 1;
      }

      // Check shift transition (every 150s)
      if (this.shiftRemainingSeconds === 0) {
        if (this.roundRemainingSeconds > 0) {
          // Advance to next shift
          this.currentShift = Math.min(4, this.currentShift + 1);
          this.shiftRemainingSeconds = this.shiftDurationSeconds;
          this.rotationAlert = true;
          // Auto-clear rotation alert after 10 seconds
          setTimeout(() => {
            this.rotationAlert = false;
            this.broadcast();
          }, 10000);
        } else {
          // Entire round finished (600 seconds)
          this.state = "FINISHED";
          this.rotationAlert = true;
          if (this.timerInterval) clearInterval(this.timerInterval);
          this.stopRound();
        }
      }

      this.broadcast();
    }, 1000);
  }
}

export const roundStore = new RoundStore();
