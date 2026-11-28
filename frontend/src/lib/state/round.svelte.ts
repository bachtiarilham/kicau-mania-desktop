import { engine } from "./engine.svelte";
import { auth } from "./auth.svelte";
import { api, message } from "$lib/services/api";
import { gantanganChannels } from "$lib/components/cameras/camera-slots";
import type { JudgeZoneScheduleItem, ParticipantMapping, RoundState, ZoneId } from "$lib/models";
export type { ParticipantMapping, ZoneId } from "$lib/models";

const zones: ZoneId[] = ["A", "B", "C", "D"];
const ranges = ["G01 – G06", "G07 – G12", "G13 – G18", "G19 – G24"];
const regularSeconds = 120;
const finalSeconds = 30;
const regularEnd = zones.length * regularSeconds;
const duration = regularEnd + zones.length * finalSeconds;
export const roundMilestones = [...Array.from({ length: zones.length + 1 }, (_, index) => index * regularSeconds), duration];
const roundPath = "/api/v1/local/round";
const storageKey = "kicau_cloud_round_v1";

export function getRoundInterval(elapsed: number) {
  const second = Math.min(elapsed, duration - 1);
  if (second < regularEnd) {
    const index = Math.floor(second / regularSeconds);
    return { index, offset: index, length: regularSeconds, end: (index + 1) * regularSeconds, final: false };
  }
  const index = Math.floor((second - regularEnd) / finalSeconds);
  return { index, offset: (zones.length - 1 + index) % zones.length, length: finalSeconds, end: regularEnd + (index + 1) * finalSeconds, final: true };
}

export function getJudgeAssignedZone(judgeIndex: number, elapsed: number): ZoneId {
  return zones[(judgeIndex + getRoundInterval(elapsed).offset) % zones.length];
}

export function getRoundProgress(elapsed: number, state: RoundState) {
  const elapsedSeconds = state === "IDLE" ? 0 : Math.max(0, Math.min(duration, elapsed));
  return {
    elapsedSeconds,
    percent: elapsedSeconds * 100 / duration,
    activeStep: roundMilestones.findLastIndex((milestone) => milestone <= elapsedSeconds),
  };
}

export function getJudgeRotationSchedule(judgeIndex: number): JudgeZoneScheduleItem[] {
  return zones.map((_, index) => {
    const shift = index + 1;
    const slot = (judgeIndex + index) % zones.length;
    return {
      shift,
      zone: zones[slot],
      range: ranges[slot],
      startClock: `0${index * 2}:00`,
      endClock: `0${shift * 2}:00`,
    };
  });
}

type RoundSelection = {
  eventId: string;
  roundId: string;
  eventName: string;
  sessionName: string;
  startTime: string;
  endTime: string;
  jurors: string[];
  jurorIds?: string[];
  participants?: ParticipantMapping[];
};

type RoundSnapshot = Omit<RoundSelection, "participants"> & {
  arenaId: string;
  participants: Pick<ParticipantMapping, "gantanganNumber" | "birdName" | "ownerName">[];
  state: "RUNNING" | "FINISHED";
  elapsedMs: number;
  runId: string;
};

class RoundStore {
  eventName = $state("");
  sessionName = $state("");
  startTime = $state("");
  endTime = $state("");
  jurors = $state<string[]>([]);
  jurorIds = $state<string[]>([]);
  eventId = $state("");
  roundId = $state("");
  participants = $state<ParticipantMapping[]>([]);
  state = $state<RoundState>("IDLE");
  checkedGantangan = $state<Record<string, Record<number, boolean>>>({});
  elapsedSeconds = $state(0);
  busy = $state(false);
  commandError = $state("");
  syncError = $state("");

  readonly totalDurationSeconds = duration;
  interval = $derived(getRoundInterval(this.elapsedSeconds));
  intervalLabel = $derived(this.interval.final ? `Putaran akhir ${this.interval.index + 1}/4` : `Shift ${this.interval.index + 1}/4`);
  roundRemainingSeconds = $derived(duration - this.elapsedSeconds);
  shiftRemainingSeconds = $derived(this.interval.end - this.elapsedSeconds);
  timeUpDialog = $derived(this.state === "RUNNING" && this.elapsedSeconds === duration);
  nextZoneElapsed = $derived(this.interval.end === regularEnd ? regularEnd + finalSeconds : this.interval.end);
  private timer: ReturnType<typeof setInterval> | undefined;
  private anchor = 0;
  private runId = "";

  constructor() {
    if (typeof localStorage === "undefined") return;
    try {
      const saved = JSON.parse(localStorage.getItem(storageKey) ?? "{}");
      this.setSelection(saved);
      this.checkedGantangan = saved.checkedGantangan ?? {};
    } catch {}
  }

