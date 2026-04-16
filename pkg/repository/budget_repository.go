package repository

import (
	"context"
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BudgetRow struct {
	ID          int64
	UserID      int64
	CategoryID  pgtype.Int8
	LimitAmount pgtype.Numeric
	Currency    string
	Period      string
	StartsAt    pgtype.Date
	EndsAt      pgtype.Date
	IsActive    bool
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
}

type BudgetProgressRow struct {
	LimitAmount pgtype.Numeric
	Spent       pgtype.Numeric
}

type BudgetRepository struct {
	pool *pgxpool.Pool
}

func NewBudgetRepository(pool *pgxpool.Pool) *BudgetRepository {
	return &BudgetRepository{pool: pool}
}

func (r *BudgetRepository) List(ctx context.Context, userID int64) ([]BudgetRow, error) {
	const q = `
SELECT id, user_id, category_id, limit_amount, currency, period, starts_at, ends_at, is_active, created_at, updated_at
FROM budgets
WHERE user_id = $1 AND is_active = TRUE
ORDER BY created_at DESC
`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]BudgetRow, 0)
	for rows.Next() {
		var b BudgetRow
		if err = rows.Scan(&b.ID, &b.UserID, &b.CategoryID, &b.LimitAmount, &b.Currency, &b.Period, &b.StartsAt, &b.EndsAt, &b.IsActive, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *BudgetRepository) Create(ctx context.Context, userID int64, categoryID *int64, limitAmount, currency, period, startsAt string, endsAt *string) (BudgetRow, error) {
	const q = `
INSERT INTO budgets (user_id, category_id, limit_amount, currency, period, starts_at, ends_at)
VALUES ($1, $2, $3, $4, $5, $6::date, $7::date)
RETURNING id, user_id, category_id, limit_amount, currency, period, starts_at, ends_at, is_active, created_at, updated_at
`
	limitNum, err := stringToNumeric(limitAmount)
	if err != nil {
		return BudgetRow{}, err
	}

	var endsAtArg interface{}
	if endsAt != nil {
		endsAtArg = *endsAt
	}

	var b BudgetRow
	err = r.pool.QueryRow(ctx, q, userID, int8FromPtr(categoryID), limitNum, currency, period, startsAt, endsAtArg).Scan(
		&b.ID, &b.UserID, &b.CategoryID, &b.LimitAmount, &b.Currency, &b.Period, &b.StartsAt, &b.EndsAt, &b.IsActive, &b.CreatedAt, &b.UpdatedAt,
	)
	return b, err
}

func (r *BudgetRepository) GetByIDForUser(ctx context.Context, id, userID int64) (BudgetRow, error) {
	const q = `
SELECT id, user_id, category_id, limit_amount, currency, period, period, starts_at, ends_at, is_active, created_at, updated_at
FROM budgets
WHERE id = $1 AND user_id = $2
`
	var b BudgetRow
	err := r.pool.QueryRow(ctx, q, id, userID).Scan(
		&b.ID, &b.UserID, &b.CategoryID, &b.LimitAmount, &b.Currency, &b.Period, &b.Period, &b.StartsAt, &b.EndsAt, &b.IsActive, &b.CreatedAt, &b.UpdatedAt,
	)
	return b, err
}

func (r *BudgetRepository) GetProgress(ctx context.Context, budgetID, userID int64) (BudgetProgressRow, error) {
	const q = `
SELECT
  b.limit_amount,
  COALESCE(SUM(t.amount), 0) AS spent
FROM budgets b
LEFT JOIN transactions t
  ON t.deleted_at IS NULL
  AND t.type = 'expense'
  AND t.transacted_at >= b.starts_at
  AND (b.ends_at IS NULL OR t.transacted_at <= b.ends_at)
  AND (b.category_id IS NULL OR t.category_id = b.category_id)
  AND t.account_id IN (
    SELECT id FROM accounts WHERE user_id = $2 AND deleted_at IS NULL
  )
WHERE b.id = $1 AND b.user_id = $2
GROUP BY b.limit_amount
`
	var row BudgetProgressRow
	err := r.pool.QueryRow(ctx, q, budgetID, userID).Scan(&row.LimitAmount, &row.Spent)
	return row, err
}

func (r *BudgetRepository) Update(ctx context.Context, id, userID int64, limitAmount, period, endsAt *string) (BudgetRow, error) {
	var limitNum *pgtype.Numeric
	if limitAmount != nil {
		n, err := stringToNumeric(*limitAmount)
		if err != nil {
			return BudgetRow{}, err
		}
		limitNum = &n
	}

	var endsAtArg interface{}
	if endsAt != nil {
		endsAtArg = *endsAt
	}

	const q = `
UPDATE budgets
SET limit_amount = COALESCE($3, limit_amount),
    period       = COALESCE($4::text, period),
    ends_at      = COALESCE($5::date, ends_at),
    updated_at   = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, category_id, limit_amount, currency, period, starts_at, ends_at, is_active, created_at, updated_at
`
	var b BudgetRow
	err := r.pool.QueryRow(ctx, q, id, userID, limitNum, textFromPtr(period), endsAtArg).Scan(
		&b.ID, &b.UserID, &b.CategoryID, &b.LimitAmount, &b.Currency, &b.Period, &b.StartsAt, &b.EndsAt, &b.IsActive, &b.CreatedAt, &b.UpdatedAt,
	)
	return b, err
}

func (r *BudgetRepository) Deactivate(ctx context.Context, id, userID int64) (int64, error) {
	const q = `UPDATE budgets SET is_active = FALSE, updated_at = NOW() WHERE id = $1 AND user_id = $2 AND is_active = TRUE`
	result, err := r.pool.Exec(ctx, q, id, userID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// BudgetProgressCalc computes remaining and percentage from a progress row.
func BudgetProgressCalc(limit, spent pgtype.Numeric) (limitStr, spentStr, remainingStr, percentStr string) {
	limitStr = numericToString4(limit)
	spentStr = numericToString4(spent)

	lRat, _ := new(big.Rat).SetString(limitStr)
	sRat, _ := new(big.Rat).SetString(spentStr)

	rem := new(big.Rat).Sub(lRat, sRat)
	remainingStr = rem.FloatString(4)

	if lRat.Sign() == 0 {
		percentStr = "0.00"
		return
	}
	pct := new(big.Rat).Mul(new(big.Rat).Quo(sRat, lRat), big.NewRat(100, 1))
	percentStr = pct.FloatString(2)
	return
}
