export function uploadPercentage(state: string, uploaded: number, total: number): number {
  if (state === "UPLOADED") return 100;
  if (!Number.isFinite(uploaded) || !Number.isFinite(total) || total <= 0) return 0;
  return Math.max(0, Math.min(99, Math.floor((uploaded / total) * 100)));
}
