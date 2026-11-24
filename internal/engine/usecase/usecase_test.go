package usecase_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configModel "kicau.desktop/cctv/internal/config/model"
	"kicau.desktop/cctv/internal/engine/gateway"
	engineUC "kicau.desktop/cctv/internal/engine/usecase"
)

type mockCloud struct{}

func (m *mockCloud) Login(ctx context.Context, email, password string) (map[string]any, error) {
	return nil, nil
}
func (m *mockCloud) Logout(ctx context.Context) error { return nil }
func (m *mockCloud) Call(ctx context.Context, method, path string, body any, out any) error {
	return nil
}

type mockConfig struct{}

func (m *mockConfig) GetConfig() configModel.Config          { return configModel.Config{} }
func (m *mockConfig) SaveConfig(cfg configModel.Config) error { return nil }
func (m *mockConfig) GetRecordingsDir() string               { return "" }
func (m *mockConfig) SetRecordingsDir(dir string) error      { return nil }
func (m *mockConfig) OpenRecordingsFolder() error            { return nil }
func (m *mockConfig) Preferences() string                    { return "" }

func TestEngineUsecase_SessionStateTracking(t *testing.T) {
	var session any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": map[string]any{"configured": true, "session": session}})
	}))
	defer server.Close()

	ctx := context.Background()
	engineGateway := gateway.New(server.URL)
	uc := engineUC.New(engineGateway, &mockCloud{}, &mockConfig{})

	uc.SetActive(true)

	if _, err := uc.Engine(ctx, "GET", "/api/v1/local/status", nil); err != nil {
		t.Fatal(err)
	}
	if uc.IsActive() {
		t.Fatal("expected active to be false after nil session")
	}

	session = map[string]any{"id": "s1", "state": "FINALIZING"}
	if _, err := uc.Engine(ctx, "GET", "/api/v1/local/status", nil); err != nil {
		t.Fatal(err)
	}
	if !uc.IsActive() {
		t.Fatal("expected active to be true for FINALIZING session")
	}

	session = map[string]any{"id": "s1", "state": "STOPPED"}
	if _, err := uc.Engine(ctx, "GET", "/api/v1/local/status", nil); err != nil {
		t.Fatal(err)
	}
	if uc.IsActive() {
		t.Fatal("expected active to be false for STOPPED session")
	}
}
