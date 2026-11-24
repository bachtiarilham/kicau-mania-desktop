package player

import (
	"context"

	"kicau.desktop/cctv/internal/player/model"
)

type PlayerDriver interface {
	Sync(id, source, header string, r model.Rect) error
	SyncWithZoom(id, source, header string, r model.Rect, zoom int) error
	Stop(id string)
	Hide()
	Mute(id string, muted bool) error
	Shutdown()
}

type PlayerUsecase interface {
	SyncPlayerWithZoom(ctx context.Context, id string, r model.Rect, zoom int) error
	StopPlayer(id string)
	HidePlayers()
	Mute(id string, muted bool) error
	PlayRecording(id string, r model.Rect) error
	Shutdown()
}
