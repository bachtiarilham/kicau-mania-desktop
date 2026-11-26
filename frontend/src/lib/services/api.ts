import { Call, Events } from "@wailsio/runtime";
export type Row = Record<string, any>;
// A single adapter isolates the Wails transport from UI/domain state. The
// generated bindings are also shipped for typed integrations and inspection.
const call = <T = any>(method: string, ...args: any[]): Promise<T> =>
  Call.ByName(
    `kicau.desktop/cctv/internal/bindings.Desktop.${method}`,
    ...args,
  );
export const api = {
  bootstrap: () => call<Row>("Bootstrap"),
  config: () => call<Row>("GetConfig"),
  saveConfig: (config: Row) => call("SaveConfig", config),
  engine: <T = any>(method: string, path: string, body: Row | null = null) =>
    call<T>("Engine", method, path, body),
  cloud: <T = any>(path: string) => call<T>("Cloud", path),
  roundAction: <T = any>(
    arenaID: string,
    roundID: string,
    action: "start" | "begin-judging" | "close",
    expectedStatus: string,
  ) => call<T>("RoundAction", arenaID, roundID, action, expectedStatus),
  login: (email: string, password: string) =>
    call<Row>("Login", email, password),
  logout: () => call("Logout"),
  pair: (arena: string) => call<Row>("Pair", arena),
  revoke: (id: string) => call("Revoke", id),
  bind: (input: Row) => call<Row>("BindSession", input),
  sessionID: () => call<string>("NewSessionID"),
  player: (id: string, rect: Row, zoom: number) =>
    call("SyncPlayerWithZoom", id, rect, zoom),
  stopPlayer: (id: string) => call("StopPlayer", id),
  hidePlayers: () => call("HidePlayers"),
  mute: (id: string, muted: boolean) => call("Mute", id, muted),
  playRecording: (id: string, rect: Row) => call("PlayRecording", id, rect),
  export: (id: string, filename = "") => call<string>("ExportRecording", id, filename),
  selectFolder: () => call<string>("SelectRecordingsFolder"),
  openFolder: () => call("OpenRecordingsFolder"),
  saveSession: (sessionID: string, sessionName: string, targetDir: string, format: string) =>
    call<string[]>("SaveSessionRecordings", sessionID, sessionName, targetDir, format),
  on: (name: string, fn: (v: any) => void) =>
    Events.On(name, (event) => fn(event.data)),
};
export async function all(path: string): Promise<Row[]> {
  const rows: Row[] = [];
  let cursor = "";
  const seen = new Set<string>();
  for (let i = 0; i < 100; i++) {
    const page = await api.cloud<Row>(
      path +
        (path.includes("?") ? "&" : "?") +
        "limit=100&cursor=" +
        encodeURIComponent(cursor),
    );
    if (page?.items === null && page.has_more === false) return rows;
    if (!Array.isArray(page?.items) || typeof page.has_more !== "boolean")
      throw Error("Pagination cloud tidak valid");
    rows.push(...page.items);
    if (!page.has_more) return rows;
    if (!page.next_cursor || seen.has(page.next_cursor))
      throw Error("Cursor cloud tidak valid");
    cursor = page.next_cursor;
    seen.add(cursor);
  }
  throw Error("Terlalu banyak data; persempit pilihan");
}

export const message = (e: unknown) =>
  e instanceof Error ? e.message : String(e);
