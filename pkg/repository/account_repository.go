package repository

import (
	"context"

	sqlc "finance-tracker/db/queries"
)

type AccountRepository struct {
	q *sqlc.Queries
}

func NewAccountRepository(q *sqlc.Queries) *AccountRepository {
	return &AccountRepository{q: q}
}

func (r *AccountRepository) Create(ctx context.Context, userID int64, name, accountType, currency, balance string) (sqlc.Account, error) {
	numeric, err := stringToNumeric(balance)
	if err != nil {
		return sqlc.Account{}, err
	}
	return r.q.CreateAccount(ctx, sqlc.CreateAccountParams{
		UserID:      userID,
		Name:        name,
		AccountType: accountType,
		Balance:     numeric,
		Currency:    currency,
	})
}

func (r *AccountRepository) ListByUser(ctx context.Context, userID int64) ([]sqlc.Account, error) {
	return r.q.ListAccountsByUser(ctx, userID)
}

func (r *AccountRepository) GetByIDForUser(ctx context.Context, accountID, userID int64) (sqlc.Account, error) {
	return r.q.GetAccountByIDForUser(ctx, sqlc.GetAccountByIDForUserParams{
		ID:     accountID,
		UserID: userID,
	})
}

func (r *AccountRepository) UpdateByIDForUser(ctx context.Context, accountID, userID int64, name, accountType, currency *string) (sqlc.Account, error) {
	return r.q.UpdateAccountByIDForUser(ctx, sqlc.UpdateAccountByIDForUserParams{
		ID:          accountID,
		UserID:      userID,
		Name:        textFromPtr(name),
		AccountType: textFromPtr(accountType),
		Currency:    textFromPtr(currency),
	})
}

func (r *AccountRepository) SoftDeleteByIDForUser(ctx context.Context, accountID, userID int64) (int64, error) {
	return r.q.SoftDeleteAccountByIDForUser(ctx, sqlc.SoftDeleteAccountByIDForUserParams{
		ID:     accountID,
		UserID: userID,
	})
}
