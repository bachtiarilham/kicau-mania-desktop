export type ZoneId = "A" | "B" | "C" | "D";

export type RoundState = "IDLE" | "RUNNING" | "FINISHED";

export interface ParticipantMapping {
  registrationId: string;
  gantanganId: string;
  gantanganNumber: number;
  birdName: string;
  birdCode?: string;
  ownerName: string;
  ownerPhone?: string;
  channelId?: string;
}

export interface JudgeZoneScheduleItem {
  shift: number;
  zone: ZoneId;
  range: string;
  startClock: string;
  endClock: string;
}
