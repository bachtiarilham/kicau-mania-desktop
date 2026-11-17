package bindings

import (
	"context"
	"encoding/json"
	"kicau.desktop/cctv/internal/localapp"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionSnapshotClearsNativeStaleActivity(t *testing.T) {
	var session any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": map[string]any{"configured": true, "session": session}})
	}))
	defer server.Close()
	d := &Desktop{ctx: context.Background(), engine: localapp.New(server.URL)}
	d.active.Store(true)
	if _, e := d.Engine("GET", "/api/v1/local/status", nil); e != nil {
		t.Fatal(e)
	}
	if d.active.Load() {
		t.Fatal("empty snapshot retained active session")
	}
	session = map[string]any{"id": "one", "state": "FINALIZING"}
	if _, e := d.Engine("GET", "/api/v1/local/status", nil); e != nil {
		t.Fatal(e)
	}
	if !d.active.Load() {
		t.Fatal("genuine finalization lost")
	}
	session = map[string]any{"id": "one", "state": "STOPPED"}
	if _, e := d.Engine("GET", "/api/v1/local/status", nil); e != nil {
		t.Fatal(e)
	}
	if d.active.Load() {
		t.Fatal("terminal snapshot retained active session")
	}
}
