package repository

import (
	"context"

	sqlc "finance-tracker/db/queries"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	q *sqlc.Queries
}

func NewUserRepository(q *sqlc.Queries) *UserRepository {
	return &UserRepository{q: q}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash, name, currency string) (sqlc.User, error) {
	return r.q.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		Currency:     currency,
	})
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *UserRepository) GetByID(ctx context.Context, userID int64) (sqlc.User, error) {
	return r.q.GetUserByID(ctx, userID)
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID int64, name, currency *string) (sqlc.User, error) {
	return r.q.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
		ID:       userID,
		Name:     textFromPtr(name),
		Currency: textFromPtr(currency),
	})
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) (sqlc.User, error) {
	return r.q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: passwordHash,
	})
}

func (r *UserRepository) UpdateRole(ctx context.Context, userID int64, role string) (sqlc.User, error) {
	return r.q.UpdateUserRole(ctx, sqlc.UpdateUserRoleParams{
		ID:   userID,
		Role: role,
	})
}

func (r *UserRepository) InsertRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt pgtype.Timestamptz) error {
	_, err := r.q.InsertRefreshToken(ctx, sqlc.InsertRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	return err
}

func (r *UserRepository) ListValidRefreshTokens(ctx context.Context) ([]sqlc.RefreshToken, error) {
	return r.q.ListValidRefreshTokens(ctx)
}

func (r *UserRepository) ListValidRefreshTokensByUser(ctx context.Context, userID int64) ([]sqlc.RefreshToken, error) {
	return r.q.ListValidRefreshTokensByUser(ctx, userID)
}

func (r *UserRepository) RevokeRefreshTokenByID(ctx context.Context, id int64) (int64, error) {
	return r.q.RevokeRefreshTokenByID(ctx, id)
}

func (r *UserRepository) RevokeRefreshTokenByIDForUser(ctx context.Context, id, userID int64) (int64, error) {
	return r.q.RevokeRefreshTokenByIDForUser(ctx, sqlc.RevokeRefreshTokenByIDForUserParams{
		ID:     id,
		UserID: userID,
	})
}
