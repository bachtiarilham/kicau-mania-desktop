<script lang="ts">
  import { roundStore } from "$lib/state/round.svelte";

  let { number, class: className = "" }: {
    number: number;
    class?: string;
  } = $props();

  const participant = $derived(roundStore.getParticipantByGantangan(number));
  const code = $derived(String(number).padStart(2, "0"));
  const storedBird = $derived(participant?.birdName?.trim());
  const storedOwner = $derived(participant?.ownerName?.trim());
  const bird = $derived(storedBird === `Gantangan ${number}` ? "" : storedBird);
  const owner = $derived(storedOwner === "Pemilik" ? "" : storedOwner);
  const label = $derived(bird || `Gantangan ${code}`);
</script>

<div class="flex min-w-0 items-center gap-1.5 {className}">
  {#if bird || owner}
    <span class="shrink-0 rounded bg-primary px-1 font-mono text-[10px] font-bold text-primary-foreground">G{code}</span>
  {/if}
  <span class="truncate text-foreground" title={owner ? `${label} (${owner})` : label}>
    {label}{#if owner} <span class="font-normal text-muted-foreground">({owner})</span>{/if}
  </span>
</div>
