package usecase

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"kicau.desktop/cctv/internal/config"
	"kicau.desktop/cctv/internal/config/model"
	"kicau.desktop/cctv/internal/config/repository"
)

type usecase struct {
	mu     sync.RWMutex
	config model.Config
	store  config.ConfigStore
}

func New(cfg model.Config, store config.ConfigStore) config.ConfigUsecase {
	return &usecase{
		config: cfg,
		store:  store,
	}
}

func (u *usecase) GetConfig() model.Config {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.config
}

func (u *usecase) SaveConfig(cfg model.Config) error {
	if err := repository.Validate(cfg); err != nil {
		return err
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	if err := u.store.Save("desktop.json", cfg, false); err != nil {
		return err
	}
	u.config = cfg
	return nil
}

func (u *usecase) GetRecordingsDir() string {
	u.mu.RLock()
	dir := u.config.RecordingsDir
	u.mu.RUnlock()

	if dir != "" {
		return dir
	}
	userProfile := os.Getenv("USERPROFILE")
	if userProfile != "" {
		return filepath.Join(userProfile, "Videos", "KicauMania")
	}
	exe, _ := os.Executable()
	return filepath.Join(filepath.Dir(exe), "recordings")
}

func (u *usecase) SetRecordingsDir(dir string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.config.RecordingsDir = dir
	return u.store.Save("desktop.json", u.config, false)
}

func (u *usecase) OpenRecordingsFolder() error {
	dir := u.GetRecordingsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return exec.Command("explorer.exe", dir).Start()
}

func (u *usecase) Preferences() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	b, _ := json.Marshal(u.config)
	return string(b)
}
