package auth

import "context"

type CloudGateway interface {
	Login(ctx context.Context, email, password string) (map[string]any, error)
	Logout(ctx context.Context) error
	Call(ctx context.Context, method, path string, body any, out any) error
}

type AuthUsecase interface {
	Login(ctx context.Context, email, password string) (map[string]any, error)
	Logout(ctx context.Context) error
	Cloud(ctx context.Context, path string) (any, error)
	RoundAction(ctx context.Context, arenaID, roundID, action, expectedStatus string) (any, error)
	Revoke(ctx context.Context, engineID string) error
}
