package usecase_test

import (
	"testing"

	"kicau.desktop/cctv/internal/config/model"
	configUC "kicau.desktop/cctv/internal/config/usecase"
)

type mockStore struct {
	data map[string]any
}

func (m *mockStore) Load(name string, v any, secret bool) error { return nil }
func (m *mockStore) Save(name string, v any, secret bool) error {
	m.data[name] = v
	return nil
}

func TestConfigUsecase(t *testing.T) {
	store := &mockStore{data: make(map[string]any)}
	initialCfg := model.Config{
		LocalURL:      "http://127.0.0.1:8090",
		RTSPBase:      "rtsp://127.0.0.1:8554",
		CloudURL:      "http://127.0.0.1:8080",
		VLC:           "C:\\Program Files\\VideoLAN\\VLC",
		LiveCacheMS:   50,
		RecordingsDir: "C:\\Videos",
	}
	uc := configUC.New(initialCfg, store)

	cfg := uc.GetConfig()
	if cfg.LocalURL != initialCfg.LocalURL {
		t.Fatalf("expected %s, got %s", initialCfg.LocalURL, cfg.LocalURL)
	}

	if dir := uc.GetRecordingsDir(); dir != "C:\\Videos" {
		t.Fatalf("expected C:\\Videos, got %s", dir)
	}

	if err := uc.SetRecordingsDir("D:\\NewRecordings"); err != nil {
		t.Fatal(err)
	}
	if dir := uc.GetRecordingsDir(); dir != "D:\\NewRecordings" {
		t.Fatalf("expected updated dir D:\\NewRecordings, got %s", dir)
	}

	pref := uc.Preferences()
	if pref == "" {
		t.Fatal("preferences string should not be empty")
	}
}
