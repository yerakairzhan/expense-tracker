package service

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlc "finance-tracker/db/queries"
	"finance-tracker/pkg/auth"
	"finance-tracker/pkg/cache"
	"finance-tracker/pkg/models"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	t.Run("returns conflict on duplicate email", func(t *testing.T) {
		svc := &AuthService{
			users: &fakeAuthUserRepo{
				createFn: func(context.Context, string, string, string, string) (sqlc.User, error) {
					return sqlc.User{}, &pgconn.PgError{Code: "23505"}
				},
			},
			refreshStore:  &fakeRefreshStore{createFn: func(context.Context, string, int64, time.Duration) error { return nil }},
			blocklist:     &fakeBlocklist{revokeFn: func(context.Context, string, time.Duration) error { return nil }},
			jwtSecret:     "secret",
			refreshPepper: "pepper",
		}

		// Act.
		got, err := svc.Register(context.Background(), models.RegisterRequest{
			Email:    "john@example.com",
			Password: "password123",
			Name:     "John",
			Currency: "USD",
		})

		// Assert.
		if got != nil {
			t.Fatal("expected nil tokens")
		}
		if err == nil || err.Message != "email already exists" {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("register issues tokens", func(t *testing.T) {
		var gotHash string
		svc := &AuthService{
			users: &fakeAuthUserRepo{
				createFn: func(_ context.Context, email, passwordHash, name, currency string) (sqlc.User, error) {
					if email != "john@example.com" || name != "John" || currency != "USD" {
						t.Fatalf("unexpected create args: %s %s %s", email, name, currency)
					}
					if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("password123")) != nil {
						t.Fatal("expected bcrypt hash")
					}
					return testUserRow(passwordHash), nil
				},
			},
			refreshStore: &fakeRefreshStore{
				createFn: func(_ context.Context, tokenHash string, userID int64, ttl time.Duration) error {
					gotHash = tokenHash
					if userID != 42 || ttl != auth.RefreshTokenTTL {
						t.Fatalf("unexpected refresh store args: %d %s", userID, ttl)
					}
					return nil
				},
			},
			blocklist:     &fakeBlocklist{revokeFn: func(context.Context, string, time.Duration) error { return nil }},
			jwtSecret:     "secret",
			refreshPepper: "pepper",
		}

		// Act.
		got, err := svc.Register(context.Background(), models.RegisterRequest{
			Email:    "john@example.com",
			Password: "password123",
			Name:     "John",
			Currency: "USD",
		})

		// Assert.
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
		if got == nil || got.AccessToken == "" || got.RefreshToken == "" || gotHash == "" {
			t.Fatalf("unexpected tokens: %#v", got)
		}
	})
}

func TestAuthService_LoginRefreshLogout(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("login rejects invalid password", func(t *testing.T) {
		svc := &AuthService{
			users:         &fakeAuthUserRepo{getByEmailFn: func(context.Context, string) (sqlc.User, error) { return testUserRow(string(hash)), nil }},
			refreshStore:  &fakeRefreshStore{createFn: func(context.Context, string, int64, time.Duration) error { return nil }},
			blocklist:     &fakeBlocklist{revokeFn: func(context.Context, string, time.Duration) error { return nil }},
			jwtSecret:     "secret",
			refreshPepper: "pepper",
		}

		// Act.
		got, appErr := svc.Login(context.Background(), models.LoginRequest{
			Email:    "john@example.com",
			Password: "wrong-password",
		})

		// Assert.
		if got != nil {
			t.Fatal("expected nil tokens")
		}
		if appErr == nil || appErr.Message != "invalid credentials" {
			t.Fatalf("unexpected error: %#v", appErr)
		}
	})

	t.Run("refresh rotates session", func(t *testing.T) {
		var rotateCalled bool
		svc := &AuthService{
			users: &fakeAuthUserRepo{},
			refreshStore: &fakeRefreshStore{
				getFn: func(context.Context, string) (*cache.RefreshSession, error) {
					return &cache.RefreshSession{UserID: 42}, nil
				},
				rotateFn: func(_ context.Context, oldTokenHash, newTokenHash string, userID int64, ttl time.Duration) error {
					rotateCalled = true
					if oldTokenHash == "" || newTokenHash == "" || userID != 42 || ttl != auth.RefreshTokenTTL {
						t.Fatalf("unexpected rotate args: %q %q %d %s", oldTokenHash, newTokenHash, userID, ttl)
					}
					return nil
				},
			},
			blocklist:     &fakeBlocklist{revokeFn: func(context.Context, string, time.Duration) error { return nil }},
			jwtSecret:     "secret",
			refreshPepper: "pepper",
		}

		// Act.
		got, appErr := svc.Refresh(context.Background(), "refresh-token")

		// Assert.
		if appErr != nil {
			t.Fatalf("unexpected error: %#v", appErr)
		}
		if got == nil || got.AccessToken == "" || got.RefreshToken == "" || !rotateCalled {
			t.Fatalf("unexpected result: %#v rotate=%v", got, rotateCalled)
		}
	})

	t.Run("refresh rejects missing session", func(t *testing.T) {
		svc := &AuthService{
			users: &fakeAuthUserRepo{},
			refreshStore: &fakeRefreshStore{
				getFn: func(context.Context, string) (*cache.RefreshSession, error) { return nil, nil },
			},
			blocklist:     &fakeBlocklist{revokeFn: func(context.Context, string, time.Duration) error { return nil }},
			jwtSecret:     "secret",
			refreshPepper: "pepper",
		}

		// Act.
		got, appErr := svc.Refresh(context.Background(), "refresh-token")

		// Assert.
		if got != nil {
			t.Fatal("expected nil tokens")
		}
		if appErr == nil || appErr.Message != "invalid refresh token" {
			t.Fatalf("unexpected error: %#v", appErr)
		}
	})

	t.Run("logout validates token subject and revokes", func(t *testing.T) {
		accessToken, tokenErr := auth.GenerateAccessToken("secret", 42, time.Now().UTC())
		if tokenErr != nil {
			t.Fatal(tokenErr)
		}
		var revoked bool
		svc := &AuthService{
			users: &fakeAuthUserRepo{},
			refreshStore: &fakeRefreshStore{
				deleteFn: func(_ context.Context, tokenHash string) error {
					if tokenHash == "" {
						t.Fatal("expected hashed refresh token")
					}
					return nil
				},
			},
			blocklist: &fakeBlocklist{
				revokeFn: func(_ context.Context, tokenID string, ttl time.Duration) error {
					revoked = true
					if tokenID == "" || ttl <= 0 {
						t.Fatalf("unexpected revoke args: %q %s", tokenID, ttl)
					}
					return nil
				},
			},
			jwtSecret:     "secret",
			refreshPepper: "pepper",
		}

		// Act.
		appErr := svc.Logout(context.Background(), 42, "refresh-token", accessToken)

		// Assert.
		if appErr != nil {
			t.Fatalf("unexpected error: %#v", appErr)
		}
		if !revoked {
			t.Fatal("expected access token revoke")
		}
	})

	t.Run("logout returns internal on refresh delete failure", func(t *testing.T) {
		accessToken, tokenErr := auth.GenerateAccessToken("secret", 42, time.Now().UTC())
		if tokenErr != nil {
			t.Fatal(tokenErr)
		}
		svc := &AuthService{
			users: &fakeAuthUserRepo{},
			refreshStore: &fakeRefreshStore{
				deleteFn: func(context.Context, string) error { return errors.New("boom") },
			},
			blocklist:     &fakeBlocklist{revokeFn: func(context.Context, string, time.Duration) error { return nil }},
			jwtSecret:     "secret",
			refreshPepper: "pepper",
		}

		// Act.
		appErr := svc.Logout(context.Background(), 42, "refresh-token", accessToken)

		// Assert.
		if appErr == nil || appErr.Message != "failed to revoke refresh token" {
			t.Fatalf("unexpected error: %#v", appErr)
		}
	})
}
