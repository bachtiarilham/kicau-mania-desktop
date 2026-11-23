package domain

import (
	"kicau.desktop/cctv/internal/model"
)

// ConfigStore defines persistence operations for desktop configuration.
type ConfigStore interface {
	Load(name string, v any, secret bool) error
	Save(name string, v any, secret bool) error
}

// ConfigUsecase encapsulates business logic for desktop preferences and directories.
type ConfigUsecase interface {
	GetConfig() model.Config
	SaveConfig(cfg model.Config) error
	GetRecordingsDir() string
	SetRecordingsDir(dir string) error
	OpenRecordingsFolder() error
	Preferences() string
}
