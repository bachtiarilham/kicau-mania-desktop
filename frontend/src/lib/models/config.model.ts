export interface DesktopConfig {
  local_url: string;
  rtsp_base: string;
  cloud_url: string;
  vlc: string;
  live_cache_ms: number;
  recordings_dir?: string;
}
