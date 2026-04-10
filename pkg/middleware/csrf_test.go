package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDoubleSubmitCSRF(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("bypasses when disabled", func(t *testing.T) {
		router := gin.New()
		router.Use(DoubleSubmitCSRF(false))
		router.POST("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		// Act.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		router.ServeHTTP(rec, req)

		// Assert.
		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rec.Code)
		}
	})

	t.Run("rejects missing cookie or header", func(t *testing.T) {
		router := gin.New()
		router.Use(DoubleSubmitCSRF(true))
		router.POST("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		// Act.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("X-CSRF-Token", "abc")
		router.ServeHTTP(rec, req)

		// Assert.
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("unexpected status: %d", rec.Code)
		}
	})

	t.Run("accepts matching values", func(t *testing.T) {
		router := gin.New()
		router.Use(DoubleSubmitCSRF(true))
		router.POST("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		// Act.
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("X-CSRF-Token", "abc")
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "abc"})
		router.ServeHTTP(rec, req)

		// Assert.
		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rec.Code)
		}
	})
}
