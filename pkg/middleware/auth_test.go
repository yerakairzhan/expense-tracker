package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"finance-tracker/pkg/auth"
	"github.com/gin-gonic/gin"
)

type fakeTokenBlocklist struct {
	isRevokedFn func(ctx context.Context, tokenID string) (bool, error)
}

func (f *fakeTokenBlocklist) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	return f.isRevokedFn(ctx, tokenID)
}

func TestAccessTokenFromHeader(t *testing.T) {
	t.Run("rejects malformed header", func(t *testing.T) {
		// Act.
		_, err := AccessTokenFromHeader("Token abc")

		// Assert.
		if err == nil || err.Message != "invalid authorization header" {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("returns bearer token", func(t *testing.T) {
		// Act.
		got, err := AccessTokenFromHeader("Bearer abc")

		// Assert.
		if err != nil || got != "abc" {
			t.Fatalf("unexpected result: %q %#v", got, err)
		}
	})
}

func TestJWTAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("aborts revoked token", func(t *testing.T) {
		raw, err := auth.GenerateAccessToken("secret", 42, time.Now().UTC())
		if err != nil {
			t.Fatal(err)
		}
		router := gin.New()
		router.Use(JWTAuth("secret", &fakeTokenBlocklist{
			isRevokedFn: func(context.Context, string) (bool, error) { return true, nil },
		}))
		router.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })

		// Act.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+raw)
		router.ServeHTTP(rec, req)

		// Assert.
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("unexpected status: %d", rec.Code)
		}
	})

	t.Run("sets user id on valid token", func(t *testing.T) {
		raw, err := auth.GenerateAccessToken("secret", 42, time.Now().UTC())
		if err != nil {
			t.Fatal(err)
		}
		router := gin.New()
		router.Use(JWTAuth("secret", &fakeTokenBlocklist{
			isRevokedFn: func(context.Context, string) (bool, error) { return false, nil },
		}))
		router.GET("/protected", func(c *gin.Context) {
			userID, ok := UserIDFromContext(c)
			if !ok || userID != 42 {
				t.Fatalf("unexpected user id: %d %v", userID, ok)
			}
			c.Status(http.StatusOK)
		})

		// Act.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+raw)
		router.ServeHTTP(rec, req)

		// Assert.
		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("maps blocklist error to internal", func(t *testing.T) {
		raw, err := auth.GenerateAccessToken("secret", 42, time.Now().UTC())
		if err != nil {
			t.Fatal(err)
		}
		router := gin.New()
		router.Use(JWTAuth("secret", &fakeTokenBlocklist{
			isRevokedFn: func(context.Context, string) (bool, error) { return false, errors.New("boom") },
		}))
		router.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })

		// Act.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+raw)
		router.ServeHTTP(rec, req)

		// Assert.
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("unexpected status: %d", rec.Code)
		}
	})
}
