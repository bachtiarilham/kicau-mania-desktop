package domain

import (
	"context"
	"kicau.desktop/cctv/internal/model"
)

// PlayerDriver abstracts native video decoding and window embedding (libvlc + Win32 HWND).
type PlayerDriver interface {
	Sync(id, source, header string, r model.Rect) error
	SyncWithZoom(id, source, header string, r model.Rect, zoom int) error
	Stop(id string)
	Hide()
	Mute(id string, muted bool) error
	Shutdown()
}

// PlayerUsecase manages video player synchronizations, zoom, active session channel guards, and playback.
type PlayerUsecase interface {
	SyncPlayerWithZoom(ctx context.Context, id string, r model.Rect, zoom int) error
	StopPlayer(id string)
	HidePlayers()
	Mute(id string, muted bool) error
	PlayRecording(id string, r model.Rect) error
	Shutdown()
}
