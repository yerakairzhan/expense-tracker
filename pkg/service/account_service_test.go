package service

import (
	"context"
	"errors"
	"testing"

	sqlc "finance-tracker/db/queries"
	"finance-tracker/pkg/models"
)

func TestAccountService_Create(t *testing.T) {
	t.Run("rejects invalid balance", func(t *testing.T) {
		svc := &AccountService{accounts: &fakeAccountRepo{}}

		// Act.
		got, err := svc.Create(context.Background(), 42, testCreateAccountRequest("0.0000"))

		// Assert.
		if got != nil {
			t.Fatal("expected nil account")
		}
		if err == nil || err.Message != "balance must be a positive decimal with up to 4 fraction digits" {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("creates account", func(t *testing.T) {
		repo := &fakeAccountRepo{
			createFn: func(_ context.Context, userID int64, name, accountType, currency, balance string) (account sqlc.Account, err error) {
				if userID != 42 || name != "Cash" || accountType != "cash" || currency != "USD" || balance != "10.5000" {
					t.Fatalf("unexpected create args: %d %s %s %s %s", userID, name, accountType, currency, balance)
				}
				return testAccountRow(), nil
			},
		}
		svc := &AccountService{accounts: repo}

		// Act.
		got, err := svc.Create(context.Background(), 42, testCreateAccountRequest("10.5000"))

		// Assert.
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
		if got == nil || got.Balance != "100.5000" {
			t.Fatalf("unexpected account: %#v", got)
		}
	})
}

func TestAccountService_UpdateAndDelete(t *testing.T) {
	t.Run("update requires a field", func(t *testing.T) {
		svc := &AccountService{accounts: &fakeAccountRepo{}}

		// Act.
		got, err := svc.Update(context.Background(), 42, 7, models.UpdateAccountRequest{})

		// Assert.
		if got != nil {
			t.Fatal("expected nil account")
		}
		if err == nil || err.Message != "at least one field is required" {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("delete maps affected zero to not found", func(t *testing.T) {
		svc := &AccountService{
			accounts: &fakeAccountRepo{
				softDeleteFn: func(context.Context, int64, int64) (int64, error) { return 0, nil },
			},
		}

		// Act.
		err := svc.Delete(context.Background(), 42, 7)

		// Assert.
		if err == nil || err.Message != "account not found" {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("delete maps repo error to internal", func(t *testing.T) {
		svc := &AccountService{
			accounts: &fakeAccountRepo{
				softDeleteFn: func(context.Context, int64, int64) (int64, error) { return 0, errors.New("boom") },
			},
		}

		// Act.
		err := svc.Delete(context.Background(), 42, 7)

		// Assert.
		if err == nil || err.Message != "failed to delete account" {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func testCreateAccountRequest(balance string) models.CreateAccountRequest {
	return models.CreateAccountRequest{
		Name:        "Cash",
		AccountType: "cash",
		Currency:    "USD",
		Balance:     balance,
	}
}
