package middleware

import (
	"strings"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/auth"
	"github.com/gin-gonic/gin"
)

const userIDContextKey = "auth_user_id"

func UserIDFromContext(c *gin.Context) (int64, bool) {
	v, ok := c.Get(userIDContextKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authz := c.GetHeader("Authorization")
		if authz == "" {
			writeAbort(c, apperror.Unauthorized("missing bearer token"))
			return
		}
		parts := strings.SplitN(authz, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			writeAbort(c, apperror.Unauthorized("invalid authorization header"))
			return
		}
		claims, err := auth.ParseAccessToken(secret, strings.TrimSpace(parts[1]))
		if err != nil {
			writeAbort(c, apperror.Unauthorized("invalid or expired token"))
			return
		}
		c.Set(userIDContextKey, claims.UserID)
		c.Next()
	}
}

func writeAbort(c *gin.Context, err *apperror.Error) {
	c.AbortWithStatusJSON(err.Status, gin.H{
		"error": gin.H{
			"code":    err.Code,
			"message": err.Message,
		},
	})
}