  private setSelection(config: Partial<RoundSelection>) {
    this.eventId = config.eventId ?? "";
    this.roundId = config.roundId ?? "";
    this.eventName = config.eventName ?? "";
    this.sessionName = config.sessionName ?? "";
    this.startTime = config.startTime ?? "";
    this.endTime = config.endTime ?? "";
    this.jurors = [...(config.jurors ?? [])];
    this.jurorIds = [...(config.jurorIds ?? [])];
    this.participants = [...(config.participants ?? [])];
  }

  private saveToStorage() {
    if (typeof localStorage === "undefined") return;
    try {
      localStorage.setItem(storageKey, JSON.stringify({
        eventId: this.eventId, roundId: this.roundId, eventName: this.eventName, sessionName: this.sessionName,
        startTime: this.startTime, endTime: this.endTime, jurors: this.jurors, jurorIds: this.jurorIds, participants: this.participants,
        checkedGantangan: this.checkedGantangan,
      }));
    } catch {}
  }

  setCloudRound(config: RoundSelection) {
    this.setSelection(config);
    this.resetRound();
    this.saveToStorage();
  }

  selectCloudEvent(eventId: string, eventName: string) {
    if (eventId && this.eventId === eventId) {
      if (this.eventName !== eventName) {
        this.eventName = eventName;
        this.saveToStorage();
      }
      return;
    }
    this.setCloudRound({ eventId, roundId: "", eventName, sessionName: "", startTime: "", endTime: "", jurors: [] });
  }

  clearCloudRound() {
    this.selectCloudEvent("", "");
  }

  getParticipantByGantangan(num: number) {
    return this.participants.find((participant) => participant.gantanganNumber === num);
  }

  isGantanganChecked(roundId: string, num: number) {
    return Boolean(roundId && this.checkedGantangan[roundId]?.[num]);
  }

  toggleGantanganCheck(roundId: string, num: number) {
    this.setAllGantanganChecks(roundId, [num], !this.isGantanganChecked(roundId, num));
  }

  setAllGantanganChecks(roundId: string, nums: number[], checked: boolean) {
    if (!roundId) return;
    const next = { ...this.checkedGantangan[roundId] };
    for (const num of nums) next[num] = checked;
    this.checkedGantangan = { ...this.checkedGantangan, [roundId]: next };
    this.saveToStorage();
  }

  isJudgeZoneChecked(roundId: string, judgeIndex: number, zone: ZoneId) {
    const item = getJudgeRotationSchedule(judgeIndex).find((shift) => shift.zone === zone);
    return Boolean(roundId && roundId === this.roundId && item && this.elapsedSeconds >= item.shift * regularSeconds);
  }

  get judgeIndex(): number | null {
    const byID = this.jurorIds.flatMap((id, index) => id && id === auth.userId ? [index] : []);
    if (byID.length > 0) return byID.length === 1 ? byID[0] : null;
    const normalize = (value: string) => value.trim().toLocaleLowerCase("id-ID").replace(/\s+/g, " ");
    const name = normalize(auth.userName);
    const matches = this.jurors.flatMap((judge, index) => name && normalize(judge) === name ? [index] : []);
    return matches.length === 1 && !this.jurorIds[matches[0]] ? matches[0] : null;
  }

  async recoverOperatorRound(arenaId: string) {
    try {
      const started = performance.now();
      const snapshot = await api.engine<RoundSnapshot | null>("GET", roundPath);
      if (auth.arenaId !== arenaId || auth.role === "juri" || this.busy || this.state !== "IDLE" ||
          !snapshot || snapshot.arenaId !== arenaId || snapshot.state !== "RUNNING") return;
      if (this.roundId !== snapshot.roundId) this.setRemoteSelection(snapshot);
      this.applyClock(snapshot, (performance.now() - started) / 2);
    } catch {
      // START and STOP report errors when the operator sends a command.
    }
  }

  private setRemoteSelection(snapshot: RoundSnapshot) {
    this.setSelection({
      ...snapshot,
      participants: snapshot.participants.map((p) => ({ ...p, registrationId: "", gantanganId: "" })),
    });
  }

