const zones = ["A", "B", "C", "D"] as const;

export function cameraSlots(channels: readonly { id?: string }[]) {
  return zones.flatMap((zone, zoneIndex) =>
    Array.from({ length: 6 }, (_, index) => {
      const position = zoneIndex * 6 + index;
      const channel = channels[position]?.id;
      return { id: channel || `slot-${position + 1}`, number: position + 1, zone, attached: !!channel };
    }),
  );
}

export type CameraSlot = ReturnType<typeof cameraSlots>[number];
