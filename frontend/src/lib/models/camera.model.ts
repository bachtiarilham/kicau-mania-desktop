export type { GantanganStream } from "./livestream.model";

export interface PlayerRect {
  x: number;
  y: number;
  width: number;
  height: number;
  clip_x?: number;
  clip_y?: number;
  clip_width?: number;
  clip_height?: number;
  visible: boolean;
  device_pixels?: boolean;
}

export interface PlayerStatus {
  channel_id: string;
  state: string;
  message?: string;
  codec?: string;
  fps: number;
  decoded: number;
  displayed: number;
  lost: number;
}
