import { describe, it, expect, vi, beforeEach } from "vitest";
const transport = vi.hoisted(() => vi.fn());
vi.mock("@wailsio/runtime", () => ({
  Call: { ByName: transport },
  Events: { On: vi.fn() },
}));
import { api, all } from "./api";
describe("desktop bindings and cloud lookup", () => {
  beforeEach(() => transport.mockReset());
  it("uses native Go service for engine mutations", async () => {
    transport.mockResolvedValue({ state: "STARTING" });
    await api.engine("POST", "/api/v1/media/sessions", { id: "session" });
    expect(transport).toHaveBeenCalledWith(
      "kicau.desktop/cctv/internal/bindings.Desktop.Engine",
      "POST",
      "/api/v1/media/sessions",
      { id: "session" },
    );
  });
  it("loads every cursor page without replacing failures with empty data", async () => {
    transport
      .mockResolvedValueOnce({
        items: [{ id: "1" }],
        has_more: true,
        next_cursor: "next",
      })
      .mockResolvedValueOnce({ items: [{ id: "2" }], has_more: false });
    expect(await all("/arenas")).toEqual([{ id: "1" }, { id: "2" }]);
    transport.mockRejectedValueOnce(Error("503"));
    await expect(all("/arenas")).rejects.toThrow("503");
  });
  it("rejects a repeating cursor", async () => {
    transport.mockResolvedValue({
      items: [],
      has_more: true,
      next_cursor: "repeat",
    });
    await expect(all("/arenas")).rejects.toThrow("Cursor");
  });
  it("keeps the selected session identity for mapping", async () => {
    const mapping = {
      session_id: "local-session",
      round_id: "cloud-round",
      mappings: [
        {
          channel_id: "g01",
          registration_id: "ticket",
          gantangan_id: "gantangan",
        },
      ],
    };
    transport.mockResolvedValue(mapping);
    await api.bind(mapping);
    expect(transport).toHaveBeenCalledWith(
      "kicau.desktop/cctv/internal/bindings.Desktop.BindSession",
      mapping,
    );
  });
});
