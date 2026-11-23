package domain

import (
	"context"
)

// CloudGateway abstracts outbound HTTP communication to the cloud backend.
type CloudGateway interface {
	Login(ctx context.Context, email, password string) (map[string]any, error)
	Logout(ctx context.Context) error
	Call(ctx context.Context, method, path string, body any, out any) error
}

// AuthUsecase handles operator and admin authentication and cloud gateway access.
type AuthUsecase interface {
	Login(ctx context.Context, email, password string) (map[string]any, error)
	Logout(ctx context.Context) error
	Cloud(ctx context.Context, path string) (any, error)
	Revoke(ctx context.Context, engineID string) error
}
