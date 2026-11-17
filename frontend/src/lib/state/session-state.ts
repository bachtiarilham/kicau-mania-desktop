import type { Row } from "$lib/services/api";

const states = new Set([
  "STARTING", "RUNNING", "DEGRADED", "STOPPING", "FINALIZING",
  "STOPPED", "INTERRUPTED", "FAILED",
]);
const activeStates = new Set(["STARTING", "RUNNING", "DEGRADED", "STOPPING", "FINALIZING"]);

export function readSession(value: unknown): Row | null {
  if (value == null) return null;
  if (typeof value !== "object" || Array.isArray(value))
    throw Error("Respons sesi engine tidak valid; hubungkan ulang engine.");
  const row = value as Row;
  if (typeof row.id !== "string" || !row.id || !states.has(row.state))
    throw Error("Identitas atau status sesi engine tidak valid; hubungkan ulang engine.");
  return row;
}

export function isActiveSession(value: Row | null): boolean {
  return !!value?.id && activeStates.has(value.state);
}
