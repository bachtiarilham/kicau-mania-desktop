package usecase

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"kicau.desktop/cctv/internal/config"
	"kicau.desktop/cctv/internal/engine"
	"kicau.desktop/cctv/internal/player"
	"kicau.desktop/cctv/internal/player/model"
)

type usecase struct {
	driver player.PlayerDriver
	engine engine.LocalEngineGateway
	config config.ConfigUsecase

	sessionMu     sync.Mutex
	cachedSession struct {
		State    string   `json:"state"`
		Channels []string `json:"channels"`
	}
	cachedSessionAt time.Time
}

func New(driver player.PlayerDriver, engine engine.LocalEngineGateway, config config.ConfigUsecase) player.PlayerUsecase {
	return &usecase{
		driver: driver,
		engine: engine,
		config: config,
	}
}

func (u *usecase) SyncPlayerWithZoom(ctx context.Context, id string, r model.Rect, zoom int) error {
	if zoom < 100 || zoom > 300 {
		return errors.New("Zoom harus antara 100% dan 300%")
	}
	if !player.ValidID(id) {
		return errors.New("ID kamera tidak valid")
	}

	u.sessionMu.Lock()
	session := u.cachedSession
	cached := time.Since(u.cachedSessionAt) < 1500*time.Millisecond
	u.sessionMu.Unlock()

	if !cached {
		var fetched struct {
			State    string   `json:"state"`
			Channels []string `json:"channels"`
		}
		if err := u.engine.Call(ctx, "GET", "/api/v1/media/sessions/current", nil, &fetched); err == nil {
			u.sessionMu.Lock()
			u.cachedSession = fetched
			u.cachedSessionAt = time.Now()
			session = fetched
			u.sessionMu.Unlock()
		}
	}

	if (session.State == "RUNNING" || session.State == "DEGRADED") && len(session.Channels) > 0 {
		found := false
		for _, v := range session.Channels {
			if v == id {
				found = true
				break
			}
		}
		if !found {
			return errors.New("Kamera bukan bagian sesi aktif")
		}
	}

	cfg := u.config.GetConfig()
	source := strings.TrimRight(cfg.RTSPBase, "/") + "/" + id
	return u.driver.SyncWithZoom(id, source, "", r, zoom)
}

func (u *usecase) StopPlayer(id string) {
	u.driver.Stop(id)
}

func (u *usecase) HidePlayers() {
	u.driver.Hide()
}

func (u *usecase) Mute(id string, muted bool) error {
	return u.driver.Mute(id, muted)
}

func (u *usecase) PlayRecording(id string, r model.Rect) error {
	if !player.ValidID(id) {
		return errors.New("ID rekaman tidak valid")
	}
	cfg := u.config.GetConfig()
	url := cfg.LocalURL + "/api/v1/recordings/" + id + "/file"
	header := "X-Local-Control: " + u.engine.Nonce()
	return u.driver.Sync(id, url, header, r)
}

func (u *usecase) Shutdown() {
	u.driver.Shutdown()
}
