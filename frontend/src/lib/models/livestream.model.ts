export interface GantanganStream {
  id: string;
  gantanganNumber: number;
  code: string;
  title: string;
  channelName: string;
  arenaName: string;
  category: string;
  zone?: "A" | "B" | "C" | "D";
  roundLabel: string;
  viewers: number;
  isLive: boolean;
  quality: string;
  tags: string[];
  description: string;
  accentColor: string;
  previewGradient: string;
  rtspUrl?: string;
}

export function formatViewerCount(count: number): string {
  if (count >= 1000) {
    const value = (count / 1000).toFixed(1).replace(/\.0$/, "");
    return `${value}K penonton`;
  }
  return `${count} penonton`;
}
