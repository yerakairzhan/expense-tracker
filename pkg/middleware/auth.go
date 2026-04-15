package middleware

import (
	"context"
	"strings"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/auth"

	"github.com/gin-gonic/gin"
)

const userIDContextKey = "auth_user_id"
const roleContextKey = "auth_role"

type tokenBlocklist interface {
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
}

func UserIDFromContext(c *gin.Context) (int64, bool) {
	v, ok := c.Get(userIDContextKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

func RoleFromContext(c *gin.Context) (string, bool) {
	v, ok := c.Get(roleContextKey)
	if !ok {
		return "", false
	}
	role, ok := v.(string)
	if !ok || strings.TrimSpace(role) == "" {
		return "", false
	}
	return role, true
}

func AccessTokenFromHeader(authz string) (string, *apperror.Error) {
	if authz == "" {
		return "", apperror.Unauthorized("missing bearer token")
	}
	parts := strings.SplitN(authz, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", apperror.Unauthorized("invalid authorization header")
	}
	return strings.TrimSpace(parts[1]), nil
}

func JWTAuth(secret string, blocklist tokenBlocklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, authErr := AccessTokenFromHeader(c.GetHeader("Authorization"))
		if authErr != nil {
			writeAbort(c, authErr)
			return
		}
		claims, err := auth.ParseAccessToken(secret, rawToken)
		if err != nil {
			writeAbort(c, apperror.Unauthorized("invalid or expired token"))
			return
		}
		if claims.ID == "" {
			writeAbort(c, apperror.Unauthorized("invalid token"))
			return
		}
		if blocklist != nil {
			revoked, err := blocklist.IsRevoked(c.Request.Context(), claims.ID)
			if err != nil {
				writeAbort(c, apperror.Internal("failed to verify token"))
				return
			}
			if revoked {
				writeAbort(c, apperror.Unauthorized("token has been revoked"))
				return
			}
		}
		c.Set(userIDContextKey, claims.UserID)
		c.Set(roleContextKey, claims.Role)
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		trimmed := strings.TrimSpace(strings.ToLower(role))
		if trimmed == "" {
			continue
		}
		allowed[trimmed] = struct{}{}
	}

	return func(c *gin.Context) {
		role, ok := RoleFromContext(c)
		if !ok {
			writeAbort(c, apperror.Forbidden("insufficient permissions"))
			return
		}
		if _, ok = allowed[strings.ToLower(role)]; !ok {
			writeAbort(c, apperror.Forbidden("insufficient permissions"))
			return
		}
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
