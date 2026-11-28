import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "$lib/services/api";
import { cameraSlots } from "$lib/components/cameras/camera-slots";
import { auth } from "./auth.svelte";
import { engine } from "./engine.svelte";
import { getJudgeAssignedZone, getJudgeRotationSchedule, getRoundInterval, getRoundProgress, roundStore } from "./round.svelte";

const round = {
  arenaId: "arena-1", eventId: "event-1", roundId: "round-1", eventName: "Piala Umi Kasum",
  sessionName: "Sesi 1", startTime: "13:00", endTime: "13:20",
  jurors: ["Juri 1", "Juri 2", "Juri 3", "Juri 4"], jurorIds: ["1", "2", "3", "4"], participants: [],
};

beforeEach(() => {
  vi.useFakeTimers({ toFake: ["setInterval", "clearInterval", "performance"] });
  roundStore.setCloudRound(round);
  auth.login("operator", "operator@example.test", "Operator", "10", round.arenaId);
  vi.spyOn(engine, "active", "get").mockReturnValue(true);
  vi.spyOn(api, "roundAction").mockResolvedValue(undefined);
});

afterEach(() => {
  roundStore.resetRound();
  auth.logout();
  vi.useRealTimers();
  vi.restoreAllMocks();
});

describe("official round timetable", () => {
  it("fills by elapsed time and preserves the actual progress on early stop", () => {
    expect(getRoundProgress(0, "IDLE")).toEqual({ activeStep: 0, percent: 0, elapsedSeconds: 0 });
    expect(getRoundProgress(60, "RUNNING")).toEqual({ activeStep: 0, percent: 10, elapsedSeconds: 60 });
    expect(getRoundProgress(120, "RUNNING")).toEqual({ activeStep: 1, percent: 20, elapsedSeconds: 120 });
    expect(getRoundProgress(510, "RUNNING")).toEqual({ activeStep: 4, percent: 85, elapsedSeconds: 510 });
    expect(getRoundProgress(600, "RUNNING")).toEqual({ activeStep: 5, percent: 100, elapsedSeconds: 600 });
    expect(getRoundProgress(240, "FINISHED")).toEqual({ activeStep: 2, percent: 40, elapsedSeconds: 240 });
  });
  it("keeps sparse camera bindings in their slots and excludes the arena camera", () => {
    const slots = cameraSlots([{ id: "gantangan24" }, { id: "arena-overview" }, { id: "gantangan03" }]);
    expect(slots).toHaveLength(24);
    expect(slots[0].attached).toBe(false);
    expect(slots[2]).toMatchObject({ id: "gantangan03", zone: "A" });
    expect(slots[23]).toMatchObject({ id: "gantangan24", zone: "D" });
    expect(slots.filter((slot) => slot.attached)).toHaveLength(2);
  });
  it.each([
    [0, "ABCD", 120], [120, "BCDA", 240], [240, "CDAB", 360], [360, "DABC", 480],
    [480, "DABC", 510], [510, "ABCD", 540], [540, "BCDA", 570], [570, "CDAB", 600],
  ])("assigns four distinct zones at elapsed second %i", (second, matrix, end) => {
    for (const elapsed of [second, end - 1]) {
      expect(Array.from({ length: 4 }, (_, index) => getJudgeAssignedZone(index, elapsed)).join("")).toBe(matrix);
      expect(getRoundInterval(elapsed).end).toBe(end);
    }
  });

  it("checks zones at the end of their regular shift and freezes at ten minutes", async () => {
    const call = vi.spyOn(api, "engine").mockResolvedValue({ ...round, state: "RUNNING", elapsedMs: 0 });
    await roundStore.startRound();
    expect(call).toHaveBeenCalledWith(
      "POST", "/api/v1/local/round/start",
      expect.objectContaining({ roundId: round.roundId, startTime: "13:00", endTime: "13:20" }),
    );
    expect(roundStore.roundRemainingSeconds).toBe(600);
    expect(roundStore.shiftRemainingSeconds).toBe(120);
    vi.advanceTimersByTime(119_900);
    expect(roundStore.isJudgeZoneChecked(round.roundId, 0, "A")).toBe(false);
    vi.advanceTimersByTime(100);
    expect(roundStore.isJudgeZoneChecked(round.roundId, 0, "A")).toBe(true);
    expect(roundStore.isJudgeZoneChecked(round.roundId, 0, "B")).toBe(false);
    vi.advanceTimersByTime(360_000);
    expect(roundStore.intervalLabel).toBe("Putaran akhir 1/4");
    expect(roundStore.shiftRemainingSeconds).toBe(30);
    for (let judge = 0; judge < 4; judge++) {
      expect(getJudgeRotationSchedule(judge).every(({ zone }) => roundStore.isJudgeZoneChecked(round.roundId, judge, zone))).toBe(true);
    }
    vi.advanceTimersByTime(180_000);
    expect(roundStore.roundRemainingSeconds).toBe(0);
    expect(roundStore.shiftRemainingSeconds).toBe(0);
    expect(roundStore.timeUpDialog).toBe(true);
    call.mockResolvedValue({ ...round, state: "FINISHED", elapsedMs: 600_000 });
    expect(await roundStore.stopRound()).toBe(true);
    expect(roundStore.timeUpDialog).toBe(false);
  });

  it("uses the arena snapshot to recover after reconnect and freezes on operator stop", async () => {
    auth.login("juri", "judge@example.test", "Nama berbeda dari jadwal", "2", round.arenaId);
    const call = vi.spyOn(api, "engine");
    call.mockResolvedValue({ ...round, state: "RUNNING", runId: "test", elapsedMs: 120_200 });
    const stopFollowing = roundStore.followArena(round.arenaId);
    try {
      await vi.advanceTimersByTimeAsync(0);
      expect(roundStore.judgeIndex).toBe(1);
      expect(roundStore.startTime).toBe("13:00");
      expect(roundStore.endTime).toBe("13:20");
      expect(getJudgeAssignedZone(1, roundStore.elapsedSeconds)).toBe("C");
      call.mockRejectedValue(new Error("offline"));
      await vi.advanceTimersByTimeAsync(1000);
      expect(roundStore.syncError).toContain("offline");
      call.mockResolvedValue({ ...round, state: "RUNNING", elapsedMs: 540_250 });
      await vi.advanceTimersByTimeAsync(1000);
      expect(roundStore.syncError).toBe("");
      expect(roundStore.intervalLabel).toBe("Putaran akhir 3/4");
      expect(roundStore.isJudgeZoneChecked(round.roundId, 1, "B")).toBe(true);
      call.mockResolvedValue({ ...round, state: "FINISHED", elapsedMs: 545_200 });
      await vi.advanceTimersByTimeAsync(2500);
      expect(roundStore.state).toBe("FINISHED");
      expect(roundStore.elapsedSeconds).toBe(545);
    } finally {
      stopFollowing();
    }
  });

  it("retains the operator's inspection checklist and clears the round when changing events", () => {
    roundStore.setAllGantanganChecks(round.roundId, [1, 2], true);
    roundStore.toggleGantanganCheck(round.roundId, 1);
    expect(roundStore.isGantanganChecked(round.roundId, 1)).toBe(false);
    expect(roundStore.isGantanganChecked(round.roundId, 2)).toBe(true);
    roundStore.selectCloudEvent(round.eventId, "Updated");
    expect(roundStore.roundId).toBe(round.roundId);
    roundStore.selectCloudEvent("event-2", "Next event");
    expect(roundStore.roundId).toBe("");
    expect(roundStore.jurors).toEqual([]);
  });
});
