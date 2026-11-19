import { describe, it, expect, beforeEach, vi } from "vitest";
import { roundStore } from "./round.svelte";

describe("roundStore", () => {
  beforeEach(() => {
    roundStore.resetRound();
    roundStore.updateConfig({
      eventName: "Piala Umi Kasum",
      sessionName: "Sesi 1",
      durationSeconds: 150,
      jurors: ["Juri 1", "Juri 2", "Juri 3", "Juri 4"],
    });
  });

  it("initializes with default config values", () => {
    expect(roundStore.eventName).toBe("Piala Umi Kasum");
    expect(roundStore.sessionName).toBe("Sesi 1");
    expect(roundStore.durationSeconds).toBe(150);
    expect(roundStore.remainingSeconds).toBe(150);
    expect(roundStore.state).toBe("IDLE");
    expect(roundStore.rotationAlert).toBe(false);
  });

  it("updates config dynamically", () => {
    roundStore.updateConfig({
      eventName: "Piala Presiden 2026",
      sessionName: "Sesi 2",
      durationSeconds: 180,
    });
    expect(roundStore.eventName).toBe("Piala Presiden 2026");
    expect(roundStore.sessionName).toBe("Sesi 2");
    expect(roundStore.durationSeconds).toBe(180);
    expect(roundStore.remainingSeconds).toBe(180);
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
