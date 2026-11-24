<script lang="ts">
  import { Volume2 } from "@lucide/svelte";

  interface Props {
    metric?: Record<string, any>;
    compact?: boolean;
  }

  let { metric, compact = false }: Props = $props();
  const gradientID = $props.id();
  const floor = -65;
  const reference = -35;
  const level = (value: number) => Math.max(0, Math.min(1, (value - floor) / -floor));
  let peak = $derived(typeof metric?.peak_dbfs === "number" && Number.isFinite(metric.peak_dbfs) ? metric.peak_dbfs : null);
  let rms = $derived(typeof metric?.rms_dbfs === "number" && Number.isFinite(metric.rms_dbfs) ? metric.rms_dbfs : null);
  let history = $derived((metric?.history || []).slice(-50) as number[]);
  let points = $derived(history.map((value, index) =>
    `${8 + ((50 - history.length + index) / 49) * 504},${43 - level(value) * 38}`,
  ).join(" "));
  let area = $derived(history.length > 1
    ? `${8 + ((50 - history.length) / 49) * 504},43 ${points} 512,43`
    : "");
</script>

<section
  class="{compact ? 'mx-1.5 mb-1 px-2 py-1 rounded-lg' : 'mx-2 mb-2 px-3 pb-1.5 pt-2 rounded-xl'} border border-border bg-muted select-none"
  aria-label="Level suara burung"
>
  <div class="{compact ? 'mb-1' : 'mb-2'} flex min-w-0 items-center justify-between gap-1.5">
    <span class="flex items-center gap-1 text-[10px] font-medium text-muted-foreground">
      <Volume2 size={12} class="text-sky-400" />
      <span>Audio Burung</span>
    </span>
    <div class="flex items-center gap-1.5 font-mono">
      <span class="text-[9px] text-muted-foreground">RMS: {rms === null ? "—" : rms.toFixed(1)}</span>
      <span class="rounded bg-muted px-1 text-[10px] font-semibold text-emerald-400">
        {peak === null ? "— dBFS" : `${peak.toFixed(1)} dBFS`}
      </span>
    </div>
  </div>

  <div
    class="relative {compact ? 'mb-1 h-1' : 'mb-2 h-1.5'} overflow-hidden rounded-full bg-accent"
    role="img"
    aria-label={peak === null ? "Belum ada metrik audio" : `Level audio ${peak.toFixed(1)} dBFS`}
  >
    <div
      class="h-full rounded-full bg-gradient-to-r from-sky-500 via-emerald-400 to-amber-400 transition-[width] duration-150"
      style:width={`${peak === null ? 0 : level(peak) * 100}%`}
    ></div>
    <span class="absolute inset-y-0 w-0.5 bg-amber-400/70" style:left={`${level(reference) * 100}%`}></span>
  </div>

  <div class="relative overflow-hidden rounded border border-border bg-muted">
    <svg
      viewBox="0 0 520 46"
      preserveAspectRatio="none"
      class="block {compact ? 'h-7 sm:h-8' : 'h-12'} w-full"
      role="img"
      aria-label="Riwayat level audio, skala minus 65 sampai 0 dBFS; garis acuan minus 35 dBFS"
    >
      <defs>
        <linearGradient id={gradientID} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#0369a1" stop-opacity="0.5" />
          <stop offset="100%" stop-color="#0369a1" stop-opacity="0.08" />
        </linearGradient>
      </defs>
      {#if area}<polygon points={area} fill={`url(#${gradientID})`} />{/if}
      <line
        x1="0"
        x2="520"
        y1={43 - level(reference) * 38}
        y2={43 - level(reference) * 38}
        stroke="#92400e"
        stroke-opacity="0.55"
        stroke-dasharray="6 6"
        vector-effect="non-scaling-stroke"
      />
      <polyline
        {points}
        fill="none"
        stroke="#38bdf8"
        stroke-width="1.25"
        stroke-linejoin="round"
        vector-effect="non-scaling-stroke"
      />
    </svg>
    <span class="absolute right-1 top-0.5 font-mono text-[7px] text-amber-400/80" title="Garis acuan visual; tidak mengubah pengaturan deteksi engine">−35 dB</span>
    <span class="absolute bottom-0.5 left-1 font-mono text-[7px] text-muted-foreground">−65 ◂ Now</span>
    {#if peak === null}
      <span class="absolute inset-0 grid place-content-center text-[9px] text-muted-foreground">
        Menunggu audio kamera
      </span>
    {/if}
  </div>
</section>
