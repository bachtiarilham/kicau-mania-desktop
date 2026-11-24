package usecase

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync/atomic"

	"kicau.desktop/cctv/internal/auth"
	authGW "kicau.desktop/cctv/internal/auth/gateway"
	"kicau.desktop/cctv/internal/config"
	"kicau.desktop/cctv/internal/engine"
)

type usecase struct {
	engine engine.LocalEngineGateway
	cloud  auth.CloudGateway
	config config.ConfigUsecase
	active atomic.Bool
}

func New(engine engine.LocalEngineGateway, cloud auth.CloudGateway, config config.ConfigUsecase) engine.EngineUsecase {
	return &usecase{
		engine: engine,
		cloud:  cloud,
		config: config,
	}
}

func (u *usecase) Bootstrap(ctx context.Context) (map[string]any, error) {
	return u.engine.Bootstrap(ctx)
}

func (u *usecase) Engine(ctx context.Context, method, path string, body map[string]any) (any, error) {
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
	err := u.engine.Call(ctx, method, path, body, &out)
	if err == nil && path == "/api/v1/local/status" {
		if status, ok := out.(map[string]any); ok {
			if status["session"] == nil {
				u.active.Store(false)
			} else if session, ok := status["session"].(map[string]any); ok {
				switch session["state"] {
				case "STARTING", "RUNNING", "DEGRADED", "STOPPING", "FINALIZING":
					u.active.Store(true)
				case "STOPPED", "INTERRUPTED", "FAILED":
					u.active.Store(false)
				}
			}
		}
	}
	if err == nil && strings.HasPrefix(path, "/api/v1/media/sessions") {
		if r, ok := out.(map[string]any); ok {
			state, _ := r["state"].(string)
			if state != "" {
				u.active.Store(state != "STOPPED" && state != "INTERRUPTED" && state != "FAILED")
			}
		} else if out == nil {
			u.active.Store(false)
		}
	}
	return out, err
}

func (u *usecase) Pair(ctx context.Context, arena string) (any, error) {
	id := authGW.ID()
	var current map[string]any
	if err := u.engine.Call(ctx, "GET", "/api/v1/desktop/cloud", nil, &current); err != nil {
		return nil, err
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
	if err := u.cloud.Call(ctx, "POST", "/media/engines/"+id, map[string]any{"arena_id": arena}, &paired); err != nil {
		return nil, err
	}
	cfg := u.config.GetConfig()
	err := u.engine.Call(ctx, "POST", "/api/v1/desktop/cloud", map[string]any{
		"base_url":  cfg.CloudURL,
		"engine_id": paired.ID,
		"arena_id":  paired.Arena,
		"token":     paired.Token,
	}, nil)
	if err != nil {
		_ = u.cloud.Call(ctx, "DELETE", "/media/engines/"+id, nil, nil)
		return nil, err
	}
	return map[string]any{"engine_id": paired.ID, "arena_id": paired.Arena}, nil
}

func (u *usecase) BindSession(ctx context.Context, in map[string]any) (any, error) {
	var out map[string]any
	if err := u.cloud.Call(ctx, "POST", "/media/ingestions", in, &out); err != nil {
		return nil, err
	}
	out["session_id"] = in["session_id"]
	if err := u.engine.Call(ctx, "POST", "/api/v1/desktop/mappings", out, nil); err != nil {
		return nil, err
	}
	var existing any
	sessionID, _ := in["session_id"].(string)
	out["existing_session"] = u.engine.Call(ctx, "GET", "/api/v1/media/sessions/"+url.PathEscape(sessionID), nil, &existing) == nil
	return out, nil
}

func (u *usecase) NewSessionID() string {
	return authGW.ID()
}

func (u *usecase) IsActive() bool {
	return u.active.Load()
}

func (u *usecase) SetActive(active bool) {
	u.active.Store(active)
}

func (u *usecase) LocalRecordings(ctx context.Context, cursor string) (any, error) {
	return u.Engine(ctx, "GET", "/api/v1/recordings?limit=25&cursor="+url.QueryEscape(cursor), nil)
}

func (u *usecase) SubscribeMetrics(ctx context.Context, onMetric func(any)) {
	u.engine.Metrics(ctx, onMetric)
}
