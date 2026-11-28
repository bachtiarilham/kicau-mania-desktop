const zones = ["A", "B", "C", "D"] as const;
export const arenaCameraID = "arena-overview";

export function gantanganCameraID(number: number) {
  return `gantangan${String(number).padStart(2, "0")}`;
}

export function gantanganChannels<T extends { id?: string }>(channels: readonly T[]) {
  const slots: (T | undefined)[] = Array(24).fill(undefined);
  const legacy: T[] = [];
  for (const channel of channels) {
    if (!channel.id || channel.id === arenaCameraID) continue;
    const match = /^gantangan(0[1-9]|1[0-9]|2[0-4])$/.exec(channel.id);
    if (match) slots[Number(match[1]) - 1] = channel;
    else legacy.push(channel);
  }
  return slots.map((channel) => channel ?? legacy.shift());
}

export function cameraSlots(channels: readonly { id?: string }[]) {
  const gantangan = gantanganChannels(channels);
  return zones.flatMap((zone, zoneIndex) =>
    Array.from({ length: 6 }, (_, index) => {
      const position = zoneIndex * 6 + index;
      const channel = gantangan[position]?.id;
      return { id: channel || `slot-${position + 1}`, number: position + 1, zone, attached: !!channel };
    }),
  );
}

export type CameraSlot = ReturnType<typeof cameraSlots>[number];
