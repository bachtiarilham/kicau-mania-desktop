package auth

import (
	"context"
	"errors"
	"strings"

	"kicau.desktop/cctv/internal/domain"
)

type usecase struct {
	cloud domain.CloudGateway
}

// New creates a new AuthUsecase implementation.
func New(cloud domain.CloudGateway) domain.AuthUsecase {
	return &usecase{
		cloud: cloud,
	}
}

func (u *usecase) Login(ctx context.Context, email, password string) (map[string]any, error) {
	profile, err := u.cloud.Login(ctx, email, password)
	if err != nil {
		return nil, err
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
		_ = u.cloud.Logout(ctx)
		return nil, errors.New("Akses cloud desktop memerlukan akun ADMIN atau OPERATOR")
	}
	return profile, nil
}

func (u *usecase) Logout(ctx context.Context) error {
	return u.cloud.Logout(ctx)
}

func (u *usecase) Cloud(ctx context.Context, path string) (any, error) {
	allowed := false
	for _, p := range []string{"/auth/me", "/arenas", "/registrations", "/media/recordings"} {
		allowed = allowed || strings.HasPrefix(path, p)
	}
	if !allowed {
		return nil, errors.New("Endpoint cloud tidak diizinkan")
	}
	var out any
	err := u.cloud.Call(ctx, "GET", path, nil, &out)
	return out, err
}

func (u *usecase) Revoke(ctx context.Context, engineID string) error {
	return u.cloud.Call(ctx, "DELETE", "/media/engines/"+engineID, nil, nil)
}
