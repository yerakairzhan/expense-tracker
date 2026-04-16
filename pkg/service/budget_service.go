package service

import (
	"context"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"
)

type BudgetService struct {
	repo budgetRepository
}

type budgetRepository interface {
	List(ctx context.Context, userID int64) ([]repository.BudgetRow, error)
	Create(ctx context.Context, userID int64, categoryID *int64, limitAmount, currency, period, startsAt string, endsAt *string) (repository.BudgetRow, error)
	GetByIDForUser(ctx context.Context, id, userID int64) (repository.BudgetRow, error)
	GetProgress(ctx context.Context, budgetID, userID int64) (repository.BudgetProgressRow, error)
	Update(ctx context.Context, id, userID int64, limitAmount, period, endsAt *string) (repository.BudgetRow, error)
	Deactivate(ctx context.Context, id, userID int64) (int64, error)
}

func NewBudgetService(repo *repository.BudgetRepository) *BudgetService {
	return &BudgetService{repo: repo}
}

func (s *BudgetService) List(ctx context.Context, userID int64) ([]models.Budget, *apperror.Error) {
	rows, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, apperror.Internal("failed to list budgets")
	}
	out := make([]models.Budget, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapBudget(r))
	}
	return out, nil
}

func (s *BudgetService) Create(ctx context.Context, userID int64, req models.CreateBudgetRequest) (*models.Budget, *apperror.Error) {
	if !isPositiveDecimal(req.LimitAmount) {
		return nil, apperror.Validation("limit_amount must be a positive decimal")
	}
	row, err := s.repo.Create(ctx, userID, req.CategoryID, req.LimitAmount, req.Currency, req.Period, req.StartsAt, req.EndsAt)
	if err != nil {
		return nil, apperror.Internal("failed to create budget")
	}
	out := mapBudget(row)
	return &out, nil
}

func (s *BudgetService) GetProgress(ctx context.Context, userID, id int64) (*models.BudgetProgress, *apperror.Error) {
	row, err := s.repo.GetProgress(ctx, id, userID)
	if err != nil {
		return nil, apperror.NotFound("budget not found")
	}
	limitStr, spentStr, remainingStr, pctStr := repository.BudgetProgressCalc(row.LimitAmount, row.Spent)
	return &models.BudgetProgress{
		BudgetID:    id,
		LimitAmount: limitStr,
		Spent:       spentStr,
		Remaining:   remainingStr,
		Percentage:  pctStr,
	}, nil
}

func (s *BudgetService) Update(ctx context.Context, userID, id int64, req models.UpdateBudgetRequest) (*models.Budget, *apperror.Error) {
	if req.LimitAmount == nil && req.Period == nil && req.EndsAt == nil {
		return nil, apperror.Validation("at least one field is required")
	}
	if req.LimitAmount != nil && !isPositiveDecimal(*req.LimitAmount) {
		return nil, apperror.Validation("limit_amount must be a positive decimal")
	}
	row, err := s.repo.Update(ctx, id, userID, req.LimitAmount, req.Period, req.EndsAt)
	if err != nil {
		return nil, apperror.NotFound("budget not found")
	}
	out := mapBudget(row)
	return &out, nil
}

func (s *BudgetService) Delete(ctx context.Context, userID, id int64) *apperror.Error {
	affected, err := s.repo.Deactivate(ctx, id, userID)
	if err != nil {
		return apperror.Internal("failed to deactivate budget")
	}
	if affected == 0 {
		return apperror.NotFound("budget not found")
	}
	return nil
}
