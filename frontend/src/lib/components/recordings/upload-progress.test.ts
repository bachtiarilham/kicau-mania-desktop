import { describe, expect, it } from "vitest";
import { uploadPercentage } from "./upload-progress";

describe("upload percentage", () => {
  it("uses acknowledged bytes and waits for Cloud confirmation before showing 100%", () => {
    expect(uploadPercentage("WAITING_MAPPING", 0, 0)).toBe(0);
    expect(uploadPercentage("UPLOADING", 32, 80)).toBe(40);
    expect(uploadPercentage("SYNCING", 80, 80)).toBe(99);
    expect(uploadPercentage("UPLOADED", 0, 0)).toBe(100);
  });

  it("keeps invalid byte counts within the progress range", () => {
    expect(uploadPercentage("UPLOADING", -1, 80)).toBe(0);
    expect(uploadPercentage("UPLOADING", 1, 0)).toBe(0);
    expect(uploadPercentage("RETRYING", 100, 80)).toBe(99);
  });
});
