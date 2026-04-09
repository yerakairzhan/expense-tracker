package repository

import (
	"context"
	"errors"
	"testing"

	sqlc "finance-tracker/db/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeTxPool struct {
	beginTxFn func(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

func (f *fakeTxPool) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return f.beginTxFn(ctx, txOptions)
}

func (f *fakeTxPool) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("unexpected Query call")
}

func (f *fakeTxPool) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("unexpected QueryRow call")
}

type fakeTx struct {
	commitCalled   bool
	rollbackCalled bool
	commitFn       func(ctx context.Context) error
}

func (f *fakeTx) Begin(context.Context) (pgx.Tx, error) { return nil, errors.New("not implemented") }
func (f *fakeTx) Commit(ctx context.Context) error {
	f.commitCalled = true
	if f.commitFn != nil {
		return f.commitFn(ctx)
	}
	return nil
}
func (f *fakeTx) Rollback(context.Context) error { f.rollbackCalled = true; return nil }
func (f *fakeTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, errors.New("not implemented")
}
func (f *fakeTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeTx) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (f *fakeTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("not implemented")
}
func (f *fakeTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeTx) QueryRow(context.Context, string, ...any) pgx.Row { return nil }
func (f *fakeTx) Conn() *pgx.Conn                                  { return nil }

type fakeTransactionQueries struct {
	withTxFn                               func(tx pgx.Tx) transactionQueries
	getAccountByIDForUserForUpdateFn       func(ctx context.Context, arg sqlc.GetAccountByIDForUserForUpdateParams) (sqlc.Account, error)
	categoryAccessibleForUserFn            func(ctx context.Context, arg sqlc.CategoryAccessibleForUserParams) (int64, error)
	createTransactionFn                    func(ctx context.Context, arg sqlc.CreateTransactionParams) (sqlc.Transaction, error)
	updateAccountBalanceDeltaByIDForUserFn func(ctx context.Context, arg sqlc.UpdateAccountBalanceDeltaByIDForUserParams) (sqlc.Account, error)
	getTransactionByIDForUserForUpdateFn   func(ctx context.Context, arg sqlc.GetTransactionByIDForUserForUpdateParams) (sqlc.Transaction, error)
	updateTransactionByIDForUserFn         func(ctx context.Context, arg sqlc.UpdateTransactionByIDForUserParams) (sqlc.Transaction, error)
	softDeleteTransactionByIDForUserFn     func(ctx context.Context, arg sqlc.SoftDeleteTransactionByIDForUserParams) (int64, error)
}

func (f *fakeTransactionQueries) WithTx(tx pgx.Tx) transactionQueries {
	if f.withTxFn != nil {
		return f.withTxFn(tx)
	}
	return f
}
func (f *fakeTransactionQueries) ListTransactionsForUser(context.Context, sqlc.ListTransactionsForUserParams) ([]sqlc.Transaction, error) {
	panic("unexpected ListTransactionsForUser call")
}
func (f *fakeTransactionQueries) GetTransactionByIDForUser(context.Context, sqlc.GetTransactionByIDForUserParams) (sqlc.Transaction, error) {
	panic("unexpected GetTransactionByIDForUser call")
}
func (f *fakeTransactionQueries) GetAccountByIDForUserForUpdate(ctx context.Context, arg sqlc.GetAccountByIDForUserForUpdateParams) (sqlc.Account, error) {
	return f.getAccountByIDForUserForUpdateFn(ctx, arg)
}
func (f *fakeTransactionQueries) CategoryAccessibleForUser(ctx context.Context, arg sqlc.CategoryAccessibleForUserParams) (int64, error) {
	if f.categoryAccessibleForUserFn == nil {
		return 0, nil
	}
	return f.categoryAccessibleForUserFn(ctx, arg)
}
func (f *fakeTransactionQueries) CreateTransaction(ctx context.Context, arg sqlc.CreateTransactionParams) (sqlc.Transaction, error) {
	return f.createTransactionFn(ctx, arg)
}
func (f *fakeTransactionQueries) UpdateAccountBalanceDeltaByIDForUser(ctx context.Context, arg sqlc.UpdateAccountBalanceDeltaByIDForUserParams) (sqlc.Account, error) {
	return f.updateAccountBalanceDeltaByIDForUserFn(ctx, arg)
}
func (f *fakeTransactionQueries) GetTransactionByIDForUserForUpdate(ctx context.Context, arg sqlc.GetTransactionByIDForUserForUpdateParams) (sqlc.Transaction, error) {
	return f.getTransactionByIDForUserForUpdateFn(ctx, arg)
}
func (f *fakeTransactionQueries) UpdateTransactionByIDForUser(ctx context.Context, arg sqlc.UpdateTransactionByIDForUserParams) (sqlc.Transaction, error) {
	return f.updateTransactionByIDForUserFn(ctx, arg)
}
func (f *fakeTransactionQueries) SoftDeleteTransactionByIDForUser(ctx context.Context, arg sqlc.SoftDeleteTransactionByIDForUserParams) (int64, error) {
	return f.softDeleteTransactionByIDForUserFn(ctx, arg)
}

