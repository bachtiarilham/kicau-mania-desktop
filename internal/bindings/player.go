package bindings

import (
	"errors"

	"kicau.desktop/cctv/internal/model"
)

func (d *Desktop) SyncPlayer(id string, r model.Rect) error {
	return d.SyncPlayerWithZoom(id, r, 100)
}

func (d *Desktop) SyncZoomPlayer(id string, r model.Rect) error {
	return d.SyncPlayerWithZoom(id, r, 200)
}

func (d *Desktop) SyncPlayerWithZoom(id string, r model.Rect, zoom int) error {
	if d.container.PlayerUC == nil {
		return errors.New("Player belum siap")
	}
	return d.container.PlayerUC.SyncPlayerWithZoom(d.ctx, id, r, zoom)
}

func (d *Desktop) StopPlayer(id string) {
	if d.container.PlayerUC != nil {
		d.container.PlayerUC.StopPlayer(id)
	}
}

func (d *Desktop) HidePlayers() {
	if d.container.PlayerUC != nil {
		d.container.PlayerUC.HidePlayers()
	}
}

func (d *Desktop) Mute(id string, muted bool) error {
	if d.container.PlayerUC == nil {
		return nil
	}
	return d.container.PlayerUC.Mute(id, muted)
}

func (d *Desktop) PlayRecording(id string, r model.Rect) error {
	if d.container.PlayerUC == nil {
		return errors.New("Player belum siap")
	}
	return d.container.PlayerUC.PlayRecording(id, r)
}
