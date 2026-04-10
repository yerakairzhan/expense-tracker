package service

import (
	"testing"
	"time"

	sqlc "finance-tracker/db/queries"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestIsPositiveDecimal(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "valid integer", value: "10", want: true},
		{name: "valid fraction", value: "10.1234", want: true},
		{name: "zero", value: "0.0000", want: false},
		{name: "too many decimals", value: "1.12345", want: false},
		{name: "negative", value: "-1", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			got := isPositiveDecimal(tt.value)

			// Assert.
			if got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestConversionHelpers(t *testing.T) {
	t.Run("stringToNumeric rejects blank", func(t *testing.T) {
		// Act.
		_, err := stringToNumeric("   ")

		// Assert.
		if err == nil || err.Error() != "amount is required" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("dateFromString rejects invalid format", func(t *testing.T) {
		// Act.
		_, err := dateFromString("01/02/2024")

		// Assert.
		if err == nil || err.Error() != "invalid date, expected YYYY-MM-DD" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("numericToString4 returns fixed scale", func(t *testing.T) {
		// Act.
		got := numericToString4(mustNumeric("12.3"))

		// Assert.
		if got != "12.3000" {
			t.Fatalf("unexpected value: %s", got)
		}
	})

	t.Run("mapTransaction preserves nullable fields", func(t *testing.T) {
		row := testTransactionRow()

		// Act.
		got := mapTransaction(row)

		// Assert.
		if got.CategoryID == nil || *got.CategoryID != 3 || got.Notes == nil || *got.Notes != "weekly" {
			t.Fatalf("unexpected transaction: %#v", got)
		}
	})
}

func TestMapUserAndAccount(t *testing.T) {
	userRow := testUserRow("hash")
	accountRow := testAccountRow()

	// Act.
	gotUser := mapUser(userRow)
	gotAccount := mapAccount(accountRow)

	// Assert.
	if gotUser.Email != "john@example.com" || gotAccount.Balance != "100.5000" {
		t.Fatalf("unexpected mapping: %#v %#v", gotUser, gotAccount)
	}
}

func TestPointerHelpers(t *testing.T) {
	value := "hello"
	intValue := int64(9)

	// Act.
	text := textFromPtr(&value)
	number := int8FromPtr(&intValue)
	emptyDate := dateToString(pgtype.Date{})
	emptyTimestamp := timestamptzToTime(pgtype.Timestamptz{})

	// Assert.
	if !text.Valid || text.String != "hello" {
		t.Fatalf("unexpected text: %#v", text)
	}
	if !number.Valid || number.Int64 != 9 {
		t.Fatalf("unexpected int: %#v", number)
	}
	if emptyDate != "" || !emptyTimestamp.Equal(time.Time{}) {
		t.Fatalf("unexpected empty conversions: %q %v", emptyDate, emptyTimestamp)
	}
}

func TestNumericToString4HandlesInvalidValue(t *testing.T) {
	// Act.
	got := numericToString4(pgtype.Numeric{})

	// Assert.
	if got != "0.0000" {
		t.Fatalf("unexpected value: %s", got)
	}
}

func TestMapUserWithGeneratedRow(t *testing.T) {
	now := testTime()
	row := sqlc.User{
		ID:        1,
		Email:     "a@b.com",
		Name:      "A",
		Currency:  "USD",
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	// Act.
	got := mapUser(row)

	// Assert.
	if got.ID != 1 || got.CreatedAt != now {
		t.Fatalf("unexpected user: %#v", got)
	}
}
