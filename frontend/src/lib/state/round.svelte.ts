import { engine } from "./engine.svelte";
import { auth } from "./auth.svelte";
import { api } from "$lib/services/api";
import type {
  JudgeZoneScheduleItem,
  ParticipantMapping,
  RoundConfig,
  RoundState,
  ZoneId,
} from "$lib/models";

export type { JudgeZoneScheduleItem, ParticipantMapping, RoundConfig, RoundState, ZoneId };

const ZONE_RANGES: Record<ZoneId, string> = {
  A: "G01 – G06",
  B: "G07 – G12",
  C: "G13 – G18",
  D: "G19 – G24",
};

export function getJudgeAssignedZone(
  judgeIndex: number,
  currentShift: number
): "A" | "B" | "C" | "D" {
  const zones: ("A" | "B" | "C" | "D")[] = ["A", "B", "C", "D"];
  const zoneIndex = (((judgeIndex % 4) + (Math.max(1, currentShift) - 1)) % 4 + 4) % 4;
  return zones[zoneIndex];
}

function parseClockSeconds(raw: string): number {
  const match = raw.trim().match(/^(\d{1,2})[:.](\d{2})(?::(\d{2}))?/);
  if (!match) return 9 * 3600;
  const h = Number(match[1]);
  const m = Number(match[2]);
  const s = Number(match[3] || "0");
  if (h >= 24 || m >= 60 || s >= 60) return 9 * 3600;
  return h * 3600 + m * 60 + s;
}

function formatClock(totalSeconds: number): string {
  const normalized = ((totalSeconds % 86400) + 86400) % 86400;
  const h = Math.floor(normalized / 3600);
  const m = Math.floor((normalized % 3600) / 60);
  const s = normalized % 60;
  const base = `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}`;
  return s === 0 ? base : `${base}:${String(s).padStart(2, "0")}`;
}

export function getJudgeRotationSchedule(
  judgeIndex: number,
  startTime = "",
  shiftDurationSeconds = 150
): JudgeZoneScheduleItem[] {
  const baseSeconds = parseClockSeconds(startTime || "09:00");
  const step = Math.max(1, shiftDurationSeconds || 150);
  const items: JudgeZoneScheduleItem[] = [];
  for (let shift = 1; shift <= 4; shift++) {
    const zone = getJudgeAssignedZone(judgeIndex, shift);
    const fromSec = baseSeconds + (shift - 1) * step;
    const toSec = baseSeconds + shift * step;
    items.push({
      shift,
      zone,
      range: ZONE_RANGES[zone],
      startClock: formatClock(fromSec),
      endClock: formatClock(toSec),
    });
  }
  return items;
}

const STORAGE_KEY = "kicau_cloud_round_v1";
const SYNC_CHANNEL = "kicau_cloud_round_sync_v1";
const defaultRoundDuration = 600;

function scheduledSeconds(start: string, end: string) {
  const parse = (time: string) => {
    const match = time.trim().match(/^(\d{1,2})[:.](\d{2})(?::(\d{2}))?(?:\s*(?:WIB|WITA|WIT))?$/i);
    if (!match) return null;
    const [, h, m, s = "00"] = match;
    if (+h >= 24 || +m >= 60 || +s >= 60) return null;
    return +h * 3600 + +m * 60 + +s;
  };
  const from = parse(start);
  const to = parse(end);
  if (from === null || to === null) return defaultRoundDuration;
  const duration = (to - from + 86400) % 86400;
  return duration >= 4 ? duration : defaultRoundDuration;
}

class RoundStore {
  eventName = $state("");
  sessionName = $state("");
  startTime = $state("");
  endTime = $state("");
  jurors = $state<string[]>([]);

  eventId = $state("");
  roundId = $state("");
  participants = $state<ParticipantMapping[]>([]);

  // Round is 10 minutes (600 seconds) divided into 4 shifts of 150 seconds each
  totalDurationSeconds = $state(600);
  shiftDurationSeconds = $state(150);
  roundRemainingSeconds = $state(600);
  shiftRemainingSeconds = $state(150);
  currentShift = $state(1);