func TestTransactionRepository_CreateForUser(t *testing.T) {
	tx := &fakeTx{}
	var gotDelta string
	repo := &TransactionRepository{
		pool: &fakeTxPool{
			beginTxFn: func(context.Context, pgx.TxOptions) (pgx.Tx, error) { return tx, nil },
		},
		q: &fakeTransactionQueries{
			getAccountByIDForUserForUpdateFn: func(context.Context, sqlc.GetAccountByIDForUserForUpdateParams) (sqlc.Account, error) {
				return sqlc.Account{ID: 7}, nil
			},
			createTransactionFn: func(context.Context, sqlc.CreateTransactionParams) (sqlc.Transaction, error) {
				return sqlc.Transaction{ID: 9}, nil
			},
			updateAccountBalanceDeltaByIDForUserFn: func(_ context.Context, arg sqlc.UpdateAccountBalanceDeltaByIDForUserParams) (sqlc.Account, error) {
				gotDelta = numericToString4(arg.Balance)
				return sqlc.Account{}, nil
			},
		},
	}

	// Act.
	got, err := repo.CreateForUser(context.Background(), 42, sqlc.CreateTransactionParams{
		AccountID: 7,
		Amount:    mustNumericRepo("18.9000"),
		Type:      "expense",
	})

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 9 || gotDelta != "-18.9000" || !tx.commitCalled {
		t.Fatalf("unexpected result: %#v delta=%s commit=%v", got, gotDelta, tx.commitCalled)
	}
}

func TestTransactionRepository_UpdateForUser(t *testing.T) {
	tx := &fakeTx{}
	var gotDelta string
	repo := &TransactionRepository{
		pool: &fakeTxPool{
			beginTxFn: func(context.Context, pgx.TxOptions) (pgx.Tx, error) { return tx, nil },
		},
		q: &fakeTransactionQueries{
			getTransactionByIDForUserForUpdateFn: func(context.Context, sqlc.GetTransactionByIDForUserForUpdateParams) (sqlc.Transaction, error) {
				return sqlc.Transaction{AccountID: 7, Amount: mustNumericRepo("10.0000"), Type: "expense"}, nil
			},
			updateTransactionByIDForUserFn: func(context.Context, sqlc.UpdateTransactionByIDForUserParams) (sqlc.Transaction, error) {
				return sqlc.Transaction{AccountID: 7, Amount: mustNumericRepo("12.5000"), Type: "expense"}, nil
			},
			updateAccountBalanceDeltaByIDForUserFn: func(_ context.Context, arg sqlc.UpdateAccountBalanceDeltaByIDForUserParams) (sqlc.Account, error) {
				gotDelta = numericToString4(arg.Balance)
				return sqlc.Account{}, nil
			},
		},
	}

	// Act.
	got, err := repo.UpdateForUser(context.Background(), 42, 9, sqlc.UpdateTransactionByIDForUserParams{})

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotDelta != "-2.5000" || !tx.commitCalled || got.Amount.Int == nil {
		t.Fatalf("unexpected result: delta=%s commit=%v tx=%#v", gotDelta, tx.commitCalled, got)
	}
}

