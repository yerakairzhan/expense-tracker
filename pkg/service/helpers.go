package service

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	sqlc "finance-tracker/db/queries"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

var decimalPattern = regexp.MustCompile(`^\d+(\.\d{1,4})?$`)

func isPositiveDecimal(v string) bool {
	if !decimalPattern.MatchString(v) {
		return false
	}
	return v != "0" && v != "0.0" && v != "0.00" && v != "0.000" && v != "0.0000"
}

func stringToNumeric(input string) (pgtype.Numeric, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return pgtype.Numeric{}, fmt.Errorf("amount is required")
	}
	var n pgtype.Numeric
	if err := n.Scan(trimmed); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("invalid numeric value")
	}
	return n, nil
}

func dateFromString(iso string) (pgtype.Date, error) {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("invalid date, expected YYYY-MM-DD")
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

func textFromPtr(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *v, Valid: true}
}

func int8FromPtr(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}

func numericToString4(n pgtype.Numeric) string {
	if !n.Valid || n.NaN || n.Int == nil {
		return "0.0000"
	}
	r := new(big.Rat).SetInt(n.Int)
	switch {
	case n.Exp > 0:
		r = r.Mul(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)))
	case n.Exp < 0:
		r = r.Quo(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)))
	}
	return r.FloatString(4)
}

func timestamptzToTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func dateToString(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

func mapUser(in sqlc.User) models.User {
	return models.User{
		ID:        in.ID,
		Email:     in.Email,
		Name:      in.Name,
		Currency:  in.Currency,
		Role:      in.Role,
		CreatedAt: timestamptzToTime(in.CreatedAt),
		UpdatedAt: timestamptzToTime(in.UpdatedAt),
	}
}

func mapAccount(in sqlc.Account) models.Account {
	return models.Account{
		ID:          in.ID,
		UserID:      in.UserID,
		Name:        in.Name,
		AccountType: in.AccountType,
		Balance:     numericToString4(in.Balance),
		Currency:    in.Currency,
		CreatedAt:   timestamptzToTime(in.CreatedAt),
		UpdatedAt:   timestamptzToTime(in.UpdatedAt),
	}
}

func mapTransaction(in sqlc.Transaction) models.Transaction {
	var categoryID *int64
	if in.CategoryID.Valid {
		v := in.CategoryID.Int64
		categoryID = &v
	}
	var notes *string
	if in.Notes.Valid {
		v := in.Notes.String
		notes = &v
	}
	return models.Transaction{
		ID:           in.ID,
		AccountID:    in.AccountID,
		CategoryID:   categoryID,
		Amount:       numericToString4(in.Amount),
		Currency:     in.Currency,
		Type:         in.Type,
		Description:  in.Description,
		Notes:        notes,
		TransactedAt: dateToString(in.TransactedAt),
		CreatedAt:    timestamptzToTime(in.CreatedAt),
		UpdatedAt:    timestamptzToTime(in.UpdatedAt),
	}
}

func mapCategory(in repository.CategoryRow) models.Category {
	var userID *int64
	if in.UserID.Valid {
		v := in.UserID.Int64
		userID = &v
	}
	var color *string
	if in.Color.Valid {
		v := in.Color.String
		color = &v
	}
	var icon *string
	if in.Icon.Valid {
		v := in.Icon.String
		icon = &v
	}
	return models.Category{
		ID:        in.ID,
		UserID:    userID,
		Name:      in.Name,
		Type:      in.Type,
		Color:     color,
		Icon:      icon,
		IsSystem:  in.IsSystem,
		CreatedAt: timestamptzToTime(in.CreatedAt),
		UpdatedAt: timestamptzToTime(in.UpdatedAt),
	}
}

func mapBudget(in repository.BudgetRow) models.Budget {
	var categoryID *int64
	if in.CategoryID.Valid {
		v := in.CategoryID.Int64
		categoryID = &v
	}
	var endsAt *string
	if in.EndsAt.Valid {
		s := dateToString(in.EndsAt)
		endsAt = &s
	}
	return models.Budget{
		ID:          in.ID,
		UserID:      in.UserID,
		CategoryID:  categoryID,
		LimitAmount: numericToString4(in.LimitAmount),
		Currency:    in.Currency,
		Period:      in.Period,
		StartsAt:    dateToString(in.StartsAt),
		EndsAt:      endsAt,
		IsActive:    in.IsActive,
		CreatedAt:   timestamptzToTime(in.CreatedAt),
		UpdatedAt:   timestamptzToTime(in.UpdatedAt),
	}
}

func mapRecurring(in repository.RecurringRow) models.RecurringPayment {
	var categoryID *int64
	if in.CategoryID.Valid {
		v := in.CategoryID.Int64
		categoryID = &v
	}
	var endsAt *string
	if in.EndsAt.Valid {
		s := dateToString(in.EndsAt)
		endsAt = &s
	}
	return models.RecurringPayment{
		ID:         in.ID,
		UserID:     in.UserID,
		AccountID:  in.AccountID,
		CategoryID: categoryID,
		Title:      in.Title,
		Amount:     numericToString4(in.Amount),
		Currency:   in.Currency,
		Frequency:  in.Frequency,
		NextRunAt:  dateToString(in.NextRunAt),
		EndsAt:     endsAt,
		IsActive:   in.IsActive,
		CreatedAt:  timestamptzToTime(in.CreatedAt),
		UpdatedAt:  timestamptzToTime(in.UpdatedAt),
	}
}

func errorAs(err error, target interface{}) bool {
	return errors.As(err, target)
}
