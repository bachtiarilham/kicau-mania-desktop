package container

import (
	"context"
	"errors"
	"os"

	"kicau.desktop/cctv/internal/auth"
	authGW "kicau.desktop/cctv/internal/auth/gateway"
	authUC "kicau.desktop/cctv/internal/auth/usecase"

	"kicau.desktop/cctv/internal/config"
	configRepo "kicau.desktop/cctv/internal/config/repository"
	configUC "kicau.desktop/cctv/internal/config/usecase"

	"kicau.desktop/cctv/internal/engine"
	engineGW "kicau.desktop/cctv/internal/engine/gateway"
	engineUC "kicau.desktop/cctv/internal/engine/usecase"

	"kicau.desktop/cctv/internal/player"
	playerModel "kicau.desktop/cctv/internal/player/model"
	playerUC "kicau.desktop/cctv/internal/player/usecase"

	"kicau.desktop/cctv/internal/recording"
	recordingUC "kicau.desktop/cctv/internal/recording/usecase"
)

type Container struct {
	Store   *configRepo.Store
	Engine  *engineGW.Client
	Cloud   *authGW.Client
	Players *player.Manager

	ConfigUC    config.ConfigUsecase
	AuthUC      auth.AuthUsecase
	EngineUC    engine.EngineUsecase
	PlayerUC    player.PlayerUsecase
	RecordingUC recording.RecordingUsecase

	Ctx    context.Context
	Cancel context.CancelFunc
}

func New() (*Container, error) {
	s, err := configRepo.New()
	if err != nil {
		return nil, err
	}
	cfg := configRepo.Default()
	if err := s.Load("desktop.json", &cfg, false); err != nil && !os.IsNotExist(err) {
		return nil, errors.New("Pengaturan desktop tidak dapat dibaca; periksa desktop.json pada LocalAppData")
	}
	if err = configRepo.Validate(cfg); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	engineClient := engineGW.New(cfg.LocalURL)
	cloudClient := authGW.New(cfg.CloudURL, s)

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

func (c *Container) AttachPlayer(windowHandle func() uintptr, onStatus func(playerModel.PlayerStatus)) {
	cfg := c.ConfigUC.GetConfig()
	c.Players = player.New(cfg.VLC, windowHandle, onStatus)
	c.Players.LiveCacheMS = cfg.LiveCacheMS
	c.PlayerUC = playerUC.New(c.Players, c.Engine, c.ConfigUC)
}
