package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecurringRow struct {
	ID         int64
	UserID     int64
	AccountID  int64
	CategoryID pgtype.Int8
	Title      string
	Amount     pgtype.Numeric
	Currency   string
	Frequency  string
	NextRunAt  pgtype.Date
	EndsAt     pgtype.Date
	IsActive   bool
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
}

type RecurringRepository struct {
	pool *pgxpool.Pool
}

func NewRecurringRepository(pool *pgxpool.Pool) *RecurringRepository {
	return &RecurringRepository{pool: pool}
}

func (r *RecurringRepository) List(ctx context.Context, userID int64) ([]RecurringRow, error) {
	const q = `
SELECT id, user_id, account_id, category_id, title, amount, currency, frequency, next_run_at, ends_at, is_active, created_at, updated_at
FROM recurring_payments
WHERE user_id = $1
ORDER BY created_at DESC
`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]RecurringRow, 0)
	for rows.Next() {
		var rr RecurringRow
		if err = rows.Scan(&rr.ID, &rr.UserID, &rr.AccountID, &rr.CategoryID, &rr.Title, &rr.Amount,
			&rr.Currency, &rr.Frequency, &rr.NextRunAt, &rr.EndsAt, &rr.IsActive, &rr.CreatedAt, &rr.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, rr)
	}
	return out, rows.Err()
}

func (r *RecurringRepository) Create(ctx context.Context, userID, accountID int64, categoryID *int64, title, amount, currency, frequency, nextRunAt string, endsAt *string) (RecurringRow, error) {
	const q = `
INSERT INTO recurring_payments (user_id, account_id, category_id, title, amount, currency, frequency, next_run_at, ends_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8::date, $9::date)
RETURNING id, user_id, account_id, category_id, title, amount, currency, frequency, next_run_at, ends_at, is_active, created_at, updated_at
`
	amountNum, err := stringToNumeric(amount)
	if err != nil {
		return RecurringRow{}, err
	}

	var endsAtArg interface{}
	if endsAt != nil {
		endsAtArg = *endsAt
	}

	var rr RecurringRow
	err = r.pool.QueryRow(ctx, q, userID, accountID, int8FromPtr(categoryID), title, amountNum, currency, frequency, nextRunAt, endsAtArg).Scan(
		&rr.ID, &rr.UserID, &rr.AccountID, &rr.CategoryID, &rr.Title, &rr.Amount,
		&rr.Currency, &rr.Frequency, &rr.NextRunAt, &rr.EndsAt, &rr.IsActive, &rr.CreatedAt, &rr.UpdatedAt,
	)
	return rr, err
}

func (r *RecurringRepository) GetByIDForUser(ctx context.Context, id, userID int64) (RecurringRow, error) {
	const q = `
SELECT id, user_id, account_id, category_id, title, amount, currency, frequency, next_run_at, ends_at, is_active, created_at, updated_at
FROM recurring_payments
WHERE id = $1 AND user_id = $2
`
	var rr RecurringRow
	err := r.pool.QueryRow(ctx, q, id, userID).Scan(
		&rr.ID, &rr.UserID, &rr.AccountID, &rr.CategoryID, &rr.Title, &rr.Amount,
		&rr.Currency, &rr.Frequency, &rr.NextRunAt, &rr.EndsAt, &rr.IsActive, &rr.CreatedAt, &rr.UpdatedAt,
	)
	return rr, err
}

func (r *RecurringRepository) Update(ctx context.Context, id, userID int64, amount, frequency, nextRunAt, endsAt *string) (RecurringRow, error) {
	var amountNum *pgtype.Numeric
	if amount != nil {
		n, err := stringToNumeric(*amount)
		if err != nil {
			return RecurringRow{}, err
		}
		amountNum = &n
	}

	var endsAtArg interface{}
	if endsAt != nil {
		endsAtArg = *endsAt
	}

	const q = `
UPDATE recurring_payments
SET amount     = COALESCE($3, amount),
    frequency  = COALESCE($4::text, frequency),
    next_run_at = COALESCE($5::date, next_run_at),
    ends_at    = COALESCE($6::date, ends_at),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, account_id, category_id, title, amount, currency, frequency, next_run_at, ends_at, is_active, created_at, updated_at
`
	var rr RecurringRow
	err := r.pool.QueryRow(ctx, q, id, userID, amountNum, textFromPtr(frequency), textFromPtr(nextRunAt), endsAtArg).Scan(
		&rr.ID, &rr.UserID, &rr.AccountID, &rr.CategoryID, &rr.Title, &rr.Amount,
		&rr.Currency, &rr.Frequency, &rr.NextRunAt, &rr.EndsAt, &rr.IsActive, &rr.CreatedAt, &rr.UpdatedAt,
	)
	return rr, err
}

func (r *RecurringRepository) Deactivate(ctx context.Context, id, userID int64) (int64, error) {
	const q = `UPDATE recurring_payments SET is_active = FALSE, updated_at = NOW() WHERE id = $1 AND user_id = $2 AND is_active = TRUE`
	result, err := r.pool.Exec(ctx, q, id, userID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
