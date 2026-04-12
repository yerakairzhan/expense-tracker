package service

import (
	"context"
	"time"

	sqlc "finance-tracker/db/queries"
	"finance-tracker/pkg/cache"
	"finance-tracker/pkg/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeAccountRepo struct {
	listByUserFn        func(ctx context.Context, userID int64) ([]sqlc.Account, error)
	createFn            func(ctx context.Context, userID int64, name, accountType, currency, balance string) (sqlc.Account, error)
	getByIDForUserFn    func(ctx context.Context, accountID, userID int64) (sqlc.Account, error)
	updateByIDForUserFn func(ctx context.Context, accountID, userID int64, name, accountType, currency *string) (sqlc.Account, error)
	softDeleteFn        func(ctx context.Context, accountID, userID int64) (int64, error)
}

func (f *fakeAccountRepo) ListByUser(ctx context.Context, userID int64) ([]sqlc.Account, error) {
	return f.listByUserFn(ctx, userID)
}

func (f *fakeAccountRepo) Create(ctx context.Context, userID int64, name, accountType, currency, balance string) (sqlc.Account, error) {
	return f.createFn(ctx, userID, name, accountType, currency, balance)
}

func (f *fakeAccountRepo) GetByIDForUser(ctx context.Context, accountID, userID int64) (sqlc.Account, error) {
	return f.getByIDForUserFn(ctx, accountID, userID)
}

func (f *fakeAccountRepo) UpdateByIDForUser(ctx context.Context, accountID, userID int64, name, accountType, currency *string) (sqlc.Account, error) {
	return f.updateByIDForUserFn(ctx, accountID, userID, name, accountType, currency)
}

func (f *fakeAccountRepo) SoftDeleteByIDForUser(ctx context.Context, accountID, userID int64) (int64, error) {
	return f.softDeleteFn(ctx, accountID, userID)
}

type fakeTransactionRepo struct {
	listForUserFn    func(ctx context.Context, params sqlc.ListTransactionsForUserParams) ([]sqlc.Transaction, error)
	createForUserFn  func(ctx context.Context, userID int64, params sqlc.CreateTransactionParams) (sqlc.Transaction, error)
	getByIDForUserFn func(ctx context.Context, txID, userID int64) (sqlc.Transaction, error)
	updateForUserFn  func(ctx context.Context, userID, txID int64, params sqlc.UpdateTransactionByIDForUserParams) (sqlc.Transaction, error)
	softDeleteFn     func(ctx context.Context, userID, txID int64) error
}

func (f *fakeTransactionRepo) ListForUser(ctx context.Context, params sqlc.ListTransactionsForUserParams) ([]sqlc.Transaction, error) {
	return f.listForUserFn(ctx, params)
}

func (f *fakeTransactionRepo) CreateForUser(ctx context.Context, userID int64, params sqlc.CreateTransactionParams) (sqlc.Transaction, error) {
	return f.createForUserFn(ctx, userID, params)
}

func (f *fakeTransactionRepo) GetByIDForUser(ctx context.Context, txID, userID int64) (sqlc.Transaction, error) {
	return f.getByIDForUserFn(ctx, txID, userID)
}

func (f *fakeTransactionRepo) UpdateForUser(ctx context.Context, userID, txID int64, params sqlc.UpdateTransactionByIDForUserParams) (sqlc.Transaction, error) {
	return f.updateForUserFn(ctx, userID, txID, params)
}

func (f *fakeTransactionRepo) SoftDeleteForUser(ctx context.Context, userID, txID int64) error {
	return f.softDeleteFn(ctx, userID, txID)
}

type fakeAnalyticsRepo struct {
	lastMonthSummaryFn           func(ctx context.Context, userID int64, start, end time.Time) (repository.AnalyticsSummaryRow, error)
	dailyProfitFn                func(ctx context.Context, userID int64, start, end time.Time) ([]repository.AnalyticsDailyProfitRow, error)
	lastMonthExpenseByCategoryFn func(ctx context.Context, userID int64, start, end time.Time) ([]repository.AnalyticsCategoryExpenseRow, error)
	monthlyProfitFn              func(ctx context.Context, userID int64, startMonth, endMonth time.Time) ([]repository.AnalyticsMonthlyProfitRow, error)
}

func (f *fakeAnalyticsRepo) LastMonthSummary(ctx context.Context, userID int64, start, end time.Time) (repository.AnalyticsSummaryRow, error) {
	return f.lastMonthSummaryFn(ctx, userID, start, end)
}

func (f *fakeAnalyticsRepo) DailyProfit(ctx context.Context, userID int64, start, end time.Time) ([]repository.AnalyticsDailyProfitRow, error) {
	return f.dailyProfitFn(ctx, userID, start, end)
}

func (f *fakeAnalyticsRepo) LastMonthExpenseByCategory(ctx context.Context, userID int64, start, end time.Time) ([]repository.AnalyticsCategoryExpenseRow, error) {
	return f.lastMonthExpenseByCategoryFn(ctx, userID, start, end)
}