  state = $state<RoundState>("IDLE");
  rotationAlert = $state(false);
  timeUpDialog = $state(false);
  overtimeSeconds = $state(0);
  checkedGantangan = $state<Record<string, Record<number, boolean>>>({});
  judgeZoneChecks = $state<Record<string, Record<string, boolean>>>({});

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
        jurors: $state.snapshot(this.jurors),
        eventId: this.eventId,
        roundId: this.roundId,
        participants: $state.snapshot(this.participants),
        totalDurationSeconds: this.totalDurationSeconds,
        shiftDurationSeconds: this.shiftDurationSeconds,
        roundRemainingSeconds: this.roundRemainingSeconds,
        shiftRemainingSeconds: this.shiftRemainingSeconds,
        currentShift: this.currentShift,
        state: this.state,
        rotationAlert: this.rotationAlert,
        timeUpDialog: this.timeUpDialog,
        overtimeSeconds: this.overtimeSeconds,
        startedAt: this.startedAt,
        checkedGantangan: $state.snapshot(this.checkedGantangan),
        judgeZoneChecks: $state.snapshot(this.judgeZoneChecks),
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
    this.eventId = payload.eventId ?? this.eventId;
    this.roundId = payload.roundId ?? this.roundId;
    this.participants = Array.isArray(payload.participants) ? payload.participants : this.participants;
    this.totalDurationSeconds = payload.totalDurationSeconds ?? this.totalDurationSeconds;
    this.shiftDurationSeconds = payload.shiftDurationSeconds ?? this.shiftDurationSeconds;
    this.roundRemainingSeconds = payload.roundRemainingSeconds ?? this.roundRemainingSeconds;
    this.shiftRemainingSeconds = payload.shiftRemainingSeconds ?? this.shiftRemainingSeconds;
    this.currentShift = payload.currentShift ?? this.currentShift;
    this.state = payload.state ?? this.state;
    this.rotationAlert = payload.rotationAlert ?? this.rotationAlert;
    this.timeUpDialog = payload.timeUpDialog ?? this.timeUpDialog;
    this.overtimeSeconds = payload.overtimeSeconds ?? this.overtimeSeconds;
    this.startedAt = payload.startedAt ?? this.startedAt;
    if (payload.checkedGantangan) this.checkedGantangan = payload.checkedGantangan;
    if (payload.judgeZoneChecks) this.judgeZoneChecks = payload.judgeZoneChecks;

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
          eventId: this.eventId,
          roundId: this.roundId,
          participants: this.participants,
          totalDurationSeconds: this.totalDurationSeconds,
          shiftDurationSeconds: this.shiftDurationSeconds,
          checkedGantangan: $state.snapshot(this.checkedGantangan),
          judgeZoneChecks: $state.snapshot(this.judgeZoneChecks),
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
        if (parsed.eventId) this.eventId = parsed.eventId;
        if (parsed.roundId) this.roundId = parsed.roundId;
        if (Array.isArray(parsed.participants)) this.participants = parsed.participants;
        if (typeof parsed.totalDurationSeconds === "number") {
          this.totalDurationSeconds = parsed.totalDurationSeconds;
          this.roundRemainingSeconds = parsed.totalDurationSeconds;
        }
        if (typeof parsed.shiftDurationSeconds === "number") {
          this.shiftDurationSeconds = parsed.shiftDurationSeconds;
          this.shiftRemainingSeconds = parsed.shiftDurationSeconds;
        }
        if (parsed.checkedGantangan && typeof parsed.checkedGantangan === "object") {
          this.checkedGantangan = parsed.checkedGantangan;
        }
        if (parsed.judgeZoneChecks && typeof parsed.judgeZoneChecks === "object") {
          this.judgeZoneChecks = parsed.judgeZoneChecks;
        }
      }
    } catch {
      // Ignore storage parse error
    }
  }

  setCloudRound(config: {
    eventId: string;
    roundId: string;
    eventName: string;
    sessionName: string;
    startTime: string;
    endTime: string;
    jurors: string[];
    participants?: ParticipantMapping[];
  }) {
    this.eventId = config.eventId;
    this.roundId = config.roundId;
    this.eventName = config.eventName;
    this.sessionName = config.sessionName;
    this.startTime = config.startTime;
    this.endTime = config.endTime;
    this.jurors = [...config.jurors];
    this.participants = [...(config.participants ?? [])];
    this.totalDurationSeconds = scheduledSeconds(config.startTime, config.endTime);
    this.shiftDurationSeconds = Math.floor(this.totalDurationSeconds / 4);
    this.saveToStorage();
    this.resetRound();
  }

  clearCloudRound() {
    this.setCloudRound({
      eventId: "", roundId: "", eventName: "", sessionName: "",
      startTime: "", endTime: "", jurors: [],
    });
  }

  getParticipantByGantangan(num: number): ParticipantMapping | undefined {
    return this.participants.find((p) => p.gantanganNumber === num);
  }

  isGantanganChecked(roundId: string, num: number): boolean {
    if (!roundId) return false;
    return Boolean(this.checkedGantangan[roundId]?.[num]);
  }

  toggleGantanganCheck(roundId: string, num: number) {
    if (!roundId) return;
    const prev = this.checkedGantangan[roundId] ?? {};
    this.checkedGantangan = {
      ...this.checkedGantangan,
      [roundId]: { ...prev, [num]: !prev[num] },
    };
    this.saveToStorage();
    this.broadcast();
  }

  setAllGantanganChecks(roundId: string, nums: number[], checked: boolean) {
    if (!roundId) return;
    const next: Record<number, boolean> = { ...(this.checkedGantangan[roundId] ?? {}) };
    for (const num of nums) {
      next[num] = checked;
    }
    this.checkedGantangan = { ...this.checkedGantangan, [roundId]: next };
    this.saveToStorage();
    this.broadcast();
  }

  isJudgeZoneChecked(roundId: string, judgeIndex: number, zone: ZoneId): boolean {
    const key = roundId || "default";
    return Boolean(this.judgeZoneChecks[key]?.[`${judgeIndex}-${zone}`]);
  }

  toggleJudgeZoneCheck(roundId: string, judgeIndex: number, zone: ZoneId) {
    const key = roundId || "default";
    const prev = this.judgeZoneChecks[key] ?? {};
    const slotKey = `${judgeIndex}-${zone}`;
    this.judgeZoneChecks = {
      ...this.judgeZoneChecks,
      [key]: { ...prev, [slotKey]: !prev[slotKey] },
    };
    this.saveToStorage();
    this.broadcast();
  }

  async startRound() {
    if (!this.roundId || this.state === "RUNNING") return;

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

    if (typeof window !== "undefined" && auth.arenaId && this.roundId) {
      void api.roundAction(auth.arenaId, this.roundId, "start", "PREPARED").catch(() => {});
    }

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

  continueRound(extraSeconds = 0) {
    if (extraSeconds > 0) {
      this.roundRemainingSeconds += extraSeconds;
      this.shiftRemainingSeconds += extraSeconds;
    }
    this.timeUpDialog = false;
    this.rotationAlert = false;
    this.broadcast();
  }

  stopRound() {
    if (this.timerInterval) {
      clearInterval(this.timerInterval);
      this.timerInterval = null;
    }
    this.state = "FINISHED";
    this.timeUpDialog = false;
    this.rotationAlert = false;
    this.broadcast();

    if (typeof window !== "undefined" && auth.arenaId && this.roundId) {
      void api.roundAction(auth.arenaId, this.roundId, "begin-judging", "RUNNING").catch(() => {});
    }

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
    this.timeUpDialog = false;
    this.overtimeSeconds = 0;
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
        if (this.roundRemainingSeconds === 0) {
          // Time is up! DO NOT auto-stop recording.
          // Trigger the dialog so field operator can choose to continue or stop.
          this.timeUpDialog = true;
          this.rotationAlert = true;
        }
      } else {
        // Overtime: keep counting up so operator knows how much extra time has elapsed
        this.overtimeSeconds += 1;
      }

      // Decrement shift seconds
      if (this.shiftRemainingSeconds > 0) {
        this.shiftRemainingSeconds -= 1;
      }

      // Check shift transition (every 150s)
      if (this.shiftRemainingSeconds === 0) {
        if (this.currentShift < 4) {
          // Advance to next shift
          this.currentShift += 1;
          this.shiftRemainingSeconds = this.currentShift === 4 ? this.roundRemainingSeconds : this.shiftDurationSeconds;
          this.rotationAlert = true;
          // Auto-clear rotation alert after 10 seconds
          setTimeout(() => {
            this.rotationAlert = false;
            this.broadcast();
          }, 10000);
        }
      }

      this.broadcast();
    }, 1000);
  }
}

export const roundStore = new RoundStore();
