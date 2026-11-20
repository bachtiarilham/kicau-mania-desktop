package bindings

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/wailsapp/wails/v3/pkg/application"
	"io"
	"kicau.desktop/cctv/internal/cloud"
	"kicau.desktop/cctv/internal/localapp"
	"kicau.desktop/cctv/internal/model"
	"kicau.desktop/cctv/internal/player"
	"kicau.desktop/cctv/internal/settings"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

type Desktop struct {
	active  atomic.Bool
	mu      sync.Mutex
	config  model.Config
	store   *settings.Store
	engine  *localapp.Client
	cloud   *cloud.Client
	players *player.Manager
	cancel  context.CancelFunc
	app     *application.App
	window  *application.WebviewWindow
	ctx     context.Context
}

func New() (*Desktop, error) {
	s, e := settings.New()
	if e != nil {
		return nil, e
	}
	cfg := settings.Default()
	if err := s.Load("desktop.json", &cfg, false); err != nil && !os.IsNotExist(err) {
		return nil, errors.New("Pengaturan desktop tidak dapat dibaca; periksa desktop.json pada LocalAppData")
	}
	if e = settings.Validate(cfg); e != nil {
		return nil, e
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Desktop{config: cfg, store: s, engine: localapp.New(cfg.LocalURL), cloud: cloud.New(cfg.CloudURL, s), ctx: ctx, cancel: cancel}, nil
}

// Attach is excluded from generated bindings because it is not an exported RPC.
func (d *Desktop) attach(a *application.App, w *application.WebviewWindow) {
	d.app = a
	d.window = w
	d.players = player.New(d.config.VLC, func() uintptr { return uintptr(w.NativeWindow()) }, func(s model.PlayerStatus) { a.Event.Emit("player:status", s) })
	d.players.LiveCacheMS = d.config.LiveCacheMS
	go d.engine.Metrics(d.ctx, func(v any) { a.Event.Emit("engine:metric", v) })
}
func (d *Desktop) GetConfig() model.Config { d.mu.Lock(); defer d.mu.Unlock(); return d.config }

// SaveConfig applies at next launch, preserving active engine sessions.
func (d *Desktop) SaveConfig(cfg model.Config) error {
	if e := settings.Validate(cfg); e != nil {
		return e
	}
	return d.store.Save("desktop.json", cfg, false)
}
func (d *Desktop) Bootstrap() (map[string]any, error) { return d.engine.Bootstrap(d.ctx) }
func (d *Desktop) Engine(method, path string, body map[string]any) (any, error) {
	if !strings.HasPrefix(path, "/api/v1/") || strings.Contains(path, "bootstrap") || strings.Contains(path, "/file") {
		return nil, errors.New("Endpoint engine tidak diizinkan")
	}
	allowed := false
	for _, p := range []string{"setup/", "media/sessions", "channels", "recordings", "desktop/", "local/status"} {
		allowed = allowed || strings.HasPrefix(strings.TrimPrefix(path, "/api/v1/"), p)
	}
	if !allowed || !(method == "GET" || method == "POST" || method == "DELETE") {
		return nil, errors.New("Operasi engine tidak diizinkan")
	}
	var out any
	e := d.engine.Call(d.ctx, method, path, body, &out)
	if e == nil && path == "/api/v1/local/status" {
		if status, ok := out.(map[string]any); ok {
			if status["session"] == nil {
				d.active.Store(false)
			} else if session, ok := status["session"].(map[string]any); ok {
				switch session["state"] {
				case "STARTING", "RUNNING", "DEGRADED", "STOPPING", "FINALIZING":
					d.active.Store(true)
				case "STOPPED", "INTERRUPTED", "FAILED":
					d.active.Store(false)
				}
			}
		}
	}
	if e == nil && strings.HasPrefix(path, "/api/v1/media/sessions") {
		if r, ok := out.(map[string]any); ok {
			state, _ := r["state"].(string)
			if state != "" {
				d.active.Store(state != "STOPPED" && state != "INTERRUPTED" && state != "FAILED")
			}
		} else if out == nil {
			d.active.Store(false)
		}
	}
	return out, e
}
func (d *Desktop) Login(email, password string) (map[string]any, error) {
	profile, e := d.cloud.Login(d.ctx, email, password)
	if e != nil {
		return nil, e
	}
	allowed := false
	if roles, ok := profile["roles"].([]any); ok {
		for _, role := range roles {
			allowed = allowed || role == "ADMIN" || role == "OPERATOR"
		}
	}
	if roles, ok := profile["arena_roles"].([]any); ok {
		for _, entry := range roles {
			if role, ok := entry.(map[string]any); ok {
				allowed = allowed || role["role"] == "ADMIN" || role["role"] == "OPERATOR"
			}
		}
	}
	if !allowed {
		_ = d.cloud.Logout(d.ctx)
		return nil, errors.New("Akses cloud desktop memerlukan akun ADMIN atau OPERATOR")
	}
	return profile, nil
}
func (d *Desktop) Logout() error { return d.cloud.Logout(d.ctx) }
func (d *Desktop) Cloud(path string) (any, error) {
	allowed := false
	for _, p := range []string{"/auth/me", "/arenas", "/registrations", "/media/recordings"} {
		allowed = allowed || strings.HasPrefix(path, p)
	}
	if !allowed {
		return nil, errors.New("Endpoint cloud tidak diizinkan")
	}
	var out any
	e := d.cloud.Call(d.ctx, "GET", path, nil, &out)
	return out, e
}
func (d *Desktop) Pair(arena string) (any, error) {
	id := cloud.ID()
	var current map[string]any
	if e := d.engine.Call(d.ctx, "GET", "/api/v1/desktop/cloud", nil, &current); e != nil {
		return nil, e
	}
	if current["paired"] == true {
		if current["arena_id"] != arena {
			return nil, errors.New("Engine ini sudah dipasangkan ke arena lain. Gunakan engine terpisah agar antrean lama tetap aman.")
		}
		if existing, ok := current["engine_id"].(string); ok && existing != "" {
			id = existing
		}
	}
	var paired struct {
		ID    string `json:"id"`
		Arena string `json:"arena_id"`
		Token string `json:"token"`
	}
	if e := d.cloud.Call(d.ctx, "POST", "/media/engines/"+id, map[string]any{"arena_id": arena}, &paired); e != nil {
		return nil, e
	}
	e := d.engine.Call(d.ctx, "POST", "/api/v1/desktop/cloud", map[string]any{"base_url": d.config.CloudURL, "engine_id": paired.ID, "arena_id": paired.Arena, "token": paired.Token}, nil)
	if e != nil {
		_ = d.cloud.Call(d.ctx, "DELETE", "/media/engines/"+id, nil, nil)
		return nil, e
	}
	return map[string]any{"engine_id": paired.ID, "arena_id": paired.Arena}, nil
}
func (d *Desktop) Revoke(engine string) error {
	return d.cloud.Call(d.ctx, "DELETE", "/media/engines/"+engine, nil, nil)
}
func (d *Desktop) BindSession(in map[string]any) (any, error) {
	var out map[string]any
	if e := d.cloud.Call(d.ctx, "POST", "/media/ingestions", in, &out); e != nil {
		return nil, e
	}
	out["session_id"] = in["session_id"]
	if e := d.engine.Call(d.ctx, "POST", "/api/v1/desktop/mappings", out, nil); e != nil {
		return nil, e
	}
	var existing any
	out["existing_session"] = d.engine.Call(d.ctx, "GET", "/api/v1/media/sessions/"+url.PathEscape(in["session_id"].(string)), nil, &existing) == nil
	return out, nil
}
func (d *Desktop) NewSessionID() string { return cloud.ID() }
func (d *Desktop) SyncPlayer(id string, r model.Rect) error {
	if !player.ValidID(id) {
		return errors.New("ID kamera tidak valid")
	}
	var session struct {
		State    string   `json:"state"`
		Channels []string `json:"channels"`
	}
	if e := d.engine.Call(d.ctx, "GET", "/api/v1/media/sessions/current", nil, &session); e != nil {
		errStr := strings.ToLower(e.Error())
		if strings.Contains(errStr, "method not allowed") || strings.Contains(errStr, "405") || strings.Contains(errStr, "404") {
			return errors.New("Kamera belum siap / sesi belum aktif")
		}
		return e
	}
	if session.State != "RUNNING" && session.State != "DEGRADED" {
		return errors.New("Mulai sesi sebelum menonton kamera")
	}
	found := false
	for _, v := range session.Channels {
		found = found || v == id
	}
	if !found {
		return errors.New("Kamera bukan bagian sesi aktif")
	}
	return d.players.Sync(id, strings.TrimRight(d.config.RTSPBase, "/")+"/"+id, "", r)
}
func (d *Desktop) StopPlayer(id string)             { d.players.Stop(id) }
func (d *Desktop) HidePlayers()                     { d.players.Hide() }
func (d *Desktop) Mute(id string, muted bool) error { return d.players.Mute(id, muted) }
func (d *Desktop) PlayRecording(id string, r model.Rect) error {
	if !player.ValidID(id) {
		return errors.New("ID rekaman tidak valid")
	}
	return d.players.Sync("recording-"+id, d.config.LocalURL+"/api/v1/recordings/"+id+"/file", "X-Local-Control: "+d.engine.Nonce(), r)
}
func (d *Desktop) ExportRecording(id string) (string, error) {
	if !player.ValidID(id) {
		return "", errors.New("ID rekaman tidak valid")
	}
	res, e := d.engine.File(d.ctx, id)
	if e != nil {
		return "", e
	}
	defer res.Body.Close()
	ext, label := ".mkv", "Video MKV"
	d.HidePlayers()
	dest, e := d.app.Dialog.SaveFile().SetFilename(id+ext).AddFilter(label, "*"+ext).AttachToWindow(d.window).PromptForSingleSelection()
	if e != nil || dest == "" {
		return "", e
	}
	f, e := os.CreateTemp(filepath.Dir(dest), ".export-*"+ext)
	if e != nil {
		return "", e
	}
	defer os.Remove(f.Name())
	_, e = io.Copy(f, res.Body)
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		return "", e
	}
	if e = os.Rename(f.Name(), dest); e != nil {
		return "", e
	}
	return dest, nil
}
func (d *Desktop) Preferences() string { b, _ := json.Marshal(d.config); return string(b) }
func (d *Desktop) LocalRecordings(cursor string) (any, error) {
	return d.Engine("GET", "/api/v1/recordings?limit=25&cursor="+url.QueryEscape(cursor), nil)
}
