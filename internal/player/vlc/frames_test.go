package vlc

import (
	"testing"
	"time"
	"unsafe"
)

func TestFrameMonitorDetectsFrozenVideoWhileVLCMayStillBePlaying(t *testing.T) {
	now := time.Now()
	m := frameMonitor{}
	if state, retry := m.observe(now, 0); state != "WAITING_VIDEO" || retry {
		t.Fatal(state, retry)
	}
	if state, _ := m.observe(now.Add(time.Second), 20); state != "PLAYING" {
		t.Fatal(state)
	}
	if state, retry := m.observe(now.Add(5*time.Second), 20); state != "STALLED" || retry {
		t.Fatal(state, retry)
	}
	if state, _ := m.observe(now.Add(6*time.Second), 100); state != "PLAYING" {
		t.Fatal("recovery ignored", state)
	}
	if state, retry := m.observe(now.Add(17*time.Second), 100); state != "RECONNECTING" || !retry {
		t.Fatal(state, retry)
	}
}
func TestFrameMonitorWaitsForFirstFrameAndHandlesCounterReset(t *testing.T) {
	now := time.Now()
	m := frameMonitor{}
	m.observe(now, 0)
	if state, retry := m.observe(now.Add(20*time.Second), 0); state != "RECONNECTING" || !retry {
		t.Fatal(state, retry)
	}
	m = frameMonitor{}
	m.observe(now, 200)
	if state, retry := m.observe(now.Add(time.Second), 0); state != "WAITING_VIDEO" || retry {
		t.Fatal("visibility/counter reset incorrectly treated as stall", state, retry)
	}
	if unsafe.Sizeof(mediaStats{}) != 60 {
		t.Fatal("LibVLC media stats ABI mismatch")
	}
}
