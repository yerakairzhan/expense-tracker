package service

import (
	"context"

	sqlc "finance-tracker/db/queries"
	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"
)

type AccountService struct {
	accounts accountRepository
}

type accountRepository interface {
	ListByUser(ctx context.Context, userID int64) ([]sqlc.Account, error)
	Create(ctx context.Context, userID int64, name, accountType, currency, balance string) (sqlc.Account, error)
	GetByIDForUser(ctx context.Context, accountID, userID int64) (sqlc.Account, error)
	UpdateByIDForUser(ctx context.Context, accountID, userID int64, name, accountType, currency *string) (sqlc.Account, error)
	SoftDeleteByIDForUser(ctx context.Context, accountID, userID int64) (int64, error)
}

func NewAccountService(accounts *repository.AccountRepository) *AccountService {
	return &AccountService{accounts: accounts}
}

func (s *AccountService) List(ctx context.Context, userID int64) ([]models.Account, *apperror.Error) {
	rows, err := s.accounts.ListByUser(ctx, userID)
	if err != nil {
		return nil, apperror.Internal("failed to list accounts")
	}
	out := make([]models.Account, 0, len(rows))
	for _, item := range rows {
		out = append(out, mapAccount(item))
	}
	return out, nil
}

func (s *AccountService) Create(ctx context.Context, userID int64, req models.CreateAccountRequest) (*models.Account, *apperror.Error) {
	if req.Balance == "" {
		req.Balance = "0.00"
	}
	if !isNonNegativeDecimal(req.Balance) {
		return nil, apperror.Validation("balance must be a non-negative decimal with up to 4 fraction digits")
	}
	row, err := s.accounts.Create(ctx, userID, req.Name, req.AccountType, req.Currency, req.Balance)
	if err != nil {
		return nil, apperror.Internal("failed to create account")
	}
	out := mapAccount(row)
	return &out, nil
}

func (s *AccountService) GetByID(ctx context.Context, userID, accountID int64) (*models.Account, *apperror.Error) {
	row, err := s.accounts.GetByIDForUser(ctx, accountID, userID)
	if err != nil {
		return nil, apperror.NotFound("account not found")
	}
	out := mapAccount(row)
	return &out, nil
}

func (s *AccountService) Update(ctx context.Context, userID, accountID int64, req models.UpdateAccountRequest) (*models.Account, *apperror.Error) {
	if req.Name == nil && req.AccountType == nil && req.Currency == nil {
		return nil, apperror.Validation("at least one field is required")
	}
	row, err := s.accounts.UpdateByIDForUser(ctx, accountID, userID, req.Name, req.AccountType, req.Currency)
	if err != nil {
		return nil, apperror.NotFound("account not found")
	}
	out := mapAccount(row)
	return &out, nil
}

func (s *AccountService) Delete(ctx context.Context, userID, accountID int64) *apperror.Error {
	affected, err := s.accounts.SoftDeleteByIDForUser(ctx, accountID, userID)
	if err != nil {
		return apperror.Internal("failed to delete account")
	}
	if affected == 0 {
		return apperror.NotFound("account not found")
	}
	return nil
}
