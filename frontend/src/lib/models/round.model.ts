export type ZoneId = "A" | "B" | "C" | "D";

export type RoundState = "IDLE" | "RUNNING" | "PAUSED" | "FINISHED";

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

export interface RoundConfig {
  eventName: string;
  sessionName: string;
  startTime: string;
  endTime: string;
  jurors: string[];
  totalDurationSeconds?: number;
  shiftDurationSeconds?: number;
  durationSeconds?: number;
  eventId?: string;
  roundId?: string;
  participants?: ParticipantMapping[];
}
