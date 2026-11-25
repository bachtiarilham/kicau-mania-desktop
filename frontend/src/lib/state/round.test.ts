import { describe, it, expect, beforeEach, vi } from "vitest";
import { roundStore, getJudgeAssignedZone, getJudgeRotationSchedule } from "./round.svelte";

describe("roundStore and judge rotation matrix", () => {
  beforeEach(() => {
    roundStore.setCloudRound({
      eventId: "event-1",
      roundId: "round-1",
      eventName: "Piala Umi Kasum",
      sessionName: "Sesi 1",
      startTime: "10:00",
      endTime: "10:10",
      jurors: ["Juri 1", "Juri 2", "Juri 3", "Juri 4"],
    });
  });

  it("initializes with default config values for 10m round and 150s shifts", () => {
    expect(roundStore.eventName).toBe("Piala Umi Kasum");
    expect(roundStore.sessionName).toBe("Sesi 1");
    expect(roundStore.totalDurationSeconds).toBe(600);
    expect(roundStore.shiftDurationSeconds).toBe(150);
    expect(roundStore.roundRemainingSeconds).toBe(600);
    expect(roundStore.shiftRemainingSeconds).toBe(150);
    expect(roundStore.currentShift).toBe(1);
    expect(roundStore.state).toBe("IDLE");
    expect(roundStore.rotationAlert).toBe(false);
  });

  it("calculates exact rotation matrix for 4 judges across 4 shifts", () => {
    // Shift 1: J1=A, J2=B, J3=C, J4=D
    expect(getJudgeAssignedZone(0, 1)).toBe("A");
    expect(getJudgeAssignedZone(1, 1)).toBe("B");
    expect(getJudgeAssignedZone(2, 1)).toBe("C");
    expect(getJudgeAssignedZone(3, 1)).toBe("D");

    // Shift 2: J1=B, J2=C, J3=D, J4=A
    expect(getJudgeAssignedZone(0, 2)).toBe("B");
    expect(getJudgeAssignedZone(1, 2)).toBe("C");
    expect(getJudgeAssignedZone(2, 2)).toBe("D");
    expect(getJudgeAssignedZone(3, 2)).toBe("A");

    // Shift 3: J1=C, J2=D, J3=A, J4=B
    expect(getJudgeAssignedZone(0, 3)).toBe("C");
    expect(getJudgeAssignedZone(1, 3)).toBe("D");
    expect(getJudgeAssignedZone(2, 3)).toBe("A");
    expect(getJudgeAssignedZone(3, 3)).toBe("B");

    // Shift 4: J1=D, J2=A, J3=B, J4=C
    expect(getJudgeAssignedZone(0, 4)).toBe("D");
    expect(getJudgeAssignedZone(1, 4)).toBe("A");
    expect(getJudgeAssignedZone(2, 4)).toBe("B");
    expect(getJudgeAssignedZone(3, 4)).toBe("C");
  });

  it("replaces cloud round configuration, including any missing judges", () => {
    roundStore.setCloudRound({
      eventId: "event-2",
      roundId: "round-2",
      eventName: "Piala Presiden 2026",
      sessionName: "Sesi 2",
      startTime: "23:50",
      endTime: "00:10",
      jurors: [],
    });
    expect(roundStore.eventName).toBe("Piala Presiden 2026");
    expect(roundStore.sessionName).toBe("Sesi 2");
    expect(roundStore.totalDurationSeconds).toBe(1200);
    expect(roundStore.roundRemainingSeconds).toBe(1200);
    expect(roundStore.shiftDurationSeconds).toBe(300);
    expect(roundStore.jurors).toEqual([]);
  });

  it("starts and stops round", () => {
    roundStore.startRound();
    expect(roundStore.state).toBe("RUNNING");
    expect(roundStore.rotationAlert).toBe(false);

    roundStore.stopRound();
    expect(roundStore.state).toBe("FINISHED");
    expect(roundStore.timeUpDialog).toBe(false);
  });

  it("supports continuing round when time is up", () => {
    roundStore.startRound();
    roundStore.roundRemainingSeconds = 0;
    roundStore.timeUpDialog = true;

    roundStore.continueRound(60);
    expect(roundStore.roundRemainingSeconds).toBe(60);
    expect(roundStore.timeUpDialog).toBe(false);
  });

  it("computes deterministic zone schedule with clock ranges for each judge", () => {
    const j1 = getJudgeRotationSchedule(0, "13:00", 150);
    expect(j1.map((s) => s.zone)).toEqual(["A", "B", "C", "D"]);
    expect(j1[0].startClock).toBe("13:00");
    expect(j1[0].endClock).toBe("13:02:30");
    expect(j1[1].startClock).toBe("13:02:30");
    expect(j1[1].endClock).toBe("13:05");

    const j2 = getJudgeRotationSchedule(1, "13:00", 150);
    expect(j2.map((s) => s.zone)).toEqual(["B", "C", "D", "A"]);
  });

  it("toggles gantangan and judge zone checklists persistently per round", () => {
    expect(roundStore.isGantanganChecked("round-1", 5)).toBe(false);
    roundStore.toggleGantanganCheck("round-1", 5);
    expect(roundStore.isGantanganChecked("round-1", 5)).toBe(true);
    roundStore.toggleGantanganCheck("round-1", 5);
    expect(roundStore.isGantanganChecked("round-1", 5)).toBe(false);

    roundStore.setAllGantanganChecks("round-1", [1, 2, 3], true);
    expect(roundStore.isGantanganChecked("round-1", 2)).toBe(true);

    expect(roundStore.isJudgeZoneChecked("round-1", 0, "A")).toBe(false);
    roundStore.toggleJudgeZoneCheck("round-1", 0, "A");
    expect(roundStore.isJudgeZoneChecked("round-1", 0, "A")).toBe(true);
  });
});
