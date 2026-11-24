package config

import (
	"kicau.desktop/cctv/internal/config/model"
)

type ConfigStore interface {
	Load(name string, v any, secret bool) error
	Save(name string, v any, secret bool) error
}

type ConfigUsecase interface {
	GetConfig() model.Config
	SaveConfig(cfg model.Config) error
	GetRecordingsDir() string
	SetRecordingsDir(dir string) error
	OpenRecordingsFolder() error
	Preferences() string
}
