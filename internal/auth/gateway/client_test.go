package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kicau.desktop/cctv/internal/config/repository"
)

func TestOpaqueSessionRefreshLogoutAndOriginIsolation(t *testing.T) {
	refreshes := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := any(map[string]any{"id": "user", "name": "Petugas"})
		if r.URL.Path == "/api/v1/auth/login" || r.URL.Path == "/api/v1/auth/refresh" {
			if r.Header.Get("Origin") == "" {
				t.Error("native origin missing")
			}
			if r.URL.Path == "/api/v1/auth/refresh" {
				cookie, e := r.Cookie("__Host-kicau_refresh")
				if e != nil || cookie.Value != "refresh" {
					t.Error("refresh cookie missing")
				}
				refreshes++
			}
			http.SetCookie(w, &http.Cookie{Name: "__Host-kicau_refresh", Value: "refresh", Path: "/", Secure: true, HttpOnly: true})
			data = map[string]any{"access_token": "opaque-access", "expires_in": 900, "user": map[string]any{"id": "user", "name": "Petugas"}}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "message": "OK", "data": data})
	}))
	defer server.Close()
	store := &repository.Store{Dir: t.TempDir()}
	client := New(server.URL+"/api/v1", store)
	client.http = server.Client()
	user, e := client.Login(context.Background(), "user@test.test", "password")
	if e != nil {
		t.Fatal(e)
	}
	if user["access_token"] != nil {
		t.Fatal("token leaked as profile")
	}
	client.session.Expires = time.Now().Add(-time.Minute)
	var profile map[string]any
	if e = client.Call(context.Background(), "GET", "/auth/me", nil, &profile); e != nil || refreshes != 1 {
		t.Fatal("refresh failed", e)
	}
	isolated := New("https://another.example.test/api/v1", store)
	if isolated.session.Refresh != "" || isolated.session.Access != "" {
		t.Fatal("credentials crossed backend origin")
	}
	if e = client.Logout(context.Background()); e != nil {
		t.Fatal(e)
	}
	if client.session.Refresh != "" {
		t.Fatal("logout retained refresh")
	}
}
