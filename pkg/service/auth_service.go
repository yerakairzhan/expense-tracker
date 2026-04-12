package service

import (
	"context"
	"time"

	sqlc "finance-tracker/db/queries"
	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/auth"
	"finance-tracker/pkg/cache"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users         authUserRepository
	jwtSecret     string
	blocklist     tokenBlocklist
	refreshStore  refreshSessionStore
	refreshPepper string
}

type authUserRepository interface {
	Create(ctx context.Context, email, passwordHash, name, currency string) (sqlc.User, error)
	GetByEmail(ctx context.Context, email string) (sqlc.User, error)
	GetByID(ctx context.Context, userID int64) (sqlc.User, error)
}

type tokenBlocklist interface {
	Revoke(ctx context.Context, tokenID string, ttl time.Duration) error
}

type refreshSessionStore interface {
	CreateRefreshSession(ctx context.Context, tokenHash string, userID int64, ttl time.Duration) error
	GetRefreshSession(ctx context.Context, tokenHash string) (*cache.RefreshSession, error)
	DeleteRefreshSession(ctx context.Context, tokenHash string) error
	RotateRefreshSession(ctx context.Context, oldTokenHash, newTokenHash string, userID int64, ttl time.Duration) error
}

func NewAuthService(
	users *repository.UserRepository,
	jwtSecret string,
	blocklist tokenBlocklist,
	refreshStore refreshSessionStore,
	refreshPepper string,
) *AuthService {
	return &AuthService{
		users:         users,
		jwtSecret:     jwtSecret,
		blocklist:     blocklist,
		refreshStore:  refreshStore,
		refreshPepper: refreshPepper,
	}
}

func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.AuthTokens, *apperror.Error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Internal("failed to hash password")
	}

	user, err := s.users.Create(ctx, req.Email, string(passwordHash), req.Name, req.Currency)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errorAs(err, &pgErr); ok && pgErr.Code == "23505" {
			return nil, apperror.Conflict("email already exists")
		}
		return nil, apperror.Internal("failed to create user")
	}

	tokens, appErr := s.issueTokens(ctx, user.ID, user.Role, time.Now().UTC())
	if appErr != nil {
		return nil, appErr
	}
	return tokens, nil
}

// Login: create refresh session in Redis (hashed token), return refresh token only for cookie setting.
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthTokens, *apperror.Error) {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Unauthorized("invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, apperror.Unauthorized("invalid credentials")
	}
	return s.issueTokens(ctx, user.ID, user.Role, time.Now().UTC())
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*models.AuthTokens, *apperror.Error) {
	oldHash, err := auth.HashRefreshToken(s.refreshPepper, refreshToken)
	if err != nil {
		return nil, apperror.Internal("failed to hash refresh token")
	}
	session, err := s.refreshStore.GetRefreshSession(ctx, oldHash)
	if err != nil || session == nil || session.UserID <= 0 {
		return nil, apperror.Unauthorized("invalid refresh token")
	}

	newRefreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, apperror.Internal("failed to issue refresh token")
	}
	newHash, err := auth.HashRefreshToken(s.refreshPepper, newRefreshToken)
	if err != nil {
		return nil, apperror.Internal("failed to hash refresh token")
	}
	if err := s.refreshStore.RotateRefreshSession(ctx, oldHash, newHash, session.UserID, auth.RefreshTokenTTL); err != nil {
		return nil, apperror.Internal("failed to rotate refresh token")
	}

	user, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, apperror.Unauthorized("invalid refresh token")
	}

	accessToken, err := auth.GenerateAccessToken(s.jwtSecret, session.UserID, user.Role, time.Now().UTC())
	if err != nil {
		return nil, apperror.Internal("failed to issue access token")
	}
	return &models.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken, // internal only; handler must put into cookie.
		ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID int64, rawRefreshToken, rawAccessToken string) *apperror.Error {
	claims, err := auth.ParseAccessToken(s.jwtSecret, rawAccessToken)
	if err != nil {
		return apperror.Unauthorized("invalid or expired token")
	}
	if claims.ID == "" || claims.ExpiresAt == nil {
		return apperror.Unauthorized("invalid token")
	}
	if claims.UserID != userID {
		return apperror.Unauthorized("token subject mismatch")
	}

	refreshHash, err := auth.HashRefreshToken(s.refreshPepper, rawRefreshToken)
	if err != nil {
		return apperror.Internal("failed to hash refresh token")
	}
	if err = s.refreshStore.DeleteRefreshSession(ctx, refreshHash); err != nil {
		return apperror.Internal("failed to revoke refresh token")
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl > 0 {
		if err = s.blocklist.Revoke(ctx, claims.ID, ttl); err != nil {
			return apperror.Internal("failed to revoke access token")
		}
	}
	return nil
}

func (s *AuthService) issueTokens(ctx context.Context, userID int64, role string, now time.Time) (*models.AuthTokens, *apperror.Error) {
	access, err := auth.GenerateAccessToken(s.jwtSecret, userID, role, now)
	if err != nil {
		return nil, apperror.Internal("failed to issue access token")
	}

	refreshRaw, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, apperror.Internal("failed to issue refresh token")
	}

	refreshHash, err := auth.HashRefreshToken(s.refreshPepper, refreshRaw)
	if err != nil {
		return nil, apperror.Internal("failed to hash refresh token")
	}

	if err = s.refreshStore.CreateRefreshSession(ctx, refreshHash, userID, auth.RefreshTokenTTL); err != nil {
		return nil, apperror.Internal("failed to store refresh token")
	}

	return &models.AuthTokens{
		AccessToken:  access,
		RefreshToken: refreshRaw,
		ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
	}, nil
}
