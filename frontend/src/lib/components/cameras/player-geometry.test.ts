import { describe, expect, it } from "vitest";
import { pixelRect } from "./player-geometry";

describe("native camera clipping", () => {
  it("rounds the viewport boundaries inward at fractional scale", () => {
    const bounds = { left: 3.4, top: 8.5, right: 101.2, bottom: 65.9 };
    const clip = { left: 4.2, top: 9.7, right: 99.8, bottom: 64.2 };
    expect(pixelRect(bounds, clip, 1.25, true)).toMatchObject({
      x: 4, y: 11, width: 123, height: 71,
      clip_x: 2, clip_y: 2, clip_width: 118, clip_height: 67,
      visible: true, device_pixels: true,
    });
  });

  it("hides a native pane entirely outside the visible card", () => {
    const bounds = { left: 10, top: 100, right: 90, bottom: 150 };
    const clip = { ...bounds, bottom: 75 };
    expect(pixelRect(bounds, clip, 1, true)).toMatchObject({ clip_height: 0, visible: false });
  });
});
