export interface RecordingItem {
  id: string;
  channel_id?: string;
  session_id?: string;
  status?: string;
  created_at?: string;
  size_bytes?: number;
  path?: string;
  filename?: string;
  [key: string]: any;
}

export interface ExportRecordingResult {
  savedPath: string;
  success: boolean;
  message?: string;
}
