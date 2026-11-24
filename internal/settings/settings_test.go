package settings_test

import (
	"path/filepath"
	"testing"

	"kicau.desktop/cctv/internal/model"
	"kicau.desktop/cctv/internal/settings"
)

func TestValidate_CloudURL(t *testing.T) {
	absVlc, _ := filepath.Abs("vlc/libvlc.dll")
	validCfg := model.Config{
		LocalURL:      "http://127.0.0.1:8090",
		RTSPBase:      "rtsp://127.0.0.1:8554",
		CloudURL:      "https://api.kicaumania.com/api/v1",
		VLC:           absVlc,
		LiveCacheMS:   50,
		RecordingsDir: "D:\\Recordings",
	}

	if err := settings.Validate(validCfg); err != nil {
		t.Fatalf("expected valid config to pass, got: %v", err)
	}

	// Local development loopback HTTP is also permitted
	localCfg := validCfg
	localCfg.CloudURL = "http://localhost:8091/api/v1"
	if err := settings.Validate(localCfg); err != nil {
		t.Fatalf("expected local loopback to pass, got: %v", err)
	}

	// Insecure non-loopback HTTP is rejected
	insecureCfg := validCfg
	insecureCfg.CloudURL = "http://api.kicaumania.com/api/v1"
	if err := settings.Validate(insecureCfg); err == nil {
		t.Fatal("expected insecure remote HTTP CloudURL to be rejected")
	}
}
