package vlc

import "time"

// Native LibVLC 3 media stats ABI: all fields are 32-bit, including floats.
type mediaStats struct {
	ReadBytes                            int32
	InputBitrate                         float32
	DemuxReadBytes                       int32
	DemuxBitrate                         float32
	DemuxCorrupted, DemuxDiscontinuity   int32
	DecodedVideo, DecodedAudio           int32
	DisplayedPictures, LostPictures      int32
	PlayedAudioBuffers, LostAudioBuffers int32
	SentPackets, SentBytes               int32
	SendBitrate                          float32
}

type frameMonitor struct {
	started, progressed time.Time
	count               int32
	seen                bool
}

func (m *frameMonitor) observe(now time.Time, count int32) (string, bool) {
	if m.started.IsZero() {
		m.started = now
		m.progressed = now
	}
	if count < m.count {
		m.count = count
		m.started = now
		m.progressed = now
		m.seen = count > 0
	}
	if count > m.count {
		m.count = count
		m.progressed = now
		m.seen = true
		return "PLAYING", false
	}
	if !m.seen {
		if now.Sub(m.started) >= 20*time.Second {
			return "RECONNECTING", true
		}
		return "WAITING_VIDEO", false
	}
	if now.Sub(m.progressed) >= 10*time.Second {
		return "RECONNECTING", true
	}
	if now.Sub(m.progressed) >= 3*time.Second {
		return "STALLED", false
	}
	return "PLAYING", false
}
