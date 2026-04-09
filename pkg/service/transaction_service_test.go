package service

import (
	"context"
	"errors"
	"testing"

	sqlc "finance-tracker/db/queries"
	"finance-tracker/pkg/models"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestTransactionService_List(t *testing.T) {
	from := "2024-01-01"
	to := "2024-01-31"
	accountID := int64(7)
	txType := "expense"

	repo := &fakeTransactionRepo{
		listForUserFn: func(_ context.Context, params sqlc.ListTransactionsForUserParams) ([]sqlc.Transaction, error) {
			if params.UserID != 42 || params.OffsetRows != 10 || params.LimitRows != 10 {
				t.Fatalf("unexpected paging params: %+v", params)
			}
			if !params.AccountID.Valid || params.AccountID.Int64 != accountID {
				t.Fatalf("unexpected account filter: %+v", params.AccountID)
			}
			if !params.Type.Valid || params.Type.String != txType {
				t.Fatalf("unexpected type filter: %+v", params.Type)
			}
			if !params.FromDate.Valid || !params.ToDate.Valid {
				t.Fatalf("expected date filters: %+v", params)
			}
			return []sqlc.Transaction{testTransactionRow()}, nil
		},
	}
	svc := &TransactionService{txRepo: repo}

	// Act.
	got, err := svc.List(context.Background(), 42, models.ListTransactionsQuery{
		AccountID: &accountID,
		Type:      &txType,
		From:      &from,
		To:        &to,
		Page:      2,
		Limit:     10,
	})

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}
	if len(got) != 1 || got[0].Amount != "18.9000" {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestTransactionService_CreateUpdateDelete(t *testing.T) {
	t.Run("create validates amount and maps no rows", func(t *testing.T) {
		svc := &TransactionService{
			txRepo: &fakeTransactionRepo{
				createForUserFn: func(context.Context, int64, sqlc.CreateTransactionParams) (sqlc.Transaction, error) {
					return sqlc.Transaction{}, errors.New("no rows in result set")
				},
			},
		}

		// Act.
		got, err := svc.Create(context.Background(), 42, models.CreateTransactionRequest{
			AccountID:    7,
			Amount:       "10.1000",
			Currency:     "USD",
			Type:         "expense",
			Description:  "Groceries",
			TransactedAt: "2024-01-02",
		})

		// Assert.
		if got != nil {
			t.Fatal("expected nil transaction")
		}
		if err == nil || err.Message != "account not found" {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("update requires a field", func(t *testing.T) {
		svc := &TransactionService{txRepo: &fakeTransactionRepo{}}

		// Act.
		got, err := svc.Update(context.Background(), 42, 9, models.UpdateTransactionRequest{})

		// Assert.
		if got != nil {
			t.Fatal("expected nil transaction")
		}
		if err == nil || err.Message != "at least one field is required" {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("update maps invalid amount", func(t *testing.T) {
		amount := "12.12345"
		svc := &TransactionService{txRepo: &fakeTransactionRepo{}}

		// Act.
		got, err := svc.Update(context.Background(), 42, 9, models.UpdateTransactionRequest{Amount: &amount})

		// Assert.
		if got != nil {
			t.Fatal("expected nil transaction")
		}
		if err == nil || err.Message != "amount must be a positive decimal with up to 4 fraction digits" {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("delete maps no rows", func(t *testing.T) {
		svc := &TransactionService{
			txRepo: &fakeTransactionRepo{
				softDeleteFn: func(context.Context, int64, int64) error { return errors.New("no rows in result set") },
			},
		}

		// Act.
		err := svc.Delete(context.Background(), 42, 9)

		// Assert.
		if err == nil || err.Message != "transaction not found" {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("create passes converted params", func(t *testing.T) {
		categoryID := int64(3)
		notes := "weekly"
		repo := &fakeTransactionRepo{
			createForUserFn: func(_ context.Context, userID int64, params sqlc.CreateTransactionParams) (sqlc.Transaction, error) {
				if userID != 42 || params.AccountID != 7 || !params.CategoryID.Valid || params.CategoryID.Int64 != categoryID {
					t.Fatalf("unexpected create args: %+v", params)
				}
				if !params.Notes.Valid || params.Notes.String != notes || !params.TransactedAt.Valid {
					t.Fatalf("unexpected create args: %+v", params)
				}
				return testTransactionRow(), nil
			},
		}
		svc := &TransactionService{txRepo: repo}

		// Act.
		got, err := svc.Create(context.Background(), 42, models.CreateTransactionRequest{
			AccountID:    7,
			CategoryID:   &categoryID,
			Amount:       "18.9000",
			Currency:     "USD",
			Type:         "expense",
			Description:  "Groceries",
			Notes:        &notes,
			TransactedAt: "2024-01-02",
		})

		// Assert.
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
		if got == nil || got.Description != "Groceries" {
			t.Fatalf("unexpected transaction: %#v", got)
		}
	})
}

func TestTransactionService_GetByIDAndInternalErrors(t *testing.T) {
	svc := &TransactionService{
		txRepo: &fakeTransactionRepo{
			getByIDForUserFn: func(context.Context, int64, int64) (sqlc.Transaction, error) {
				return sqlc.Transaction{}, errors.New("missing")
			},
			listForUserFn: func(context.Context, sqlc.ListTransactionsForUserParams) ([]sqlc.Transaction, error) {
				return nil, errors.New("boom")
			},
			updateForUserFn: func(context.Context, int64, int64, sqlc.UpdateTransactionByIDForUserParams) (sqlc.Transaction, error) {
				return sqlc.Transaction{}, errors.New("boom")
			},
		},
	}

	// Act.
	got, getErr := svc.GetByID(context.Background(), 42, 9)
	_, listErr := svc.List(context.Background(), 42, models.ListTransactionsQuery{Page: 1, Limit: 20})
	amount := "12.0000"
	_, updateErr := svc.Update(context.Background(), 42, 9, models.UpdateTransactionRequest{Amount: &amount})

	// Assert.
	if got != nil {
		t.Fatal("expected nil transaction")
	}
	if getErr == nil || getErr.Message != "transaction not found" {
		t.Fatalf("unexpected get error: %#v", getErr)
	}
	if listErr == nil || listErr.Message != "failed to list transactions" {
		t.Fatalf("unexpected list error: %#v", listErr)
	}
	if updateErr == nil || updateErr.Message != "failed to update transaction" {
		t.Fatalf("unexpected update error: %#v", updateErr)
	}
}

func TestTransactionService_UpdatePassesNumericAmount(t *testing.T) {
	repo := &fakeTransactionRepo{
		updateForUserFn: func(_ context.Context, userID, txID int64, params sqlc.UpdateTransactionByIDForUserParams) (sqlc.Transaction, error) {
			if userID != 42 || txID != 9 {
				t.Fatalf("unexpected ids: user=%d tx=%d", userID, txID)
			}
			if !params.Amount.Valid || params.Amount.NaN || params.Amount.Int == nil {
				t.Fatalf("expected numeric amount: %+v", params.Amount)
			}
			if !params.Notes.Valid || params.Notes.String != "updated" {
				t.Fatalf("unexpected notes: %+v", params.Notes)
			}
			return testTransactionRow(), nil
		},
	}
	svc := &TransactionService{txRepo: repo}
	amount := "20.0000"
	notes := "updated"

	// Act.
	got, err := svc.Update(context.Background(), 42, 9, models.UpdateTransactionRequest{
		Amount: &amount,
		Notes:  &notes,
	})

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}
	if got == nil || got.Amount != "18.9000" {
		t.Fatalf("unexpected transaction: %#v", got)
	}
}

func TestTransactionService_ListRejectsInvalidDate(t *testing.T) {
	svc := &TransactionService{txRepo: &fakeTransactionRepo{}}
	from := "2024/01/01"

	// Act.
	got, err := svc.List(context.Background(), 42, models.ListTransactionsQuery{
		From:  &from,
		Page:  1,
		Limit: 20,
	})

	// Assert.
	if got != nil {
		t.Fatal("expected nil list")
	}
	if err == nil || err.Message != "invalid date, expected YYYY-MM-DD" {
		t.Fatalf("unexpected error: %#v", err)
	}
}

func TestTransactionService_UpdateClearsOptionalFields(t *testing.T) {
	categoryID := int64(3)
	repo := &fakeTransactionRepo{
		updateForUserFn: func(_ context.Context, _, _ int64, params sqlc.UpdateTransactionByIDForUserParams) (sqlc.Transaction, error) {
			if !params.CategoryID.Valid || params.CategoryID.Int64 != categoryID {
				t.Fatalf("unexpected category: %+v", params.CategoryID)
			}
			if params.Amount != (pgtype.Numeric{}) {
				t.Fatalf("expected zero numeric amount: %+v", params.Amount)
			}
			return testTransactionRow(), nil
		},
	}
	svc := &TransactionService{txRepo: repo}

	// Act.
	got, err := svc.Update(context.Background(), 42, 9, models.UpdateTransactionRequest{CategoryID: &categoryID})

	// Assert.
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}
	if got == nil {
		t.Fatal("expected transaction")
	}
}
