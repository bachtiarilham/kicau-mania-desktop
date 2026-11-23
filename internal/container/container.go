package container

import (
	"context"
	"errors"
	"os"

	"kicau.desktop/cctv/internal/cloud"
	"kicau.desktop/cctv/internal/domain"
	"kicau.desktop/cctv/internal/localapp"
	"kicau.desktop/cctv/internal/model"
	"kicau.desktop/cctv/internal/player"
	"kicau.desktop/cctv/internal/settings"
	authUC "kicau.desktop/cctv/internal/usecase/auth"
	configUC "kicau.desktop/cctv/internal/usecase/config"
	engineUC "kicau.desktop/cctv/internal/usecase/engine"
	mediaUC "kicau.desktop/cctv/internal/usecase/media"
	recordingUC "kicau.desktop/cctv/internal/usecase/recording"
)

// Container manages the dependency injection lifecycle for the application.
type Container struct {
	Store   *settings.Store
	Engine  *localapp.Client
	Cloud   *cloud.Client
	Players *player.Manager

	ConfigUC    domain.ConfigUsecase
	AuthUC      domain.AuthUsecase
	EngineUC    domain.EngineUsecase
	PlayerUC    domain.PlayerUsecase
	RecordingUC domain.RecordingUsecase

	Ctx    context.Context
	Cancel context.CancelFunc
}

// New initializes configuration, repositories, and usecases.
func New() (*Container, error) {
	s, err := settings.New()
	if err != nil {
		return nil, err
	}
	cfg := settings.Default()
	if err := s.Load("desktop.json", &cfg, false); err != nil && !os.IsNotExist(err) {
		return nil, errors.New("Pengaturan desktop tidak dapat dibaca; periksa desktop.json pada LocalAppData")
	}
	if err = settings.Validate(cfg); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	engineClient := localapp.New(cfg.LocalURL)
	cloudClient := cloud.New(cfg.CloudURL, s)

	cfgUC := configUC.New(cfg, s)
	authUCase := authUC.New(cloudClient)
	engineUCase := engineUC.New(engineClient, cloudClient, cfgUC)
	recordingUCase := recordingUC.New(engineClient, cfgUC)

	c := &Container{
		Store:       s,
		Engine:      engineClient,
		Cloud:       cloudClient,
		ConfigUC:    cfgUC,
		AuthUC:      authUCase,
		EngineUC:    engineUCase,
		RecordingUC: recordingUCase,
		Ctx:         ctx,
		Cancel:      cancel,
	}
	return c, nil
}

// AttachPlayer initializes VLC player hardware driver and player usecase once window handle is available.
func (c *Container) AttachPlayer(windowHandle func() uintptr, onStatus func(model.PlayerStatus)) {
	cfg := c.ConfigUC.GetConfig()
	c.Players = player.New(cfg.VLC, windowHandle, onStatus)
	c.Players.LiveCacheMS = cfg.LiveCacheMS
	c.PlayerUC = mediaUC.New(c.Players, c.Engine, c.ConfigUC)
}
