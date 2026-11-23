<script lang="ts">
  import { ZoomIn } from "@lucide/svelte";

  let { value, onChange }: { value: number; onChange: (value: number) => void } = $props();
</script>

<label class="mx-2 flex items-center gap-1.5 pb-0.5 text-zinc-300">
  <ZoomIn class="size-3 shrink-0" aria-hidden="true" />
  <input
    type="range"
    min="100"
    max="300"
    step="10"
    {value}
    oninput={(event) => onChange(event.currentTarget.valueAsNumber)}
    aria-label="Zoom kamera"
    aria-valuetext={`${value}%`}
    class="h-4 min-w-0 flex-1 cursor-pointer accent-sky-400"
  />
  <span class="w-8 shrink-0 text-right font-mono text-[9px]">{value}%</span>
</label>
