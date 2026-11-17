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
  login: (email: string, password: string) =>
    call<Row>("Login", email, password),
  logout: () => call("Logout"),
  pair: (arena: string) => call<Row>("Pair", arena),
  revoke: (id: string) => call("Revoke", id),
  bind: (input: Row) => call<Row>("BindSession", input),
  sessionID: () => call<string>("NewSessionID"),
  player: (id: string, rect: Row) => call("SyncPlayer", id, rect),
  stopPlayer: (id: string) => call("StopPlayer", id),
  hidePlayers: () => call("HidePlayers"),
  mute: (id: string, muted: boolean) => call("Mute", id, muted),
  playRecording: (id: string, rect: Row) => call("PlayRecording", id, rect),
  export: (id: string) => call<string>("ExportRecording", id),
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
    if (!Array.isArray(page.items) || typeof page.has_more !== "boolean")
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
export async function job(
  input: Row,
  onUpdate: (v: Row) => void,
  signal?: AbortSignal,
): Promise<any> {
  let value = await api.engine<Row>("POST", "/api/v1/setup/jobs", input);
  onUpdate(value);
  while (value.state === "RUNNING") {
    await new Promise<void>((resolve, reject) => {
      const timer = setTimeout(resolve, 1000);
      signal?.addEventListener(
        "abort",
        () => {
          clearTimeout(timer);
          reject(Error("Pemantauan dihentikan"));
        },
        { once: true },
      );
    });
    value = await api.engine<Row>("GET", `/api/v1/setup/jobs/${value.id}`);
    onUpdate(value);
  }
  if (value.state !== "SUCCEEDED")
    throw Error(
      [value.message, value.advice].filter(Boolean).join(". ") ||
        "Pekerjaan gagal",
    );
  return value.result;
}
export const message = (e: unknown) =>
  e instanceof Error ? e.message : String(e);
