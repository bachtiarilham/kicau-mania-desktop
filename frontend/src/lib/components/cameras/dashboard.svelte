<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import {
    Camera,
    Maximize2,
    Volume2,
    VolumeX,
    ArrowLeft,
  } from "@lucide/svelte";
  import NativePlayer from "./native-player.svelte";
  import AudioMeter from "../metrics/audio-meter.svelte";
  import { api } from "$lib/services/api";
  import { engine } from "$lib/state/engine.svelte";
  const STORAGE_KEY = "kicau_dashboard_block";
  let columns = $state(3),
    block = $state(
      (typeof localStorage !== "undefined" &&
        localStorage.getItem(STORAGE_KEY)) ||
        "all",
    ),
    focus = $state(""),
    audio = $state("");

  $effect(() => {
    if (typeof localStorage !== "undefined") {
      localStorage.setItem(STORAGE_KEY, block);
    }
    if (block.startsWith("area_")) {
      columns = 3;
    }
  });

  let channels = $derived(
    engine.channels.filter((_, i) => {
      if (block === "area_a") return i >= 0 && i < 6;
      if (block === "area_b") return i >= 6 && i < 12;
      if (block === "area_c") return i >= 12 && i < 18;
      if (block === "area_d") return i >= 18 && i < 24;
      if (block === "first") return i < 12;
      if (block === "second") return i >= 12;
      return true;
    }),
  );
  async function sound(id: string) {
    const muted = audio === id;
    audio = muted ? "" : id;
    await api.mute(id, muted);
  }
</script>

<div class="section-head">
  <div>
    <div class="mb-2 text-[10px] tracking-[.2em] text-primary">
      ARENA SURVEILLANCE
    </div>
    <h1>{focus || "Pantau setiap gantangan."}</h1>
    <p class="subtitle">
      Live lokal, audio, dan kondisi kamera dalam satu layar.
    </p>
  </div>
  <div class="flex gap-2">
    {#if focus}<Button variant="outline" onclick={() => (focus = "")}
        ><ArrowLeft size={14} /> Semua kamera</Button
      >{:else}<select
        onfocus={() => (engine.overlay = true)}
        onblur={() => (engine.overlay = false)}
        onchange={() => (engine.overlay = false)}
        aria-label="Blok kamera"
        bind:value={block}
      >
        <option value="all">Semua gantangan (24 Kamera)</option>
        <option value="area_a">Layar 1: Area A (6 Kamera)</option>
        <option value="area_b">Layar 2: Area B (6 Kamera)</option>
        <option value="area_c">Layar 3: Area C (6 Kamera)</option>
        <option value="area_d">Layar 4: Area D (6 Kamera)</option>
        <option value="first">Gantangan 1–12</option>
        <option value="second">Gantangan 13–24</option>
      </select>
      <select
        onfocus={() => (engine.overlay = true)}
        onblur={() => (engine.overlay = false)}
        onchange={() => (engine.overlay = false)}
        aria-label="Kolom grid"
        bind:value={columns}
        >{#each [1, 2, 3, 4, 6] as n}<option value={n}>{n} kolom</option
          >{/each}</select
      >{/if}
  </div>
</div>
{#if !channels.length}<div class="empty">
    <Camera class="mx-auto mb-4 text-primary" size={36} />
    <h2>Belum ada kamera</h2>
    <p class="subtitle">
      Buka Konfigurasi NVR untuk mendaftarkan dan menguji sumber.
    </p>
  </div>{:else}<div
    class="grid gap-4"
    style:grid-template-columns={`repeat(${focus ? 1 : columns},minmax(0,1fr))`}
  >
    {#each channels as channel (channel.id)}<article
        class:hidden={!!focus && focus !== channel.id}
        class="overflow-hidden rounded-xl border border-border bg-card"
      >
        <header class="flex items-center justify-between gap-2 px-3 py-2.5">
          <div class="flex min-w-0 items-center gap-2">
            <span
              class="h-1.5 w-1.5 rounded-full"
              class:bg-primary={engine.playerStatus[channel.id]?.state ===
                "PLAYING"}
              class:bg-zinc-600={engine.playerStatus[channel.id]?.state !==
                "PLAYING"}
            ></span><span class="truncate text-xs font-semibold"
              >{channel.id}</span
            >
          </div>
          <div class="flex shrink-0 items-center gap-1">
            <span class="text-[9px] text-muted-foreground"
              >{engine.running
                ? engine.playerStatus[channel.id]?.state || "CONNECTING"
                : "STANDBY"}{#if engine.running && engine.playerStatus[channel.id]?.fps !== undefined}
                · {Number(engine.playerStatus[channel.id].fps).toFixed(1)} FPS{/if}</span
            ><Button
              variant="ghost"
              size="icon-xs"
              aria-label="Audio kamera"
              onclick={() => sound(channel.id)}
              >{#if audio === channel.id}<Volume2 size={13} />{:else}<VolumeX
                  size={13}
                />{/if}</Button
            ><Button
              variant="ghost"
              size="icon-xs"
              aria-label="Fokus kamera"
              onclick={() => (focus = focus === channel.id ? "" : channel.id)}
              ><Maximize2 size={13} /></Button
            >
          </div>
        </header>
        {#if engine.running}<NativePlayer
            id={channel.id}
            visible={!focus || focus === channel.id}
          />{:else}<div
            class="grid aspect-video place-content-center bg-[#090f10] text-xs text-muted-foreground"
          >
            Mulai sesi untuk melihat live
          </div>{/if}<AudioMeter metric={engine.metrics[channel.id]} />
      </article>{/each}
  </div>{/if}
