package service

import (
	"context"
	"errors"
	"testing"

	sqlc "finance-tracker/db/queries"
	"finance-tracker/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_UpdateMe(t *testing.T) {
	t.Run("requires a field", func(t *testing.T) {
		svc := &UserService{users: &fakeUserRepo{}}

		// Act.
		got, err := svc.UpdateMe(context.Background(), 42, models.UpdateMeRequest{})

		// Assert.
		if got != nil {
			t.Fatal("expected nil user")
		}
		if err == nil || err.Message != "at least one field is required" {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("updates profile", func(t *testing.T) {
		name := "Jane"
		svc := &UserService{
			users: &fakeUserRepo{
				updateProfileFn: func(_ context.Context, userID int64, gotName, gotCurrency *string) (sqlc.User, error) {
					if userID != 42 || gotName == nil || *gotName != name || gotCurrency != nil {
						t.Fatalf("unexpected update args: %d %#v %#v", userID, gotName, gotCurrency)
					}
					row := testUserRow("hash")
					row.Name = name
					return row, nil
				},
			},
		}

		// Act.
		got, err := svc.UpdateMe(context.Background(), 42, models.UpdateMeRequest{Name: &name})

		// Assert.
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
		if got == nil || got.Name != "Jane" {
			t.Fatalf("unexpected user: %#v", got)
		}
	})
}

func TestUserService_ChangePassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("rejects wrong current password", func(t *testing.T) {
		svc := &UserService{
			users: &fakeUserRepo{
				getByIDFn: func(context.Context, int64) (sqlc.User, error) { return testUserRow(string(hash)), nil },
			},
		}

		// Act.
		appErr := svc.ChangePassword(context.Background(), 42, models.ChangePasswordRequest{
			CurrentPassword: "wrong",
			NewPassword:     "newpassword123",
		})

		// Assert.
		if appErr == nil || appErr.Message != "current password is incorrect" {
			t.Fatalf("unexpected error: %#v", appErr)
		}
	})

	t.Run("updates password", func(t *testing.T) {
		var updatedHash string
		svc := &UserService{
			users: &fakeUserRepo{
				getByIDFn: func(context.Context, int64) (sqlc.User, error) { return testUserRow(string(hash)), nil },
				updatePasswordFn: func(_ context.Context, userID int64, passwordHash string) (sqlc.User, error) {
					updatedHash = passwordHash
					if userID != 42 {
						t.Fatalf("unexpected user id: %d", userID)
					}
					return testUserRow(passwordHash), nil
				},
			},
		}

		// Act.
		appErr := svc.ChangePassword(context.Background(), 42, models.ChangePasswordRequest{
			CurrentPassword: "password123",
			NewPassword:     "newpassword123",
		})

		// Assert.
		if appErr != nil {
			t.Fatalf("unexpected error: %#v", appErr)
		}
		if updatedHash == "" || bcrypt.CompareHashAndPassword([]byte(updatedHash), []byte("newpassword123")) != nil {
			t.Fatal("expected updated bcrypt hash")
		}
	})

	t.Run("maps update failure", func(t *testing.T) {
		svc := &UserService{
			users: &fakeUserRepo{
				getByIDFn:        func(context.Context, int64) (sqlc.User, error) { return testUserRow(string(hash)), nil },
				updatePasswordFn: func(context.Context, int64, string) (sqlc.User, error) { return sqlc.User{}, errors.New("boom") },
			},
		}

		// Act.
		appErr := svc.ChangePassword(context.Background(), 42, models.ChangePasswordRequest{
			CurrentPassword: "password123",
			NewPassword:     "newpassword123",
		})

		// Assert.
		if appErr == nil || appErr.Message != "failed to update password" {
			t.Fatalf("unexpected error: %#v", appErr)
		}
	})
}
