package bindings_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kicau.desktop/cctv/internal/bindings"
	configModel "kicau.desktop/cctv/internal/config/model"
	configUC "kicau.desktop/cctv/internal/config/usecase"
	"kicau.desktop/cctv/internal/container"
	engineGW "kicau.desktop/cctv/internal/engine/gateway"
	engineUC "kicau.desktop/cctv/internal/engine/usecase"
)

type mockStore struct{}

func (m *mockStore) Load(name string, v any, secret bool) error { return nil }
func (m *mockStore) Save(name string, v any, secret bool) error { return nil }

type mockCloud struct{}

func (m *mockCloud) Login(ctx context.Context, email, password string) (map[string]any, error) {
	return nil, nil
}
func (m *mockCloud) Logout(ctx context.Context) error { return nil }
func (m *mockCloud) Call(ctx context.Context, method, path string, body any, out any) error {
	return nil
}

func TestSessionSnapshotClearsNativeStaleActivity(t *testing.T) {
	var session any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": map[string]any{"configured": true, "session": session}})
	}))
	defer server.Close()

	ctx := context.Background()
	engineClient := engineGW.New(server.URL)
	cfg := configModel.Config{}
	cfgUCase := configUC.New(cfg, &mockStore{})
	engineUCase := engineUC.New(engineClient, &mockCloud{}, cfgUCase)

	c := &container.Container{
		Engine:   engineClient,
		EngineUC: engineUCase,
		ConfigUC: cfgUCase,
		Ctx:      ctx,
		Cancel:   func() {},
	}
	d := bindings.NewWithContainer(c)
	d.SetActive(true)

	if _, e := d.Engine("GET", "/api/v1/local/status", nil); e != nil {
		t.Fatal(e)
	}
	if d.IsActive() {
		t.Fatal("empty snapshot retained active session")
	}
	session = map[string]any{"id": "one", "state": "FINALIZING"}
	if _, e := d.Engine("GET", "/api/v1/local/status", nil); e != nil {
		t.Fatal(e)
	}
	if !d.IsActive() {
		t.Fatal("genuine finalization lost")
	}
	session = map[string]any{"id": "one", "state": "STOPPED"}
	if _, e := d.Engine("GET", "/api/v1/local/status", nil); e != nil {
		t.Fatal(e)
	}
	if d.IsActive() {
		t.Fatal("terminal snapshot retained active session")
	}
}
