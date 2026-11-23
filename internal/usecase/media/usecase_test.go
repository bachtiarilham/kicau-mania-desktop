package media_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"kicau.desktop/cctv/internal/model"
	mediaUC "kicau.desktop/cctv/internal/usecase/media"
)

type mockPlayerDriver struct {
	syncedWithZoom bool
	zoomLevel      int
	muted          bool
	stopped        bool
	hidden         bool
}

func (m *mockPlayerDriver) Sync(id, source, header string, r model.Rect) error { return nil }
func (m *mockPlayerDriver) SyncWithZoom(id, source, header string, r model.Rect, zoom int) error {
	m.syncedWithZoom = true
	m.zoomLevel = zoom
	return nil
}
func (m *mockPlayerDriver) Stop(id string)             { m.stopped = true }
func (m *mockPlayerDriver) Hide()                     { m.hidden = true }
func (m *mockPlayerDriver) Mute(id string, muted bool) error {
	m.muted = muted
	return nil
}
func (m *mockPlayerDriver) Shutdown() {}

type mockEngine struct{}

func (m *mockEngine) Bootstrap(ctx context.Context) (map[string]any, error) { return nil, nil }
func (m *mockEngine) Call(ctx context.Context, method, path string, body any, out any) error {
	return nil
}
func (m *mockEngine) File(ctx context.Context, id string) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("dummy")),
	}, nil
}
func (m *mockEngine) Metrics(ctx context.Context, onMetric func(any)) {}
func (m *mockEngine) Nonce() string                                   { return "test-nonce" }

type mockConfig struct {
	cfg model.Config
}

func (m *mockConfig) GetConfig() model.Config               { return m.cfg }
func (m *mockConfig) SaveConfig(cfg model.Config) error     { return nil }
func (m *mockConfig) GetRecordingsDir() string              { return "" }
func (m *mockConfig) SetRecordingsDir(dir string) error     { return nil }
func (m *mockConfig) OpenRecordingsFolder() error           { return nil }
func (m *mockConfig) Preferences() string                   { return "" }

func TestPlayerUsecase_ZoomValidation(t *testing.T) {
	driver := &mockPlayerDriver{}
	engine := &mockEngine{}
	cfg := &mockConfig{cfg: model.Config{RTSPBase: "rtsp://127.0.0.1:8554"}}
	uc := mediaUC.New(driver, engine, cfg)

	ctx := context.Background()
	r := model.Rect{Width: 640, Height: 480}

	// 1. Zoom < 100 should fail
	if err := uc.SyncPlayerWithZoom(ctx, "cam01", r, 50); err == nil {
		t.Fatal("expected error for zoom < 100")
	}

	// 2. Zoom > 300 should fail
	if err := uc.SyncPlayerWithZoom(ctx, "cam01", r, 350); err == nil {
		t.Fatal("expected error for zoom > 300")
	}

	// 3. Valid zoom (200)
	if err := uc.SyncPlayerWithZoom(ctx, "cam01", r, 200); err != nil {
		t.Fatalf("unexpected error for valid zoom: %v", err)
	}
	if !driver.syncedWithZoom || driver.zoomLevel != 200 {
		t.Fatalf("driver was not synced with zoom 200: %v, level: %d", driver.syncedWithZoom, driver.zoomLevel)
	}

	// 4. Mute and Stop
	if err := uc.Mute("cam01", true); err != nil {
		t.Fatal(err)
	}
	if !driver.muted {
		t.Fatal("expected driver muted to be true")
	}

	uc.StopPlayer("cam01")
	if !driver.stopped {
		t.Fatal("expected driver stopped to be true")
	}
}
