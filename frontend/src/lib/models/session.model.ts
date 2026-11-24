export type SessionState =
  | "STARTING"
  | "RUNNING"
  | "DEGRADED"
  | "STOPPING"
  | "FINALIZING"
  | "STOPPED"
  | "INTERRUPTED"
  | "FAILED";

export interface EngineSession {
  id: string;
  state: SessionState;
  event_id?: string;
  round_id?: string;
  session_name?: string;
  created_at?: string;
  [key: string]: any;
}

export interface SaveSessionRequest {
  sessionID: string;
  sessionName: string;
  targetDir: string;
  format: string;
}
