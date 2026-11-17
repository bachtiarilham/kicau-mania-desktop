import { beforeEach, describe, expect, it, vi } from "vitest";
const transport = vi.hoisted(() => ({ engine: vi.fn(), bootstrap: vi.fn() }));
vi.mock("$lib/services/api", () => ({
  api: transport,
  message: (e: Error) => e.message,
}));
import { engine } from "./engine.svelte";
describe("engine setup and failure states", () => {
  beforeEach(() => {
    transport.engine.mockReset();
    transport.bootstrap.mockReset();
    engine.connected = false;
    engine.channels = [];
    engine.session = null;
    engine.error = "";
    engine.notice = "";
    engine.busy = false;
  });
  it("recognizes a reachable engine that explicitly has no NVR configuration", async () => {
    transport.bootstrap.mockResolvedValue({ networks: [] });
    transport.engine.mockImplementation(async (_: string, path: string) =>
      path === "/api/v1/local/status"
        ? { configured: false }
        : { paired: false },
    );
    await engine.refresh();
    expect(engine.connected).toBe(true);
    expect(engine.channels).toEqual([]);
    expect(engine.notice).toContain("Konfigurasikan NVR");
    expect(transport.engine).not.toHaveBeenCalledWith(
      "GET",
      "/api/v1/channels",
    );
  });
  it("retains service failures instead of pretending the engine has no cameras", async () => {
    transport.bootstrap.mockRejectedValue(Error("Engine tidak tersedia"));
    await engine.refresh();
    expect(engine.connected).toBe(false);
    expect(engine.error).toContain("tidak tersedia");
  });
  it("clears obsolete Stop controls when the engine snapshot reports no session", async () => {
    engine.connected = true;
    engine.session = { id: "previous", state: "FINALIZING" };
    transport.engine.mockImplementation(async (_: string, path: string) => {
      if (path === "/api/v1/local/status") return { configured: true, session: null };
      if (path === "/api/v1/channels") return [{ id: "g01" }];
      return { paired: false };
    });
    await engine.refresh();
    expect(engine.session).toBeNull();
    expect(engine.active).toBe(false);
    expect(engine.connected).toBe(true);
    expect(engine.channels).toHaveLength(1);
    expect(transport.engine).not.toHaveBeenCalledWith("GET", "/api/v1/media/sessions/current");
  });
  it("keeps genuine finalization locked until a terminal snapshot arrives", async () => {
    engine.connected = true;
    let state = "FINALIZING";
    transport.engine.mockImplementation(async (_: string, path: string) => {
      if (path === "/api/v1/local/status") return { configured: true, session: { id: "session", state } };
      if (path === "/api/v1/channels") return [{ id: "g01" }];
      return { paired: false };
    });
    await engine.refresh();
    expect(engine.active).toBe(true);
    state = "STOPPED";
    await engine.refresh();
    expect(engine.active).toBe(false);
  });
  it("does not treat malformed session objects as an active session", async () => {
    engine.connected = true;
    transport.engine.mockResolvedValue({ configured: true, session: {} });
    await engine.refresh();
    expect(engine.connected).toBe(false);
    expect(engine.error).toContain("sesi engine tidak valid");
    expect(engine.active).toBe(false);
  });
  it("does not overwrite a newly started session with an older polling response", async () => {
    engine.connected = true;
    engine.channels = [{ id: "g01" }];
    let resolveStatus!: (value: unknown) => void;
    transport.engine.mockImplementation(async (method: string, path: string) => {
      if (method === "POST") return { id: "new-session", state: "STARTING" };
      if (path === "/api/v1/local/status") return new Promise(resolve => { resolveStatus = resolve; });
      return [];
    });
    const poll = engine.refresh();
    await engine.start("new-session");
    resolveStatus({ configured: true, session: null });
    await poll;
    expect(engine.session?.id).toBe("new-session");
    expect(engine.active).toBe(true);
  });
});