  followArena(arenaId: string) {
    let disposed = false;
    let pending = false;
    this.syncError = "Menghubungkan ke server arena…";
    const refresh = async () => {
      if (pending) return;
      pending = true;
      const requestStarted = performance.now();
      try {
        const snapshot = await api.engine<RoundSnapshot | null>("GET", roundPath);
        if (disposed) return;
        this.syncError = "";
        if (!snapshot || snapshot.arenaId !== arenaId) {
          this.setSelection({});
          this.resetRound();
          return;
        }
        if (this.roundId !== snapshot.roundId || this.runId !== snapshot.runId) {
          this.setRemoteSelection(snapshot);
        }
        this.applyClock(snapshot, (performance.now() - requestStarted) / 2);
      } catch (error) {
        if (!disposed) this.syncError = `Sinkronisasi ronde terputus · ${message(error)}`;
      } finally {
        pending = false;
      }
    };
    void refresh();
    const poll = setInterval(() => void refresh(), 1000);
    return () => {
      disposed = true;
      clearInterval(poll);
      this.stopTicker();
      this.syncError = "";
    };
  }

  async startRound() {
    if (!this.roundId || this.state === "RUNNING" || this.busy) return;
    this.busy = true;
    this.commandError = "";
    let started = false;
    try {
      const selection = {
        arenaId: auth.arenaId, eventId: this.eventId, roundId: this.roundId,
        eventName: this.eventName, sessionName: this.sessionName, jurors: $state.snapshot(this.jurors),
        startTime: this.startTime, endTime: this.endTime,
        jurorIds: $state.snapshot(this.jurorIds),
        participants: this.participants.map(({ gantanganNumber, birdName, ownerName }) => ({ gantanganNumber, birdName, ownerName })),
      };
      const requestStarted = performance.now();
      const snapshot = await api.engine<RoundSnapshot>("POST", roundPath + "/start", selection);
      this.applyClock(snapshot, (performance.now() - requestStarted) / 2);
      started = true;
      void api.roundAction(selection.arenaId, selection.roundId, "start", "PREPARED").catch(() => {});
      if (engine.active || engine.busy) return;
      if (!engine.draftID) engine.draftID = await api.sessionID();
      if (engine.pairing.engine_id && engine.pairing.arena_id === selection.arenaId) {
        const channels = gantanganChannels(engine.channels);
        const mappings = this.participants.flatMap((participant) => {
          const channel = channels[participant.gantanganNumber - 1];
          return channel?.id && participant.registrationId && participant.gantanganId
            ? [{ channel_id: channel.id, registration_id: participant.registrationId, gantangan_id: participant.gantanganId }]
            : [];
        });
        if (mappings.length) {
          await api.bind({ engine_id: engine.pairing.engine_id, session_id: engine.draftID, round_id: selection.roundId, mappings }).catch(() => {
            engine.notice = "Pemetaan peserta ke rekaman belum tersinkron ke cloud.";
          });
        }
      }
      await engine.start(engine.draftID);
      engine.draftID = "";
    } catch (error) {
      this.commandError = `${started ? "Ronde berjalan, rekaman gagal dimulai" : "Ronde gagal dimulai"} · ${message(error)}`;
    } finally {
      this.busy = false;
    }
  }

  async stopRound(): Promise<boolean> {
    if (this.busy || this.state !== "RUNNING") return false;
    this.busy = true;
    this.commandError = "";
    try {
      const ids = { arenaId: auth.arenaId, roundId: this.roundId };
      const snapshot = await api.engine<RoundSnapshot>("POST", roundPath + "/stop", ids);
      this.applyClock(snapshot);
      void api.roundAction(ids.arenaId, ids.roundId, "begin-judging", "RUNNING").catch(() => {});
      if (engine.running && !engine.busy) void engine.stop().catch((error) => {
        this.commandError = `Ronde selesai, rekaman gagal dihentikan · ${message(error)}`;
      });
      return true;
    } catch (error) {
      this.commandError = `Ronde gagal dihentikan · ${message(error)}`;
      return false;
    } finally {
      this.busy = false;
    }
  }

  resetRound() {
    this.stopTicker();
    this.state = "IDLE";
    this.elapsedSeconds = 0;
    this.commandError = "";
    this.runId = "";
  }

  private applyClock(snapshot: RoundSnapshot, transitMs = 0) {
    this.state = snapshot.state;
    const elapsedMs = snapshot.elapsedMs + (this.state === "RUNNING" ? transitMs : 0);
    this.anchor = performance.now() - elapsedMs;
    this.runId = snapshot.runId;
    this.elapsedSeconds = Math.min(duration, Math.floor(elapsedMs / 1000));
    if (this.state !== "RUNNING" || this.elapsedSeconds === duration) {
      this.stopTicker();
    } else if (!this.timer) {
      this.timer = setInterval(() => {
        this.elapsedSeconds = Math.min(duration, Math.floor((performance.now() - this.anchor) / 1000));
        if (this.elapsedSeconds === duration) this.stopTicker();
      }, 100);
    }
  }

  private stopTicker() {
    clearInterval(this.timer);
    this.timer = undefined;
  }
}

export const roundStore = new RoundStore();
