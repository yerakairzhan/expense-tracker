package service

import (
	"context"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"
)

type RecurringService struct {
	repo recurringRepository
}

type recurringRepository interface {
	List(ctx context.Context, userID int64) ([]repository.RecurringRow, error)
	Create(ctx context.Context, userID, accountID int64, categoryID *int64, title, amount, currency, frequency, nextRunAt string, endsAt *string) (repository.RecurringRow, error)
	GetByIDForUser(ctx context.Context, id, userID int64) (repository.RecurringRow, error)
	Update(ctx context.Context, id, userID int64, amount, frequency, nextRunAt, endsAt *string) (repository.RecurringRow, error)
	Deactivate(ctx context.Context, id, userID int64) (int64, error)
}

func NewRecurringService(repo *repository.RecurringRepository) *RecurringService {
	return &RecurringService{repo: repo}
}

func (s *RecurringService) List(ctx context.Context, userID int64) ([]models.RecurringPayment, *apperror.Error) {
	rows, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, apperror.Internal("failed to list recurring payments")
	}
	out := make([]models.RecurringPayment, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapRecurring(r))
	}
	return out, nil
}

func (s *RecurringService) Create(ctx context.Context, userID int64, req models.CreateRecurringRequest) (*models.RecurringPayment, *apperror.Error) {
	if !isPositiveDecimal(req.Amount) {
		return nil, apperror.Validation("amount must be a positive decimal")
	}
	row, err := s.repo.Create(ctx, userID, req.AccountID, req.CategoryID, req.Title, req.Amount, req.Currency, req.Frequency, req.NextRunAt, req.EndsAt)
	if err != nil {
		return nil, apperror.Internal("failed to create recurring payment")
	}
	out := mapRecurring(row)
	return &out, nil
}

func (s *RecurringService) Update(ctx context.Context, userID, id int64, req models.UpdateRecurringRequest) (*models.RecurringPayment, *apperror.Error) {
	if req.Amount == nil && req.Frequency == nil && req.NextRunAt == nil && req.EndsAt == nil {
		return nil, apperror.Validation("at least one field is required")
	}
	if req.Amount != nil && !isPositiveDecimal(*req.Amount) {
		return nil, apperror.Validation("amount must be a positive decimal")
	}
	row, err := s.repo.Update(ctx, id, userID, req.Amount, req.Frequency, req.NextRunAt, req.EndsAt)
	if err != nil {
		return nil, apperror.NotFound("recurring payment not found")
	}
	out := mapRecurring(row)
	return &out, nil
}

func (s *RecurringService) Delete(ctx context.Context, userID, id int64) *apperror.Error {
	affected, err := s.repo.Deactivate(ctx, id, userID)
	if err != nil {
		return apperror.Internal("failed to deactivate recurring payment")
	}
	if affected == 0 {
		return apperror.NotFound("recurring payment not found")
	}
	return nil
}
