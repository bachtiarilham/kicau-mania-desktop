package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBootstrapKeepsNonceInGoAndForwardsItOnlyToEngine(t *testing.T) {
	nonce := strings.Repeat("a", 64)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := any(nil)
		if r.URL.Path == "/api/v1/setup/bootstrap" {
			data = map[string]any{"nonce": nonce, "networks": []any{}}
		} else if r.Header.Get("X-Local-Control") != nonce {
			t.Error("local nonce missing")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": data})
	}))
	defer server.Close()
	c := New(server.URL)
	b, e := c.Bootstrap(context.Background())
	if e != nil || b["nonce"] != nil {
		t.Fatal("unsafe bootstrap", e)
	}
	if e = c.Call(context.Background(), "POST", "/api/v1/media/sessions", map[string]any{}, nil); e != nil {
		t.Fatal(e)
	}
	if e = c.Call(context.Background(), "GET", "https://evil.test/", nil, nil); e == nil {
		t.Fatal("foreign endpoint allowed")
	}
}
