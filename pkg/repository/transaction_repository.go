package repository

import (
	"context"
	"fmt"
	"math/big"

	sqlc "finance-tracker/db/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewTransactionRepository(pool *pgxpool.Pool, q *sqlc.Queries) *TransactionRepository {
	return &TransactionRepository{pool: pool, q: q}
}

func (r *TransactionRepository) ListForUser(ctx context.Context, params sqlc.ListTransactionsForUserParams) ([]sqlc.Transaction, error) {
	return r.q.ListTransactionsForUser(ctx, params)
}

func (r *TransactionRepository) GetByIDForUser(ctx context.Context, txID, userID int64) (sqlc.Transaction, error) {
	return r.q.GetTransactionByIDForUser(ctx, sqlc.GetTransactionByIDForUserParams{
		ID:     txID,
		UserID: userID,
	})
}

func (r *TransactionRepository) CreateForUser(
	ctx context.Context,
	userID int64,
	params sqlc.CreateTransactionParams,
) (sqlc.Transaction, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return sqlc.Transaction{}, err
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)

	_, err = qtx.GetAccountByIDForUserForUpdate(ctx, sqlc.GetAccountByIDForUserForUpdateParams{
		ID:     params.AccountID,
		UserID: userID,
	})
	if err != nil {
		return sqlc.Transaction{}, err
	}
	if params.CategoryID.Valid {
		if _, err = qtx.CategoryAccessibleForUser(ctx, sqlc.CategoryAccessibleForUserParams{
			ID:     params.CategoryID.Int64,
			UserID: pgtype.Int8{Int64: userID, Valid: true},
		}); err != nil {
			return sqlc.Transaction{}, err
		}
	}

	created, err := qtx.CreateTransaction(ctx, params)
	if err != nil {
		return sqlc.Transaction{}, err
	}

	delta, err := signedAmount(params.Amount, params.Type)
	if err != nil {
		return sqlc.Transaction{}, err
	}
	deltaNum, err := stringToNumeric(delta)
	if err != nil {
		return sqlc.Transaction{}, err
	}

	_, err = qtx.UpdateAccountBalanceDeltaByIDForUser(ctx, sqlc.UpdateAccountBalanceDeltaByIDForUserParams{
		ID:      params.AccountID,
		UserID:  userID,
		Balance: deltaNum,
	})
	if err != nil {
		return sqlc.Transaction{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.Transaction{}, err
	}
	return created, nil
}

func (r *TransactionRepository) UpdateForUser(
	ctx context.Context,
	userID, txID int64,
	params sqlc.UpdateTransactionByIDForUserParams,
) (sqlc.Transaction, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return sqlc.Transaction{}, err
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)

	current, err := qtx.GetTransactionByIDForUserForUpdate(ctx, sqlc.GetTransactionByIDForUserForUpdateParams{
		ID:     txID,
		UserID: userID,
	})
	if err != nil {
		return sqlc.Transaction{}, err
	}
	if params.CategoryID.Valid {
		if _, err = qtx.CategoryAccessibleForUser(ctx, sqlc.CategoryAccessibleForUserParams{
			ID:     params.CategoryID.Int64,
			UserID: pgtype.Int8{Int64: userID, Valid: true},
		}); err != nil {
			return sqlc.Transaction{}, err
		}
	}

	params.ID = txID
	params.UserID = userID
	updated, err := qtx.UpdateTransactionByIDForUser(ctx, params)
	if err != nil {
		return sqlc.Transaction{}, err
	}

	oldSigned, err := signedAmount(current.Amount, current.Type)
	if err != nil {
		return sqlc.Transaction{}, err
	}
	newSigned, err := signedAmount(updated.Amount, updated.Type)
	if err != nil {
		return sqlc.Transaction{}, err
	}
	delta, err := subtractDecimalStrings(newSigned, oldSigned)
	if err != nil {
		return sqlc.Transaction{}, err
	}

	if delta != "0.0000" {
		deltaNum, err := stringToNumeric(delta)
		if err != nil {
			return sqlc.Transaction{}, err
		}
		_, err = qtx.UpdateAccountBalanceDeltaByIDForUser(ctx, sqlc.UpdateAccountBalanceDeltaByIDForUserParams{
			ID:      current.AccountID,
			UserID:  userID,
			Balance: deltaNum,
		})
		if err != nil {
			return sqlc.Transaction{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.Transaction{}, err
	}
	return updated, nil
}

func (r *TransactionRepository) SoftDeleteForUser(ctx context.Context, userID, txID int64) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)
	current, err := qtx.GetTransactionByIDForUserForUpdate(ctx, sqlc.GetTransactionByIDForUserForUpdateParams{
		ID:     txID,
		UserID: userID,
	})
	if err != nil {
		return err
	}

	affected, err := qtx.SoftDeleteTransactionByIDForUser(ctx, sqlc.SoftDeleteTransactionByIDForUserParams{
		ID:     txID,
		UserID: userID,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return pgx.ErrNoRows
	}

	currentSigned, err := signedAmount(current.Amount, current.Type)
	if err != nil {
		return err
	}
	delta, err := subtractDecimalStrings("0.0000", currentSigned)
	if err != nil {
		return err
	}
	deltaNum, err := stringToNumeric(delta)
	if err != nil {
		return err
	}

	_, err = qtx.UpdateAccountBalanceDeltaByIDForUser(ctx, sqlc.UpdateAccountBalanceDeltaByIDForUserParams{
		ID:      current.AccountID,
		UserID:  userID,
		Balance: deltaNum,
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func signedAmount(amount pgtype.Numeric, txType string) (string, error) {
	raw := numericToString4(amount)
	if txType == "income" {
		return raw, nil
	}
	if txType == "expense" || txType == "transfer" {
		if raw == "0.0000" {
			return raw, nil
		}
		return "-" + raw, nil
	}
	return "", fmt.Errorf("invalid transaction type")
}

func subtractDecimalStrings(left, right string) (string, error) {
	l, ok := new(big.Rat).SetString(left)
	if !ok {
		return "", fmt.Errorf("invalid decimal value")
	}
	r, ok := new(big.Rat).SetString(right)
	if !ok {
		return "", fmt.Errorf("invalid decimal value")
	}
	return new(big.Rat).Sub(l, r).FloatString(4), nil
}
