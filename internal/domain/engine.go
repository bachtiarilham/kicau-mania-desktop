package domain

import (
	"context"
	"net/http"
)

// LocalEngineGateway abstracts HTTP communication to the local-app RTSP/recording engine.
type LocalEngineGateway interface {
	Bootstrap(ctx context.Context) (map[string]any, error)
	Call(ctx context.Context, method, path string, body any, out any) error
	File(ctx context.Context, id string) (*http.Response, error)
	Metrics(ctx context.Context, onMetric func(any))
	Nonce() string
}

// EngineUsecase coordinates communication with local engine, active session lifecycle, and pairing.
type EngineUsecase interface {
	Bootstrap(ctx context.Context) (map[string]any, error)
	Engine(ctx context.Context, method, path string, body map[string]any) (any, error)
	Pair(ctx context.Context, arena string) (any, error)
	BindSession(ctx context.Context, in map[string]any) (any, error)
	NewSessionID() string
	IsActive() bool
	SetActive(active bool)
	LocalRecordings(ctx context.Context, cursor string) (any, error)
	SubscribeMetrics(ctx context.Context, onMetric func(any))
}
