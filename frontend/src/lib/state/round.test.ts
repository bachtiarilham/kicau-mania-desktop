import { describe, it, expect, beforeEach, vi } from "vitest";
import { roundStore, getJudgeAssignedZone } from "./round.svelte";

describe("roundStore and judge rotation matrix", () => {
  beforeEach(() => {
    roundStore.resetRound();
    roundStore.updateConfig({
      eventName: "Piala Umi Kasum",
      sessionName: "Sesi 1",
      totalDurationSeconds: 600,
      shiftDurationSeconds: 150,
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

  it("updates config dynamically", () => {
    roundStore.updateConfig({
      eventName: "Piala Presiden 2026",
      sessionName: "Sesi 2",
      totalDurationSeconds: 600,
      shiftDurationSeconds: 150,
    });
    expect(roundStore.eventName).toBe("Piala Presiden 2026");
    expect(roundStore.sessionName).toBe("Sesi 2");
    expect(roundStore.totalDurationSeconds).toBe(600);
    expect(roundStore.roundRemainingSeconds).toBe(600);
    expect(roundStore.shiftDurationSeconds).toBe(150);
  });

  it("starts and stops round", () => {
    roundStore.startRound();
    expect(roundStore.state).toBe("RUNNING");
    expect(roundStore.rotationAlert).toBe(false);

    roundStore.stopRound();
    expect(roundStore.state).toBe("FINISHED");
    expect(roundStore.rotationAlert).toBe(true);
  });
});
