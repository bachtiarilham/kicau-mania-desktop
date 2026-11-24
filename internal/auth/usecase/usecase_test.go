package usecase_test

import (
	"context"
	"testing"

	authUC "kicau.desktop/cctv/internal/auth/usecase"
)

type mockCloud struct {
	loginRes  map[string]any
	loginErr  error
	loggedOut bool
}

func (m *mockCloud) Login(ctx context.Context, email, password string) (map[string]any, error) {
	return m.loginRes, m.loginErr
}

func (m *mockCloud) Logout(ctx context.Context) error {
	m.loggedOut = true
	return nil
}

func (m *mockCloud) Call(ctx context.Context, method, path string, body any, out any) error {
	return nil
}

func TestAuthUsecase_Login(t *testing.T) {
	ctx := context.Background()

	mock := &mockCloud{
		loginRes: map[string]any{"roles": []any{"MEMBER"}},
	}
	uc := authUC.New(mock)
	_, err := uc.Login(ctx, "test@example.com", "pass")
	if err == nil {
		t.Fatal("expected error for non-staff role, got nil")
	}
	if !mock.loggedOut {
		t.Fatal("expected logout on rejected role")
	}

	mockAdmin := &mockCloud{
		loginRes: map[string]any{"roles": []any{"ADMIN"}},
	}
	ucAdmin := authUC.New(mockAdmin)
	profile, err := ucAdmin.Login(ctx, "admin@example.com", "pass")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if profile == nil {
		t.Fatal("expected profile, got nil")
	}

	mockJudge := &mockCloud{
		loginRes: map[string]any{"roles": []any{"JUDGE"}},
	}
	ucJudge := authUC.New(mockJudge)
	profileJudge, err := ucJudge.Login(ctx, "judge@example.com", "pass")
	if err != nil {
		t.Fatalf("expected judge success, got error: %v", err)
	}
	if profileJudge == nil {
		t.Fatal("expected judge profile, got nil")
	}

	if _, err := ucAdmin.Cloud(ctx, "/dangerous/endpoint"); err == nil {
		t.Fatal("expected forbidden path error")
	}
	if _, err := ucAdmin.Cloud(ctx, "/auth/me"); err != nil {
		t.Fatalf("expected allowed /auth/me, got error: %v", err)
	}
}