func (f *fakeAnalyticsRepo) MonthlyProfit(ctx context.Context, userID int64, startMonth, endMonth time.Time) ([]repository.AnalyticsMonthlyProfitRow, error) {
	return f.monthlyProfitFn(ctx, userID, startMonth, endMonth)
}

type fakeAuthUserRepo struct {
	createFn     func(ctx context.Context, email, passwordHash, name, currency string) (sqlc.User, error)
	getByEmailFn func(ctx context.Context, email string) (sqlc.User, error)
	getByIDFn    func(ctx context.Context, userID int64) (sqlc.User, error)
}

func (f *fakeAuthUserRepo) Create(ctx context.Context, email, passwordHash, name, currency string) (sqlc.User, error) {
	return f.createFn(ctx, email, passwordHash, name, currency)
}

func (f *fakeAuthUserRepo) GetByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return f.getByEmailFn(ctx, email)
}

func (f *fakeAuthUserRepo) GetByID(ctx context.Context, userID int64) (sqlc.User, error) {
	return f.getByIDFn(ctx, userID)
}

type fakeBlocklist struct {
	revokeFn func(ctx context.Context, tokenID string, ttl time.Duration) error
}

func (f *fakeBlocklist) Revoke(ctx context.Context, tokenID string, ttl time.Duration) error {
	return f.revokeFn(ctx, tokenID, ttl)
}

type fakeRefreshStore struct {
	createFn func(ctx context.Context, tokenHash string, userID int64, ttl time.Duration) error
	getFn    func(ctx context.Context, tokenHash string) (*cache.RefreshSession, error)
	deleteFn func(ctx context.Context, tokenHash string) error
	rotateFn func(ctx context.Context, oldTokenHash, newTokenHash string, userID int64, ttl time.Duration) error
}

func (f *fakeRefreshStore) CreateRefreshSession(ctx context.Context, tokenHash string, userID int64, ttl time.Duration) error {
	return f.createFn(ctx, tokenHash, userID, ttl)
}

func (f *fakeRefreshStore) GetRefreshSession(ctx context.Context, tokenHash string) (*cache.RefreshSession, error) {
	return f.getFn(ctx, tokenHash)
}

func (f *fakeRefreshStore) DeleteRefreshSession(ctx context.Context, tokenHash string) error {
	return f.deleteFn(ctx, tokenHash)
}

func (f *fakeRefreshStore) RotateRefreshSession(ctx context.Context, oldTokenHash, newTokenHash string, userID int64, ttl time.Duration) error {
	return f.rotateFn(ctx, oldTokenHash, newTokenHash, userID, ttl)
}

type fakeUserRepo struct {
	getByIDFn        func(ctx context.Context, userID int64) (sqlc.User, error)
	updateProfileFn  func(ctx context.Context, userID int64, name, currency *string) (sqlc.User, error)
	updatePasswordFn func(ctx context.Context, userID int64, passwordHash string) (sqlc.User, error)
	updateRoleFn     func(ctx context.Context, userID int64, role string) (sqlc.User, error)
}

func (f *fakeUserRepo) GetByID(ctx context.Context, userID int64) (sqlc.User, error) {
	return f.getByIDFn(ctx, userID)
}

func (f *fakeUserRepo) UpdateProfile(ctx context.Context, userID int64, name, currency *string) (sqlc.User, error) {
	return f.updateProfileFn(ctx, userID, name, currency)
}

func (f *fakeUserRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) (sqlc.User, error) {
	return f.updatePasswordFn(ctx, userID, passwordHash)
}

func (f *fakeUserRepo) UpdateRole(ctx context.Context, userID int64, role string) (sqlc.User, error) {
	return f.updateRoleFn(ctx, userID, role)
}

type fakePinger struct {
	pingFn func(ctx context.Context) error
}

func (f *fakePinger) Ping(ctx context.Context) error {
	return f.pingFn(ctx)
}

func mustNumeric(value string) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.Scan(value); err != nil {
		panic(err)
	}
	return n
}

func testTime() time.Time {
	return time.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC)
}

func testAccountRow() sqlc.Account {
	now := testTime()
	return sqlc.Account{
		ID:          7,
		UserID:      42,
		Name:        "Cash",
		AccountType: "cash",
		Balance:     mustNumeric("100.5000"),
		Currency:    "USD",
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	}
}

func testTransactionRow() sqlc.Transaction {
	now := testTime()
	return sqlc.Transaction{
		ID:           9,
		AccountID:    7,
		CategoryID:   pgtype.Int8{Int64: 3, Valid: true},
		Amount:       mustNumeric("18.9000"),
		Currency:     "USD",
		Type:         "expense",
		Description:  "Groceries",
		Notes:        pgtype.Text{String: "weekly", Valid: true},
		TransactedAt: pgtype.Date{Time: time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC), Valid: true},
		CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	}
}

func testUserRow(passwordHash string) sqlc.User {
	now := testTime()
	return sqlc.User{
		ID:           42,
		Email:        "john@example.com",
		PasswordHash: passwordHash,
		Name:         "John",
		Currency:     "USD",
		Role:         "user",
		CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	}
}
