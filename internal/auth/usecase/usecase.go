package usecase

import (
	"context"
	"errors"
	"strings"

	"kicau.desktop/cctv/internal/auth"
)

type usecase struct {
	cloud auth.CloudGateway
}

func New(cloud auth.CloudGateway) auth.AuthUsecase {
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
	isRoleAllowed := func(role string) bool {
		r := strings.ToUpper(role)
		return r == "ADMIN" || r == "OPERATOR" || r == "JUDGE" || r == "JURI"
	}
	if scope, ok := profile["active_scope"].(map[string]any); ok {
		role, _ := scope["role"].(string)
		if !isRoleAllowed(role) {
			_ = u.cloud.Logout(ctx)
			return nil, errors.New("Scope cloud aktif tidak diizinkan untuk desktop")
		}
		return profile, nil
	}
	if roles, ok := profile["roles"].([]any); ok {
		for _, role := range roles {
			if rStr, ok := role.(string); ok && isRoleAllowed(rStr) {
				allowed = true
				break
			}
		}
	}
	if !allowed {
		if roles, ok := profile["arena_roles"].([]any); ok {
			for _, entry := range roles {
				if roleMap, ok := entry.(map[string]any); ok {
					if rStr, ok := roleMap["role"].(string); ok && isRoleAllowed(rStr) {
						allowed = true
						break
					}
				}
			}
		}
	}
	if !allowed {
		_ = u.cloud.Logout(ctx)
		return nil, errors.New("Akses cloud desktop memerlukan akun ADMIN, OPERATOR, atau JURI/JUDGE")
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