func TestTransactionRepository_SoftDeleteForUser(t *testing.T) {
	tx := &fakeTx{}
	var gotDelta string
	repo := &TransactionRepository{
		pool: &fakeTxPool{
			beginTxFn: func(context.Context, pgx.TxOptions) (pgx.Tx, error) { return tx, nil },
		},
		q: &fakeTransactionQueries{
			getTransactionByIDForUserForUpdateFn: func(context.Context, sqlc.GetTransactionByIDForUserForUpdateParams) (sqlc.Transaction, error) {
				return sqlc.Transaction{AccountID: 7, Amount: mustNumericRepo("10.0000"), Type: "income"}, nil
			},
			softDeleteTransactionByIDForUserFn: func(context.Context, sqlc.SoftDeleteTransactionByIDForUserParams) (int64, error) {
				return 1, nil
			},
			updateAccountBalanceDeltaByIDForUserFn: func(_ context.Context, arg sqlc.UpdateAccountBalanceDeltaByIDForUserParams) (sqlc.Account, error) {
				gotDelta = numericToString4(arg.Balance)
				return sqlc.Account{}, nil
			},
		},
	}

	// Act.
	err := repo.SoftDeleteForUser(context.Background(), 42, 9)

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotDelta != "-10.0000" || !tx.commitCalled {
		t.Fatalf("unexpected delta/commit: %s %v", gotDelta, tx.commitCalled)
	}
}

func TestTransactionRepository_SoftDeleteForUserMapsMissingRow(t *testing.T) {
	tx := &fakeTx{}
	repo := &TransactionRepository{
		pool: &fakeTxPool{
			beginTxFn: func(context.Context, pgx.TxOptions) (pgx.Tx, error) { return tx, nil },
		},
		q: &fakeTransactionQueries{
			getTransactionByIDForUserForUpdateFn: func(context.Context, sqlc.GetTransactionByIDForUserForUpdateParams) (sqlc.Transaction, error) {
				return sqlc.Transaction{AccountID: 7, Amount: mustNumericRepo("10.0000"), Type: "income"}, nil
			},
			softDeleteTransactionByIDForUserFn: func(context.Context, sqlc.SoftDeleteTransactionByIDForUserParams) (int64, error) {
				return 0, nil
			},
		},
	}

	// Act.
	err := repo.SoftDeleteForUser(context.Background(), 42, 9)

	// Assert.
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
}

func TestTransactionRepositoryDecimalHelpers(t *testing.T) {
	t.Run("signedAmount handles types", func(t *testing.T) {
		// Act.
		income, err := signedAmount(mustNumericRepo("10.0000"), "income")
		if err != nil {
			t.Fatal(err)
		}
		expense, err := signedAmount(mustNumericRepo("10.0000"), "expense")
		if err != nil {
			t.Fatal(err)
		}

		// Assert.
		if income != "10.0000" || expense != "-10.0000" {
			t.Fatalf("unexpected signed values: %s %s", income, expense)
		}
	})

	t.Run("subtractDecimalStrings subtracts fixed scale", func(t *testing.T) {
		// Act.
		got, err := subtractDecimalStrings("12.5000", "10.0000")

		// Assert.
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "2.5000" {
			t.Fatalf("unexpected value: %s", got)
		}
	})
}

func mustNumericRepo(value string) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.Scan(value); err != nil {
		panic(err)
	}
	return n
}
